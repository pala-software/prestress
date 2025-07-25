package prestress

import "github.com/jackc/pgx/v5"

type Core struct {
	// All operations should be registered to here.
	ops Registry[AnyOperation]
}

func (feature Core) Operations() *Registry[AnyOperation] {
	return &feature.ops
}

func CoreFromEnv() *Core {
	feature := &Core{}
	// No configuration at this time
	return feature
}

func (feature *Core) Provider() any {
	return func(
		conn *pgx.Conn,
	) (
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
}

func (feature *Core) Invoker() any {
	return func() error {
		return nil
	}
}
