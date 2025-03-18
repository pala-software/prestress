package prestress

import (
	"net/http"
)

// TODO: Test
func (server *Server) StartHttpServer() error {
	var err error

	mux := http.NewServeMux()

	mux.HandleFunc(
		"OPTIONS /{schema}/{table}",
		server.handleTableOptions,
	)
	mux.HandleFunc(
		"OPTIONS /{schema}/{table}/subscription",
		server.handleSubscriptionOptions,
	)

	mux.HandleFunc("GET /{schema}/{table}", server.handleFind)
	mux.HandleFunc("POST /{schema}/{table}", server.handleCreate)
	mux.HandleFunc("PATCH /{schema}/{table}", server.handleUpdate)
	mux.HandleFunc("DELETE /{schema}/{table}", server.handleDelete)
	mux.HandleFunc("GET /{schema}/{table}/subscription", server.handleSubscription)

	server.HTTP = &http.Server{
		Addr:    ":8080", // TODO: Allow configuring
		Handler: server.handleCors(mux),
	}
	err = server.HTTP.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func (server Server) handleTableOptions(
	writer http.ResponseWriter,
	request *http.Request,
) {
	writer.Header().Set("Allow", "OPTIONS, GET, POST, PATCH, DELETE")
	writer.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PATCH, DELETE")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	writer.WriteHeader(204)
}

func (server Server) handleSubscriptionOptions(
	writer http.ResponseWriter,
	request *http.Request,
) {
	writer.Header().Set("Allow", "OPTIONS, GET")
	writer.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	writer.WriteHeader(204)
}
