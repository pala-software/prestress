package prestress

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

// TODO: Test
func (server Server) Create(
	ctx context.Context,
	auth AuthenticationResult,
	schema string,
	table string,
	data map[string]interface{},
) error {
	var err error

	tx, err := server.Begin(ctx, auth, schema)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		_, err = tx.ExecContext(
			ctx,
			fmt.Sprintf(
				"INSERT INTO %s.%s DEFAULT VALUES",
				pq.QuoteIdentifier(schema),
				pq.QuoteIdentifier(table),
			),
		)
		if err != nil {
			return err
		}
	} else {
		columns := make([]string, 0, len(data))
		values := make([]interface{}, 0, len(data))
		placeholders := make([]string, 0, len(data))
		n := 1
		for key, value := range data {
			columns = append(columns, pq.QuoteIdentifier(key))
			values = append(values, value)
			placeholders = append(placeholders, "$"+strconv.Itoa(n))
			n++
		}

		_, err = tx.ExecContext(
			ctx,
			fmt.Sprintf(
				"INSERT INTO %s.%s (%s) VALUES (%s)",
				pq.QuoteIdentifier(schema),
				pq.QuoteIdentifier(table),
				strings.Join(columns, ", "),
				strings.Join(placeholders, ", "),
			),
			values...,
		)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// TODO: Test
func (server Server) handleCreate(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var err error

	schema := request.PathValue("schema")
	table := request.PathValue("table")

	auth := server.Authenticate(writer, request)
	if auth == nil {
		return
	}

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

	var data map[string]interface{}
	err = json.Unmarshal(msg, &data)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(400)
		return
	}

	err = server.Create(request.Context(), *auth, schema, table, data)
	if err != nil {
		handleOperationError(writer, err)
		return
	}

	writer.WriteHeader(201)
}
