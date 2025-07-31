package main

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/pala-software/prestress/pkg/migrator"
)

func migrate() (err error) {
	c, err := container()
	if err != nil {
		return
	}

	err = c.Invoke(func(mig *migrator.Migrator, pool *pgxpool.Pool) error {
		return mig.Migrate(pool)
	})
	if err != nil {
		return
	}

	fmt.Println("Database is up to date!")
	return
}
