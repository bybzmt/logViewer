package find

import (
	"io"
	"time"

	"golang.org/x/crypto/ssh"
)

func NewClientSSH(sh *ssh.Client, cmd string) (Client, error) {
	se, err := sh.NewSession()
	if err != nil {
		return nil, err
	}

	w, err := se.StdinPipe()
	if err != nil {
		return nil, err
	}

	r, err := se.StdoutPipe()
	if err != nil {
		return nil, err
	}

	se.Start(cmd)

	conn := connSSH{
		r:   r,
		w:   w,
		cmd: se,
		cancel: func() {
			se.Signal(ssh.SIGKILL)
		},
	}

	return newClient(&conn), nil
}

type connSSH struct {
	r      io.Reader
	w      io.WriteCloser
	cancel func()
	timer  *time.Timer
	cmd    *ssh.Session
}

func (c *connSSH) Read(b []byte) (n int, err error) {
	return c.r.Read(b)
}

func (c *connSSH) Write(b []byte) (n int, err error) {
	return c.w.Write(b)
}

func (c *connSSH) Close() error {
	c.w.Close()

	if c.timer != nil {
		c.timer.Stop()
	}

	c.cancel()
	c.cmd.Close()

	return nil
}

func (c *connSSH) SetDeadline(t time.Time) error {
	if c.timer != nil {
		c.timer.Stop()
	}

	c.timer = time.AfterFunc(time.Until(t), c.cancel)
	return nil
}
