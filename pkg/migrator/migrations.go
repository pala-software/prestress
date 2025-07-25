package migrator

import (
	"embed"
	"io/fs"
	"os"
)

//go:embed migrations/*.sql
var migrations embed.FS

func RegisterPrestressMigrations(mig *Migrator) (err error) {
	dir, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return
	}

	mig.Targets.Register(MigrationTarget{
		Name:      "prestress",
		Directory: dir,
	})
	return
}

func RegisterMigrationsFromEnv(mig *Migrator) (err error) {
	dir := os.Getenv("PRESTRESS_MIGRATIONS")
	if dir == "" {
		return
	}

	mig.Targets.Register(MigrationTarget{
		Name:      "app",
		Directory: os.DirFS(dir),
	})
	return
}
