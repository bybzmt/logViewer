package find

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type OP uint8

const mask uint32 = 0x00ffffff

const (
	OP_EXIT OP = iota + 1
	OP_OK
	OP_EOF
	OP_PING
	OP_MSG
	OP_ERR
	//列出文件列表
	OP_GLOB
	//查找文件
	OP_GREP
	OP_READ
	//状态报告
	OP_STAT
	OP_GZIP = 1 << 7
)

func read(r io.Reader) (OP, []byte) {
	var data uint32

	err := binary.Read(r, binary.BigEndian, &data)
	if err != nil {
		panic(ErrorIO(fmt.Errorf("read err: %s", err)))
	}

	op := OP(data >> 24)
	len := data & mask

	if len == 0 {
		return op, nil
	}

	var buf bytes.Buffer

	_, err = io.CopyN(&buf, r, int64(len))
	if err != nil {
		panic(ErrorIO(fmt.Errorf("read err: %s", err)))
	}

	if op&OP_GZIP != 0 {
		zr, err := gzip.NewReader(&buf)
		if err != nil {
			panic(err)
		}

		var buf2 bytes.Buffer

		if _, e := io.Copy(&buf2, zr); e != nil {
			panic(e)
		}

		if e := zr.Close(); e != nil {
			panic(e)
		}

		op &^= OP_GZIP

		return op, buf2.Bytes()
	}

	return op, buf.Bytes()
}

func write(w io.Writer, op OP, data []byte) {
	l := uint32(len(data))

	if l >= mask {
		panic(writeDataBig)
	}

	if l > 50 {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)

		if _, e := zw.Write(data); e != nil {
			panic(e)
		}

		if e := zw.Close(); e != nil {
			panic(e)
		}

		l = uint32(buf.Len())
		data = buf.Bytes()

		op |= OP_GZIP
	}

	code := (uint32(op) << 24) | l

	err := binary.Write(w, binary.BigEndian, code)
	if err != nil {
		panic(ErrorIO(fmt.Errorf("write err: %s", err)))
	}

	if l == 0 {
		return
	}

	_, err = w.Write(data)
	if err != nil {
		panic(ErrorIO(fmt.Errorf("write err: %s", err)))
	}
}

func writeJson(w io.Writer, op OP, v interface{}, err error) {
	if err != nil {
		write(w, OP_ERR, []byte(err.Error()))
		return
	}

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		panic(ErrorProtocol(fmt.Errorf("json encoder %s", err)))
	}

	write(w, op, buf.Bytes())
}

func readJson(r io.Reader, op OP, v interface{}) error {
	op2, buf := read(r)

	if op2 == OP_ERR {
		return errors.New(string(buf))
	}

	if op == op2 {
		err := json.Unmarshal(buf, v)
		if err != nil {
			panic(ErrorProtocol(fmt.Errorf("json decoder %s", err)))
		}
		return nil
	}

	panic(unexpectedOP(op))
}
