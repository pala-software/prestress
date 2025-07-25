package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"

	"gitlab.com/pala-software/prestress/pkg/prestress"
)

type corsHandler struct {
	AllowedOrigins string
	Next           http.Handler
}

func (handler corsHandler) ServeHTTP(
	writer http.ResponseWriter,
	request *http.Request,
) {
	if handler.AllowedOrigins != "" {
		writer.Header().Set("Access-Control-Allow-Origin", handler.AllowedOrigins)
	}

	handler.Next.ServeHTTP(writer, request)
}

func startHttpServer(
	mux *http.ServeMux,
	lifecycle *prestress.Lifecycle,
) (err error) {
	// Start listening
	// TODO: Allow configuring port
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	defer func() {
		for _, hook := range lifecycle.Shutdown.Value() {
			hookErr := hook()
			err = errors.Join(err, hookErr)
		}
	}()

	// Emit start event
	for _, hook := range lifecycle.Start.Value() {
		err = hook()
		if err != nil {
			return
		}
	}

	// Serve HTTP
	srv := &http.Server{
		Handler: corsHandler{
			AllowedOrigins: os.Getenv("PRESTRESS_ALLOWED_ORIGINS"),
			Next:           mux,
		},
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.Serve(listener)
	}()

	fmt.Println("Server started on port 8080!")

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = srv.Shutdown(context.Background())
	return
}
