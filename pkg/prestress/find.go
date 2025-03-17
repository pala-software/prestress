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
	Done func()
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
		Done: func() {
			rows.Close()
			err := tx.Commit(ctx)
			if err != nil {
				fmt.Println(err)
			}
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

	first := true
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
		handleOperationError(writer, err)
		return
	}

	if first {
		writer.WriteHeader(200)
		writer.Write([]byte("[]"))
	}
}
