package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Subscription struct {
	Change chan Change
}

func (server Server) Subscribe(
	ctx context.Context,
	auth authenticationResult,
	schema string,
	table string,
) (*Subscription, error) {
	var err error

	baseCtx := context.WithoutCancel(ctx)
	subCtx, cancel := context.WithCancel(baseCtx)

	conn, err := server.DB.Conn(subCtx)
	if err != nil {
		cancel()
		return nil, err
	}

	encodedVariables, err := json.Marshal(auth.Variables)
	if err != nil {
		conn.Close()
		cancel()
		return nil, err
	}

	var subId int
	err = conn.QueryRowContext(
		ctx,
		"SELECT watcher.setup_subscription($1, $2, $3, $4)",
		auth.Role,
		schema,
		table,
		encodedVariables,
	).Scan(&subId)
	if err != nil {
		conn.Close()
		cancel()
		return nil, err
	}

	subscription := &Subscription{
		Change: make(chan Change),
	}
	server.subscriptions[subId] = subscription

	context.AfterFunc(ctx, func() {
		_, err := conn.ExecContext(
			subCtx,
			"SELECT watcher.teardown_subscription($1)",
			subId,
		)
		if err != nil {
			fmt.Println(err)
		}
		delete(server.subscriptions, subId)
		conn.Close()
		cancel()
	})

	return subscription, nil
}

func (server Server) handleSubscription(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var err error

	schema := request.PathValue("schema")
	table := request.PathValue("table")

	auth := server.authenticate(writer, request)
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
		// TODO: Handle error better
		fmt.Println(err)
		writer.WriteHeader(500)
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
