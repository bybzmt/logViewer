package main

import (
	"io/fs"
	"logViewer/core/protocol"
	"os"
	"strings"
)

func openFile(file string) (*os.File, error) {
	for _, dir := range limitDirs {
		if strings.HasPrefix(file, dir) {
			return os.Open(file)
		}
	}

	return nil, protocol.AccessDenied
}

func listDirFiles(dir string) ([]string, error) {
	for _, pre := range limitDirs {
		if strings.HasPrefix(dir, pre) {
			f := os.DirFS(dir)
			return fs.Glob(f, "*")
		}
	}

	return nil, protocol.AccessDenied
}
