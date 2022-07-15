package tcp

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"time"
)

type Client struct {
	Addr string
	c    net.Conn
	rw   *bufio.ReadWriter
}

func Dial(addr string) (*Client, error) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	var rs Client
	rs.c = c
	w := bufio.NewWriter(c)
	r := bufio.NewReader(c)
	rs.rw = bufio.NewReadWriter(r, w)
	return &rs, nil
}

func (rs *Client) Glob(pattern string) ([]string, error) {

	write(rs.rw, OP_GLOB, []byte(pattern))

	if err := rs.rw.Flush(); err != nil {
		panic(ErrorIO(err))
	}

	files := []string{}

	if err := readJson(rs.rw, OP_GLOB, &files); err != nil {
		return nil, err
	}

	return files, nil
}

func (rs *Client) Open(m *Match) error {
	writeJson(rs.rw, OP_GREP, m, nil)

	if err := rs.rw.Flush(); err != nil {
		panic(ErrorIO(err))
	}

	return nil
}

func (rs *Client) Close() {
	write(rs.rw, OP_EXIT, nil)

	if err := rs.rw.Flush(); err != nil {
		log.Println("close", err)
	}

	rs.c.Close()
}

func (rs *Client) Read() ([]byte, error) {
	rs.c.SetDeadline(time.Now().Add(time.Second * 5))

	write(rs.rw, OP_NEXT, nil)
	if err := rs.rw.Flush(); err != nil {
		panic(ErrorIO(err))
	}

	op, buf := read(rs.rw)

	defer func() {
		if err := rs.rw.Flush(); err != nil {
			panic(ErrorIO(err))
		}
	}()

	switch op {
	case OP_EXIT:
		write(rs.rw, OP_EXIT, nil)

		return nil, io.EOF
	case OP_ERR:
		write(rs.rw, OP_EXIT, nil)

		return nil, errors.New(string(buf))
	case OP_MSG:
		write(rs.rw, OP_EXIT, nil)

		return buf, nil
	default:
		panic(UnexpectedOP)
	}
}
