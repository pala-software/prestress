package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"

	"gitlab.com/pala-software/prestress/pkg/prestress"
)

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
		Handler: mux,
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.Serve(listener)
	}()

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
