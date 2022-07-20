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
	ID         uint64 `boltholdKey:"ID"`
	Note       string
	Files      string
	TimeRegex  string
	TimeLayout string
	Contains   string
	Regex      string
	Decoder    string
	ServerID   uint64
}
