package prestress

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO: Test
func (server *Server) ConnectToDatabase() error {
	var err error

	server.DB, err = pgxpool.New(context.Background(), server.DbConnStr)
	if err != nil {
		return err
	}

	return nil
}
