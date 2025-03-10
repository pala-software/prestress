package prestress

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
)

type Change struct {
	RowOperation string          `json:"op"`
	RowKey       json.RawMessage `json:"key"`
	RowData      json.RawMessage `json:"data"`
}

type Subscription struct {
	Change chan Change
}

// TODO: Test
func (server Server) Subscribe(
	ctx context.Context,
	auth AuthenticationResult,
	schema string,
	table string,
) (*Subscription, error) {
	var err error

	tx, err := server.Begin(ctx, auth, schema)
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

	conn, err := server.DB.Acquire(subCtx)
	if err != nil {
		cancel()
		return nil, err
	}

	encodedVariables, err := json.Marshal(auth.Variables)
	if err != nil {
		conn.Release()
		cancel()
		return nil, err
	}

	var subId int
	err = server.DB.QueryRow(
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
	server.subscriptions[subId] = subscription

	context.AfterFunc(ctx, func() {
		_, err := server.DB.Exec(
			subCtx,
			"SELECT prestress.teardown_subscription($1)",
			subId,
		)
		if err != nil {
			fmt.Println(err)
		}
		delete(server.subscriptions, subId)
		conn.Release()
		cancel()
	})

	return subscription, nil
}

// TODO: Test
func (server Server) handleSubscription(
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

	subscription, err := server.Subscribe(
		request.Context(),
		*auth,
		schema,
		table,
	)
	if err != nil {
		handleOperationError(writer, err)
		return
	}

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	responseController := http.NewResponseController(writer)

	for {
		select {
		case <-request.Context().Done():
			return

		case change := <-subscription.Change:
			encodedChange, err := json.Marshal(change)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Prepend message with a "data: " prefix
			encodedChange = append([]byte("event: change\ndata: "), encodedChange...)

			// End the message with two newline characters
			encodedChange = append(encodedChange, 0x0A, 0x0A)

			_, err = writer.Write(encodedChange)
			if err != nil {
				fmt.Println(err)
				return
			}

			err = responseController.Flush()
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// TODO: Test
func (server Server) collectChanges(ctx context.Context, tx pgx.Tx) error {
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

		subscription, exists := server.subscriptions[subId]
		if !exists {
			fmt.Println("change collected for subscription that does not exist")
			continue
		}

		subscription.Change <- change
	}

	return nil
}
