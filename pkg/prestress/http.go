package prestress

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
)

type Middleware func(http.Handler) http.Handler

func (server *Server) HTTP() *http.ServeMux {
	if server.serveMux == nil {
		server.serveMux = http.NewServeMux()
	}
	return server.serveMux
}

func (server *Server) AddMiddleware(middleware Middleware) {
	server.middleware = append(server.middleware, middleware)
}

func (server *Server) StartHttpServer() (err error) {
	// Apply middleware
	handler := http.Handler(server.HTTP())
	for _, middleware := range server.middleware {
		handler = middleware(handler)
	}

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
		eventErr := server.Emit(ServerShutdownEvent{Context: context.Background()})
		err = errors.Join(err, eventErr)
	}()

	// Emit start event
	err = server.Emit(ServerStartEvent{Context: ctx})
	if err != nil {
		return
	}

	// Serve HTTP
	srv := &http.Server{
		Handler: handler,
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
