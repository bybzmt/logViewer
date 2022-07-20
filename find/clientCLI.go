package find

import (
	"context"
	"io"
	"os"
	"os/exec"
	"time"
)

func NewClientCLI(cmd string) (Client, error) {
	ctx, cancel := context.WithCancel(context.Background())
	c := exec.CommandContext(ctx, cmd)
	c.Stderr = os.Stderr

	w, err := c.StdinPipe()
	if err != nil {
		return nil, err
	}

	r, err := c.StdoutPipe()
	if err != nil {
		return nil, err
	}

	conn := connCli{
		r:      r,
		w:      w,
		cancel: cancel,
		cmd:    c,
		ctx:    ctx,
	}

	err = c.Start()
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Millisecond * 100)

	return newClient(&conn), nil
}

type connCli struct {
	r      io.ReadCloser
	w      io.WriteCloser
	cancel context.CancelFunc
	ctx    context.Context
	timer  *time.Timer
	cmd    *exec.Cmd
}

func (c *connCli) Read(b []byte) (n int, err error) {
	return c.r.Read(b)
}

func (c *connCli) Write(b []byte) (n int, err error) {
	return c.w.Write(b)
}

func (c *connCli) Close() error {
	c.r.Close()
	c.w.Close()

	if c.timer != nil {
		c.timer.Stop()
	}

	select {
	case <-c.ctx.Done():
	default:
		c.cancel()
	}

	return nil
}

func (c *connCli) SetDeadline(t time.Time) error {
	if c.timer != nil {
		c.timer.Stop()
	}

	c.timer = time.AfterFunc(time.Until(t), c.cancel)
	return nil
}
