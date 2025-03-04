package server

import (
	"net/http"
)

func (server *Server) startHttpServer() error {
	var err error

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{schema}/{table}", server.handleFind)
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
