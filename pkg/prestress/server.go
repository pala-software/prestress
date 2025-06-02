package prestress

import (
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	DbConnStr     string
	Authenticator Authenticator
	DB            *pgxpool.Pool
	serveMux      *http.ServeMux
	listeners     []EventListener
	middleware    []Middleware
	migrations    []MigrationTarget
}

// Construct Server and read configuration from environment variables.
func ServerFromEnv() Server {
	server := Server{}
	server.DbConnStr = os.Getenv("PRESTRESS_DB")
	if server.DbConnStr == "" {
		panic("empty or unset PRESTRESS_DB")
	}
	return server
}

func (server *Server) ApplyFeatures(features ...Feature) error {
	var err error
	for _, feature := range features {
		err = feature.Apply(server)
		if err != nil {
			return err
		}
	}
	return nil
}

func (server *Server) Start() error {
	if err := server.ConnectToDatabase(); err != nil {
		return err
	}
	if err := server.StartHttpServer(); err != nil {
		return err
	}
	return nil
}
