package main

import (
	"log"
	"logViewer/find/tcp"
)

func main() {
	log.Println("runing.")

	s := tcp.Server{
		Addr: "127.0.0.2:7000",
	}
	s.Init()
	s.Run()
}
