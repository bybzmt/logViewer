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
	c    Conn
	rw   *bufio.ReadWriter
}

func Dial(addr string) (*Client, error) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewClient(c)
}

func NewClient(c Conn) (*Client, error) {
	var rs Client
	rs.c = c
	w := bufio.NewWriter(c)
	r := bufio.NewReader(c)
	rs.rw = bufio.NewReadWriter(r, w)
	return &rs, nil
}

func (c *Client) Glob(pattern string) (out []string, err error) {
	defer tryErr(&err)

	write(c.rw, OP_GLOB, []byte(pattern))

	if e := c.rw.Flush(); e != nil {
		panic(ErrorIO(e))
	}

	if err := readJson(c.rw, OP_GLOB, &out); err != nil {
		return nil, err
	}

	return
}

func (c *Client) Open(m *Match) (err error) {
	defer tryErr(&err)

	writeJson(c.rw, OP_GREP, m, nil)

	if e := c.rw.Flush(); e != nil {
		panic(ErrorIO(e))
	}

	return
}

func (c *Client) Close() {
	write(c.rw, OP_EXIT, nil)

	if e := c.rw.Flush(); e != nil {
		log.Println("close", e)
	}

	c.c.Close()
}

func tryErr(err *error) {
	switch p := recover(); e := p.(type) {
	case ErrorIO:
		*err = e
	case ErrorProtocol:
		*err = e
	default:
		panic(p)
	}
}

func (c *Client) Read() (data []byte, err error) {
	c.c.SetDeadline(time.Now().Add(time.Second * 5))

	defer tryErr(&err)

	write(c.rw, OP_NEXT, nil)

	if e := c.rw.Flush(); e != nil {
		panic(ErrorIO(e))
	}

	op, buf := read(c.rw)

	switch op {
	case OP_EOF:
		write(c.rw, OP_EXIT, nil)

		err = io.EOF
	case OP_ERR:
		write(c.rw, OP_EXIT, nil)

		err = errors.New(string(buf))
	case OP_MSG:
		write(c.rw, OP_EXIT, nil)

		data = buf
	default:
		panic(unexpectedOP(op))
	}

	if e := c.rw.Flush(); e != nil {
		panic(ErrorIO(e))
	}

	return
}
