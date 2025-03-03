package server

import (
	"database/sql"
	"embed"
	"fmt"
)

//go:embed migrations/*.sql
var migrations embed.FS

func (server Server) RunMigrations() error {
	var err error

	err = server.readConfiguration()
	if err != nil {
		return err
	}

	err = server.connectToDatabase()
	if err != nil {
		return err
	}

	var initialized sql.NullBool
	err = server.DB.QueryRow(
		`SELECT TRUE
		FROM pg_namespace
		WHERE nspname = 'meta'`,
	).Scan(&initialized)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	version := ""
	if initialized.Valid && initialized.Bool {
		var variable sql.NullString
		err = server.DB.QueryRow(
			`SELECT value
			FROM meta.variable
			WHERE name = 'database_version'`,
		).Scan(&variable)
		if err != nil {
			return err
		}

		if variable.Valid {
			version = variable.String
		}
	}

	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if name <= version {
			// Skip older and current versions.
			continue
		}

		migration, err := migrations.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}

		fmt.Printf("Running migration %s...\n", name)
		_, err = server.DB.Query(string(migration))
		if err != nil {
			return err
		}

		_, err = server.DB.Query(
			`INSERT INTO meta.variable (name, value)
			VALUES ('database_version', $1)
			ON CONFLICT (name) DO UPDATE SET value = $1`,
			name,
		)
		if err != nil {
			return err
		}
	}

	fmt.Println("Database is up to date!")
	return nil
}
