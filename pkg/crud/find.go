package crud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

const FindOperation = "Find"

type FindResult struct {
	Rows pgx.Rows
	Done func()
}

func (feature Crud) Find(
	ctx context.Context,
	auth prestress.AuthenticationResult,
	schema string,
	table string,
	where Where,
	limit int,
	offset int,
) (*FindResult, error) {
	var err error

	err = feature.server.Emit(BeforeBeginOperationEvent{
		OperationName: FindOperation,
		Auth:          &auth,
		Schema:        &schema,
		Table:         &table,
		Context:       ctx,
	})
	if err != nil {
		return nil, err
	}

	tx, err := feature.server.Begin(ctx, auth, schema)
	if err != nil {
		return nil, err
	}

	err = feature.server.Emit(AfterBeginOperationEvent{
		OperationName: FindOperation,
		Auth:          auth,
		Schema:        schema,
		Table:         table,
		Transaction:   tx,
		Context:       ctx,
	})
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(
		ctx,
		fmt.Sprintf(
			"SELECT to_json(t) FROM %s AS t %s LIMIT %d OFFSET %d",
			pgx.Identifier{schema, table}.Sanitize(),
			where.String("t", 1),
			limit,
			offset,
		),
		where.Values()...,
	)
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	return &FindResult{
		Rows: rows,
		Done: func() {
			rows.Close()

			err = feature.server.Emit(BeforeCommitOperationEvent{
				OperationName: FindOperation,
				Auth:          auth,
				Schema:        schema,
				Table:         table,
				Transaction:   tx,
				Context:       ctx,
			})
			if err != nil {
				fmt.Println(err)
				tx.Rollback(ctx)
				return
			}

			err := tx.Commit(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}

			err = feature.server.Emit(AfterCommitOperationEvent{
				OperationName: FindOperation,
				Auth:          auth,
				Schema:        schema,
				Table:         table,
				Context:       ctx,
			})
			if err != nil {
				fmt.Println(err)
				return
			}
		},
	}, nil
}

func (feature Crud) handleFind(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var err error

	schema := request.PathValue("schema")
	table := request.PathValue("table")
	query := request.URL.Query()

	where := ParseWhere(query)

	limit := 100
	if query.Has("limit") {
		limit, err = strconv.Atoi(query.Get("limit"))
		if err != nil {
			writer.WriteHeader(400)
			return
		}
	}

	offset := 0
	if query.Has("offset") {
		offset, err = strconv.Atoi(query.Get("offset"))
		if err != nil {
			writer.WriteHeader(400)
			return
		}
	}

	auth := feature.server.Authenticate(writer, request)
	if auth == nil {
		return
	}

	result, err := feature.Find(
		request.Context(),
		*auth,
		schema,
		table,
		where,
		limit,
		offset,
	)
	if err != nil {
		prestress.HandleDatabaseError(writer, err)
		return
	}

	first := true
	row := json.RawMessage{}
	defer result.Done()
	for result.Rows.Next() {
		err := result.Rows.Scan(&row)
		if err != nil {
			fmt.Println(err)
			return
		}

		encodedRow, err := json.Marshal(row)
		if err != nil {
			fmt.Println(err)
			return
		}

		if first {
			first = false
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(200)
			writer.Write([]byte("["))
			defer writer.Write([]byte("]"))
		} else {
			writer.Write([]byte(","))
		}

		_, err = writer.Write(encodedRow)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	err = result.Rows.Err()
	if err != nil {
		prestress.HandleDatabaseError(writer, err)
		return
	}

	if first {
		writer.WriteHeader(200)
		writer.Write([]byte("[]"))
	}
}
