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
	defer ctx.close()

	for ctx.run {
		op := protocol.ReadOP(ctx.r)

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

		h(ctx)

		if err := ctx.w.Flush(); err != nil {
			panic(protocol.ErrorIO(err))
		}
	}
}

func (ctx *matchCtx) close() {
	for _, m := range ctx.result.all {
		m.file.Close()
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
	ctx.mode = 1

	file := protocol.ReadString(ctx.r)
	f, err := openFile(file)
	if err != nil {
		protocol.WriteError(ctx.w, err)
	}

	match := &matcher{file: f}
	ctx.result.all = append(ctx.result.all, match)
	protocol.WriteOP(ctx.w, protocol.OP_OK)
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
	num := protocol.ReadInt64(ctx.r)

	i := len(ctx.result.all)
	if i == 0 {
		protocol.WriteError(ctx.w, protocol.NotOpenFile)
		return
	}

	ctx.result.all[i].seek = num
	protocol.WriteOK(ctx.w)
}

func op_set_time_parser(ctx *matchCtx) {
	i := len(ctx.result.all) - 1

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
	i := len(ctx.result.all) - 1

	strs := protocol.ReadStrings(ctx.r)
	protocol.WriteOK(ctx.w)

	ctx.result.all[i].filters = append(ctx.result.all[i].filters, filterContains(strs))
}

func op_add_not_match(ctx *matchCtx) {
	i := len(ctx.result.all) - 1

	strs := protocol.ReadStrings(ctx.r)
	protocol.WriteOK(ctx.w)

	ctx.result.all[i].filters = append(ctx.result.all[i].filters, filterNot(filterContains(strs)))
}

func op_add_regexp(ctx *matchCtx) {
	i := len(ctx.result.all) - 1

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
	i := len(ctx.result.all) - 1

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
	ctx.mode = 2

	startMatch(&ctx.result, ctx.w)
}

func op_stat(ctx *matchCtx) {
}
