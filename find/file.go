package find

import (
	"io"
	"os"
	"sort"

	"bufio"
)

type Filter func([]byte) bool
type TimeParser func([]byte) (int64, bool)

type File struct {
	Name       string
	f          *os.File
	r          *bufio.Reader
	TimeParser TimeParser
	Filters    []Filter
	startTime  int64
	endTime    int64
	tailf      bool
	LineCount  int64
	MatchCount int64
	missLine   int
	bufSize    uint32
	line       *line
}

type line struct {
	time int64
	data []byte
}

func (m *File) reset() {
	m.r.Reset(m.f)
	m.line = nil
}

func (m *File) readLine() (*line, error) {
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

		var t int64
		var ok bool = true

		if m.TimeParser != nil {
			t, ok = m.TimeParser(data)
		}

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
			m.missLine++

			if m.line != nil {
				m.line.data = append(m.line.data, data...)
			}
		}
	}
}

func (m *File) timeMinMax() (min int64, max int64, err error) {
	var maxl, minl *line

	maxl, err = m.maxLine()
	if err != nil {
		return
	}

	m.f.Seek(0, io.SeekStart)
	m.reset()

	minl, err = m.readLine()
	if err != nil {
		return
	}

	m.f.Seek(0, io.SeekStart)
	m.reset()

	return minl.time, maxl.time, nil
}

func (m *File) maxLine() (*line, error) {
	m.f.Seek(int64(-10*m.r.Size()), io.SeekEnd)
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

func (m *File) start() error {
	var err error
	m.f, err = os.Open(m.Name)
	if err != nil {
		return err
	}

	m.r = bufio.NewReaderSize(m.f, int(m.bufSize))

	if m.tailf {
		_, err := m.f.Seek(0, io.SeekEnd)
		return err
	}

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

	seek_max, err := m.f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	seek, err := m.findSeek(0, seek_max)
	if err != nil {
		return err
	}

	m.reset()
	_, err = m.f.Seek(seek, io.SeekStart)
	return err
}

func (m *File) findSeek(min, max int64) (int64, error) {
	diff := (max - min) / 2

	if diff < int64(m.r.Size()*10) {
		return min, nil
	}

	seek := min + diff

	m.reset()
	_, err := m.f.Seek(seek, io.SeekStart)
	if err != nil {
		return 0, err
	}

	l, err := m.readLine()
	if err != nil {
		return 0, io.EOF
	}

	if l.time < m.startTime {
		return m.findSeek(seek, max)
	}

	return m.findSeek(min, seek)
}

func (m *File) matchByTime() (*line, error) {
	for {
		l, err := m.readLine()
		if err != nil {
			return nil, err
		}

		m.LineCount++

		if !m.tailf {
			if l.time > m.endTime {
				return nil, io.EOF
			}

			if l.time < m.startTime {
				continue
			}
		}

		return l, nil
	}
}

type result struct {
	line *line
	file *File
}

type results []result

func (rs results) Len() int {
	return len(rs)
}

func (rs results) Less(i, j int) bool {
	return rs[i].line.time < rs[j].line.time
}

func (rs results) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

func (rs *results) matchByTime() (*line, *File, error) {
	old := (*rs)[0].line
	m := (*rs)[0].file

	l, err := (*rs)[0].file.matchByTime()
	if err != nil {
		*rs = (*rs)[1:]

		if err == io.EOF {
			if old != nil {
				return old, m, nil
			}
		}

		return nil, nil, err
	}

	(*rs)[0].line = l

	sort.Stable(*rs)

	return old, m, nil
}
