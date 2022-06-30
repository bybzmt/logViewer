package main

import (
	"flag"
	"log"
	"logViewer/core"
)

var addr = flag.String("addr", "127.0.0.1:8080", "listen addr:port")

func main() {
	var ui core.Ui
	ui.Init()

	ui.HttpServer.Addr = *addr

	err := ui.HttpServer.ListenAndServe()
	if err != nil {
		log.Fatalln("ListenAndServe", err)
	}
}
