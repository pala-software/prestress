package crud

import (
	"net/http"

	"gitlab.com/pala-software/prestress/pkg/prestress"
)

type Crud struct {
	server *prestress.Server
}

// Construct CRUD Feature and read configuration from environment variables.
func CrudFromEnv() *Crud {
	feature := Crud{}
	// No configuration at this time
	return &feature
}

func (feature *Crud) Apply(server *prestress.Server) error {
	feature.server = server
	mux := feature.server.HTTP()
	mux.HandleFunc("OPTIONS /{schema}/{table}", feature.handleTableOptions)
	mux.HandleFunc("GET /{schema}/{table}", feature.handleFind)
	mux.HandleFunc("POST /{schema}/{table}", feature.handleCreate)
	mux.HandleFunc("PATCH /{schema}/{table}", feature.handleUpdate)
	mux.HandleFunc("DELETE /{schema}/{table}", feature.handleDelete)
	return nil
}

func (feature Crud) handleTableOptions(
	writer http.ResponseWriter,
	request *http.Request,
) {
	writer.Header().Set("Allow", "OPTIONS, GET, POST, PATCH, DELETE")
	writer.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PATCH, DELETE")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	writer.WriteHeader(204)
}
