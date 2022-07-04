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
	msg    chan []byte
	err    chan error
	op     chan protocol.OP
	run    bool
	mode   uint8
}

type matchServer struct {
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
		c:   c,
		r:   bufio.NewReader(c),
		w:   bufio.NewWriter(c),
		run: true,
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

func (ctx *matchCtx) close() {
	defer ctx.c.Close()

	ctx.run = false

	ctx.result.close()
}

func op_ping(ctx *matchCtx) {
	protocol.WriteOP(ctx.w, protocol.OP_PONG)
}

func op_exit(ctx *matchCtx) {
	ctx.run = false
}

func op_list(ctx *matchCtx) {
	file := protocol.ReadString(ctx.r)
	files, err := listDirFiles(file)
	if err != nil {
		ctx.err <- err
		return
	}

	protocol.RespListDir(ctx.w, files)
}

func op_open(ctx *matchCtx) {
	file := protocol.ReadString(ctx.r)
	f, err := openFile(file)
	if err != nil {
		ctx.err <- err
		return
	}

	ctx.mode = 1
	match := &matcher{file: f}
	ctx.result.all = append(ctx.result.all, match)
}

func op_set_starttime(ctx *matchCtx) {
	num := protocol.ReadInt64(ctx.r)
	ctx.result.limit = num
}

func op_set_stoptime(ctx *matchCtx) {
	num := protocol.ReadInt64(ctx.r)
	ctx.result.limit = num
}

func op_set_limit(ctx *matchCtx) {
	num := protocol.ReadUint16(ctx.r)
	ctx.result.limit = int64(num)
}

func op_set_buf(ctx *matchCtx) {
	num := protocol.ReadUint16(ctx.r)
	ctx.result.bufSize = int(num)
}

func op_set_seek(ctx *matchCtx) {
	i := len(ctx.result.all) - 1

	num := protocol.ReadInt64(ctx.r)
	ctx.result.all[i].seek = num
}

func op_set_time_parser(ctx *matchCtx) {
	i := len(ctx.result.all) - 1

	expr := protocol.ReadString(ctx.r)
	timeLayout := protocol.ReadString(ctx.r)

	reg, err := perlRegexp(expr)
	if err != nil {
		ctx.err <- err
		return
	}

	ctx.result.all[i].timeParser = timeParserRegexp(reg, timeLayout)
}

func op_add_match(ctx *matchCtx) {
	i := len(ctx.result.all) - 1

	strs := protocol.ReadStrings(ctx.r)

	ctx.result.all[i].filters = append(ctx.result.all[i].filters, filterContains(strs))
}

func op_add_not_match(ctx *matchCtx) {
	i := len(ctx.result.all) - 1

	strs := protocol.ReadStrings(ctx.r)

	ctx.result.all[i].filters = append(ctx.result.all[i].filters, filterNot(filterContains(strs)))
}

func op_add_regexp(ctx *matchCtx) {
	i := len(ctx.result.all) - 1

	expr := protocol.ReadString(ctx.r)
	reg, err := perlRegexp(expr)
	if err != nil {
		ctx.err <- err
		return
	}

	ctx.result.all[i].filters = append(ctx.result.all[i].filters, filterRegexp(reg))
}

func op_add_not_regexp(ctx *matchCtx) {
	i := len(ctx.result.all) - 1

	expr := protocol.ReadString(ctx.r)
	reg, err := perlRegexp(expr)
	if err != nil {
		ctx.err <- err
		return
	}

	ctx.result.all[i].filters = append(ctx.result.all[i].filters, filterNot(filterRegexp(reg)))
}

func op_start(ctx *matchCtx) {
	ctx.mode = 2

	err := ctx.result.init()
	if err != nil {
		ctx.err <- err
		return
	}

	protocol.WriteOK(ctx.w)

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
			l, err := ctx.result.match()
			if err != nil {
				ctx.err <- err
				return
			}
			ctx.msg <- l.data
		}
	}()
}

func op_stat(ctx *matchCtx) {
}
