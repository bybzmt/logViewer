package find

import (
	"net"
)

func Dial(addr string) (Client, error) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return newClient(c), nil
}
