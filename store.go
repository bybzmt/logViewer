package main

type ServerConfig struct {
	ID         uint64 `boltholdKey:"ID"`
	Note       string
	Addr       string
	User       string
	Passwd     string
	UsePwd     bool
	PrivateKey string
}

type ViewLog struct {
	ID        uint64 `boltholdKey:"ID"`
	Note      string
	Files     string
	Separator byte
	LineMatch string
	Match     []string
	Decoder   string
	BeginTime int64
	StopTime  int64
}
