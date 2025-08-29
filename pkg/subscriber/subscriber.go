package subscriber

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
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
	feature := &Subscriber{
		CrudRootPath: "/data",
	}
	// No configuration at this time
	return feature
}

func (feature *Subscriber) Provider() any {
	return func(
		pool *pgxpool.Pool,
		begin *prestress.BeginOperation,
		create *crud.CreateOperation,
		update *crud.UpdateOperation,
		delete *crud.DeleteOperation,
		core *prestress.Core,
	) (
		self *Subscriber,
		subscribe *SubscribeOperation,
	) {
		self = feature
		subscribe = NewSubscribeOperation(pool,
			begin,
			create,
			update,
			delete,
		)
		core.Operations().Register(subscribe)
		return
	}
}

func (feature *Subscriber) Invoker() any {
	return func(
		subscribe *SubscribeOperation,
		mux *http.ServeMux,
		mig *migrator.Migrator,
	) (err error) {
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
