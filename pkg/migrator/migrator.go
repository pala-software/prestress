package migrator

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

type Migrator struct {
	Targets prestress.Registry[MigrationTarget]
}

func MigratorFromEnv() *Migrator {
	return &Migrator{}
}

func (mig Migrator) Migrate(pool *pgxpool.Pool) (err error) {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		return
	}

	defer conn.Release()

	for _, target := range mig.Targets.Value() {
		err = target.Migrate(conn, false)
		if err != nil {
			return
		}
	}

	return
}

func (mig *Migrator) Provider() any {
	return func() (self *Migrator) {
		self = mig
		return
	}
}

func (mig *Migrator) Invoker() any {
	return func() (err error) {
		err = mig.RegisterMigrations(mig)
		if err != nil {
			return
		}

		return
	}
}

func (Migrator) RegisterMigrations(mig *Migrator) (err error) {
	err = RegisterPrestressMigrations(mig)
	if err != nil {
		return
	}

	err = RegisterMigrationsFromEnv(mig)
	if err != nil {
		return
	}

	return
}
