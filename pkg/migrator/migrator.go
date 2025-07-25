package migrator

import (
	"github.com/jackc/pgx/v5"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

type Migrator struct {
	Targets prestress.Registry[MigrationTarget]
}

func MigratorFromEnv() *Migrator {
	return &Migrator{}
}

func (mig Migrator) Migrate(conn *pgx.Conn) (err error) {
	for _, target := range mig.Targets.Value() {
		err = target.Migrate(conn, false)
		if err != nil {
			return
		}
	}

	return
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

func (feature Migrator) Provider() any {
	return feature.Register
}

func (mig *Migrator) Register() (self *Migrator, err error) {
	self = mig

	err = mig.RegisterMigrations(mig)
	if err != nil {
		return
	}

	return
}
