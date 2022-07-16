package tcp

import (
	"errors"
	"fmt"
	"time"
)

type ErrorIO error
type ErrorProtocol error
type ErrorAccessDenied error

func unexpectedOP(op OP) ErrorProtocol {
	return ErrorProtocol(fmt.Errorf("unexpected op:%d", op))
}

var AccessDenied = ErrorAccessDenied(errors.New("access denied"))
var NotOpenFile = ErrorAccessDenied(errors.New("not open file"))

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
