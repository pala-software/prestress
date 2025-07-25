package main

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
)

func databaseFromEnv() (conn *pgx.Conn, err error) {
	connStr := os.Getenv("PRESTRESS_DB")
	conn, err = pgx.Connect(context.Background(), connStr)
	return
}
