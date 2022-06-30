package main

import (
	"log"
	"net"

	"github.com/integrii/flaggy"
)

var runing = true

func main() {

	var dirs []string
	var addr string

	flaggy.StringSlice(&dirs, "", "dir", "limit dir")
	flaggy.String(&addr, "", "addr", "listen addr:port")

	if len(dirs) < 1 {
		dirs = append(dirs, "/")
	}

	if addr == "" {
		addr = "127.0.0.2:7000"
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}

	for runing {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go service(conn)
	}

}

func service(c net.Conn) {
}
