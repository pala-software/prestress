package server

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
)

//go:embed migrations/*.sql
var migrations embed.FS

func (server Server) Migrate(target string, dir fs.FS, forceRunAll bool) error {
	var err error
	var initialized sql.NullBool
	err = server.DB.QueryRow(
		`SELECT TRUE
		FROM pg_namespace
		WHERE nspname = 'palakit'`,
	).Scan(&initialized)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if (!initialized.Valid || !initialized.Bool) && target != "palakit" {
		return fmt.Errorf(
			"cannot migrate %s, initial palakit migration has not been run",
			target,
		)
	}

	version := ""
	if initialized.Valid && initialized.Bool {
		var variable sql.NullString
		err = server.DB.QueryRow(
			`SELECT value
			FROM palakit.database_variable
			WHERE name = $1`,
			target+"_version",
		).Scan(&variable)
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		if variable.Valid {
			version = variable.String
		}
	}

	entries, err := fs.ReadDir(dir, ".")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if name <= version && !forceRunAll {
			// Skip older and current versions.
			continue
		}

		migration, err := fs.ReadFile(dir, name)
		if err != nil {
			return err
		}

		fmt.Printf("Running migration %s...\n", name)
		_, err = server.DB.Exec(string(migration))
		if err != nil {
			return err
		}

		_, err = server.DB.Exec(
			`INSERT INTO palakit.database_variable (name, value)
			VALUES ('palakit_version', $1)
			ON CONFLICT (name) DO UPDATE SET value = $1`,
			name,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (server Server) MigratePalakit() error {
	dir, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return err
	}

	return server.Migrate("palakit", dir, false)
}

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

	err = server.MigratePalakit()
	if err != nil {
		return err
	}

	if server.migrationDir != "" {
		dir := os.DirFS(server.migrationDir)
		err = server.Migrate("app", dir, false)
		if err != nil {
			return err
		}
	}

	fmt.Println("Database is up to date!")
	return nil
}
