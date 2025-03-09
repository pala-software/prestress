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
	migrationDir     string
	disableAuth      bool
	clientId         string
	clientSecret     string
	introspectionUrl *url.URL

	// Connections
	DB   *sql.DB
	HTTP *http.Server

	// State
	subscriptions map[int]*Subscription
}

// TODO: Test
func (server Server) Start() error {
	// Initialize state
	server.subscriptions = make(map[int]*Subscription)

	if err := server.readConfiguration(); err != nil {
		return err
	}
	if err := server.connectToDatabase(); err != nil {
		return err
	}
	go server.listenForChange()
	if err := server.startHttpServer(); err != nil {
		return err
	}
	return nil
}
