package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Change struct {
	SubscriptionId int             `json:"subscription_id"`
	RowKey         json.RawMessage `json:"row_key"`
	RowData        json.RawMessage `json:"row_data"`
	RowOperation   string          `json:"row_operation"`
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
		panic(err)
	}

	for {
		notification := <-listener.Notify
		if notification == nil {
			continue
		}

		rows, err := server.DB.Query(`
			SELECT
        subscription_id,
        row_key AS key,
        row_data AS data,
        row_operation AS op
      FROM watcher.change
		`)
		if err != nil {
			fmt.Println(err)
			continue
		}

		change := Change{}
		for rows.Next() {
			err = rows.Scan(
				&change.SubscriptionId,
				&change.RowKey,
				&change.RowData,
				&change.RowOperation,
			)
			if err != nil {
				fmt.Println(err)
				continue
			}

			subscription, exists := server.subscriptions[change.SubscriptionId]
			if !exists {
				continue
			}

			subscription.Change <- change
		}

		_, err = server.DB.Exec("TRUNCATE watcher.change")
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}
