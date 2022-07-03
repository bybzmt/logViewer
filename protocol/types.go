package protocol

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

func ReadUint16(r io.Reader) uint16 {
	var num uint16

	err := binary.Read(r, binary.BigEndian, &num)
	if err != nil {
		panic(ErrorIO(err))
	}

	return num
}

func WriteUint16(w io.Writer, num uint16) {
	err := binary.Write(w, binary.BigEndian, num)
	if err != nil {
		panic(ErrorIO(err))
	}
}

func ReadOP(r io.Reader) OP {
	op := ReadUint16(r)
	return OP(op)
}

func WriteOP(w io.Writer, op OP) {
	WriteUint16(w, uint16(op))
}

func WriteOK(w io.Writer) {
	WriteOP(w, OP_OK)
}

func ExpectedOP(r io.Reader, needs ...OP) OP {
	has := ReadOP(r)
	for _, need := range needs {
		if has == need {
			return has
		}
	}
	panic(UnexpectedOP)
}

func ReadString(r io.Reader) string {
	buf := ReadBytes(r)
	return string(buf)
}

func ReadError(r io.Reader) error {
	str := ReadString(r)
	return errors.New(str)
}

func WriteError(w io.Writer, err error) {
	WriteOP(w, OP_ERR)
	WriteString(w, err.Error())
}

func WriteString(w io.Writer, data string) {
	buf := []byte(data)
	WriteBytes(w, buf)
}

func ReadBytes(r io.Reader) []byte {
	var len uint16

	err := binary.Read(r, binary.BigEndian, &len)
	if err != nil {
		panic(ErrorIO(err))
	}

	if len == 0 {
		return nil
	}

	buf := make([]byte, len)

	_, err = io.ReadFull(r, buf)
	if err != nil {
		panic(ErrorIO(err))
	}

	return buf
}

func WriteBytes(w io.Writer, buf []byte) {
	var l uint16 = uint16(len(buf))

	err := binary.Write(w, binary.BigEndian, l)
	if err != nil {
		panic(ErrorIO(err))
	}

	if l == 0 {
		return
	}

	_, err = w.Write(buf)
	if err != nil {
		panic(ErrorIO(err))
	}
}

func ReadStrings(r io.Reader) []string {
	var len uint16

	err := binary.Read(r, binary.BigEndian, &len)
	if err != nil {
		panic(ErrorIO(err))
	}

	var out []string

	for len > 0 {
		out = append(out, ReadString(r))
		len--
	}

	return out
}

func WriteStrings(w io.Writer, strs []string) {
	var l uint16 = uint16(len(strs))

	err := binary.Write(w, binary.BigEndian, l)
	if err != nil {
		panic(ErrorIO(err))
	}

	for _, str := range strs {
		WriteString(w, str)
	}
}
