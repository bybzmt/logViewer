package main

import (
	"flag"
	"log"
)

var addr = flag.String("addr", "127.0.0.1:8080", "listen addr:port")

func main() {
	var ui Ui
	ui.init()

	ui.httpServer.Addr = *addr

	err := ui.httpServer.ListenAndServe()
	if err != nil {
		log.Fatalln("ListenAndServe", err)
	}
}
