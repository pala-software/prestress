package prestress

import "github.com/jackc/pgx/v5"

type Core struct {
	// All operations should be registered to here.
	ops Registry[AnyOperation]
}

func CoreFromEnv() *Core {
	feature := &Core{}
	// No configuration at this time
	return feature
}

func (feature Core) Provider() any {
	return feature.Register
}

func (feature Core) Operations() *Registry[AnyOperation] {
	return &feature.ops
}

func (feature *Core) Register(conn *pgx.Conn) (
	self *Core,
	lifecycle *Lifecycle,
	begin *BeginOperation,
) {
	self = feature
	lifecycle = NewLifecycle()
	begin = NewBeginOperation(conn)

	feature.Operations().Register(begin)

	return
}
