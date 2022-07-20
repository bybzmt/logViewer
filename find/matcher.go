package find

import (
	"io"

	"github.com/fsnotify/fsnotify"
)

const DEFAULT_LINE_BUF = 1024 * 128

type Matcher struct {
	results    results
	All        []File
	StartTime  int64
	EndTime    int64
	MatchCount int64
	BufSize    uint32
	Tailf      bool
	watcher    *fsnotify.Watcher
}

func (rs *Matcher) Match() ([]byte, error) {
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

	var err error
	if rs.Tailf {
		rs.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			return err
		}
	}

	for _, m := range rs.All {
		m.startTime = rs.StartTime
		m.endTime = rs.EndTime
		m.bufSize = rs.BufSize
		m.tailf = rs.Tailf

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

		if rs.Tailf {
			rs.watcher.Add(m.Name)
		}

		rs.results = append(rs.results, result{
			line: l,
			file: &m,
		})
	}

	return nil
}

type Stat struct {
	Seek int64
	All  int64
}

func (rs *Matcher) Stat() (s Stat, err error) {
	for _, m := range rs.All {
		var old int64 = 0
		var end int64 = 0

		old, err = m.f.Seek(0, io.SeekCurrent)
		if err != nil {
			return
		}

		end, err = m.f.Seek(0, io.SeekEnd)
		if err != nil {
			return
		}

		_, err = m.f.Seek(old, io.SeekStart)
		if err != nil {
			return
		}

		s.Seek += old
		s.All += end
	}

	return
}

func (rs *Matcher) Close() {
	for _, m := range rs.All {
		m.f.Close()
	}
}

func NewMatcher(m *MatchParam) (*Matcher, error) {
	f := Matcher{
		StartTime: m.StartTime,
		EndTime:   m.EndTime,
		BufSize:   m.BufSize,
	}

	for _, fi := range m.Files {
		reg, err := PerlRegexp(fi.TimeRegex)
		if err != nil {
			return nil, err
		}

		var fs []Filter
		for _, keys := range fi.Contains {
			fs = append(fs, FilterContains(keys))
		}

		f.All = append(f.All, File{
			Name:       fi.Name,
			Filters:    fs,
			TimeParser: TimeParserRegexp(reg, fi.TimeLayout),
		})
	}

	err := f.Init()
	if err != nil {
		return nil, err
	}

	return &f, nil
}
