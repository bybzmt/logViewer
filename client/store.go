package client

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
	Separator uint8
	LineMatch string
	Filter    string
	Decoder   string
}
