package prestress

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// TODO: Test
func (server Server) Create(
	ctx context.Context,
	auth AuthenticationResult,
	schema string,
	table string,
	data map[string]any,
) error {
	var err error

	tx, err := server.Begin(ctx, auth, schema)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		_, err = tx.Exec(
			ctx,
			fmt.Sprintf(
				"INSERT INTO %s DEFAULT VALUES",
				pgx.Identifier{schema, table}.Sanitize(),
			),
		)
		if err != nil {
			return err
		}
	} else {
		columns := make([]string, 0, len(data))
		values := make([]any, 0, len(data))
		placeholders := make([]string, 0, len(data))
		n := 1
		for key, value := range data {
			columns = append(columns, pgx.Identifier{key}.Sanitize())
			values = append(values, value)
			placeholders = append(placeholders, "$"+strconv.Itoa(n))
			n++
		}

		_, err = tx.Exec(
			ctx,
			fmt.Sprintf(
				"INSERT INTO %s (%s) VALUES (%s)",
				pgx.Identifier{schema, table}.Sanitize(),
				strings.Join(columns, ", "),
				strings.Join(placeholders, ", "),
			),
			values...,
		)
		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
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

	var data map[string]any
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
