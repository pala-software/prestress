package prestress

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

type FindResult struct {
	Rows pgx.Rows
	Done func() error
}

func (server Server) Find(
	ctx context.Context,
	auth AuthenticationResult,
	schema string,
	table string,
	where Where,
	limit int,
	offset int,
) (*FindResult, error) {
	var err error

	tx, err := server.Begin(ctx, auth, schema)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(
		ctx,
		fmt.Sprintf(
			"SELECT * FROM %s AS t %s LIMIT %d OFFSET %d",
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
		Done: func() error {
			rows.Close()
			return tx.Commit(ctx)
		},
	}, nil
}

// TODO: Test
func (server Server) handleFind(
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

	auth := server.Authenticate(writer, request)
	if auth == nil {
		return
	}

	result, err := server.Find(
		request.Context(),
		*auth,
		schema,
		table,
		where,
		limit,
		offset,
	)
	if err != nil {
		handleOperationError(writer, err)
		return
	}

	writer.WriteHeader(200)
	columns := result.Rows.FieldDescriptions()

	row := make(map[string]any, len(columns))
	defer result.Done()
	for result.Rows.Next() {
		values, err := result.Rows.Values()
		if err != nil {
			fmt.Println(err)
			return
		}

		for index, column := range columns {
			row[column.Name] = values[index]
		}

		encodedRow, err := json.Marshal(row)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Add newline at the end of JSON object
		encodedRow = append(encodedRow, 0x0A)

		_, err = writer.Write(encodedRow)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
