package main

import (
	"log"
	"logViewer/find"
	"os"
)

func main() {
	log.Println("runing.")

	s := find.ServerCLI{}
	s.Init()
	s.Run(os.Stdin, os.Stdout)
}
