package main

import (
	"logViewer/find/tcp"

	"github.com/integrii/flaggy"
)

func main() {
	tcp := tcp.MatchServer{}
	cli := cliServer{}

	ft := flaggy.NewSubcommand("tcp")
	ft.String(&tcp.Addr, "", "addr", "listen addr:port")
	ft.StringSlice(&tcp.Dirs, "", "dir", "limit dir")

	fc := flaggy.NewSubcommand("cli")
	//fc.StringSlice(&cli.files, "", "file", "open file")
	fc.AddPositionalValue(&cli.glob, "file", 1, true, "open log file eg: xxx.log")
	fc.String(&cli.timeRegex, "", "timeRegex", "time regex")
	fc.String(&cli.timeLayout, "", "timeLayout", "time layout")
	fc.String(&cli.start, "", "start", "start time")
	fc.String(&cli.stop, "", "stop", "stop time")
	fc.StringSlice(&cli.matchs, "", "match", "match keyword")
	fc.Int(&cli.limit, "", "limit", "max match data num")

	flaggy.AttachSubcommand(ft, 1)
	flaggy.AttachSubcommand(fc, 1)
	flaggy.SetDescription("log viewer")
	flaggy.Parse()

	if fc.Used {
		cli.run()
		return
	}

	tcp.Run()
}
