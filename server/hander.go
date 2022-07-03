package main

import (
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
	msg    chan int
	run    bool
}

type matchServer struct {
	handler map[protocol.OP]oPhandler
}

func (s *matchServer) init() {
	s.handler[protocol.OP_EXIT] = op_exit
	s.handler[protocol.OP_PING] = op_ping
	s.handler[protocol.OP_LIST] = op_list
	s.handler[protocol.OP_OPEN] = op_open
	s.handler[protocol.SET_LINE_BUF] = op_set_buf
	s.handler[protocol.SET_LIMIT] = op_set_limit
	s.handler[protocol.SET_TIME_PARSER] = op_set_time_parser
	s.handler[protocol.ADD_MATCH] = op_add_match
	s.handler[protocol.ADD_NOT_MATCH] = op_add_not_match
	s.handler[protocol.ADD_REGEXP] = op_add_regexp
	s.handler[protocol.ADD_NOT_REGEXP] = op_add_not_regexp
}

func (s *matchServer) serviceWapper(c net.Conn) {
	defer c.Close()

	ctx := matchCtx{
		c:   c,
		r:   bufio.NewReader(c),
		w:   bufio.NewWriter(c),
		run: true,
	}

	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()

	s.service(&ctx)
}

func (s *matchServer) service(ctx *matchCtx) {
	for ctx.run {
		op := protocol.ReadOP(ctx.r)

		h, ok := s.handler[op]
		if !ok {
			panic(protocol.UnexpectedOP)
		}

		h(ctx)

		if err := ctx.w.Flush(); err != nil {
			panic(protocol.ErrorIO(err))
		}
	}
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

	protocol.RespListDir(ctx.w, files, err)
}

func op_open(ctx *matchCtx) {
	file := protocol.ReadString(ctx.r)
	f, err := openFile(file)
	if err != nil {
		protocol.WriteError(ctx.w, err)
	}

	match := &matcher{file: f}
	ctx.result.all = append(ctx.result.all, match)
	protocol.WriteOP(ctx.w, protocol.OP_OK)
}

func op_set_limit(ctx *matchCtx) {
	num := protocol.ReadUint16(ctx.r)
	ctx.result.limit = int64(num)
}

func op_set_buf(ctx *matchCtx) {
	num := protocol.ReadUint16(ctx.r)
	ctx.result.bufSize = int64(num)
}

func op_set_time_parser(ctx *matchCtx) {
	i := len(ctx.result.all)
	if i == 0 {
		return
	}

	expr := protocol.ReadString(ctx.r)
	timeLayout := protocol.ReadString(ctx.r)

	reg, err := perlRegexp(expr)
	if err != nil {
		protocol.WriteError(ctx.w, err)
		return
	}

	protocol.WriteOK(ctx.w)
	ctx.result.all[i].timeParser = timeParserRegexp(reg, timeLayout)
}

func op_add_match(ctx *matchCtx) {
	i := len(ctx.result.all)
	if i == 0 {
		return
	}

	strs := protocol.ReadStrings(ctx.r)
	protocol.WriteOK(ctx.w)

	ctx.result.all[i].filters = append(ctx.result.all[i].filters, filterContains(strs))
}

func op_add_not_match(ctx *matchCtx) {
	i := len(ctx.result.all)
	if i == 0 {
		return
	}

	strs := protocol.ReadStrings(ctx.r)
	protocol.WriteOK(ctx.w)

	ctx.result.all[i].filters = append(ctx.result.all[i].filters, filterNot(filterContains(strs)))
}

func op_add_regexp(ctx *matchCtx) {
	i := len(ctx.result.all)
	if i == 0 {
		return
	}

	expr := protocol.ReadString(ctx.r)
	reg, err := perlRegexp(expr)
	if err != nil {
		protocol.WriteError(ctx.w, err)
		return
	}

	protocol.WriteOK(ctx.w)
	ctx.result.all[i].filters = append(ctx.result.all[i].filters, filterRegexp(reg))
}

func op_add_not_regexp(ctx *matchCtx) {
	i := len(ctx.result.all)
	if i == 0 {
		return
	}

	expr := protocol.ReadString(ctx.r)
	reg, err := perlRegexp(expr)
	if err != nil {
		protocol.WriteError(ctx.w, err)
		return
	}

	protocol.WriteOK(ctx.w)
	ctx.result.all[i].filters = append(ctx.result.all[i].filters, filterNot(filterRegexp(reg)))
}

func op_start(ctx *matchCtx) {

	startMatch(&ctx.result, ctx.w)
}
