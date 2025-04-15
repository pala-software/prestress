package subscribe

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

func (feature Subscribe) Subscribe(
	ctx context.Context,
	auth prestress.AuthenticationResult,
	schema string,
	table string,
) (*Subscription, error) {
	var err error

	tx, err := feature.server.Begin(ctx, auth, schema)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		ctx,
		fmt.Sprintf(
			"SELECT 1 FROM %s",
			pgx.Identifier{schema, table}.Sanitize(),
		),
	)
	if err != nil {
		return nil, err
	}

	err = tx.Rollback(ctx)
	if err != nil {
		return nil, err
	}

	baseCtx := context.WithoutCancel(ctx)
	subCtx, cancel := context.WithCancel(baseCtx)

	conn, err := feature.server.DB.Acquire(subCtx)
	if err != nil {
		cancel()
		return nil, err
	}

	variables := map[string]any{}
	if auth.Variables != nil {
		variables = auth.Variables
	}

	encodedVariables, err := json.Marshal(variables)
	if err != nil {
		conn.Release()
		cancel()
		return nil, err
	}

	var subId int
	err = conn.QueryRow(
		subCtx,
		"SELECT prestress.setup_subscription($1, $2, $3, $4)",
		auth.Role,
		schema,
		table,
		encodedVariables,
	).Scan(&subId)
	if err != nil {
		conn.Release()
		cancel()
		return nil, err
	}

	subscription := &Subscription{
		Change: make(chan Change),
	}
	feature.subscriptions[subId] = subscription

	context.AfterFunc(ctx, func() {
		_, err := conn.Exec(
			subCtx,
			"SELECT prestress.teardown_subscription($1)",
			subId,
		)
		if err != nil {
			fmt.Println(err)
		}
		delete(feature.subscriptions, subId)
		conn.Release()
		cancel()
	})

	return subscription, nil
}

func (feature Subscribe) collectChanges(ctx context.Context, tx pgx.Tx) error {
	var err error

	rows, err := tx.Query(
		ctx,
		`SELECT
			subscription_id,
			row_key,
			row_data,
			row_operation
		FROM
			pg_temp.prestress_change`,
	)
	if err != nil {
		return err
	}

	defer rows.Close()
	var subId int
	var change Change
	for rows.Next() {
		err = rows.Scan(
			&subId,
			&change.RowKey,
			&change.RowData,
			&change.RowOperation,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		subscription, exists := feature.subscriptions[subId]
		if !exists {
			fmt.Println("change collected for subscription that does not exist")
			continue
		}

		subscription.Change <- change
	}

	return nil
}
