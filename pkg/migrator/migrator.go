package migrator

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

type Migrator struct {
	Targets prestress.Registry[Migratable]
}

func MigratorFromEnv() *Migrator {
	return &Migrator{}
}

func (mig Migrator) Migrate(pool *pgxpool.Pool) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := pool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return
	}

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
		err = RegisterPrestressMigrations(mig)
		if err != nil {
			return
		}

		return
	}
}
