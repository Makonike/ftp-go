package main

import (
	"log"
	"net"
	"strconv"
)

const (
	DefaultPort = 2121
)

var (
	port = ":"
)

func init() {
	err := SetupSetting()
	InitAdapter()
	if err != nil {
		log.Fatalf("Init Server Error")
	}
}

func main() {
	if ServerSetting.port == 0 {
		ServerSetting.port = DefaultPort
	}
	log.Printf("starting up ftp server")
	port = port + strconv.FormatInt(int64(ServerSetting.port), 10)
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("listen tcp for port: %s error: %s\n", port, err)
	}

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("accept tcp for port: %s error: %s\n", port, err)
			continue
		}
		log.Printf("connection from %v established.\n", c.RemoteAddr())
		// 处理连接
		HandleConnection(c)
	}
}
