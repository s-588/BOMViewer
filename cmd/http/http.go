package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/s-588/BOMViewer/web/templates"
)

type Server struct {
	ctx    context.Context
	cancel context.CancelFunc
	mux    *http.ServeMux
	Port   string
}

func NewServer(cancel context.CancelFunc, port string) *Server {
	return &Server{
		cancel: cancel,
		mux:    http.NewServeMux(),
		Port:   port,
	}
}

func (s *Server) Start() error {
	s.setupPaths()
	return http.ListenAndServe(s.Port, s.mux)
}

func (s *Server) setupPaths() {
	s.mux.Handle("/static/", http.FileServer(http.Dir("web/")))
	s.mux.HandleFunc("GET /welcome", s.welcomePage)
	s.mux.HandleFunc("POST /debug/exit", s.stop)
}

func (s *Server) stop(w http.ResponseWriter, r *http.Request) {
	s.cancel()
}

func (s *Server) welcomePage(w http.ResponseWriter, r *http.Request) {
	ctx, _ := context.WithTimeout(r.Context(), time.Second*10)
	err := templates.Index(r.Context()).Render(ctx, w)
	if err != nil {
		slog.Error("can't render welcome page", "error", err)
	}
}
