package main

import (
	"log"
	"logViewer/find/protocol"
	"net"
)

func tcpRun(tcp *protocol.MatchServer) {
	if len(tcp.Dirs) < 1 {
		tcp.Dirs = append(tcp.Dirs, "/")
	}

	if tcp.Addr == "" {
		tcp.Addr = "127.0.0.2:7000"
	}

	l, err := net.Listen("tcp", tcp.Addr)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Listen", tcp.Addr)

	tcp.Init()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go tcp.ServiceWapper(conn)
	}
}
