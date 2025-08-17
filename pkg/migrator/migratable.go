package migrator

import "github.com/jackc/pgx/v5/pgxpool"

type Migratable interface {
	Migrate(
		conn *pgxpool.Conn,
		forceRunAll bool,
	) error
}
