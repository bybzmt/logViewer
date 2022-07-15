package tcp

import (
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"bufio"
	"logViewer/find"
)

type oPhandler func(ctx *matchCtx)

type matchCtx struct {
	c       net.Conn
	rw      *bufio.ReadWriter
	matcher find.Matcher
	dirs    []string
	msg     chan []byte
	err     chan error
	op      chan OP
	run     bool
	mode    uint8
}

func (ctx *matchCtx) close() {
	defer ctx.c.Close()
}

type MatchServer struct {
	Addr string
	Dirs []string
}

func (s *MatchServer) Service(c net.Conn) {

	r := bufio.NewReader(c)
	w := bufio.NewWriter(c)

	ctx := matchCtx{
		c:  c,
		rw: bufio.NewReadWriter(r, w),
	}
	defer ctx.close()

	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()

	s.service(&ctx)
}

func (s *MatchServer) service(ctx *matchCtx) {
	ctx.c.SetDeadline(time.Now().Add(time.Second * 5))

	op, buf := read(ctx.rw)

	switch op {
	case OP_GLOB:
		s.serviceGlob(ctx, buf)
	case OP_GREP:
		s.serviceGrep(ctx, buf)
	default:
		panic(UnexpectedOP)
	}

	if err := ctx.rw.Flush(); err != nil {
		panic(ErrorIO(err))
	}
}

func (s *MatchServer) Glob(name string) ([]string, error) {
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

func (s *MatchServer) newMatch(m *Match) (*find.Matcher, error) {
	f := find.Matcher{
		StartTime: m.StartTime,
		EndTime:   m.EndTime,
		Limit:     m.Limit,
		BufSize:   m.BufSize,
	}

	for _, fi := range m.Files {
		if !s.hasPrefix(fi.Name) {
			return nil, AccessDenied
		}

		reg, err := find.PerlRegexp(fi.TimeRegex)
		if err != nil {
			return nil, err
		}

		var fs []find.Filter
		for _, keys := range fi.Contains {
			fs = append(fs, find.FilterContains(keys))
		}

		f.All = append(f.All, find.File{
			Name:       fi.Name,
			Filters:    fs,
			TimeParser: find.TimeParserRegexp(reg, fi.TimeLayout),
		})
	}

	return &f, nil
}

func (s *MatchServer) hasPrefix(name string) bool {
	for _, dir := range s.Dirs {
		if strings.HasPrefix(name, dir) {
			return true
		}
	}

	return false
}

func (s *MatchServer) serviceGlob(ctx *matchCtx, buf []byte) {
	dirs, err := s.Glob(string(buf))

	writeJson(ctx.rw, OP_GLOB, dirs, err)
}

func (s *MatchServer) serviceGrep(ctx *matchCtx, buf []byte) {
	var m Match
	toJson(buf, &m)

	//log.Printf("match %#v\n", m)

	f, err := s.newMatch(&m)
	if err != nil {
		write(ctx.rw, OP_ERR, []byte(err.Error()))
		return
	}
	defer f.Close()

	err = f.Init()
	if err != nil {
		write(ctx.rw, OP_ERR, []byte(err.Error()))
		return
	}

	for {
		ctx.c.SetDeadline(time.Now().Add(time.Second * 5))

		op, _ := read(ctx.rw)

		switch op {
		case OP_NEXT:
			d, err := f.Match()
			log.Println(string(d))

			if err != nil {
				if err == io.EOF {
					write(ctx.rw, OP_EXIT, nil)
				} else {
					write(ctx.rw, OP_ERR, []byte(err.Error()))
				}
				return
			}

			write(ctx.rw, OP_MSG, d)

			if err := ctx.rw.Flush(); err != nil {
				panic(ErrorIO(err))
			}

		case OP_EXIT:
			return
		case OP_STAT:
			return
		default:
			panic(UnexpectedOP)
		}
	}

}

func (s *MatchServer) Run() {
	if len(s.Dirs) < 1 {
		s.Dirs = append(s.Dirs, "/")
	}

	if s.Addr == "" {
		s.Addr = "127.0.0.2:7000"
	}

	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Listen", s.Addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go s.Service(conn)
	}
}
