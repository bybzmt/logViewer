package main

import (
	"io"
	"os"
	"sort"

	"bufio"
)

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

func (m *matcher) reset() {
	m.r.Reset(m.file)
	m.line = nil
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

func (m *matcher) timeMinMax() (min int64, max int64, err error) {
	var maxl, minl *line

	maxl, err = m.maxLine()
	if err != nil {
		return
	}

	m.file.Seek(m.seek, io.SeekStart)
	m.reset()

	minl, err = m.readLine()
	if err != nil {
		return
	}

	m.file.Seek(m.seek, io.SeekStart)
	m.reset()

	return minl.time, maxl.time, nil
}

func (m *matcher) maxLine() (*line, error) {
	m.file.Seek(int64(-10*m.r.Size()), io.SeekEnd)
	m.reset()

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
	min, max, err := m.timeMinMax()
	if err != nil {
		return err
	}

	if m.startTime > max {
		return io.EOF
	}

	if m.endTime < min {
		return io.EOF
	}

	if m.startTime <= min && min < m.endTime {
		return nil
	}

	seek_max, err := m.file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	seek, err := m.findSeek(m.seek, seek_max)
	if err != nil {
		return err
	}

	m.file.Seek(seek, io.SeekStart)
	m.reset()

	return nil
}

func (m *matcher) findSeek(min, max int64) (int64, error) {
	diff := (max - min) / 2

	if diff < int64(m.r.Size()*10) {
		return min, nil
	}

	seek := min + diff

	m.file.Seek(seek, io.SeekStart)
	m.reset()

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

type matchResult struct {
	line       []*line
	matcher    []*matcher
	all        []*matcher
	lineCount  int64
	matchCount int64
	limit      int64
	bufSize    int
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
	for _, m := range rs.all {
		m.r = bufio.NewReaderSize(m.file, rs.bufSize)

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

		rs.line = append(rs.line, l)
		rs.matcher = append(rs.matcher, m)
	}

	return nil
}

func (rs *matchResult) close() {
	for _, m := range rs.all {
		m.file.Close()
	}
}
