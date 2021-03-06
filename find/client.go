package find

import (
	"bufio"
	"errors"
	"io"
	"time"
)

type client struct {
	c       Conn
	rw      *bufio.ReadWriter
	Timeout time.Duration
	err     error
}

func newClient(c Conn) Client {
	var rs client
	rs.c = c
	w := bufio.NewWriter(c)
	r := bufio.NewReader(c)
	rs.rw = bufio.NewReadWriter(r, w)

	if rs.Timeout < 1 {
		rs.Timeout = time.Second * 5
	}

	return &rs
}

func (c *client) Glob(pattern string) (out []string, err error) {
	defer c.tryErr(&err)

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

func (c *client) Open(m *MatchParam) (err error) {
	defer c.tryErr(&err)

	c.c.SetDeadline(time.Now().Add(c.Timeout))

	writeJson(c.rw, OP_GREP, m, nil)

	if e := c.rw.Flush(); e != nil {
		panic(ErrorIO(e))
	}

	op, buf := read(c.rw)

	switch op {
	case OP_ERR:
		err = errors.New(string(buf))
	case OP_OK:
	default:
		panic(unexpectedOP(op))
	}

	return
}

func (c *client) Read() (data []byte, err error) {
	defer c.tryErr(&err)

	c.c.SetDeadline(time.Now().Add(c.Timeout))

	write(c.rw, OP_READ, nil)

	if e := c.rw.Flush(); e != nil {
		panic(ErrorIO(e))
	}

	op, buf := read(c.rw)

	switch op {
	case OP_EOF:
		err = io.EOF
	case OP_ERR:
		err = errors.New(string(buf))
	case OP_MSG:
		data = buf
	default:
		panic(unexpectedOP(op))
	}

	return
}

func (c *client) Ping() (err error) {
	defer c.tryErr(&err)

	write(c.rw, OP_PING, nil)
	if e := c.rw.Flush(); e != nil {
		panic(ErrorIO(e))
	}

	op, buf := read(c.rw)

	switch op {
	case OP_ERR:
		err = errors.New(string(buf))
	case OP_OK:
		return nil
	default:
		panic(unexpectedOP(op))
	}

	return
}

func (c *client) Close() (err error) {
	defer func() {
		err = c.c.Close()
		c.tryErr(&err)
	}()

	if c.err == nil {
		c.c.SetDeadline(time.Now().Add(c.Timeout))

		write(c.rw, OP_EXIT, nil)

		if e := c.rw.Flush(); e != nil {
			panic(ErrorIO(e))
		}
	}

	return
}

func (c *client) tryErr(err *error) {
	switch p := recover(); e := p.(type) {
	case ErrorIO:
		c.err = e
		*err = e
	case ErrorProtocol:
		c.err = e
		*err = e
	case nil:
	default:
		panic(p)
	}
}
