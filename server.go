package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"time"
)

const (
	DataUrl    = "https://data.ambition.io"
	storageDir = "uploads"
)

// ConnectionConfig 连接配置
type ConnectionConfig struct {
	DataConnectionAddr string // 远程连接地址
	Filename           string // 文件名
}

func HandleConnection(c net.Conn) {
	defer func(c net.Conn) {
		err := c.Close()
		if err != nil {
			log.Printf("connection from: %v close error: %s\n", c.RemoteAddr(), err)
		}
	}(c)

	sendMsg(c, FtpServerReady)
	user := AuthUser{}
	msg := getMsg(c)
	cmd, args, err := parseCommand(msg)
	if err != nil {
		sendMsg(c, SyntaxErr)
	}
	if cmd == "OPTS" {
		if args == "UTF8 ON" {
			// ok
			sendMsg(c, CmdOk)
		}
	}

	for {
		msg := getMsg(c)
		resp := handleLogin(msg, &user)
		sendMsg(c, resp)
		// 验证用户是否合法，如果合法就正式进入服务
		if user.valid {
			break
		}
	}

	config := ConnectionConfig{}
	for {
		cmd := getMsg(c)
		resp, err := handleCommand(cmd, &config, &user, c)
		if err != nil {
			break
		}
		sendMsg(c, resp)
		time.Sleep(100 * time.Millisecond)
	}

}

// 解析命令
func handleCommand(in string, ch *ConnectionConfig, user *AuthUser, c net.Conn) (string, error) {
	in = strings.TrimSpace(in)
	cmd, args, err := parseCommand(in)
	if err != nil {
		fmt.Printf("%s from %v: %s\n", SyntaxErr, c.RemoteAddr(), err)
		return SyntaxErr, err
	}

	ignoredCommand := []string{
		"CDUP",
		"RMD",
		"RNFR",
		"RNTO",
		"SITE",
		"STAT",
	}

	notImplemented := []string{
		"EPSV",
		"EPRT",
	}
	switch {
	case stringInList(cmd, ignoredCommand):
		return CmdNotImplementd, nil
	case stringInList(cmd, notImplemented):
		return CmdNotImplementd, nil
	case cmd == "NOOP":
		return CmdOk, nil
	case cmd == "SYST":
		return SysType, nil
	case cmd == "XMKD":
		makeDir(ch, user.username, args)
		return CmdOk, nil
	case cmd == "STOR":
		ch.Filename = stripDirectory(args)
		readPortData(ch, user.username, c)
		return TxfrCompleteOk, nil
	case cmd == "FEAT":
		return FeatResponse, nil
	case cmd == "NLST" || cmd == "LIST":
		return CmdOk, nil
	case cmd == "HELP":
		return CmdOk, nil
	case cmd == "XPWD":
		sendMsg(c, PwdResponse)
		// 显示当前目录路径

		return CmdOk, nil
	case cmd == "TYPE" && args == "I":
		return TypeSetOk, nil
	case cmd == "PORT":
		ch.DataConnectionAddr = parsePortArgs(args)
		return PortOk, nil
	case cmd == "PASV":
		return CmdNotImplementd, nil
	case cmd == "QUIT":
		return GoodbyeMsg, nil
	}
	return SyntaxErr, nil
}

func makeDir(ch *ConnectionConfig, username string, arg string) {
	fmt.Printf("connecting to %v\n", ch.DataConnectionAddr)
	var err error

	err = os.MkdirAll(path.Join(storageDir, username, arg), 0777)
	if err != nil {
		if err != nil {
			fmt.Printf("create dir error %s\n", err)
			return
		}
	}
}

func readPortData(ch *ConnectionConfig, username string, out net.Conn) {
	fmt.Printf("connecting to %v\n", ch.DataConnectionAddr)
	var err error

	c, err := net.Dial("tcp", ch.DataConnectionAddr)
	if err != nil {
		fmt.Printf("dial connect failed %s\n", err)
		return
	}

	err = c.SetReadDeadline(time.Now().Add(time.Minute))
	if err != nil {
		fmt.Printf("setDeadLine error %s\n", err)
		return
	}

	defer func(c net.Conn) {
		err = c.Close()
		if err != nil {
			fmt.Printf("connection %v close error %s\n", c.RemoteAddr(), err)
		}
	}(c)

	sendMsg(out, DataCnxAlreadyOpenStartXfr)

	err = os.MkdirAll(path.Join(storageDir, username), 0777)
	if err != nil {
		fmt.Printf("create dir error %s\n", err)
		return
	}

	outPutName := getFileName(username, ch.Filename)
	file, err := os.Create(outPutName)
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			fmt.Printf("close file error %s\n", err)
		}
	}(file)
	if err != nil {
		fmt.Printf("create file %s error %s", outPutName, err)
		return
	}

	reader := bufio.NewReader(c)
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Printf("read error %s", err)
			break
		}
		// not exist or not data in buf
		if n == 0 {
			break
		}
		// write into file
		if _, err := file.Write(buf[:n]); err != nil {
			fmt.Printf("read error %s", err)
			break
		}
	}
}

func getFileName(username, filename string) string {
	return path.Join(storageDir, username, filename)
}

func getMsg(c net.Conn) string {
	bufc := bufio.NewReader(c)
	for {
		l, err := bufc.ReadBytes('\n')
		ls := string(l)
		if err != nil {
			fmt.Printf("get msg from %v error %s\n", c.RemoteAddr(), err)
			err := c.Close()
			if err != nil {
				fmt.Printf("close connection from %v error %s\n", c.RemoteAddr(), err)
				return ""
			}
			break
		}
		fmt.Printf("received: %s\n", ls)
		return strings.TrimRight(ls, "\r")
	}
	return ""
}

func sendMsg(c net.Conn, msg string) {
	fmt.Printf("send msg: %s\n", msg)
	_, err := io.WriteString(c, msg)
	if err != nil {
		fmt.Printf("%v send msg error: %s\n", c.RemoteAddr(), err)
		return
	}
}
