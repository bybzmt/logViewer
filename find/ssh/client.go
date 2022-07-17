package ssh

import (
	"io"
	"logViewer/find"
	"time"

	"golang.org/x/crypto/ssh"
)

func SSHRun(sh *ssh.Client, cmd string) (*find.Client, error) {
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

	conn := rw2{
		r:   r,
		w:   w,
		cmd: se,
		cancel: func() {
			se.Signal(ssh.SIGKILL)
		},
	}

	return find.NewClient(&conn), nil
}

type rw2 struct {
	r      io.Reader
	w      io.WriteCloser
	cancel func()
	timer  *time.Timer
	cmd    *ssh.Session
}

func (c *rw2) Read(b []byte) (n int, err error) {
	return c.r.Read(b)
}

func (c *rw2) Write(b []byte) (n int, err error) {
	return c.w.Write(b)
}

func (c *rw2) Close() error {
	c.w.Close()

	if c.timer != nil {
		c.timer.Stop()
	}

	c.cancel()
	c.cmd.Close()

	return nil
}

func (c *rw2) SetDeadline(t time.Time) error {
	if c.timer != nil {
		c.timer.Stop()
	}

	c.timer = time.AfterFunc(time.Until(t), c.cancel)
	return nil
}
