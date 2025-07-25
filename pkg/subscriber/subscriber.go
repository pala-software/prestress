package subscriber

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"gitlab.com/pala-software/prestress/pkg/crud"
	"gitlab.com/pala-software/prestress/pkg/migrator"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Subscriber struct {
	// If root path is adjusted for CRUD feature, this must be set to match that
	// path.
	CrudRootPath string
}

// Construct Subscribe Feature and read configuration from environment
// variables.
func SubscriberFromEnv() *Subscriber {
	feature := &Subscriber{}
	// No configuration at this time
	return feature
}

func (feature Subscriber) Provider() any {
	return feature.Register
}

func (feature *Subscriber) Register(
	begin *prestress.BeginOperation,
	create *crud.CreateOperation,
	update *crud.UpdateOperation,
	delete *crud.DeleteOperation,
	core *prestress.Core,
	mux *http.ServeMux,
	mig *migrator.Migrator,
) (self *Subscriber, subscribe *SubscribeOperation, err error) {
	self = feature
	subscribe = NewSubscribeOperation(begin, create, update, delete)
	core.Operations().Register(subscribe)

	err = feature.RegisterRoutes(mux, subscribe)
	if err != nil {
		return
	}

	err = feature.RegisterMigrations(mig)
	if err != nil {
		return
	}

	return
}

func (feature Subscriber) RegisterRoutes(
	mux *http.ServeMux,
	subscribe *SubscribeOperation,
) (err error) {
	rootPath := "/"
	if feature.CrudRootPath != "" {
		rootPath = feature.CrudRootPath
	}
	if rootPath == "/" {
		rootPath = ""
	}

	mux.HandleFunc(
		fmt.Sprintf(
			"OPTIONS %s/{schema}/{table}/subscription",
			feature.CrudRootPath,
		),
		handleSubscriptionOptions,
	)
	mux.Handle(
		fmt.Sprintf(
			"GET %s/{schema}/{table}/subscription",
			feature.CrudRootPath,
		),
		subscribe,
	)

	return
}

func (Subscriber) RegisterMigrations(mig *migrator.Migrator) (err error) {
	dir, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return
	}

	mig.Targets.Register(migrator.MigrationTarget{
		Name:      "subscribe",
		Directory: dir,
	})
	return
}

func handleSubscriptionOptions(
	writer http.ResponseWriter,
	request *http.Request,
) {
	writer.Header().Set("Allow", "OPTIONS, GET")
	writer.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	writer.WriteHeader(204)
}
