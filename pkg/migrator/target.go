package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MigrationTarget struct {
	Name      string
	Directory fs.FS
}

func (target MigrationTarget) Migrate(
	conn *pgxpool.Conn,
	forceRunAll bool,
) error {
	var err error
	ctx := context.Background()

	var initialized *bool
	err = conn.QueryRow(
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
		err = conn.QueryRow(
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
		_, err = conn.Exec(ctx, string(migration))
		if err != nil {
			return err
		}

		_, err = conn.Exec(
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
