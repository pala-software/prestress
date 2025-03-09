package server

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"testing"
)

var server Server

//go:embed find_test.sql
var findTestMigrations embed.FS

func TestMain(m *testing.M) {
	var err error

	server = Server{}
	server.environment = Development
	server.dbConnStr = "dbname=palakit_test"
	server.disableAuth = true

	err = server.connectToDatabase()
	if err != nil {
		log.Fatalln(err)
	}

	err = server.MigratePalakit()
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

func expectValues(rows *sql.Rows, expectedValues []string) error {
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

func TestFindWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	auth := authenticationResult{
		Role:      "anonymous",
		Variables: map[string]interface{}{},
	}
	schema := "find_test"
	table := "test"
	filters := FilterMap{}

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
	auth := authenticationResult{
		Role:      "anonymous",
		Variables: map[string]interface{}{},
	}
	schema := "find_test"
	table := "test"
	filters := FilterMap{}

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
	auth := authenticationResult{
		Role:      "anonymous",
		Variables: map[string]interface{}{},
	}
	schema := "find_test"
	table := "test"
	filters := FilterMap{
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
