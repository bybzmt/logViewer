package client

import (
	"log"
	"regexp"
	"regexp/syntax"
)

var crossRegexp *regexp.Regexp

func init() {
	reg, err := syntax.Parse("^(http|https)://(127.0.0.1|localhost)(:\\d+)?", syntax.Perl)
	if err != nil {
		log.Panicln(err)
	}

	crossRegexp, err = regexp.Compile(reg.String())
	if err != nil {
		log.Panicln(err)
	}
}
