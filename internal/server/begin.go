package server

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

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

	_, err = tx.ExecContext(
		ctx,
		`CREATE TEMPORARY TABLE pg_temp.authorization_variable
			(name TEXT PRIMARY KEY, value TEXT)
		ON COMMIT DROP`,
	)
	if err != nil {
		return tx, err
	}

	stmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO pg_temp.authorization_variable
			(name, value)
		VALUES 
			($1, $2)`,
	)
	if err != nil {
		return tx, err
	}

	for name, value := range auth.Token {
		_, err = stmt.ExecContext(ctx, name, value)
		if err != nil {
			return tx, err
		}
	}

	return tx, nil
}
