package prestress_test

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"gitlab.com/pala-software/prestress/pkg/prestress"
)

//go:embed main_test.sql
var mainTestMigrations embed.FS

func TestMain(m *testing.M) {
	var err error

	server = prestress.Server{}
	server.Environment = prestress.Development
	server.DbConnStr = "dbname=prestress_test"
	server.DisableAuth = true

	err = server.ConnectToDatabase()
	if err != nil {
		log.Fatalln(err)
	}

	err = server.MigratePrestress()
	if err != nil {
		log.Fatalln(err)
	}

	err = server.Migrate("main_test", mainTestMigrations, true)
	if err != nil {
		log.Fatalln(err)
	}

	code := m.Run()
	os.Exit(code)
}

func expectItems(
	ctx context.Context,
	table string,
	where prestress.Where,
	expectedValues []string,
) error {
	auth := prestress.AuthenticationResult{
		Role:      "anonymous",
		Variables: map[string]any{},
	}
	schema := "test"

	result, err := server.Find(ctx, auth, schema, table, where, 100, 0)
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
