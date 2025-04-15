package subscribe

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.com/pala-software/prestress/pkg/prestress"
)

func (feature Subscribe) handleSubscription(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var err error

	schema := request.PathValue("schema")
	table := request.PathValue("table")

	auth := feature.server.Authenticate(writer, request)
	if auth == nil {
		return
	}

	subscription, err := feature.Subscribe(
		request.Context(),
		*auth,
		schema,
		table,
	)
	if err != nil {
		prestress.HandleDatabaseError(writer, err)
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

func (feature Subscribe) handleSubscriptionOptions(
	writer http.ResponseWriter,
	request *http.Request,
) {
	writer.Header().Set("Allow", "OPTIONS, GET")
	writer.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	writer.WriteHeader(204)
}
