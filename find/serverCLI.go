package find

import (
	"os"
	"time"
)

type ServerCLI struct {
	Server
}

func (s *ServerCLI) Run(stdin, stdout *os.File) {
	s.Init()

	c := connFile{
		r: stdin,
		w: stdout,
	}

	s.Service(&c)
}

type connFile struct {
	r *os.File
	w *os.File
}

func (c *connFile) Read(b []byte) (n int, err error) {
	return c.r.Read(b)
}

func (c *connFile) Write(b []byte) (n int, err error) {
	return c.w.Write(b)
}

func (c *connFile) Close() error {
	c.r.Close()
	c.w.Close()
	return nil
}

func (c *connFile) SetDeadline(t time.Time) error {
	c.r.SetReadDeadline(t)
	c.w.SetWriteDeadline(t)
	return nil
}
