package prestress

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

type FilterMap map[string]string

func (server Server) Find(
	ctx context.Context,
	auth AuthenticationResult,
	schema string,
	table string,
	filters FilterMap,
) (pgx.Rows, error) {
	var err error

	tx, err := server.Begin(ctx, auth, schema)
	if err != nil {
		return nil, err
	}

	var where []string
	n := 1
	for column := range filters {
		where = append(
			where,
			fmt.Sprintf(
				"%s = %s",
				pgx.Identifier{column}.Sanitize(),
				"$"+strconv.Itoa(n),
			),
		)
		n++
	}

	whereStr := ""
	if len(where) > 0 {
		whereStr = "WHERE " + strings.Join(where, " AND ")
	}

	values := make([]any, 0, len(filters))
	for _, value := range filters {
		values = append(values, value)
	}

	rows, err := tx.Query(
		ctx,
		fmt.Sprintf(
			"SELECT * FROM %s %s",
			pgx.Identifier{schema, table}.Sanitize(),
			whereStr,
		),
		values...,
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

	filters := make(FilterMap, len(query))
	for key, values := range query {
		if len(values) == 0 {
			continue
		}

		filters[key] = values[0]
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
		filters,
	)
	if err != nil {
		handleOperationError(writer, err)
		return
	}

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
