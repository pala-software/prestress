package prestress

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// TODO: Test
func (server Server) Delete(
	ctx context.Context,
	auth AuthenticationResult,
	schema string,
	table string,
	where Where,
) error {
	var err error

	tx, err := server.Begin(ctx, auth, schema)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		fmt.Sprintf(
			"DELETE FROM %s AS t %s",
			pgx.Identifier{schema, table}.Sanitize(),
			where.String("t", 1),
		),
		where.Values()...,
	)
	if err != nil {
		return err
	}

	err = server.collectChanges(ctx, tx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

// TODO: Test
func (server Server) handleDelete(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var err error

	schema := request.PathValue("schema")
	table := request.PathValue("table")
	query := request.URL.Query()

	where := make(Where, len(query))
	for key, values := range query {
		if len(values) == 0 {
			continue
		}

		where[key] = values[0]
	}

	auth := server.Authenticate(writer, request)
	if auth == nil {
		return
	}

	err = server.Delete(
		request.Context(),
		*auth,
		schema,
		table,
		where,
	)
	if err != nil {
		handleOperationError(writer, err)
		return
	}

	writer.WriteHeader(200)
}
