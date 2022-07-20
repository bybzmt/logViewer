package main

import (
	"log"
	"logViewer/find"
)

func main() {
	log.Println("runing.")

	s := find.ServerTCP{
		Addr: "127.0.0.2:7000",
	}
	s.Init()
	s.Run()
}
