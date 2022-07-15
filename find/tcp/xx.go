package tcp

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

type OP uint8

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
	OP_NEXT
	//状态报告
	OP_STAT
)

func read(r io.Reader) (OP, []byte) {
	var data uint32

	err := binary.Read(r, binary.BigEndian, &data)
	if err != nil {
		panic(ErrorIO(err))
	}

	op := OP(data >> 24)
	len := data & 0x0fff

	if len == 0 {
		return op, nil
	}

	var buf bytes.Buffer

	_, err = io.CopyN(&buf, r, int64(len))
	if err != nil {
		panic(ErrorIO(err))
	}

	if op == OP_MSG {
		zr, err := gzip.NewReader(&buf)
		if err != nil {
			panic(err)
		}

		var buf2 bytes.Buffer

		if _, err := io.Copy(&buf2, zr); err != nil {
			panic(err)
		}

		if err := zr.Close(); err != nil {
			panic(err)
		}

		return op, buf2.Bytes()
	}

	return op, buf.Bytes()
}

func write(w io.Writer, op OP, data []byte) {
	l := uint32(len(data))

	if l > 0 {
		if op == OP_MSG {
			var buf bytes.Buffer
			zw := gzip.NewWriter(&buf)

			if _, err := zw.Write(data); err != nil {
				panic(err)
			}

			if err := zw.Close(); err != nil {
				panic(err)
			}

			l = uint32(buf.Len())
			data = buf.Bytes()
		}
	}

	code := (uint32(op) << 24) | l

	err := binary.Write(w, binary.BigEndian, &code)
	if err != nil {
		panic(ErrorIO(err))
	}

	if l == 0 {
		return
	}

	_, err = w.Write(data)
	if err != nil {
		panic(ErrorIO(err))
	}
}

func writeJson(w io.Writer, op OP, v interface{}, err error) {
	if err != nil {
		write(w, OP_ERR, []byte(err.Error()))
		return
	}

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		panic(err)
	}

	write(w, OP_GLOB, buf.Bytes())
}

func toJson(buf []byte, v interface{}) {
	if err := json.NewDecoder(bytes.NewBuffer(buf)).Decode(v); err != nil {
		panic(err)
	}
}

func readJson(r io.Reader, op OP, v interface{}) error {
	op2, buf := read(r)

	if op2 == OP_ERR {
		return errors.New(string(buf))
	}

	if op == op2 {
		toJson(buf, v)
	}

	panic(UnexpectedOP)
}
