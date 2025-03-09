package prestress

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
)

//go:embed migrations/*.sql
var migrations embed.FS

// TODO: Test
func (server Server) Migrate(target string, dir fs.FS, forceRunAll bool) error {
	var err error
	var initialized sql.NullBool
	err = server.DB.QueryRow(
		`SELECT TRUE
		FROM pg_namespace
		WHERE nspname = 'prestress'`,
	).Scan(&initialized)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if (!initialized.Valid || !initialized.Bool) && target != "prestress" {
		return fmt.Errorf(
			"cannot migrate %s, initial prestress migration has not been run",
			target,
		)
	}

	version := ""
	if initialized.Valid && initialized.Bool {
		var variable sql.NullString
		err = server.DB.QueryRow(
			`SELECT value
			FROM prestress.database_variable
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
			`INSERT INTO prestress.database_variable (name, value)
			VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE SET value = $2`,
			target+"_version",
			name,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: Test
func (server Server) MigratePrestress() error {
	dir, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return err
	}

	return server.Migrate("prestress", dir, false)
}

// TODO: Test
func (server Server) RunMigrations() error {
	var err error

	err = server.ReadConfiguration()
	if err != nil {
		return err
	}

	err = server.ConnectToDatabase()
	if err != nil {
		return err
	}

	err = server.MigratePrestress()
	if err != nil {
		return err
	}

	if server.MigrationDir != "" {
		dir := os.DirFS(server.MigrationDir)
		err = server.Migrate("app", dir, false)
		if err != nil {
			return err
		}
	}

	fmt.Println("Database is up to date!")
	return nil
}
