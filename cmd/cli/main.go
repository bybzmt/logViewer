package main

import (
	"log"
	"logViewer/find/cli"
	"os"
)

func main() {
	log.Println("runing.")

	s := cli.Server{}
	s.Init()
	s.Run(os.Stdin, os.Stdout)
}
