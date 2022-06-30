package protocol

import (
	"encoding/binary"
	"io"
)

func ReadString(r io.Reader) string {
	buf := ReadBytes(r)
	return string(buf)
}

func ReadBytes(r io.Reader) []byte {
	var len uint16

	err := binary.Read(r, binary.BigEndian, &len)
	if err != nil {
		panic("readString " + err.Error())
	}

	if len == 0 {
		return nil
	}

	buf := make([]byte, len)

	_, err = io.ReadFull(r, buf)
	if err != nil {
		panic("readString " + err.Error())
	}

	return buf
}
