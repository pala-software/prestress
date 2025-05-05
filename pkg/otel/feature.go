package otel

import (
	"context"

	"gitlab.com/pala-software/prestress/pkg/prestress"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type OTel struct {
}

func OTelFromEnv() *OTel {
	return &OTel{}
}

func (feature OTel) Apply(server *prestress.Server) error {
	otelShutdown, err := setupOTelSDK(context.Background())
	if err != nil {
		return err
	}

	server.AddMiddleware(otelhttp.NewMiddleware("server"))

	server.OnEvent(func(event prestress.Event) error {
		switch event.(type) {
		case prestress.ServerShutdownEvent:
			return otelShutdown(context.Background())
		default:
			return nil
		}
	})

	return nil
}
