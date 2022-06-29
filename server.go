package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"strconv"
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
	NowPath            string // 当前工作目录/路径
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
	config.NowPath = "/"
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
		// rm
		"RMD",
		// rm -rf
		"RNFR",
		// rename to, maybe use mv
		"RNTO",
		// locate
		"SITE",
		// show file info
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
	case cmd == "CWD":
		// 进入某个目录
		b := solveCwd(ch, args, user.username)
		if b {
			return CmdOk, nil
		}
		return SyntaxErr, nil
	case cmd == "STOR":
		// 存储文件
		ch.Filename = stripDirectory(args)
		readPortData(ch, user.username, c)
		return TxfrCompleteOk, nil
	case cmd == "RETR":
		// 读取文件
		readData(ch, user.username, c, args)
		return CmdOk, nil
	case cmd == "FEAT":
		return FeatResponse, nil
	case cmd == "NLST" || cmd == "LIST":
		// 显示该目录下的所有文件夹以及文件
		showLs(c, ch, user.username)
		return CmdOk, nil
	case cmd == "HELP":
		// 显示服务端help
		return CmdOk, nil
	case cmd == "XPWD":
		sendMsg(c, PwdResponse)
		// 显示当前目录路径
		showPwd(ch, c)
		return CmdOk, nil
	case cmd == "TYPE" && args == "I":
		return TypeSetOk, nil
	case cmd == "PORT":
		// 解析绑定ip端口
		ch.DataConnectionAddr = parsePortArgs(args)
		return PortOk, nil
	case cmd == "PASV":
		// TODO
		return CmdNotImplementd, nil
	case cmd == "QUIT":
		return GoodbyeMsg, nil
	}
	return SyntaxErr, nil
}

func readData(ch *ConnectionConfig, username string, in net.Conn, args string) {
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

	sendMsg(in, DataCnxAlreadyOpenStartXfr)
	pwd, _ := os.Getwd()
	path2 := path.Join(pwd, storageDir, username, ch.NowPath, args)
	path2 = strings.Replace(path2, "/", "\\", -1)
	fi, err := os.OpenFile(path2, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	defer func(fi *os.File) {
		err := fi.Close()
		if err != nil {
			fmt.Printf("file close error %s", err)
			return
		}
	}(fi)
	if err != nil {
		fmt.Printf("read file error %s", err)
		return
	}
	buf := make([]byte, 1024)
	for {
		n, err := fi.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Printf("read error %s", err)
			break
		}
		if n == 0 {
			break
		}
		if _, err := c.Write(buf[:n]); err != nil {
			fmt.Printf("write error %s", err)
			break
		}
	}
}

func showListName(ch *ConnectionConfig, username string) ([]string, error) {
	fi, err := showListInfo(ch, username)
	if err != nil {
		return nil, err
	}
	res := make([]string, len(fi))
	for i := range fi {
		res = append(res, fi[i].Name())
	}
	return res, nil
}

func showListInfoName(ch *ConnectionConfig, username string) ([]fs.FileInfo, []string, error) {
	fi, err := showListInfo(ch, username)
	if err != nil {
		return nil, nil, err
	}
	res := make([]string, len(fi))
	for i := range fi {
		res = append(res, fi[i].Name())
	}
	return fi, res, nil
}

func showLs(c net.Conn, ch *ConnectionConfig, username string) {
	fis, err := showListInfo(ch, username)
	if err != nil {
		fmt.Printf("ls error %s", err)
		return
	}
	// TODO: 优化
	for _, v := range fis {
		sendMsg(c, fmt.Sprintf("%s", formatFileInfo(v)))
	}
}

func formatFileInfo(info fs.FileInfo) string {
	res := "filename: " + info.Name() + "; isDir: " + strconv.FormatBool(info.IsDir()) + "; size: " + strconv.FormatFloat(float64(info.Size()), 'E', -1, 32) + "; modyTime: " + info.ModTime().String() + "\r"
	return res
}

func showListInfo(ch *ConnectionConfig, username string) ([]fs.FileInfo, error) {
	pwd, _ := os.Getwd()
	//获取文件或目录相关信息
	path2 := path.Join(pwd, storageDir, username, ch.NowPath)
	path2 = strings.Replace(path2, "/", "\\", -1)
	fileInfoList, err := ioutil.ReadDir(path2)
	if err != nil {
		return nil, err
	}
	return fileInfoList, nil
}

func solveCwd(ch *ConnectionConfig, args string, username string) bool {
	// 返回上级目录
	if args == ".." {
		if ch.NowPath == "/" {
			return false
		}
		p := path.Base(ch.NowPath)
		if p == "/" || p == "." {
			return false
		}
		ch.NowPath = strings.TrimRight(ch.NowPath, "/"+p)
		if ch.NowPath == "" {
			ch.NowPath = "/"
		}
		return true
	}
	fi, err := showListInfo(ch, username)
	if err != nil {
		fmt.Printf("solve cmd error %s", err)
	}
	b := false
	for i := 0; i < len(fi); i++ {
		if fi[i].Name() == args && fi[i].IsDir() {
			b = true
			break
		}
	}
	if b {
		ch.NowPath = path.Join(ch.NowPath, args)
		return true
	}
	return false
}

func showPwd(ch *ConnectionConfig, c net.Conn) {
	// 截取至xxx目录
	sendMsg(c, path.Join(ch.NowPath, "\n"))
}

func makeDir(ch *ConnectionConfig, username string, arg string) {
	fmt.Printf("connecting to %v\n", ch.DataConnectionAddr)
	var err error
	path2 := path.Join(storageDir, username, ch.NowPath, arg)
	path2 = strings.Replace(path2, "/", "\\", -1)
	err = os.MkdirAll(path2, 0777)
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
	path2 := path.Join(storageDir, username, ch.NowPath)
	path2 = strings.Replace(path2, "/", "\\", -1)
	err = os.MkdirAll(path2, 0777)
	if err != nil {
		fmt.Printf("create dir error %s\n", err)
		return
	}

	outPutName := getFileName(username, ch.Filename, ch)
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

func getFileName(username, filename string, ch *ConnectionConfig) string {
	path2 := path.Join(storageDir, username, ch.NowPath, filename)
	path2 = strings.Replace(path2, "/", "\\", -1)
	return path2
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
