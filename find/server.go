package find

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
	"time"

	"bufio"
)

type matchCtx struct {
	c       Conn
	rw      *bufio.ReadWriter
	matcher *Matcher
}

func (ctx *matchCtx) close() {
	defer ctx.c.Close()

	if ctx.matcher != nil {
		ctx.matcher.Close()
	}
}

type Server struct {
	Dirs    []string
	Timeout time.Duration
}

func (s *Server) Init() {
	if len(s.Dirs) == 0 {
		s.Dirs = append(s.Dirs, "/")
	}

	if s.Timeout == 0 {
		s.Timeout = time.Second * 5
	}
}

func (s *Server) Service(c Conn) {
	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()

	if s.Timeout < 1 {
		s.Timeout = time.Second * 5
	}

	r := bufio.NewReader(c)
	w := bufio.NewWriter(c)

	ctx := matchCtx{
		c:  c,
		rw: bufio.NewReadWriter(r, w),
	}
	defer ctx.close()

	s.service(&ctx)
}

func (s *Server) service(ctx *matchCtx) {
	for {
		ctx.c.SetDeadline(time.Now().Add(s.Timeout))

		op, buf := read(ctx.rw)

		switch op {
		case OP_GLOB:
			s.serviceGlob(ctx, buf)
		case OP_GREP:
			s.serviceGrep(ctx, buf)
		case OP_READ:
			s.serviceRead(ctx, buf)
		case OP_STAT:
			s.serviceStat(ctx, buf)
		case OP_EXIT:
			return
		default:
			panic(unexpectedOP(op))
		}

		if err := ctx.rw.Flush(); err != nil {
			panic(ErrorIO(err))
		}
	}
}

func (s *Server) Glob(name string) ([]string, error) {
	dirs, err := fs.Glob(os.DirFS(""), strings.TrimLeft(name, "/"))
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(dirs))

	for _, dir := range dirs {
		dir = "/" + dir
		if s.hasPrefix(dir) {
			out = append(out, dir)
		}
	}

	return out, nil
}

func (s *Server) newMatch(m *MatchParam) (*Matcher, error) {
	f := Matcher{
		StartTime: m.StartTime,
		EndTime:   m.EndTime,
		Limit:     m.Limit,
		BufSize:   m.BufSize,
	}

	for _, fi := range m.Files {
		if !s.hasPrefix(fi.Name) {
			return nil, ErrorUser(fmt.Errorf("access denied file:%s", fi.Name))
		}

		reg, err := PerlRegexp(fi.TimeRegex)
		if err != nil {
			return nil, err
		}

		var fs []Filter
		for _, keys := range fi.Contains {
			fs = append(fs, FilterContains(keys))
		}

		f.All = append(f.All, File{
			Name:       fi.Name,
			Filters:    fs,
			TimeParser: TimeParserRegexp(reg, fi.TimeLayout),
		})
	}

	return &f, nil
}

func (s *Server) hasPrefix(name string) bool {
	for _, dir := range s.Dirs {
		if strings.HasPrefix(name, dir) {
			return true
		}
	}

	return false
}

func (s *Server) serviceGlob(ctx *matchCtx, buf []byte) {
	dirs, err := s.Glob(string(buf))

	writeJson(ctx.rw, OP_GLOB, dirs, err)
}

func (s *Server) serviceGrep(ctx *matchCtx, buf []byte) {

	if ctx.matcher != nil {
		write(ctx.rw, OP_ERR, []byte(repeatOpenFile.Error()))
		return
	}

	var m MatchParam
	err := json.Unmarshal(buf, &m)
	if err != nil {
		panic(ErrorProtocol(err))
	}

	//log.Printf("match %#v\n", m)

	f, err := s.newMatch(&m)
	if err != nil {
		write(ctx.rw, OP_ERR, []byte(err.Error()))
		return
	}

	err = f.Init()
	if err != nil {
		f.Close()

		write(ctx.rw, OP_ERR, []byte(err.Error()))
		return
	}

	ctx.matcher = f
}

func (s *Server) serviceRead(ctx *matchCtx, buf []byte) {

	if ctx.matcher == nil {
		write(ctx.rw, OP_ERR, []byte(notOpenFile.Error()))
		return
	}

	d, err := ctx.matcher.Match()
	log.Println(string(d))

	if err != nil {
		if err == io.EOF {
			write(ctx.rw, OP_EOF, nil)
		} else {
			write(ctx.rw, OP_ERR, []byte(err.Error()))
		}
		return
	}

	write(ctx.rw, OP_MSG, d)
}

func (s *Server) serviceStat(ctx *matchCtx, buf []byte) {

	if ctx.matcher == nil {
		write(ctx.rw, OP_ERR, []byte(notOpenFile.Error()))
		return
	}

	out, err := ctx.matcher.Stat()

	writeJson(ctx.rw, OP_STAT, out, err)
}
