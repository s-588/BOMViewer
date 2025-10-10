package http

import "net/http"

type Server struct {
	Port string
}

func NewServer(port string) *Server {
	return &Server{
		Port: port,
	}
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.Port, nil)
}
