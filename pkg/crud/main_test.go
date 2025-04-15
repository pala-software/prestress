package crud_test

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"gitlab.com/pala-software/prestress/pkg/crud"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

//go:embed main_test.sql
var migrations embed.FS

var feature crud.Crud
var server prestress.Server

func TestMain(m *testing.M) {
	var err error

	server.DbConnStr = os.Getenv("PRESTRESS_TEST_DB")

	err = server.ApplyFeatures(&feature)
	if err != nil {
		log.Fatalln(err)
	}

	err = server.ConnectToDatabase()
	if err != nil {
		log.Fatalln(err)
	}

	server.AddMigration("crud_test", migrations)
	err = server.MigratePrestress()
	if err != nil {
		log.Fatalln(err)
	}

	code := m.Run()
	os.Exit(code)
}

func expectItems(
	ctx context.Context,
	table string,
	where crud.Where,
	expectedValues []string,
) error {
	auth := prestress.AuthenticationResult{
		Role:      "anonymous",
		Variables: map[string]any{},
	}
	schema := "test"

	result, err := feature.Find(ctx, auth, schema, table, where, 100, 0)
	if err != nil {
		return err
	}

	if result == nil {
		return errors.New("result is nil")
	}

	defer result.Done()
	count := 0
	for _, expectedValue := range expectedValues {
		if !result.Rows.Next() {
			return fmt.Errorf(
				"expected %d rows, got %d",
				len(expectedValues),
				count,
			)
		}

		var actualValue string
		err = result.Rows.Scan(&actualValue)
		if err != nil {
			return err
		}

		if actualValue != expectedValue {
			return fmt.Errorf(
				"expected value '%s', got '%s'",
				expectedValue,
				actualValue,
			)
		}

		count++
	}

	if result.Rows.Next() {
		return fmt.Errorf(
			"expected %d rows, got too many",
			len(expectedValues),
		)
	}

	return nil
}
