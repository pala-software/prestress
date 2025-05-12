package crud

import (
	"fmt"
	"net/http"

	"gitlab.com/pala-software/prestress/pkg/prestress"
)

type Crud struct {
	// Root path where CRUD resources can be accessed.
	RootPath string

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

	rootPath := "/"
	if feature.RootPath != "" {
		rootPath = feature.RootPath
	}
	if rootPath == "/" {
		rootPath = ""
	}

	mux := feature.server.HTTP()
	mux.HandleFunc(
		fmt.Sprintf("OPTIONS %s/{schema}/{table}", rootPath),
		feature.handleTableOptions,
	)
	mux.HandleFunc(
		fmt.Sprintf("GET %s/{schema}/{table}", rootPath),
		feature.handleFind,
	)
	mux.HandleFunc(
		fmt.Sprintf("POST %s/{schema}/{table}", rootPath),
		feature.handleCreate,
	)
	mux.HandleFunc(
		fmt.Sprintf("PATCH %s/{schema}/{table}", rootPath),
		feature.handleUpdate,
	)
	mux.HandleFunc(
		fmt.Sprintf("DELETE %s/{schema}/{table}", rootPath),
		feature.handleDelete,
	)

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
