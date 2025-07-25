package crud

import (
	"fmt"
	"net/http"

	"gitlab.com/pala-software/prestress/pkg/prestress"
)

type Crud struct {
	// Root path where CRUD resources can be accessed.
	RootPath string
}

// Construct CRUD Feature and read configuration from environment variables.
func CrudFromEnv() *Crud {
	feature := Crud{}
	// No configuration at this time
	return &feature
}

func (feature *Crud) Provider() any {
	return func(
		begin *prestress.BeginOperation,
	) (
		self *Crud,
		find *FindOperation,
		create *CreateOperation,
		update *UpdateOperation,
		delete *DeleteOperation,
	) {
		self = feature
		find = NewFindOperation(begin)
		create = NewCreateOperation(begin)
		update = NewUpdateOperation(begin)
		delete = NewDeleteOperation(begin)
		return
	}
}

func (feature *Crud) Invoker() any {
	return func(
		core *prestress.Core,
		mux *http.ServeMux,
		find *FindOperation,
		create *CreateOperation,
		update *UpdateOperation,
		delete *DeleteOperation,
	) (err error) {
		core.Operations().Register(find)
		core.Operations().Register(create)
		core.Operations().Register(update)
		core.Operations().Register(delete)

		err = feature.RegisterRoutes(
			mux,
			find,
			create,
			update,
			delete,
		)
		if err != nil {
			return
		}

		return
	}
}

func (feature Crud) RegisterRoutes(
	mux *http.ServeMux,
	find *FindOperation,
	create *CreateOperation,
	update *UpdateOperation,
	delete *DeleteOperation,
) (err error) {
	rootPath := "/"
	if feature.RootPath != "" {
		rootPath = feature.RootPath
	}
	if rootPath == "/" {
		rootPath = ""
	}

	mux.HandleFunc(
		fmt.Sprintf("OPTIONS %s/{schema}/{table}", rootPath),
		handleTableOptions,
	)
	mux.Handle(
		fmt.Sprintf("GET %s/{schema}/{table}", rootPath),
		find,
	)
	mux.Handle(
		fmt.Sprintf("POST %s/{schema}/{table}", rootPath),
		create,
	)
	mux.Handle(
		fmt.Sprintf("PATCH %s/{schema}/{table}", rootPath),
		update,
	)
	mux.Handle(
		fmt.Sprintf("DELETE %s/{schema}/{table}", rootPath),
		delete,
	)

	return
}

func handleTableOptions(
	writer http.ResponseWriter,
	request *http.Request,
) {
	writer.Header().Set("Allow", "OPTIONS, GET, POST, PATCH, DELETE")
	writer.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PATCH, DELETE")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	writer.WriteHeader(204)
}
