package server

import (
	"fmt"
	"net/url"
	"os"
)

type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
)

// TODO: Test
func (env Environment) IsValid() bool {
	switch env {
	case Development, Production:
		return true
	default:
		return false
	}
}

// TODO: Test
func (server *Server) readConfiguration() error {
	var err error

	server.environment = Environment(os.Getenv("PALAKIT_ENVIRONMENT"))
	if !server.environment.IsValid() {
		return fmt.Errorf("invalid PALAKIT_ENVIRONMENT '%s'", server.environment)
	}

	server.dbConnStr = os.Getenv("PALAKIT_DB_CONNECTION_STRING")
	server.migrationDir = os.Getenv("PALAKIT_MIGRATION_DIRECTORY")

	server.disableAuth = os.Getenv("PALAKIT_AUTH_DISABLE") == "1"
	if server.disableAuth && server.environment != Development {
		return fmt.Errorf("cannot disable authentication on non-development environments")
	}

	if !server.disableAuth {
		server.introspectionUrl, err = url.Parse(os.Getenv("PALAKIT_AUTH_INTROSPECTION_URL"))
		if err != nil {
			return err
		}
		if server.introspectionUrl.String() == "" {
			return fmt.Errorf("empty or unset PALAKIT_AUTH_INTROSPECTION_URL")
		}

		server.clientId = os.Getenv("PALAKIT_AUTH_CLIENT_ID")
		if server.clientId == "" {
			return fmt.Errorf("empty or unset PALAKIT_AUTH_CLIENT_ID")
		}

		server.clientSecret = os.Getenv("PALAKIT_AUTH_CLIENT_SECRET")
		if server.clientSecret == "" {
			return fmt.Errorf("empty or unset PALAKIT_AUTH_CLIENT_SECRET")
		}
	}

	return nil
}
