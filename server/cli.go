package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"
)

type cliServer struct {
	glob       string
	timeRegex  string
	timeLayout string
	files      []string
	matchs     []string
	start      string
	stop       string
}

func (cli *cliServer) run() {
	var rs matchResult

	if cli.timeLayout == "" {
		log.Fatalln("time Layout can not empty")
	}
	if cli.timeRegex == "" {
		log.Fatalln("time Regex can not empty")
	}

	if cli.start == "" {
		cli.start = time.Unix(0, 0).Format("2006-01-02T15:04:05")
	}
	if cli.stop == "" {
		cli.stop = time.Now().Format("2006-01-02T15:04:05")
	}

	s1, err := time.Parse("2006-01-02T15:04:05", cli.start)
	if err != nil {
		log.Fatalln(err)
	}

	s2, err := time.Parse("2006-01-02T15:04:05", cli.stop)
	if err != nil {
		log.Fatalln(err)
	}

	reg, err := perlRegexp(cli.timeRegex)
	if err != nil {
		log.Fatalln(err)
	}

	defer rs.close()

	rs.bufSize = 1024 * 16

	cli.files, err = fs.Glob(os.DirFS("./"), cli.glob)
	if err != nil {
		log.Fatalln(err)
	}

	if len(cli.files) == 0 {
		log.Fatalln("log file not found")
	}

	log.Println("files:", cli.files)
	log.Println("timeRegex:", cli.timeRegex)
	log.Println("timeLayout:", cli.timeLayout)
	log.Println("start:", s1.Format(time.RFC3339))
	log.Println(" stop:", s2.Format(time.RFC3339))
	log.Println("keyword:", cli.matchs)

	for _, name := range cli.files {
		f, err := os.Open(name)
		if err != nil {
			log.Fatalln(err)
		} else {
			m := &matcher{
				file:       f,
				timeParser: timeParserRegexp(reg, cli.timeLayout),
				startTime:  s1.Unix(),
				endTime:    s2.Unix(),
			}

			if len(cli.matchs) > 0 {
				m.filters = append(m.filters, filterContains(cli.matchs))
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
