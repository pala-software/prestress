package subscribe

import (
	"embed"
	"fmt"
	"io/fs"

	"gitlab.com/pala-software/prestress/pkg/crud"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Subscribe struct {
	// If root path is adjusted for CRUD feature, this must be set to match that
	// path.
	CrudRootPath string

	server        *prestress.Server
	subscriptions map[int]*Subscription
}

// Construct Subscribe Feature and read configuration from environment
// variables.
func SubscribeFromEnv() *Subscribe {
	feature := Subscribe{}
	// No configuration at this time
	return &feature
}

func (feature Subscribe) Apply(server *prestress.Server) error {
	feature.server = server
	feature.subscriptions = map[int]*Subscription{}

	rootPath := "/"
	if feature.CrudRootPath != "" {
		rootPath = feature.CrudRootPath
	}
	if rootPath == "/" {
		rootPath = ""
	}

	http := server.HTTP()
	http.HandleFunc(
		fmt.Sprintf(
			"OPTIONS %s/{schema}/{table}/subscription",
			feature.CrudRootPath,
		),
		feature.handleSubscriptionOptions,
	)
	http.HandleFunc(
		fmt.Sprintf(
			"GET %s/{schema}/{table}/subscription",
			feature.CrudRootPath,
		),
		feature.handleSubscription,
	)

	dir, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return err
	}
	feature.server.AddMigration("subscribe", dir)

	server.OnEvent(func(event prestress.Event) error {
		switch event := event.(type) {
		case crud.AfterBeginOperationEvent:
			_, err := event.Transaction.Exec(
				event.Context,
				`CREATE TEMPORARY TABLE IF NOT EXISTS pg_temp.prestress_change
				OF prestress.change
				ON COMMIT DELETE ROWS`,
			)
			return err

		case crud.BeforeCommitOperationEvent:
			return feature.collectChanges(
				event.Context,
				event.Transaction,
			)

		default:
			return nil
		}
	})

	return nil
}
