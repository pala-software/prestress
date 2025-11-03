package param

import (
	"embed"
	"encoding/json"
	"io/fs"
	"strings"

	"gitlab.com/pala-software/prestress/pkg/migrator"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Param struct{}

func ParamFromEnv() *Param {
	feature := &Param{}
	return feature
}

func (feature *Param) Provider() any {
	return func() (self *Param) {
		self = feature
		return
	}
}

func (feature *Param) Invoker() any {
	return func(
		mig *migrator.Migrator,
		begin *prestress.BeginOperation,
	) (err error) {
		err = feature.RegisterMigrations(mig)
		if err != nil {
			return
		}

		err = feature.RegisterHooks(begin)
		if err != nil {
			return
		}

		return
	}
}

func (feature *Param) Set(ctx prestress.OperationContext, key string, value string) {
	pmap, ok := ctx.Variables["params"].(ParamMap)
	if ok {
		pmap[key] = value
	} else {
		ctx.Variables["params"] = ParamMap{
			key: value,
		}
	}
}

func (*Param) RegisterMigrations(mig *migrator.Migrator) (err error) {
	dir, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return
	}

	mig.Targets.Register(migrator.MigrationTarget{
		Name:      "param",
		Directory: dir,
	})
	return
}

func (feature *Param) RegisterHooks(
	begin *prestress.BeginOperation,
) (err error) {
	begin.Before().Register(func(
		initCtx prestress.OperationContext,
		initParams prestress.EmptyOperationParams,
	) (
		ctx prestress.OperationContext,
		params prestress.EmptyOperationParams,
		err error,
	) {
		ctx = initCtx
		params = initParams

		if ctx.Request != nil {
			for key, values := range ctx.Request.URL.Query() {
				var found bool

				key, found = strings.CutPrefix(key, "param[")
				if !found {
					continue
				}

				key, found = strings.CutSuffix(key, "]")
				if !found {
					continue
				}

				if len(values) == 0 {
					continue
				}

				feature.Set(ctx, key, values[0])
			}
		}

		return
	})

	begin.After().Register(func(
		_ prestress.OperationContext,
		_ prestress.EmptyOperationParams,
		initCtx prestress.OperationContext,
	) (ctx prestress.OperationContext, err error) {
		ctx = initCtx

		pmap, ok := ctx.Variables["params"].(ParamMap)
		if !ok {
			return
		}

		encodedParams, err := json.Marshal(pmap)
		if err != nil {
			return
		}

		_, err = ctx.Tx.Exec(
			ctx,
			"SELECT prestress.set_params($1)",
			encodedParams,
		)
		if err != nil {
			return
		}

		return
	})

	return
}
