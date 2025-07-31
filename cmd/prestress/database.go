package main

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func databaseFromEnv() (pool *pgxpool.Pool, err error) {
	connStr := os.Getenv("PRESTRESS_DB")
	pool, err = pgxpool.New(context.Background(), connStr)
	return
}
