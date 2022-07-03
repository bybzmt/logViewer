package main

import (
	"flag"
	"log"

	"logViewer/client"

	"github.com/webview/webview"
)

func runUI() {
	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("Basic Example")
	w.SetSize(800, 600, webview.HintNone)
	//w.SetHtml("Thanks for using webview!")
	w.Navigate("http://" + *addr)
	w.Run()
}

var addr = flag.String("addr", "127.0.0.1:8080", "listen addr:port")

func main() {
	var ui client.Ui
	ui.Init()

	ui.HttpServer.Addr = *addr

	go func() {
		err := ui.HttpServer.ListenAndServe()
		if err != nil {
			log.Fatalln("ListenAndServe", err)
		}
	}()

	runUI()
}
