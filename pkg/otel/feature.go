package otel

import (
	"context"
	"os"

	"gitlab.com/pala-software/prestress/pkg/prestress"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var name = "gitlab.com/pala-software/prestress/pkg/otel"
var logger = otelslog.NewLogger(name)

type OTel struct {
	TracesEnabled  bool
	MetricsEnabled bool
	LogsEnabled    bool
}

func OTelFromEnv() *OTel {
	feature := OTel{}
	feature.TracesEnabled = os.Getenv("PRESTRESS_OTEL_TRACES_ENABLE") == "1"
	feature.MetricsEnabled = os.Getenv("PRESTRESS_OTEL_METRICS_ENABLE") == "1"
	feature.LogsEnabled = os.Getenv("PRESTRESS_OTEL_METRICS_ENABLE") == "1"
	return &feature
}

func (feature OTel) Apply(server *prestress.Server) error {
	otelShutdown, err := feature.setup(context.Background())
	if err != nil {
		return err
	}

	server.AddMiddleware(otelhttp.NewMiddleware("server"))

	server.OnEvent(func(event prestress.Event) error {
		detailsMap := event.Details()
		detailsSlice := make([]any, len(detailsMap)*2)
		index := 0
		for key, val := range detailsMap {
			detailsSlice[index+0] = key
			detailsSlice[index+1] = val
			index += 2
		}
		logger.Info(event.Event(), detailsSlice...)

		switch event.(type) {
		case prestress.ServerShutdownEvent:
			return otelShutdown(context.Background())
		default:
			return nil
		}
	})

	return nil
}
