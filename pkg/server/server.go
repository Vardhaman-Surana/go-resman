package server

import (
	"errors"
	"github.com/vds/go-resman/pkg/database"
)

type Server struct {
	DB database.Database
}

func NewServer(data database.Database) (*Server, error) {
	if data == nil {
		return nil, errors.New("server expects a valid database instance")
	}
	return &Server{DB: data}, nil
}

func (server *Server) Start() (*Router, error) {
	router, err := NewRouter(server.DB)
	if err != nil {
		return nil, err
	}
	r := router.Create()
	return r, nil
}
