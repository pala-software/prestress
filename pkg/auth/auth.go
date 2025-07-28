package auth

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5"
	"gitlab.com/pala-software/prestress/pkg/migrator"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Authentication struct{}

// Construct Subscribe Feature and read configuration from environment
// variables.
func AuthenticationFromEnv() *Authentication {
	feature := &Authentication{}
	// No configuration at this time
	return feature
}

func (feature *Authentication) Provider() any {
	return func() (self *Authentication) {
		self = feature
		return
	}
}

func (feature *Authentication) Invoker() any {
	return func(
		mig *migrator.Migrator,
		authenticator Authenticator,
		begin *prestress.BeginOperation,
	) (err error) {

		err = feature.RegisterMigrations(mig)
		if err != nil {
			return
		}

		err = feature.RegisterHooks(authenticator, begin)
		if err != nil {
			return
		}

		return
	}
}

func (Authentication) RegisterMigrations(mig *migrator.Migrator) (err error) {
	dir, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return
	}

	mig.Targets.Register(migrator.MigrationTarget{
		Name:      "auth",
		Directory: dir,
	})
	return
}

func (Authentication) RegisterHooks(
	authenticator Authenticator,
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

		if ctx.Request == nil {
			// No authentication when not called from HTTP.
			return
		}

		auth, err := authenticator.Authenticate(ctx.Request)
		if err != nil {
			return
		}

		ctx.Variables["auth"] = auth
		return
	})

	begin.After().Register(func(
		_ prestress.OperationContext,
		params prestress.EmptyOperationParams,
		initCtx prestress.OperationContext,
	) (ctx prestress.OperationContext, err error) {
		ctx = initCtx

		auth, ok := ctx.Variables["auth"].(*AuthenticationResult)
		if !ok {
			return
		}

		_, err = ctx.Tx.Exec(
			ctx,
			fmt.Sprintf(
				"SET LOCAL role TO %s",
				pgx.Identifier{auth.Role}.Sanitize(),
			),
		)
		if err != nil {
			return
		}

		variables := map[string]any{}
		if auth.Variables != nil {
			variables = auth.Variables
		}

		encodedVariables, err := json.Marshal(variables)
		if err != nil {
			return
		}

		_, err = ctx.Tx.Exec(
			ctx,
			"SELECT prestress.begin_authorized($1)",
			encodedVariables,
		)
		if err != nil {
			return
		}

		return
	})

	return
}
