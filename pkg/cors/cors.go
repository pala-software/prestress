package cors

import (
	"net/http"
	"os"

	"gitlab.com/pala-software/prestress/pkg/prestress"
)

type Cors struct {
	AllowedOrigins string
}

// Construct CORS Feature and read configuration from environment variables.
func CorsFromEnv() *Cors {
	feature := Cors{}
	feature.AllowedOrigins = os.Getenv("PRESTRESS_ALLOWED_ORIGINS")
	return &feature
}

func (cors Cors) Apply(server *prestress.Server) error {
	server.AddMiddleware(cors.applyCors)
	return nil
}

func (cors Cors) applyCors(handler http.Handler) http.Handler {
	return corsHandler{cors, handler}
}
