package main

import (
	"github.com/integrii/flaggy"
)

func main() {
	tcp := matchServer{}
	cli := cliServer{}

	ft := flaggy.NewSubcommand("tcp")
	ft.String(&tcp.addr, "", "addr", "listen addr:port")
	ft.StringSlice(&tcp.dirs, "", "dir", "limit dir")

	fc := flaggy.NewSubcommand("cli")
	//fc.StringSlice(&cli.files, "", "file", "open file")
	fc.AddPositionalValue(&cli.glob, "file", 1, true, "open log file eg: xxx.log")
	fc.String(&cli.timeRegex, "", "timeRegex", "time regex")
	fc.String(&cli.timeLayout, "", "timeLayout", "time layout")
	fc.String(&cli.start, "", "start", "start time")
	fc.String(&cli.stop, "", "stop", "stop time")
	fc.StringSlice(&cli.matchs, "", "match", "match keyword")

	flaggy.AttachSubcommand(ft, 1)
	flaggy.AttachSubcommand(fc, 1)
	flaggy.SetDescription("log viewer")
	flaggy.Parse()

	if fc.Used {
		cli.run()
		return
	}

	tcp.run()
}
