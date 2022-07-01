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
	SEPARATOR_LINUX uint8 = iota
	SEPARATOR_WIN
	SEPARATOR_MAC
)

const (
	OP_EXIT OP = iota
	OP_PING
	MSG_PONG
	//列出文件列表
	OP_LIST
	RESP_LIST
	//列表响应文件状态
	MSG_STATE
	//行起始确定
	OP_LINE_MATCH
	OP_LINE_REGEX
	//打开文件
	OP_OPEN
	RESP_OPEN
	//启动操作
	OP_START
	//响应结束
	MSG_END
	//取消动作
	OP_CANCEL
	//时间段 start(int64) + end(int64)
	OP_TIME
	//seek uint64
	OP_SEEK
	//过滤字符串
	OP_MATCH
	OP_MATCH_OR
	OP_REGEXP
	//进度报告
	OP_PROGRESS
	MSG_PROGRESS
	//查找方向
	OP_REVERSE
	//找查数量
	OP_LIMIT
	//输出速度
	OP_SPEED
	//跟随文件变化
	OP_TAILF
	OP_TOFILE
	//安静模式
	OP_QUIET
	MSG_LINE
	//统计器
	OP_COUNT_MATCH
	OP_COUNT_REGEX
	OP_COUNT
	MSG_COUNT
)

func RespOpenFile(w io.Writer, err error) {
	WriteOP(w, RESP_OPEN)
	WriteError(w, err)
}

func ReadRespOpenFile(r io.Reader) error {
	ExpectedOP(r, RESP_OPEN)
	return ReadError(r)
}

func RespListDir(w io.Writer, files []string, err error) {
	WriteOP(w, RESP_LIST)

	WriteError(w, err)

	if err == nil {
		WriteStrings(w, files)
	}
}

func ReadRespListDir(r io.Reader) ([]string, error) {
	ExpectedOP(r, RESP_OPEN)

	err := ReadError(r)
	if err != nil {
		return nil, err
	}

	strs := ReadStrings(r)
	return strs, nil
}
