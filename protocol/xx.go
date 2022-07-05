package protocol

import (
	"io"
)

type OP uint16

type VarChar struct {
	Len  uint16
	Data []uint8
}
type CharArray struct {
	Len  uint16
	Data []VarChar
}

const (
	OP_EXIT OP = iota
	OP_RESET
	OP_PING
	OP_PONG
	OP_OK
	OP_MSG
	OP_ERR
	//列出文件列表
	OP_LIST
	RESP_LIST
	//打开文件
	OP_OPEN
	//启动操作
	OP_START
	//响应结束
	OP_STOP
	//状态报告
	OP_STAT
	RESP_STAT
	//时间段
	SET_STARTTIME
	SET_STOPTIME
	//seek uint64
	SET_SEEK
	//最大行大小
	SET_LINE_BUF
	//找查数量
	SET_LIMIT
	//行时间解析器
	SET_TIME_PARSER
	//过滤字符串
	ADD_MATCH
	ADD_NOT_MATCH
	ADD_REGEXP
	ADD_NOT_REGEXP
	//跟随文件变化
	SET_TAILF
	//输出速度
	SET_SPEED
	//安静模式
	SET_QUIET
)

func RespOpenFile(w io.Writer, err error) {
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteOP(w, OP_OK)
}

func ReadRespOpenFile(r io.Reader) (file_id uint16, err error) {
	op := ExpectedOP(r, OP_ERR, OP_OK)
	if op == OP_ERR {
		return 0, ReadError(r)
	}

	file_id = ReadUint16(r)
	return
}

func RespListDir(w io.Writer, files []string) {
	WriteOP(w, RESP_LIST)
	WriteStrings(w, files)
}

func ReadRespListDir(r io.Reader) ([]string, error) {
	op := ExpectedOP(r, RESP_LIST, OP_ERR)

	if op == OP_ERR {
		err := ReadError(r)
		return nil, err
	}

	strs := ReadStrings(r)
	return strs, nil
}
