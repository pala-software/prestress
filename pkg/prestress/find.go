package prestress

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func (server Server) Find(
	ctx context.Context,
	auth AuthenticationResult,
	schema string,
	table string,
	where Where,
) (pgx.Rows, error) {
	var err error

	tx, err := server.Begin(ctx, auth, schema)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(
		ctx,
		fmt.Sprintf(
			"SELECT * FROM %s AS t %s",
			pgx.Identifier{schema, table}.Sanitize(),
			where.String("t", 1),
		),
		where.Values()...,
	)
	if err != nil {
		return rows, err
	}

	return rows, nil
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

	rows, err := server.Find(
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
	columns := rows.FieldDescriptions()

	row := make(map[string]any, len(columns))
	defer rows.Close()
	for rows.Next() {
		values, err := rows.Values()
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
