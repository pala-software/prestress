package prestress

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var ErrForbiddenSchema = errors.New("forbidden schema")

// TODO: Test
func (server Server) Begin(
	ctx context.Context,
	auth AuthenticationResult,
	schema string,
) (pgx.Tx, error) {
	var err error

	if schema == "pg_temp" {
		return nil, ErrForbiddenSchema
	}

	tx, err := server.DB.Begin(ctx)
	if err != nil {
		return tx, err
	}

	_, err = tx.Exec(
		ctx,
		fmt.Sprintf(
			// pg_temp is set to last in search_path so that we don't accidentally or
			// in any case query temporary tables implicitly.
			"SET LOCAL search_path TO %s, pg_temp",
			pgx.Identifier{schema}.Sanitize(),
		),
	)
	if err != nil {
		return tx, err
	}

	_, err = tx.Exec(
		ctx,
		fmt.Sprintf(
			"SET LOCAL role TO %s",
			pgx.Identifier{auth.Role}.Sanitize(),
		),
	)
	if err != nil {
		return tx, err
	}

	encodedVariables, err := json.Marshal(auth.Variables)
	if err != nil {
		return tx, err
	}

	_, err = tx.Exec(
		ctx,
		`SELECT prestress.begin_authorized($1)`,
		encodedVariables,
	)
	if err != nil {
		return tx, err
	}

	return tx, nil
}
