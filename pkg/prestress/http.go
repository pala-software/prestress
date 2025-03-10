package prestress

import (
	"net/http"
)

// TODO: Test
func (server *Server) StartHttpServer() error {
	var err error

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{schema}/{table}", server.handleFind)
	mux.HandleFunc("POST /{schema}/{table}", server.handleCreate)
	mux.HandleFunc("PATCH /{schema}/{table}", server.handleUpdate)
	mux.HandleFunc("DELETE /{schema}/{table}", server.handleDelete)
	mux.HandleFunc("GET /{schema}/{table}/subscription", server.handleSubscription)

	server.HTTP = &http.Server{
		Addr:    ":8080", // TODO: Allow configuring
		Handler: mux,
	}
	err = server.HTTP.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}
