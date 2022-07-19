package find

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"regexp/syntax"
	"strings"
	"time"
)

func PerlRegexp(reg string) (*regexp.Regexp, error) {
	r, e := syntax.Parse(reg, syntax.Perl)
	if e != nil {
		return nil, fmt.Errorf("PerlRegexp: %s %s", reg, e)
	}

	return regexp.Compile(r.String())
}

func TimeParserRegexp(reg *regexp.Regexp, timeLayout string) TimeParser {
	return func(b []byte) (int64, bool) {
		f := reg.Find(b)
		if f == nil {
			return 0, false
		}

		t, err := time.Parse(timeLayout, string(f))
		if err != nil {
			return 0, false
		}

		return t.Unix(), true
	}
}

func FilterRegexp(reg *regexp.Regexp) Filter {
	return func(b []byte) bool {
		return reg.Match(b)
	}
}

func FilterContains(strs []string) Filter {
	return func(b []byte) bool {
		for _, str := range strs {
			if bytes.Contains(b, []byte(str)) {
				return true
			}
		}

		return false
	}
}

func FilterNot(fn Filter) Filter {
	return func(b []byte) bool {
		return !fn(b)
	}
}

var month = "(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sept|Oct|Nov|Dec)"
var week1 = "(Mon|Tues|Wed|Thur|Fri|Sat|Sun)"
var week2 = "(Monday|Tuesday|Wednesday|Thursday|Friday|Saterday|Sunday)"

//08/Jul/2022:10:27:06 +0800
var TimeLayouts = []string{
	`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[\+-]\d{2}:\d{2}`, "2006-01-02T15:04:05Z07:00",
	`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}[\+-]\d{2}:\d{2}`, "2006-01-02 15:04:05Z07:00",
	`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`, "2006-01-02T15:04:05",
	`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`, "2006-01-02 15:04:05",
	`^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}`, "2006/01/02 15:04:05",
	`^\d{2}/` + month + `/\d{4l}:\d{2}:\d{2}:\d{2} [\+-]\d{4}`, "02/Jan/2006:15:04:05 Z0700",
}

func TimeLayoutMatch(txt []byte) (int, error) {
	m := len(TimeLayouts)

	for i := 0; i < m; i += 2 {
		reg, err := PerlRegexp(TimeLayouts[i])
		if err != nil {
			panic(err)
		}

		if reg.Match(txt) {
			return i, nil
		}
	}

	return 0, errors.New("not found time layout")
}

func ParseTime(t string) (tx time.Time, e error) {
	t1 := strings.TrimSpace(t)

	i, err := TimeLayoutMatch([]byte(t1))
	if err != nil {
		e = err
		return
	}

	reg, err := PerlRegexp(TimeLayouts[i])
	if err != nil {
		panic(err)
	}

	t2 := reg.Find([]byte(t1))
	if len(t2) == 0 {
		panic(errors.New("ParseTime"))
	}

	return time.ParseInLocation(TimeLayouts[i+1], string(t2), time.Local)
}
