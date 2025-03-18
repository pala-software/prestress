package prestress

import (
	"net/http"
)

func (server Server) handleCors(handler http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if server.AllowedOrigins != "" {
			writer.Header().Set("Access-Control-Allow-Origin", server.AllowedOrigins)
		}
		handler.ServeHTTP(writer, request)
	}
}
