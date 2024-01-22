package web

import (
	"log"
	"net/http"
)

func NewServer(name, addr string) *Server {
	router := NewRouter(new(http.ServeMux))
	return &Server{
		Name:   name,
		Addr:   addr,
		Mux:    router.ServeMux,
		Router: router,
	}
}

type Server struct {
	Name   string
	Addr   string
	Router *Router
	Mux    *http.ServeMux
}

func (s *Server) Run() error {
	log.Println(s.Name, "启动中", s.Addr)
	return http.ListenAndServe(s.Addr, s.Mux)
}

func NewServers() *Servers {
	return &Servers{
		servers: make(NameServerMap, 3),
	}
}

type (
	NameServerMap map[string]*Server
	Servers       struct {
		servers NameServerMap
	}
)

func (s *Servers) New(name, addr string) *Server {
	s.Init(3)
	s.servers[name] = NewServer(name, addr)
	return s.servers[name]
}

func (s *Servers) Init(cap int) *Servers {
	if s.servers == nil {
		s.servers = make(NameServerMap, cap)
	}
	return s
}
