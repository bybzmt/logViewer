package tcp

import (
	"encoding/binary"
	"errors"
	"io"
)

type ErrorIO error
type ErrorProtocol error
type ErrorAccessDenied error

var UnexpectedOP = ErrorProtocol(errors.New("unexpected op"))
var AccessDenied = ErrorAccessDenied(errors.New("access denied"))
var NotOpenFile = ErrorAccessDenied(errors.New("not open file"))

func readUint16(r io.Reader) (uint16, error) {
	var num uint16

	err := binary.Read(r, binary.BigEndian, &num)
	if err != nil {
		return 0, err
	}

	return num, nil
}

func writeUint16(w io.Writer, num uint16) error {
	return binary.Write(w, binary.BigEndian, num)
}

func readOP(r io.Reader) (OP, error) {
	op, err := readUint16(r)
	return OP(op), err
}

func writeOP(w io.Writer, op OP) error {
	return writeUint16(w, uint16(op))
}

func readString(r io.Reader) (string, error) {
	buf, err := readBytes(r)
	return string(buf), err
}

func writeString(w io.Writer, data string) error {
	return writeBytes(w, []byte(data))
}

func readBytes(r io.Reader) ([]byte, error) {
	var len uint32

	err := binary.Read(r, binary.BigEndian, &len)
	if err != nil {
		return nil, err
	}

	if len == 0 {
		return nil, nil
	}

	buf := make([]byte, len)

	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func writeBytes(w io.Writer, buf []byte) error {
	var l uint32 = uint32(len(buf))

	err := binary.Write(w, binary.BigEndian, l)
	if err != nil {
		return err
	}

	if l == 0 {
		return nil
	}

	_, err = w.Write(buf)
	return err
}

func readErr(r io.Reader) error {
	str, err := readString(r)
	if err != nil {
		return err
	}
	return errors.New(str)
}

func writeErr(w io.Writer, err error) error {
	e := writeOP(w, OP_ERR)
	if e != nil {
		return e
	}
	return writeString(w, err.Error())
}
