package find

import (
	"log"
	"net"
)

type ServerTCP struct {
	Server
	Addr string
}

func (s *ServerTCP) Run() {
	if s.Addr == "" {
		s.Addr = "127.0.0.2:7000"
	}

	s.Init()

	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Listen", s.Addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go s.Service(conn)
	}
}
