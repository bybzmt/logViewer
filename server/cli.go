package main

import (
	"fmt"
	"io/fs"
	"log"
	"logViewer/find"
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
	limit      int
}

func (cli *cliServer) run() {
	var rs find.Matcher

	if cli.timeLayout == "" {
		log.Fatalln("time Layout can not empty")
	}
	if cli.timeRegex == "" {
		log.Fatalln("time Regex can not empty")
	}

	if cli.start == "" {
		cli.start = time.Unix(0, 0).Format("2006-01-02T15:04:05Z07:00")
	}
	if cli.stop == "" {
		cli.stop = time.Now().Format("2006-01-02T15:04:05Z07:00")
	}

	s1, err := time.Parse("2006-01-02T15:04:05Z07:00", cli.start)
	if err != nil {
		log.Fatalln(err)
	}

	s2, err := time.Parse("2006-01-02T15:04:05Z07:00", cli.stop)
	if err != nil {
		log.Fatalln(err)
	}

	reg, err := find.PerlRegexp(cli.timeRegex)
	if err != nil {
		log.Fatalln(err)
	}

	defer rs.Close()

	rs.BufSize = 1024 * 16

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

	rs.StartTime = s1.Unix()
	rs.EndTime = s2.Unix()

	for _, name := range cli.files {
		f, err := os.Open(name)
		if err != nil {
			log.Fatalln(err)
		} else {
			m := &find.File{
				File:       f,
				TimeParser: find.TimeParserRegexp(reg, cli.timeLayout),
			}

			if len(cli.matchs) > 0 {
				m.Filters = append(m.Filters, find.FilterContains(cli.matchs))
			}

			rs.All = append(rs.All, m)
		}
	}

	err = rs.Init()
	if err != nil {
		log.Fatalln(err)
	}

	for {
		data, err := rs.Match()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(string(data))

		if cli.limit > 0 {
			cli.limit--
			if cli.limit == 0 {
				break
			}
		}
	}
}
