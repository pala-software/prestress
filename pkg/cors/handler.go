package cors

import (
	"net/http"
)

type corsHandler struct {
	cors    Cors
	handler http.Handler
}

func (handler corsHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if handler.cors.AllowedOrigins != "" {
		writer.Header().Set("Access-Control-Allow-Origin", handler.cors.AllowedOrigins)
	}
	handler.handler.ServeHTTP(writer, request)
}
