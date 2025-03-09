package palakit

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
func (server *Server) ReadConfiguration() error {
	var err error

	server.Environment = Environment(os.Getenv("PALAKIT_ENVIRONMENT"))
	if !server.Environment.IsValid() {
		return fmt.Errorf("invalid PALAKIT_ENVIRONMENT '%s'", server.Environment)
	}

	server.DbConnStr = os.Getenv("PALAKIT_DB_CONNECTION_STRING")
	server.MigrationDir = os.Getenv("PALAKIT_MIGRATION_DIRECTORY")

	server.DisableAuth = os.Getenv("PALAKIT_AUTH_DISABLE") == "1"
	if server.DisableAuth && server.Environment != Development {
		return fmt.Errorf("cannot disable authentication on non-development environments")
	}

	if !server.DisableAuth {
		server.IntrospectionUrl, err = url.Parse(os.Getenv("PALAKIT_AUTH_INTROSPECTION_URL"))
		if err != nil {
			return err
		}
		if server.IntrospectionUrl.String() == "" {
			return fmt.Errorf("empty or unset PALAKIT_AUTH_INTROSPECTION_URL")
		}

		server.ClientId = os.Getenv("PALAKIT_AUTH_CLIENT_ID")
		if server.ClientId == "" {
			return fmt.Errorf("empty or unset PALAKIT_AUTH_CLIENT_ID")
		}

		server.ClientSecret = os.Getenv("PALAKIT_AUTH_CLIENT_SECRET")
		if server.ClientSecret == "" {
			return fmt.Errorf("empty or unset PALAKIT_AUTH_CLIENT_SECRET")
		}
	}

	return nil
}
