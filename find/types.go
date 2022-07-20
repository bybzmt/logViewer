package find

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

type FileParam struct {
	Name        string
	TimeRegex   string
	TimeLayout  string
	Contains    [][]string
	Regex       []string
	ContainsNot [][]string
	RegexNot    []string
}

type MatchParam struct {
	Files     []FileParam
	StartTime int64
	EndTime   int64
	BufSize   uint32
}

type Conn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	SetDeadline(t time.Time) error
}

type Client interface {
	Glob(pattern string) (out []string, err error)
	Open(m *MatchParam) (err error)
	Read() (data []byte, err error)
	Ping() (err error)
	Close() (err error)
}
