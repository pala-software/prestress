package prestress_test

import (
	"context"
	"embed"
	"log"
	"os"
	"testing"

	"gitlab.com/pala-software/prestress/pkg/prestress"
)

var server prestress.Server

//go:embed find_test.sql
var findTestMigrations embed.FS

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
		Variables: map[string]interface{}{},
	}
	schema := "find_test"
	table := "test"
	filters := prestress.FilterMap{}

	cancel()
	rows, err := server.Find(ctx, auth, schema, table, filters)

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
		Variables: map[string]interface{}{},
	}
	schema := "find_test"
	table := "test"
	filters := prestress.FilterMap{}

	rows, err := server.Find(ctx, auth, schema, table, filters)
	if err != nil {
		t.Error(err)
	}

	if rows == nil {
		t.Error("rows is nil")
		return
	}

	err = expectValues(rows, []string{"1", "2"})
	if err != nil {
		t.Error(err)
		return
	}

	err = rows.Close()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestFindWithFilter(t *testing.T) {
	ctx := context.Background()
	auth := prestress.AuthenticationResult{
		Role:      "anonymous",
		Variables: map[string]interface{}{},
	}
	schema := "find_test"
	table := "test"
	filters := prestress.FilterMap{
		"test": "1",
	}

	rows, err := server.Find(ctx, auth, schema, table, filters)
	if err != nil {
		t.Error(err)
	}

	if rows == nil {
		t.Error("rows is nil")
		return
	}

	err = expectValues(rows, []string{"1"})
	if err != nil {
		t.Error(err)
		return
	}

	err = rows.Close()
	if err != nil {
		t.Error(err)
		return
	}
}
