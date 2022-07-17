package find

import (
	"bufio"
	"errors"
	"io"
	"time"
)

type Client struct {
	c       Conn
	rw      *bufio.ReadWriter
	Timeout time.Duration
}

func NewClient(c Conn) *Client {
	var rs Client
	rs.c = c
	w := bufio.NewWriter(c)
	r := bufio.NewReader(c)
	rs.rw = bufio.NewReadWriter(r, w)
	return &rs
}

func (c *Client) Glob(pattern string) (out []string, err error) {
	defer tryErr(&err)

	c.c.SetDeadline(time.Now().Add(c.Timeout))

	write(c.rw, OP_GLOB, []byte(pattern))

	if e := c.rw.Flush(); e != nil {
		panic(ErrorIO(e))
	}

	if err := readJson(c.rw, OP_GLOB, &out); err != nil {
		return nil, err
	}

	return
}

func (c *Client) Open(m *MatchParam) (err error) {
	defer tryErr(&err)

	c.c.SetDeadline(time.Now().Add(c.Timeout))

	writeJson(c.rw, OP_GREP, m, nil)

	if e := c.rw.Flush(); e != nil {
		panic(ErrorIO(e))
	}

	return
}

func (c *Client) Read() (data []byte, err error) {
	defer tryErr(&err)

	c.c.SetDeadline(time.Now().Add(c.Timeout))

	write(c.rw, OP_READ, nil)

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

func (c *Client) Close() (err error) {
	defer tryErr(&err)

	c.c.SetDeadline(time.Now().Add(c.Timeout))

	write(c.rw, OP_EXIT, nil)

	if e := c.rw.Flush(); e != nil {
		panic(ErrorIO(e))
	}

	c.c.Close()

	return
}
