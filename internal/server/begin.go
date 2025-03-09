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
			pq.QuoteLiteral(auth.RoleName),
		),
	)
	if err != nil {
		return tx, err
	}

	_, err = tx.ExecContext(
		ctx,
		`CREATE TEMPORARY TABLE pg_temp.variable (name TEXT PRIMARY KEY, value TEXT)
		ON COMMIT DROP`,
	)
	if err != nil {
		return tx, err
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO pg_temp.variable (name, value) VALUES ('uid', $1)",
		auth.UserId,
	)
	if err != nil {
		return tx, err
	}

	return tx, nil
}
