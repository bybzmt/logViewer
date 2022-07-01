package main

import (
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"strings"

	"bufio"
	"logViewer/core/protocol"

	"github.com/integrii/flaggy"
)

var runing = true
var limitDirs []string

func main() {

	var addr string

	flaggy.StringSlice(&limitDirs, "", "dir", "limit dir")
	flaggy.String(&addr, "", "addr", "listen addr:port")

	if len(limitDirs) < 1 {
		limitDirs = append(limitDirs, "/")
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
	w := bufio.NewWriter(c)

	var match matcher

	for {
		op = protocol.ReadOP(r)

		switch op {
		case protocol.OP_EXIT:
			runing = false
			return
		case protocol.OP_OPEN:
			file := protocol.ReadString(r)
			f, err := openFile(file)
			if err == nil {
				files = append(files, f)
			}

			protocol.RespOpenFile(w, err)

		case protocol.OP_LIST:
			file := protocol.ReadString(r)
			files, err := listDirFiles(file)

			protocol.RespListDir(w, files, err)

		case protocol.OP_CANCEL:

		case protocol.OP_START:
			startMatch(&match, files, w)

		}

		if err := w.Flush(); err != nil {
			panic(protocol.ErrorIO(err))
		}
	}
}

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

type Filter func([]byte) bool
type TimerParser func([]byte) (int64, bool)

type matcher struct {
	files       *os.File
	r           *bufio.Reader
	timerParser TimerParser
	filters     []Filter
	startTime   int64
	endTime     int64
	tailf       bool
	seek        int64
	limit       uint16
	count       uint32
	buf         *line
}

type line struct {
	time int64
	data []byte
}

func (m *matcher) Match() (*line, error) {
	var l line

	for {
		buf, prefix, err := m.r.ReadLine()
		if err != nil {
			return nil, err
		}

		if !prefix {
			l.data = buf
			break
		}
	}

	for {

		t, ok := m.timerParser(l.data)
		if ok {
			new := line{
				time: t,
				data: buf,
			}

			if m.buf != nil {
				old := m.buf
				m.buf = &new
				return old
			}
		}
	}

}

func startMatch(m *matcher, fs []*os.File, w io.Writer) {

	var rs []*bufio.Reader

	for _, f := range fs {
		r := bufio.NewReaderSize(f, 1024*128)
		rs = append(rs, r)
	}

	for _, r := range rs {
		buf, _, err := r.ReadLine()
		if err != nil {

			var new []*bufio.Reader
			for _, r2 := range rs {
				if r != r2 {
					new = append(new, r2)
				}
			}
			rs = new

		}

	}

}
