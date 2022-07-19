package cli

import (
	"context"
	"io"
	"log"
	"logViewer/find"
	"os"
	"os/exec"
	"time"
)

func Dial(cmd string) (*find.Client, error) {
	ctx, cancel := context.WithCancel(context.Background())
	c := exec.CommandContext(ctx, cmd)
	c.Stderr = os.Stderr

	log.Println(1)
	w, err := c.StdinPipe()
	if err != nil {
		return nil, err
	}

	log.Println(2)
	r, err := c.StdoutPipe()
	if err != nil {
		return nil, err
	}

	conn := rw2{
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

	log.Println(3)
	return find.NewClient(&conn), nil
}

type rw2 struct {
	r      io.ReadCloser
	w      io.WriteCloser
	cancel context.CancelFunc
	ctx    context.Context
	timer  *time.Timer
	cmd    *exec.Cmd
}

func (c *rw2) Read(b []byte) (n int, err error) {
	return c.r.Read(b)
}

func (c *rw2) Write(b []byte) (n int, err error) {
	return c.w.Write(b)
}

func (c *rw2) Close() error {
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

func (c *rw2) SetDeadline(t time.Time) error {
	if c.timer != nil {
		c.timer.Stop()
	}

	c.timer = time.AfterFunc(time.Until(t), c.cancel)
	return nil
}
