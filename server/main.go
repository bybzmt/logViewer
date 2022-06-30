package main

import (
	"errors"
	"log"
	"net"
	"os"
	"strings"

	"bufio"
	"encoding/binary"
	"logViewer/core/protocol"

	"github.com/integrii/flaggy"
)

var runing = true
var dirs []string

func main() {

	var addr string

	flaggy.StringSlice(&dirs, "", "dir", "limit dir")
	flaggy.String(&addr, "", "addr", "listen addr:port")

	if len(dirs) < 1 {
		dirs = append(dirs, "/")
	}

	if addr == "" {
		addr = "127.0.0.2:7000"
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Listen", addr)

	for runing {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go service(conn)
	}

}

func serviceWapper(c net.Conn) {
	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()

	defer c.Close()

	service(c)
}

func service(c net.Conn) {
	var op protocol.OP

	r := bufio.NewReader(c)

	var files []*os.File

	err := binary.Read(r, binary.BigEndian, &op)
	if err != nil {
		panic("read op " + err.Error())
	}

	switch op {
	case protocol.OP_EXIT:
		runing = false
		return
	case protocol.OP_OPEN:
		file := protocol.ReadString(r)
		f, err := openFile(file)
		if err != nil {
			files = append(files, f)
		}

		err = binary.Write(c, binary.BigEndian, protocol.RESP_OPEN)
		if err != nil {
			panic("Write " + err.Error())
		}

	}
}

func openFile(file string) (*os.File, error) {
	for _, dir := range dirs {
		if strings.HasPrefix(file, dir) {
			return os.Open(file)
		}
	}

	return nil, errors.New("file no perms")
}
