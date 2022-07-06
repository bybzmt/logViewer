package protocol

import (
	"log"
	"logViewer/find"
)

func op_ping(ctx *matchCtx) {
	WriteOP(ctx.w, OP_PONG)
}

func op_exit(ctx *matchCtx) {
	ctx.run = false
}

func op_list(ctx *matchCtx) {
	file := ReadString(ctx.r)
	files, err := listDirFiles(file, ctx.dirs)
	if err != nil {
		ctx.err <- err
		return
	}

	RespListDir(ctx.w, files)
}

func op_open(ctx *matchCtx) {
	file := ReadString(ctx.r)

	match := &find.File{Name: file}

	ctx.matcher.All = append(ctx.matcher.All, match)
}

func op_set_starttime(ctx *matchCtx) {
	num := ReadInt64(ctx.r)
	ctx.matcher.StartTime = num
}

func op_set_stoptime(ctx *matchCtx) {
	num := ReadInt64(ctx.r)
	ctx.matcher.EndTime = num
}

func op_set_limit(ctx *matchCtx) {
	num := ReadUint16(ctx.r)
	ctx.matcher.Limit = int64(num)
}

func op_set_buf(ctx *matchCtx) {
	num := ReadUint16(ctx.r)
	ctx.matcher.BufSize = num
}

func op_set_time_parser(ctx *matchCtx) {
	i := len(ctx.matcher.All) - 1

	expr := ReadString(ctx.r)
	timeLayout := ReadString(ctx.r)

	reg, err := find.PerlRegexp(expr)
	if err != nil {
		ctx.err <- err
		return
	}

	ctx.matcher.All[i].TimeParser = find.TimeParserRegexp(reg, timeLayout)
}

func op_add_match(ctx *matchCtx) {
	i := len(ctx.matcher.All) - 1

	strs := ReadStrings(ctx.r)

	ctx.matcher.All[i].Filters = append(ctx.matcher.All[i].Filters, find.FilterContains(strs))
}

func op_add_not_match(ctx *matchCtx) {
	i := len(ctx.matcher.All) - 1

	strs := ReadStrings(ctx.r)

	ctx.matcher.All[i].Filters = append(ctx.matcher.All[i].Filters, find.FilterNot(find.FilterContains(strs)))
}

func op_add_regexp(ctx *matchCtx) {
	i := len(ctx.matcher.All) - 1

	expr := ReadString(ctx.r)
	reg, err := find.PerlRegexp(expr)
	if err != nil {
		ctx.err <- err
		return
	}

	ctx.matcher.All[i].Filters = append(ctx.matcher.All[i].Filters, find.FilterRegexp(reg))
}

func op_add_not_regexp(ctx *matchCtx) {
	i := len(ctx.matcher.All) - 1

	expr := ReadString(ctx.r)
	reg, err := find.PerlRegexp(expr)
	if err != nil {
		ctx.err <- err
		return
	}

	ctx.matcher.All[i].Filters = append(ctx.matcher.All[i].Filters, find.FilterNot(find.FilterRegexp(reg)))
}

func op_start(ctx *matchCtx) {
	ctx.mode = 2

	err := ctx.matcher.Init()
	if err != nil {
		ctx.err <- err
		return
	}

	WriteOK(ctx.w)

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
			data, err := ctx.matcher.Match()
			if err != nil {
				ctx.err <- err
				return
			}
			ctx.msg <- data
		}
	}()
}

func op_stat(ctx *matchCtx) {
}
