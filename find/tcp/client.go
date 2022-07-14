package tcp

import (
	"bufio"
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
	err := writeOP(rs.rw, OP_GLOB)
	if err != nil {
		return nil, err
	}
	err = writeString(rs.rw, pattern)
	if err != nil {
		return nil, err
	}

	err = rs.rw.Flush()
	if err != nil {
		return nil, err
	}

	return readRespListDir(rs.rw)
}

func (rs *Client) Open(m *Match) error {
	err := writeGrep(rs.rw, m)
	if err != nil {
		return err
	}

	err = rs.rw.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (rs *Client) Close() {
	rs.c.Close()
}

func (rs *Client) Read() ([]byte, error) {
	rs.c.SetReadDeadline(time.Now().Add(time.Second * 3))

	op, err := readOP(rs.rw)
	if err != nil {
		return nil, err
	}

	switch op {
	case OP_EXIT:
		err = writeOP(rs.rw, OP_EXIT)
		if err != nil {
			log.Println("exit op write", err)
		}

		return nil, io.EOF
	case OP_ERR:
		return nil, readErr(rs.rw)
	case OP_MSG:
		return readBytes(rs.rw)
	default:
		return nil, UnexpectedOP
	}
}
