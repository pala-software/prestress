package subscriber_test

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/pala-software/prestress/pkg/auth"
	"gitlab.com/pala-software/prestress/pkg/crud"
	"gitlab.com/pala-software/prestress/pkg/migrator"
	"gitlab.com/pala-software/prestress/pkg/prestress"
	"gitlab.com/pala-software/prestress/pkg/subscriber"
	"go.uber.org/dig"
)

//go:embed subscriber_test.sql
var migrations embed.FS

var container *dig.Container

var features = []prestress.Feature{
	&prestress.Core{},
	&migrator.Migrator{},
	&crud.Crud{},
	&auth.Authentication{},
	&subscriber.Subscriber{},
}

type NoopAuthenticator struct{}

func (NoopAuthenticator) Authenticate(*http.Request) (*auth.AuthenticationResult, error) {
	return &auth.AuthenticationResult{Role: auth.AnonymousRole}, nil
}

func TestMain(m *testing.M) {
	var err error

	container, err = newContainer()
	if err != nil {
		log.Fatalln(err)
	}

	err = container.Invoke(runTestMigrations)
	if err != nil {
		log.Fatalln(err)
	}

	// Emit start event
	err = container.Invoke(func(lifecycle *prestress.Lifecycle) (err error) {
		for _, hook := range lifecycle.Start.Value() {
			err = hook()
			if err != nil {
				return
			}
		}
		return
	})
	if err != nil {
		log.Fatalln(err)
	}

	code := m.Run()

	// Emit shutdown event
	err = container.Invoke(func(lifecycle *prestress.Lifecycle) (err error) {
		for _, hook := range lifecycle.Shutdown.Value() {
			err = hook()
			if err != nil {
				return
			}
		}
		return
	})
	if err != nil {
		log.Fatalln(err)
	}

	os.Exit(code)
}

func newContainer() (c *dig.Container, err error) {
	c = dig.New()

	err = c.Provide(http.NewServeMux)
	if err != nil {
		return
	}

	err = c.Provide(databaseFromEnv)
	if err != nil {
		return
	}

	err = c.Provide(func() auth.Authenticator {
		return new(NoopAuthenticator)
	})

	for _, feature := range features {
		err = c.Provide(feature.Provider())
		if err != nil {
			return
		}
	}

	for _, feature := range features {
		err = c.Invoke(feature.Invoker())
		if err != nil {
			return
		}
	}

	return
}

func databaseFromEnv() (pool *pgxpool.Pool, err error) {
	connStr := os.Getenv("PRESTRESS_TEST_DB")
	pool, err = pgxpool.New(context.Background(), connStr)
	return
}

func runTestMigrations(mig *migrator.Migrator, pool *pgxpool.Pool) (err error) {
	err = mig.Migrate(pool)
	if err != nil {
		return
	}

	migration := migrator.MigrationTarget{
		Name:      "subscriber_test",
		Directory: migrations,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return
	}

	defer conn.Release()

	// Run test migrations forcefully
	err = migration.Migrate(conn, true)
	if err != nil {
		return
	}

	return
}

func begin(initCtx context.Context) (ctx prestress.OperationContext, err error) {
	err = container.Invoke(func(begin *prestress.BeginOperation, authenticator auth.Authenticator) (err error) {
		ctx, err = begin.Begin(initCtx, "subscriber", nil)
		ctx.Variables["auth"], err = authenticator.Authenticate(nil)
		if err != nil {
			return
		}

		return
	})
	return
}
