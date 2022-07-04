package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/integrii/flaggy"
)

var limitDirs []string
var timeRegex string
var timeLayout string
var files []string
var matchs []string
var starttime string
var stoptime string

func main() {

	var addr string

	flaggy.String(&addr, "", "addr", "listen addr:port")
	flaggy.StringSlice(&limitDirs, "", "dir", "limit dir")

	flaggy.StringSlice(&files, "", "file", "cli log file")
	flaggy.String(&timeRegex, "", "timeRegex", "cli time regex")
	flaggy.String(&timeLayout, "", "timeLayout", "cli time layout")
	flaggy.String(&starttime, "", "start", "cli start")
	flaggy.String(&stoptime, "", "stop", "cli stop")
	flaggy.StringSlice(&matchs, "", "match", "cli match keyword")
	flaggy.Parse()

	if len(files) > 0 {
		m_cli()
		return
	}

	m_tcp(addr)
}

func m_cli() {
	var rs matchResult

	if timeLayout == "" {
		log.Fatalln("time Layout can not empty")
	}
	if timeRegex == "" {
		log.Fatalln("time Regex can not empty")
	}
	if starttime == "" {
		log.Fatalln("start time can not empty")
	}
	if stoptime == "" {
		log.Fatalln("stop time can not empty")
	}

	s1, err := time.Parse("2006-01-02T15:04:05", starttime)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("start", s1)

	s2, err := time.Parse("2006-01-02T15:04:05", stoptime)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("stop", s2)

	reg, err := perlRegexp(timeRegex)
	if err != nil {
		log.Fatalln(err)
	}

	defer rs.close()

	rs.bufSize = 1024 * 16

	for _, name := range files {
		f, err := os.Open(name)
		if err != nil {
			log.Fatalln(err)
		} else {
			m := &matcher{
				file:       f,
				timeParser: timeParserRegexp(reg, timeLayout),
				startTime:  s1.Unix(),
				endTime:    s2.Unix(),
			}

			if len(matchs) > 0 {
				m.filters = append(m.filters, filterContains(matchs))
			}

			rs.all = append(rs.all, m)
		}
	}

	err = rs.init()
	if err != nil {
		log.Fatalln(err)
	}

	for {
		l, err := rs.match()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(string(l.data))
	}
}

func m_tcp(addr string) {
	if len(limitDirs) < 1 {
		limitDirs = append(limitDirs, "/")
	}

	if addr == "" {
		addr = "127.0.0.2:7000"
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Listen", addr)

	s := matchServer{}
	s.init()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go s.serviceWapper(conn)
	}

}
