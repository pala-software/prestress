package prestress

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/jackc/pgx/v5"
)

//go:embed migrations/*.sql
var migrations embed.FS

type migrationTarget struct {
	Name      string
	Directory fs.FS
}

func (target migrationTarget) Migrate(server Server, forceRunAll bool) error {
	var err error
	ctx := context.Background()

	var initialized *bool
	err = server.DB.QueryRow(
		ctx,
		`SELECT TRUE
		FROM pg_namespace
		WHERE nspname = 'prestress'`,
	).Scan(&initialized)
	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	if (initialized == nil || !*initialized) && target.Name != "prestress" {
		return fmt.Errorf(
			"cannot migrate %s, initial prestress migration has not been run",
			target.Name,
		)
	}

	version := ""
	if initialized != nil && *initialized {
		var variable sql.NullString
		err = server.DB.QueryRow(
			ctx,
			`SELECT value
			FROM prestress.database_variable
			WHERE name = $1`,
			target.Name+"_version",
		).Scan(&variable)
		if err != nil && err != pgx.ErrNoRows {
			return err
		}

		if variable.Valid {
			version = variable.String
		}
	}

	entries, err := fs.ReadDir(target.Directory, ".")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if name <= version && !forceRunAll {
			// Skip older and current versions.
			continue
		}

		migration, err := fs.ReadFile(target.Directory, name)
		if err != nil {
			return err
		}

		fmt.Printf("Running migration for %s: %s\n", target.Name, name)
		_, err = server.DB.Exec(ctx, string(migration))
		if err != nil {
			return err
		}

		_, err = server.DB.Exec(
			ctx,
			`INSERT INTO prestress.database_variable (name, value)
			VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE SET value = $2`,
			target.Name+"_version",
			name,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (server *Server) AddMigration(name string, directory fs.FS) {
	server.migrations = append(
		server.migrations,
		migrationTarget{Name: name, Directory: directory},
	)
}

func (server Server) MigratePrestress() error {
	var err error

	dir, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return err
	}

	target := migrationTarget{
		Name:      "prestress",
		Directory: dir,
	}
	err = target.Migrate(server, false)
	if err != nil {
		return err
	}

	for _, target := range server.migrations {
		err = target.Migrate(server, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (server Server) MigrateApp() error {
	target := migrationTarget{
		Name:      "app",
		Directory: os.DirFS(server.MigrationDir),
	}
	return target.Migrate(server, false)
}

func (server Server) RunMigrations() error {
	var err error

	err = server.ConnectToDatabase()
	if err != nil {
		return err
	}

	err = server.MigratePrestress()
	if err != nil {
		return err
	}

	if server.MigrationDir != "" {
		server.MigrateApp()
	}

	fmt.Println("Database is up to date!")
	return nil
}
