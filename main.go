package main

import (
	"log"
	"net"
)

const (
	port = ":2121"
)

func main() {
	log.Printf("starting up ftp server")
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
