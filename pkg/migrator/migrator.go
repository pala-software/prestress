package migrator

import (
	"os"

	"gitlab.com/pala-software/prestress/pkg/prestress"
)

type Migrator struct {
	MigrationDir string
}

func MigratorFromEnv() *Migrator {
	return &Migrator{
		MigrationDir: os.Getenv("PRESTRESS_MIGRATIONS"),
	}
}

func (feature Migrator) Apply(server *prestress.Server) error {
	if feature.MigrationDir != "" {
		server.AddMigration("app", os.DirFS(feature.MigrationDir))
	}
	return nil
}
