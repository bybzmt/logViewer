package main

import (
	"log"

	"logViewer/protocol"
)

func op_ping(ctx *matchCtx) {
	protocol.WriteOP(ctx.w, protocol.OP_PONG)
}

func op_exit(ctx *matchCtx) {
	ctx.run = false
}

func op_list(ctx *matchCtx) {
	file := protocol.ReadString(ctx.r)
	files, err := listDirFiles(file, ctx.dirs)
	if err != nil {
		ctx.err <- err
		return
	}

	protocol.RespListDir(ctx.w, files)
}

func op_open(ctx *matchCtx) {
	file := protocol.ReadString(ctx.r)
	f, err := openFile(file, ctx.dirs)
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
