package main

import (
	"io"
	"log"
	"net"

	"bufio"
	"logViewer/protocol"
)

type oPhandler func(ctx *matchCtx)

type matchCtx struct {
	c      net.Conn
	r      *bufio.Reader
	w      *bufio.Writer
	result matchResult
	dirs   []string
	msg    chan []byte
	err    chan error
	op     chan protocol.OP
	run    bool
	mode   uint8
}

func (ctx *matchCtx) close() {
	defer ctx.c.Close()

	ctx.run = false

	ctx.result.close()
}

type matchServer struct {
	addr     string
	dirs     []string
	handler  map[protocol.OP]oPhandler
	setting  map[protocol.OP]oPhandler
	matching map[protocol.OP]oPhandler
}

func (s *matchServer) init() {
	s.handler = make(map[protocol.OP]oPhandler)
	s.setting = make(map[protocol.OP]oPhandler)
	s.matching = make(map[protocol.OP]oPhandler)

	s.handler[protocol.OP_EXIT] = op_exit
	s.handler[protocol.OP_PING] = op_ping
	s.handler[protocol.OP_LIST] = op_list
	s.handler[protocol.OP_OPEN] = op_open

	s.setting[protocol.OP_EXIT] = op_exit
	s.setting[protocol.OP_PING] = op_ping
	s.setting[protocol.OP_OPEN] = op_open
	s.setting[protocol.SET_STARTTIME] = op_set_starttime
	s.setting[protocol.SET_STOPTIME] = op_set_stoptime
	s.setting[protocol.SET_SEEK] = op_set_seek
	s.setting[protocol.SET_LINE_BUF] = op_set_buf
	s.setting[protocol.SET_LIMIT] = op_set_limit
	s.setting[protocol.SET_TIME_PARSER] = op_set_time_parser
	s.setting[protocol.ADD_MATCH] = op_add_match
	s.setting[protocol.ADD_NOT_MATCH] = op_add_not_match
	s.setting[protocol.ADD_REGEXP] = op_add_regexp
	s.setting[protocol.ADD_NOT_REGEXP] = op_add_not_regexp
	s.setting[protocol.OP_START] = op_start

	s.matching[protocol.OP_EXIT] = op_exit
	s.matching[protocol.OP_PING] = op_ping
	s.matching[protocol.OP_STAT] = op_stat
}

func (s *matchServer) serviceWapper(c net.Conn) {
	ctx := matchCtx{
		c:    c,
		r:    bufio.NewReader(c),
		w:    bufio.NewWriter(c),
		run:  true,
		dirs: s.dirs,
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

func (s *matchServer) getHander(ctx *matchCtx, op protocol.OP) oPhandler {
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
		panic(protocol.UnexpectedOP)
	}

	return h
}

func (s *matchServer) service(ctx *matchCtx) {

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
			op := protocol.ReadOP(ctx.r)
			ctx.op <- op
		}
	}()

	for ctx.run {
		select {
		case e := <-ctx.err:
			if e == io.EOF {
				protocol.WriteOP(ctx.w, protocol.OP_EXIT)
			} else {
				protocol.WriteError(ctx.w, e)
			}
			ctx.run = false

		case op := <-ctx.op:
			h := s.getHander(ctx, op)
			h(ctx)

		case d := <-ctx.msg:
			_, err := ctx.w.Write(d)
			if err != nil {
				panic(err)
			}
		}

		if err := ctx.w.Flush(); err != nil {
			panic(protocol.ErrorIO(err))
		}
	}
}

func (tcp *matchServer) run() {
	if len(tcp.dirs) < 1 {
		tcp.dirs = append(tcp.dirs, "/")
	}

	if tcp.addr == "" {
		tcp.addr = "127.0.0.2:7000"
	}

	l, err := net.Listen("tcp", tcp.addr)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Listen", tcp.addr)

	s := matchServer{}
	s.init()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go s.serviceWapper(conn)
	}
}
