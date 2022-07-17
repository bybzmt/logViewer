package main

import (
	"logViewer/find/cli"
	"os"
)

func main() {
	s := cli.Server{}
	s.Init()
	s.Run(os.Stdin, os.Stdout)
}
