package cli

import (
	"logViewer/find"
	"os"
	"time"
)

type Server struct {
	find.Server
}

func (s *Server) Run(stdin, stdout *os.File) {
	s.Init()

	c := rw{
		r: stdin,
		w: stdout,
	}

	s.Service(&c)
}

type rw struct {
	r *os.File
	w *os.File
}

func (c *rw) Read(b []byte) (n int, err error) {
	return c.r.Read(b)
}

func (c *rw) Write(b []byte) (n int, err error) {
	return c.w.Write(b)
}

func (c *rw) Close() error {
	c.r.Close()
	c.w.Close()
	return nil
}

func (c *rw) SetDeadline(t time.Time) error {
	c.r.SetReadDeadline(t)
	c.w.SetWriteDeadline(t)
	return nil
}
