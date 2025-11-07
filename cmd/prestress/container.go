package main

import (
	"net/http"

	"gitlab.com/pala-software/prestress/pkg/auth"
	"gitlab.com/pala-software/prestress/pkg/crud"
	"gitlab.com/pala-software/prestress/pkg/migrator"
	"gitlab.com/pala-software/prestress/pkg/oauth"
	"gitlab.com/pala-software/prestress/pkg/otel"
	"gitlab.com/pala-software/prestress/pkg/param"
	"gitlab.com/pala-software/prestress/pkg/prestress"
	"gitlab.com/pala-software/prestress/pkg/subscriber"
	"go.uber.org/dig"
)

var features = []prestress.Feature{
	prestress.CoreFromEnv(),
	migrator.MigratorFromEnv(),
	crud.CrudFromEnv(),
	subscriber.SubscriberFromEnv(),
	auth.AuthenticationFromEnv(),
	oauth.OAuthFromEnv(),
	param.ParamFromEnv(),
	otel.OTelFromEnv(),
}

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

	for _, feature := range features {
		err = c.Provide(feature.Provider())
		if err != nil {
			return
		}
	}

	for _, feature := range features {
		err = c.Invoke(feature.Invoker())
		if err != nil {
			return
		}
	}

	return
}
