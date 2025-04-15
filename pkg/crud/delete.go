package crud

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

const DeleteOperation = "Delete"

func (feature Crud) Delete(
	ctx context.Context,
	auth prestress.AuthenticationResult,
	schema string,
	table string,
	where Where,
) error {
	var err error

	err = feature.server.Emit(BeforeBeginOperationEvent{
		OperationName: DeleteOperation,
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
		OperationName: DeleteOperation,
		Auth:          auth,
		Schema:        schema,
		Table:         table,
		Transaction:   tx,
		Context:       ctx,
	})
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
		tx.Rollback(ctx)
		return err
	}

	err = feature.server.Emit(BeforeCommitOperationEvent{
		OperationName: DeleteOperation,
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
		OperationName: DeleteOperation,
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

func (feature Crud) handleDelete(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var err error

	schema := request.PathValue("schema")
	table := request.PathValue("table")
	query := request.URL.Query()

	where := ParseWhere(query)

	auth := feature.server.Authenticate(writer, request)
	if auth == nil {
		return
	}

	err = feature.Delete(
		request.Context(),
		*auth,
		schema,
		table,
		where,
	)
	if err != nil {
		prestress.HandleDatabaseError(writer, err)
		return
	}

	writer.WriteHeader(204)
}
