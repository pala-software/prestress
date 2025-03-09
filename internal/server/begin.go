package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
)

// TODO: Test
func (server Server) Begin(
	ctx context.Context,
	auth authenticationResult,
	schema string,
) (*sql.Tx, error) {
	var err error

	tx, err := server.DB.BeginTx(ctx, nil)
	if err != nil {
		return tx, err
	}

	_, err = tx.ExecContext(
		ctx,
		fmt.Sprintf(
			// pg_temp is set to last in search_path so that we don't accidentally or
			// in any case query temporary tables implicitly.
			"SET LOCAL search_path TO %s, 'pg_temp'",
			pq.QuoteLiteral(schema),
		),
	)
	if err != nil {
		return tx, err
	}

	_, err = tx.ExecContext(
		ctx,
		fmt.Sprintf(
			"SET LOCAL role TO %s",
			pq.QuoteLiteral(auth.Role),
		),
	)
	if err != nil {
		return tx, err
	}

	encodedVariables, err := json.Marshal(auth.Variables)
	if err != nil {
		return tx, err
	}

	_, err = tx.ExecContext(
		ctx,
		`SELECT palakit.begin_authorized($1)`,
		encodedVariables,
	)
	if err != nil {
		return tx, err
	}

	return tx, nil
}
