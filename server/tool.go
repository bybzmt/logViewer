package main

import (
	"io/fs"
	"logViewer/protocol"
	"os"
	"strings"
)

func openFile(file string, pre []string) (*os.File, error) {
	for _, dir := range pre {
		if strings.HasPrefix(file, dir) {
			return os.Open(file)
		}
	}

	return nil, protocol.AccessDenied
}

func listDirFiles(dir string, pre []string) ([]string, error) {
	for _, pre := range pre {
		if strings.HasPrefix(dir, pre) {
			f := os.DirFS(dir)
			return fs.Glob(f, "*")
		}
	}

	return nil, protocol.AccessDenied
}
