package tcp

import (
	"logViewer/find"
	"net"
)

func Dial(addr string) (*find.Client, error) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return find.NewClient(c), nil
}
