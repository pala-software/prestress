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

const UpdateOperation = "Update"

func (feature Crud) Update(
	ctx context.Context,
	auth prestress.AuthenticationResult,
	schema string,
	table string,
	where Where,
	data map[string]any,
) error {
	var err error

	if len(data) == 0 {
		return nil
	}

	err = feature.server.Emit(BeforeBeginOperationEvent{
		OperationName: UpdateOperation,
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
		OperationName: UpdateOperation,
		Auth:          auth,
		Schema:        schema,
		Table:         table,
		Transaction:   tx,
		Context:       ctx,
	})
	if err != nil {
		return err
	}

	patch := make([]string, 0, len(data))
	n := 1
	for column := range data {
		patch = append(
			patch,
			fmt.Sprintf(
				"%s = %s",
				pgx.Identifier{column}.Sanitize(),
				"$"+strconv.Itoa(n),
			),
		)
		n++
	}

	values := make([]any, 0, len(patch)+len(where))
	for _, value := range data {
		values = append(values, value)
	}
	values = append(values, where.Values()...)

	_, err = tx.Exec(
		ctx,
		fmt.Sprintf(
			"UPDATE %s AS t SET %s %s",
			pgx.Identifier{schema, table}.Sanitize(),
			strings.Join(patch, ", "),
			where.String("t", len(patch)+1),
		),
		values...,
	)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	err = feature.server.Emit(BeforeCommitOperationEvent{
		OperationName: UpdateOperation,
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
		OperationName: UpdateOperation,
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

func (feature Crud) handleUpdate(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var err error

	schema := request.PathValue("schema")
	table := request.PathValue("table")
	query := request.URL.Query()

	where := ParseWhere(query)

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

	err = feature.Update(
		request.Context(),
		*auth,
		schema,
		table,
		where,
		data,
	)
	if err != nil {
		prestress.HandleDatabaseError(writer, err)
		return
	}

	writer.WriteHeader(204)
}
