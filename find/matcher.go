package find

import (
	"io"
)

const DEFAULT_LINE_BUF = 1024 * 128

type Matcher struct {
	results    results
	All        []File
	StartTime  int64
	EndTime    int64
	MatchCount int64
	Limit      uint16
	BufSize    uint32
}

func (rs *Matcher) Match() ([]byte, error) {
	if rs.Limit > 0 {
		if rs.MatchCount >= int64(rs.Limit) {
			return nil, io.EOF
		}
	}

	for len(rs.results) > 0 {
		l, m, err := rs.results.matchByTime()
		if err != nil {
			return nil, err
		}

		ok := true
		for _, filter := range m.Filters {
			if !filter(l.data) {
				ok = false
				break
			}
		}

		if ok {
			m.MatchCount++
			rs.MatchCount++

			return l.data, nil
		}
	}

	return nil, io.EOF
}

func (rs *Matcher) Init() error {
	if rs.BufSize == 0 {
		rs.BufSize = DEFAULT_LINE_BUF
	}

	for _, m := range rs.All {
		m.startTime = rs.StartTime
		m.endTime = rs.EndTime
		m.bufSize = rs.BufSize

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

		rs.results = append(rs.results, result{
			line: l,
			file: &m,
		})
	}

	return nil
}

func (rs *Matcher) Close() {
	for _, m := range rs.All {
		m.f.Close()
	}
}
