package prestress

import (
	"net/http"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	// Configuration
	Environment      Environment
	DbConnStr        string
	MigrationDir     string
	AllowedOrigins   string
	DisableAuth      bool
	ClientId         string
	ClientSecret     string
	IntrospectionUrl *url.URL

	// Connections
	DB   *pgxpool.Pool
	HTTP *http.Server

	// State
	subscriptions map[int]*Subscription
}

// TODO: Test
func (server Server) Start() error {
	if err := server.ReadConfiguration(); err != nil {
		return err
	}
	if err := server.ConnectToDatabase(); err != nil {
		return err
	}
	if err := server.StartHttpServer(); err != nil {
		return err
	}
	return nil
}
