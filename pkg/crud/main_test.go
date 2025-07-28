package crud_test

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/pala-software/prestress/pkg/crud"
	"gitlab.com/pala-software/prestress/pkg/migrator"
	"gitlab.com/pala-software/prestress/pkg/prestress"
	"go.uber.org/dig"
)

//go:embed main_test.sql
var migrations embed.FS

var container *dig.Container

var features = []prestress.Feature{
	&prestress.Core{},
	&migrator.Migrator{},
	&crud.Crud{},
}

type Item struct {
	Value string `json:"value"`
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

	code := m.Run()
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
		Name:      "crud_test",
		Directory: migrations,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := pool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return
	}

	// Run test migrations forcefully
	err = migration.Migrate(conn, true)
	if err != nil {
		return
	}

	return
}

func begin(initCtx context.Context) (ctx prestress.OperationContext, err error) {
	err = container.Invoke(func(begin *prestress.BeginOperation) (err error) {
		ctx, err = begin.Begin(initCtx, "test", nil)
		return
	})
	return
}

func expectItems(
	ctx prestress.OperationContext,
	table string,
	where crud.Where,
	expectedValues []string,
) (err error) {
	var result pgx.Rows
	err = container.Invoke(func(find *crud.FindOperation) (err error) {
		params := crud.FindParams{
			Table:  table,
			Where:  where,
			Limit:  100,
			Offset: 0,
		}
		result, err = find.Execute(ctx, params)
		return
	})
	if err != nil {
		return
	}

	if result == nil {
		return errors.New("result is nil")
	}

	defer result.Close()
	count := 0
	for _, expectedValue := range expectedValues {
		if !result.Next() {
			return fmt.Errorf(
				"expected %d rows, got %d",
				len(expectedValues),
				count,
			)
		}

		var actualItem Item
		err = result.Scan(&actualItem)
		if err != nil {
			return
		}

		if actualItem.Value != expectedValue {
			return fmt.Errorf(
				"expected value '%s', got '%s'",
				expectedValue,
				actualItem.Value,
			)
		}

		count++
	}

	if result.Next() {
		return fmt.Errorf(
			"expected %d rows, got too many",
			len(expectedValues),
		)
	}

	return
}
