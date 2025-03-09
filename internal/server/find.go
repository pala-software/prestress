package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/lib/pq"
)

type FilterMap map[string]string

func (server Server) Find(
	ctx context.Context,
	auth authenticationResult,
	schema string,
	table string,
	filters FilterMap,
) (*sql.Rows, error) {
	var err error

	tx, err := server.Begin(ctx, auth, schema)
	if err != nil {
		return nil, err
	}

	var where []string
	for column, value := range filters {
		where = append(
			where,
			fmt.Sprintf(
				"%s = %s",
				pq.QuoteIdentifier(column),
				pq.QuoteLiteral(value),
			),
		)
	}

	whereStr := ""
	if len(where) > 0 {
		whereStr = "WHERE " + strings.Join(where, " AND ")
	}

	rows, err := tx.QueryContext(
		ctx,
		fmt.Sprintf(
			"SELECT * FROM %s.%s %s",
			pq.QuoteIdentifier(schema),
			pq.QuoteIdentifier(table),
			whereStr,
		),
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

	auth := server.authenticate(writer, request)
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
		// TODO: Handle error better
		fmt.Println(err)
		writer.WriteHeader(500)
		return
	}

	defer rows.Close()

	columns, err := rows.ColumnTypes()
	if err != nil {
		// TODO: Handle error better
		fmt.Println(err)
		writer.WriteHeader(500)
		return
	}

	values := make([]interface{}, len(columns))
	for index, column := range columns {
		values[index] = reflect.New(column.ScanType())
	}

	scanArgs := make([]interface{}, len(columns))
	for index := range columns {
		scanArgs[index] = &values[index]
	}

	result := make(map[string]interface{}, len(columns))
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			fmt.Println(err)
			return
		}

		for index, column := range columns {
			result[column.Name()] = values[index]
		}
		encodedRow, err := json.Marshal(result)
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
