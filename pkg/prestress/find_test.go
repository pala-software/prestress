package prestress_test

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

var server prestress.Server

//go:embed find_test.sql
var findTestMigrations embed.FS

func expectValues(rows pgx.Rows, expectedValues []string) error {
	var err error
	count := 0
	for _, expectedValue := range expectedValues {
		if !rows.Next() {
			return fmt.Errorf(
				"expected %d rows, got %d",
				len(expectedValues),
				count,
			)
		}

		var actualValue string
		err = rows.Scan(&actualValue)
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

	if rows.Next() {
		return fmt.Errorf(
			"expected %d rows, got too many",
			len(expectedValues),
		)
	}

	return nil
}

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

	err = server.Migrate("find_test", findTestMigrations, true)
	if err != nil {
		log.Fatalln(err)
	}

	code := m.Run()
	os.Exit(code)
}

func TestFindWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	auth := prestress.AuthenticationResult{
		Role:      "anonymous",
		Variables: map[string]any{},
	}
	schema := "find_test"
	table := "test"
	where := prestress.Where{}

	cancel()
	rows, err := server.Find(ctx, auth, schema, table, where, 100, 0)

	if rows != nil {
		t.Error("rows should be nil")
		return
	}

	if err != context.Canceled {
		t.Logf(
			"expected error to be '%v', got '%v'",
			context.Canceled,
			err,
		)
		return
	}
}

func TestFindAll(t *testing.T) {
	ctx := context.Background()
	auth := prestress.AuthenticationResult{
		Role:      "anonymous",
		Variables: map[string]any{},
	}
	schema := "find_test"
	table := "test"
	where := prestress.Where{}

	result, err := server.Find(ctx, auth, schema, table, where, 100, 0)
	if err != nil {
		t.Error(err)
	}

	if result == nil {
		t.Error("result is nil")
		return
	}

	defer result.Done()
	err = expectValues(result.Rows, []string{"1", "2"})
	if err != nil {
		t.Error(err)
		return
	}
}

func TestFindWithFilter(t *testing.T) {
	ctx := context.Background()
	auth := prestress.AuthenticationResult{
		Role:      "anonymous",
		Variables: map[string]any{},
	}
	schema := "find_test"
	table := "test"
	where := prestress.Where{
		"test": "1",
	}

	result, err := server.Find(ctx, auth, schema, table, where, 100, 0)
	if err != nil {
		t.Error(err)
	}

	if result == nil {
		t.Error("result is nil")
		return
	}

	defer result.Done()
	err = expectValues(result.Rows, []string{"1"})
	if err != nil {
		t.Error(err)
		return
	}
}
