package tcp

import (
	"log"
	"logViewer/find"
	"net"
)

type Server struct {
	find.Server
	Addr string
}

func (s *Server) Run() {
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
