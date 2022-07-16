package tcp

import (
	"errors"
	"fmt"
	"time"
)

type ErrorIO error
type ErrorProtocol error
type ErrorUser error

func unexpectedOP(op OP) ErrorProtocol {
	return ErrorProtocol(fmt.Errorf("unexpected op:%d", op))
}

var writeDataBig = ErrorProtocol(fmt.Errorf("writedata exceed %d", mask))
var notOpenFile = ErrorUser(errors.New("not open file"))
var repeatOpenFile = ErrorUser(errors.New("repeat open file"))

type File struct {
	Name        string
	TimeRegex   string
	TimeLayout  string
	Contains    [][]string
	Regex       []string
	ContainsNot [][]string
	RegexNot    []string
}

type Match struct {
	Files     []File
	StartTime int64
	EndTime   int64
	Limit     uint16
	BufSize   uint32
}

type Conn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	SetDeadline(t time.Time) error
}

func tryErr(err *error) {
	switch p := recover(); e := p.(type) {
	case ErrorIO:
		*err = e
	case ErrorProtocol:
		*err = e
	default:
		panic(p)
	}
}
