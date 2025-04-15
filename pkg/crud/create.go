package crud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

const CreateOperation = "Create"

func (feature Crud) Create(
	ctx context.Context,
	auth prestress.AuthenticationResult,
	schema string,
	table string,
	data map[string]any,
) error {
	var err error

	err = feature.server.Emit(BeforeBeginOperationEvent{
		OperationName: CreateOperation,
		Auth:          &auth,
		Schema:        &schema,
		Table:         &table,
		Context:       ctx,
	})
	if err != nil {
		return err
	}

	tx, err := feature.server.Begin(ctx, auth, schema)
	if err != nil {
		return err
	}

	err = feature.server.Emit(AfterBeginOperationEvent{
		OperationName: CreateOperation,
		Auth:          auth,
		Schema:        schema,
		Table:         table,
		Transaction:   tx,
		Context:       ctx,
	})
	if err != nil {
		return err
	}

	if len(data) == 0 {
		_, err = tx.Exec(
			ctx,
			fmt.Sprintf(
				"INSERT INTO %s DEFAULT VALUES",
				pgx.Identifier{schema, table}.Sanitize(),
			),
		)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
	} else {
		columns := make([]string, 0, len(data))
		values := make([]any, 0, len(data))
		placeholders := make([]string, 0, len(data))
		n := 1
		for key, value := range data {
			columns = append(columns, pgx.Identifier{key}.Sanitize())
			values = append(values, value)
			placeholders = append(placeholders, "$"+strconv.Itoa(n))
			n++
		}

		_, err = tx.Exec(
			ctx,
			fmt.Sprintf(
				"INSERT INTO %s (%s) VALUES (%s)",
				pgx.Identifier{schema, table}.Sanitize(),
				strings.Join(columns, ", "),
				strings.Join(placeholders, ", "),
			),
			values...,
		)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
	}

	err = feature.server.Emit(BeforeCommitOperationEvent{
		OperationName: CreateOperation,
		Auth:          auth,
		Schema:        schema,
		Table:         table,
		Transaction:   tx,
		Context:       ctx,
	})
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	err = feature.server.Emit(AfterCommitOperationEvent{
		OperationName: CreateOperation,
		Auth:          auth,
		Schema:        schema,
		Table:         table,
		Context:       ctx,
	})
	if err != nil {
		return err
	}

	return nil
}

func (feature Crud) handleCreate(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var err error

	schema := request.PathValue("schema")
	table := request.PathValue("table")

	if request.Body == nil {
		writer.WriteHeader(400)
		return
	}

	msg, err := io.ReadAll(request.Body)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(500)
		return
	}

	var data map[string]any
	err = json.Unmarshal(msg, &data)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(400)
		return
	}

	auth := feature.server.Authenticate(writer, request)
	if auth == nil {
		return
	}

	err = feature.Create(request.Context(), *auth, schema, table, data)
	if err != nil {
		prestress.HandleDatabaseError(writer, err)
		return
	}

	writer.WriteHeader(204)
}
