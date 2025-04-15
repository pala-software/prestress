package prestress

import (
	"net/http"
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

func (server *Server) StartHttpServer() error {
	var err error

	// Apply middleware
	handler := http.Handler(server.HTTP())
	for _, middleware := range server.middleware {
		handler = middleware(handler)
	}

	// Start server
	srv := &http.Server{
		Addr:    ":8080", // TODO: Allow configuring
		Handler: handler,
	}
	err = srv.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}
