package client

import (
	"bufio"
	"logViewer/protocol"
	"net"
)

type matchFile struct {
	id           uint16
	name         string
	startTime    int64
	endTime      int64
	timeRegex    string
	timeLayout   string
	contains     [][]string
	regex        []string
	contains_not [][]string
	regex_not    []string
	lineCount    int64
	matchCount   int64
	seek         int64
	size         int64
}

type matchResult struct {
	addr    string
	files   []matchFile
	limit   int64
	bufSize int
	msg     chan []byte
	c       net.Conn
	rw      *bufio.ReadWriter
}

func (rs *matchResult) dial() error {
	c, err := net.Dial("tcp", rs.addr)
	if err != nil {
		return err
	}

	rs.c = c
	w := bufio.NewWriter(c)
	r := bufio.NewReader(c)
	rs.rw = bufio.NewReadWriter(r, w)
	return nil
}

func (rs *matchResult) list(dir string) ([]string, error) {
	protocol.WriteOP(rs.rw, protocol.OP_LIST)
	protocol.WriteString(rs.rw, dir)

	err := rs.rw.Flush()
	if err != nil {
		return nil, err
	}

	return protocol.ReadRespListDir(rs.rw)
}

func (rs *matchResult) open() error {

	for _, f := range rs.files {
		protocol.WriteOP(rs.rw, protocol.OP_OPEN)
		protocol.WriteString(rs.rw, f.name)
		protocol.WriteOP(rs.rw, protocol.SET_STARTTIME)
		protocol.WriteInt64(rs.rw, f.startTime)
		protocol.WriteOP(rs.rw, protocol.SET_STOPTIME)
		protocol.WriteInt64(rs.rw, f.endTime)
		protocol.WriteOP(rs.rw, protocol.SET_TIME_PARSER)
		protocol.WriteString(rs.rw, f.timeRegex)
		protocol.WriteString(rs.rw, f.timeLayout)

		for _, strs := range f.contains {
			protocol.WriteOP(rs.rw, protocol.ADD_MATCH)
			protocol.WriteStrings(rs.rw, strs)
		}
		for _, str := range f.regex {
			protocol.WriteOP(rs.rw, protocol.ADD_REGEXP)
			protocol.WriteString(rs.rw, str)
		}
		for _, strs := range f.contains_not {
			protocol.WriteOP(rs.rw, protocol.ADD_NOT_MATCH)
			protocol.WriteStrings(rs.rw, strs)
		}
		for _, str := range f.regex_not {
			protocol.WriteOP(rs.rw, protocol.ADD_NOT_REGEXP)
			protocol.WriteString(rs.rw, str)
		}
		err := rs.rw.Flush()
		if err != nil {
			return err
		}

		fid, err := protocol.ReadRespOpenFile(rs.rw)
		if err != nil {
			return err
		}
		f.id = fid
	}

	protocol.WriteOP(rs.rw, protocol.SET_LINE_BUF)
	protocol.WriteInt64(rs.rw, int64(rs.bufSize))
	protocol.WriteOP(rs.rw, protocol.SET_LIMIT)
	protocol.WriteInt64(rs.rw, int64(rs.limit))

	err := rs.rw.Flush()
	if err != nil {
		return err
	}

	return nil
}
