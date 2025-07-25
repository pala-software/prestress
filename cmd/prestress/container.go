package main

import (
	"net/http"

	"gitlab.com/pala-software/prestress/pkg/auth"
	"gitlab.com/pala-software/prestress/pkg/crud"
	"gitlab.com/pala-software/prestress/pkg/migrator"
	"gitlab.com/pala-software/prestress/pkg/oauth"
	"gitlab.com/pala-software/prestress/pkg/prestress"
	"gitlab.com/pala-software/prestress/pkg/subscriber"
	"go.uber.org/dig"
)

func container() (c *dig.Container, err error) {
	c = dig.New()

	err = c.Provide(http.NewServeMux)
	if err != nil {
		return
	}

	err = c.Provide(databaseFromEnv)
	if err != nil {
		return
	}

	err = c.Provide(prestress.CoreFromEnv().Provider())
	if err != nil {
		return
	}

	err = c.Provide(migrator.MigratorFromEnv().Provider())
	if err != nil {
		return
	}

	err = c.Provide(crud.CrudFromEnv().Provider())
	if err != nil {
		return
	}

	err = c.Provide(subscriber.SubscriberFromEnv().Provider())
	if err != nil {
		return
	}

	err = c.Provide(auth.AuthenticationFromEnv().Provider())
	if err != nil {
		return
	}

	err = c.Provide(oauth.OAuthFromEnv().Provider())
	if err != nil {
		return
	}

	return
}
