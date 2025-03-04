package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Subscription struct {
	Id     int
	Change chan Change
}

func (server Server) Subscribe(
	auth authenticationResult,
	schema string,
	table string,
) (*Subscription, error) {
	var err error

	var subId int
	err = server.DB.QueryRow(
		`INSERT INTO watcher.subscription (role_name, table_schema, table_name)
		VALUES ($1, $2, $3)
		RETURNING id`,
		auth.RoleName,
		schema,
		table,
	).Scan(&subId)
	if err != nil {
		return nil, err
	}

	subscription := &Subscription{
		Id:     subId,
		Change: make(chan Change),
	}
	server.subscriptions[subId] = subscription
	return subscription, nil
}

func (server Server) Unsubscribe(subscription *Subscription) error {
	_, err := server.DB.Exec(
		"DELETE FROM watcher.subscription WHERE id = $1",
		subscription.Id,
	)
	return err
}

func (server Server) handleSubscription(writer http.ResponseWriter, request *http.Request) {
	var err error

	schema := request.PathValue("schema")
	table := request.PathValue("table")

	auth := server.authenticate(writer, request)
	if auth == nil {
		return
	}

	subscription, err := server.Subscribe(*auth, schema, table)
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

	defer server.Unsubscribe(subscription)

	for {
		select {
		case <-request.Context().Done():
			err := server.Unsubscribe(subscription)
			if err != nil {
				fmt.Println(err)
			}
			return

		case change := <-subscription.Change:
			encodedChange, err := json.Marshal(change)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Add newline at the end of JSON object
			encodedChange = append(encodedChange, 0x0A)

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
