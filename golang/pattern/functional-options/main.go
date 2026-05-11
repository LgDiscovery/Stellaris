package main

import (
	"fmt"
	"time"
)

type Server struct {
	host    string
	port    int
	timeout time.Duration
	maxConn int
}

type Option func(*Server)

func WithHost(host string) Option {
	return func(s *Server) {
		s.host = host
	}
}

func WithPort(port int) Option {
	return func(s *Server) {
		s.port = port
	}
}

func WithTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.timeout = d
	}
}

func WithMaxConnections(n int) Option {
	return func(s *Server) {
		s.maxConn = n
	}
}

func NewServer(opts ...Option) *Server {
	s := &Server{
		host:    "localhost",
		port:    80,
		timeout: 5 * time.Second,
		maxConn: 100,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func main() {
	s := NewServer(
		WithHost("localhost"),
		WithPort(8000),
		WithMaxConnections(100),
		WithTimeout(5*time.Second))
	fmt.Printf("%+v\n", s)
}
