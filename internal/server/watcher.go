package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
)

type Change struct {
	RowOperation string          `json:"op"`
	RowKey       json.RawMessage `json:"key"`
	RowData      json.RawMessage `json:"data"`
}

func (server Server) listenForChange() {
	minReconn := 10 * time.Second
	maxReconn := 1 * time.Minute

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err)
		}
	}

	listener := pq.NewListener(server.dbConnStr, minReconn, maxReconn, reportProblem)
	err := listener.Listen("change")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		notification := <-listener.Notify
		if notification == nil {
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())

		rows, err := server.DB.QueryContext(
			ctx,
			`SELECT
        subscription_id,
        row_key AS key,
        row_data AS data,
        row_operation AS op
      FROM watcher.change`,
		)
		if err != nil {
			fmt.Println(err)
			cancel()
			continue
		}

		change := Change{}
		for rows.Next() {
			var subscriptionId int
			err = rows.Scan(
				&subscriptionId,
				&change.RowKey,
				&change.RowData,
				&change.RowOperation,
			)
			if err != nil {
				fmt.Println(err)
				continue
			}

			subscription, exists := server.subscriptions[subscriptionId]
			if !exists {
				continue
			}

			subscription.Change <- change
		}

		rows.Close()

		_, err = server.DB.ExecContext(ctx, "TRUNCATE watcher.change")
		if err != nil {
			fmt.Println(err)
			cancel()
			continue
		}
	}
}
