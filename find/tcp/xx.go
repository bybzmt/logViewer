package tcp

import (
	"encoding/json"
	"errors"
	"io"
)

type OP uint16

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

const (
	OP_EXIT OP = iota
	OP_OK
	OP_MSG
	OP_ERR
	//列出文件列表
	OP_GLOB
	//查找文件
	OP_GREP
	//启动操作
	OP_START
	//状态报告
	OP_STAT
)

func respListDir(w io.Writer, files []string, err error) error {
	if err != nil {
		err = writeOP(w, OP_ERR)
		if err != nil {
			return err
		}
		return nil
	}

	err = writeOP(w, OP_OK)
	if err != nil {
		return err
	}

	return writeStrings(w, files)
}

func readRespListDir(r io.Reader) ([]string, error) {
	op, err := readOP(r)
	if err != nil {
		return nil, err
	}

	if op == OP_ERR {
		str, err := readString(r)
		if err != nil {
			return nil, err
		}

		return nil, errors.New(str)
	}

	if op == OP_OK {
		return readStrings(r)
	}

	return nil, UnexpectedOP
}

func readGrep(r io.Reader) (*Match, error) {
	var m Match
	de := json.NewDecoder(r)
	err := de.Decode(&m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func writeGrep(w io.Writer, m *Match) error {
	err := writeOP(w, OP_GREP)
	if err != nil {
		return err
	}

	en := json.NewEncoder(w)
	return en.Encode(m)
}
