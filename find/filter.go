package find

import (
	"bytes"
	"regexp"
	"regexp/syntax"
	"time"
)

func PerlRegexp(reg string) (*regexp.Regexp, error) {
	r, e := syntax.Parse(reg, syntax.Perl)
	if e != nil {
		return nil, e
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
