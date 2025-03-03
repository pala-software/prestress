package server

import (
	"database/sql"
	"net/http"
	"net/url"
)

type Server struct {
	// Configuration
	environment      Environment
	dbConnStr        string
	disableAuth      bool
	clientId         string
	clientSecret     string
	introspectionUrl *url.URL

	// Connections
	DB   *sql.DB
	HTTP *http.Server
}

func (server Server) Start() error {
	if err := server.readConfiguration(); err != nil {
		return err
	}
	if err := server.connectToDatabase(); err != nil {
		return err
	}
	if err := server.startHttpServer(); err != nil {
		return err
	}
	return nil
}
