package migrator

type Migratable interface {
	RegisterMigrations(*Migrator) error
}
