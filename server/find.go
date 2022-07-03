package main

import (
	"io"
	"os"
	"sort"

	"bufio"
)

const MAX_LINE_BUF = 1024 * 128

type Filter func([]byte) bool
type TimeParser func([]byte) (int64, bool)

type matcher struct {
	file       *os.File
	r          *bufio.Reader
	timeParser TimeParser
	filters    []Filter
	startTime  int64
	endTime    int64
	tailf      bool
	seek       int64
	lineCount  int64
	matchCount int64
	line       *line
}

type line struct {
	time int64
	data []byte
}

func (m *matcher) readLine() (*line, error) {
	for {
		data, prefix, err := m.r.ReadLine()
		if err != nil {
			if err == io.EOF {
				if m.line != nil {
					old := m.line
					m.line = nil
					return old, nil
				}
			}
			return nil, err
		}

		if prefix {
			continue
		}

		t, ok := m.timeParser(data)
		if ok {
			old := m.line

			m.line = &line{
				time: t,
				data: data,
			}

			if old != nil {
				return old, nil
			}
		} else {
			if m.line != nil {
				m.line.data = append(m.line.data, data...)
			}
		}
	}
}

func (m *matcher) timeArea() (min int64, max int64, err error) {
	var maxl, minl *line

	maxl, err = m.maxLine()
	if err != nil {
		return
	}

	m.file.Seek(m.seek, os.SEEK_SET)
	m.r.Reset(m.file)

	minl, err = m.readLine()
	if err != nil {
		return
	}

	m.file.Seek(m.seek, os.SEEK_SET)
	m.r.Reset(m.file)

	return minl.time, maxl.time, nil
}

func (m *matcher) maxLine() (*line, error) {
	m.file.Seek(-10*MAX_LINE_BUF, os.SEEK_END)
	m.r.Reset(m.file)

	var endline *line

	for {
		l, err := m.readLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		endline = l
	}

	if endline == nil {
		return nil, io.EOF
	}

	return endline, nil
}

func (m *matcher) start() error {
	min, max, err := m.timeArea()
	if err != nil {
		return err
	}

	if m.startTime > max {
		return io.EOF
	}

	if m.endTime < min {
		return io.EOF
	}

	if m.startTime >= min && min < m.endTime {
		return nil
	}

	seek_min := m.seek

	st, err := m.file.Stat()
	if err != nil {
		return err
	}
	seek_max := st.Size()

	seek, err := m.findSeek(seek_min, seek_max)
	if err != nil {
		return err
	}

	m.file.Seek(seek, os.SEEK_SET)
	m.r.Reset(m.file)

	return nil
}

func (m *matcher) findSeek(min, max int64) (int64, error) {
	diff := (max - min) / 2

	if diff < MAX_LINE_BUF*10 {
		return min, nil
	}

	seek := min + diff

	m.file.Seek(seek, os.SEEK_SET)
	m.r.Reset(m.file)

	l, err := m.readLine()
	if err != nil {
		return 0, io.EOF
	}

	if l.time < m.startTime {
		return m.findSeek(seek, max)
	}

	return m.findSeek(min, seek)
}

func (m *matcher) matchByTime() (*line, error) {
	for {
		l, err := m.readLine()
		if err != nil {
			return nil, err
		}

		if l.time > m.endTime {
			return nil, io.EOF
		}

		if l.time < m.startTime {
			continue
		}

		return l, nil
	}
}
func (m *matcher) close() error {
	return m.file.Close()
}

type matchResult struct {
	line       []*line
	matcher    []*matcher
	all        []*matcher
	lineCount  int64
	matchCount int64
	limit      int64
	bufSize    int64
}

func (m *matchResult) Len() int {
	return len(m.matcher)
}
func (m *matchResult) Less(i, j int) bool {
	return m.line[i].time < m.line[j].time
}
func (m *matchResult) Swap(i, j int) {
	m.line[i], m.line[j] = m.line[j], m.line[i]
	m.matcher[i], m.matcher[j] = m.matcher[j], m.matcher[i]
}

func (rs *matchResult) matchByTime() (*line, *matcher, error) {
	old := rs.line[0]
	m := rs.matcher[0]

	l, err := rs.matcher[0].matchByTime()
	if err != nil {
		rs.matcher[0].close()

		rs.line = rs.line[1:]
		rs.matcher = rs.matcher[1:]

		if err == io.EOF {
			if old != nil {
				return old, m, nil
			}
		}

		return nil, nil, err
	}

	rs.line[0] = l
	sort.Stable(rs)

	return old, m, nil
}

func (rs *matchResult) match() (*line, error) {
	for len(rs.matcher) > 0 {
		l, m, err := rs.matchByTime()
		if err != nil {
			return nil, err
		}

		m.lineCount++
		rs.lineCount++

		ok := true
		for _, filter := range m.filters {
			if filter(l.data) == false {
				ok = false
				break
			}
		}

		if ok {
			m.matchCount++
			rs.matchCount++

			return l, nil
		}
	}

	return nil, io.EOF
}

func (rs *matchResult) init() error {
	for i, m := range rs.all {
		err := m.start()
		if err != nil {
			if err == io.EOF {
				continue
			}
			return err
		}

		l, err := m.matchByTime()
		if err != nil {
			if err == io.EOF {
				continue
			}
			return err
		}

		rs.line[i] = l
		rs.matcher[i] = m
	}

	return nil
}

func startMatch(ms *matchResult, w io.Writer) {
	var rs = matchResult{}

	err := rs.init()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	for {
		l, err := rs.match()
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(l.data)

		if rs.limit > 0 {
			if rs.matchCount >= rs.limit {
				return
			}
		}
	}
}
