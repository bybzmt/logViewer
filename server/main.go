package main

import (
	"log"
	"net"

	"github.com/integrii/flaggy"
)

var runing = true
var limitDirs []string

func main() {

	var addr string

	flaggy.StringSlice(&limitDirs, "", "dir", "limit dir")
	flaggy.String(&addr, "", "addr", "listen addr:port")

	if len(limitDirs) < 1 {
		limitDirs = append(limitDirs, "/")
	}

	if addr == "" {
		addr = "127.0.0.2:7000"
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Listen", addr)

	s := matchServer{}
	s.init()

	for runing {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go s.serviceWapper(conn)
	}

}
