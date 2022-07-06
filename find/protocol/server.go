package protocol

import (
	"io"
	"log"
	"net"

	"bufio"
	"logViewer/find"
)

type oPhandler func(ctx *matchCtx)

type matchCtx struct {
	c       net.Conn
	r       *bufio.Reader
	w       *bufio.Writer
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

	ctx.run = false

	ctx.matcher.Close()
}

type MatchServer struct {
	Addr     string
	Dirs     []string
	handler  map[OP]oPhandler
	setting  map[OP]oPhandler
	matching map[OP]oPhandler
}

func (s *MatchServer) init() {
	s.handler = make(map[OP]oPhandler)
	s.setting = make(map[OP]oPhandler)
	s.matching = make(map[OP]oPhandler)

	s.handler[OP_EXIT] = op_exit
	s.handler[OP_PING] = op_ping
	s.handler[OP_LIST] = op_list
	s.handler[OP_OPEN] = op_open

	s.setting[OP_EXIT] = op_exit
	s.setting[OP_PING] = op_ping
	s.setting[OP_OPEN] = op_open
	s.setting[SET_STARTTIME] = op_set_starttime
	s.setting[SET_STOPTIME] = op_set_stoptime
	s.setting[SET_LINE_BUF] = op_set_buf
	s.setting[SET_LIMIT] = op_set_limit
	s.setting[SET_TIME_PARSER] = op_set_time_parser
	s.setting[ADD_MATCH] = op_add_match
	s.setting[ADD_NOT_MATCH] = op_add_not_match
	s.setting[ADD_REGEXP] = op_add_regexp
	s.setting[ADD_NOT_REGEXP] = op_add_not_regexp
	s.setting[OP_START] = op_start

	s.matching[OP_EXIT] = op_exit
	s.matching[OP_PING] = op_ping
	s.matching[OP_STAT] = op_stat
}

func (s *MatchServer) serviceWapper(c net.Conn) {
	ctx := matchCtx{
		c:    c,
		r:    bufio.NewReader(c),
		w:    bufio.NewWriter(c),
		run:  true,
		dirs: s.Dirs,
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

func (s *MatchServer) getHander(ctx *matchCtx, op OP) oPhandler {
	var h oPhandler
	var ok bool

	switch ctx.mode {
	case 0:
		h, ok = s.handler[op]
	case 1:
		h, ok = s.setting[op]
	case 2:
		h, ok = s.matching[op]
	}

	if !ok {
		panic(UnexpectedOP)
	}

	return h
}

func (s *MatchServer) service(ctx *matchCtx) {

	go func() {
		defer func() {
			err := recover()
			if err != nil {
				if e, ok := err.(error); ok {
					ctx.err <- e
				} else {
					log.Println(err)
				}
			}
		}()

		for ctx.run {
			op := ReadOP(ctx.r)
			ctx.op <- op
		}
	}()

	for ctx.run {
		select {
		case e := <-ctx.err:
			if e == io.EOF {
				WriteOP(ctx.w, OP_EXIT)
			} else {
				WriteError(ctx.w, e)
			}
			ctx.run = false

		case op := <-ctx.op:
			h := s.getHander(ctx, op)
			h(ctx)

		case d := <-ctx.msg:
			WriteOP(ctx.w, OP_MSG)
			WriteBytes(ctx.w, d)
		}

		if err := ctx.w.Flush(); err != nil {
			panic(ErrorIO(err))
		}
	}
}
