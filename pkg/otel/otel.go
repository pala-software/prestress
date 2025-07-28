package otel

import (
	"context"
	"maps"
	"net/http"
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

func (feature OTel) Middleware() func(http.Handler) http.Handler {
	return otelhttp.NewMiddleware("server")
}

func (feature *OTel) Provider() any {
	return func() (self *OTel) {
		self = feature
		return
	}
}

func (feature *OTel) Invoker() any {
	return func(
		lifecycle *prestress.Lifecycle,
		core *prestress.Core,
	) (err error) {
		err = feature.RegisterHooks(lifecycle, core)
		if err != nil {
			return
		}

		return
	}
}

func (feature OTel) RegisterHooks(
	lifecycle *prestress.Lifecycle,
	core *prestress.Core,
) (err error) {
	otelShutdown, err := feature.setup(context.Background())
	if err != nil {
		return
	}

	lifecycle.Start.Register(func() error {
		logger.Info("Start")
		return nil
	})

	lifecycle.Shutdown.Register(func() error {
		logger.Info("Shutdown")
		return otelShutdown(context.Background())
	})

	for _, op := range core.Operations().Value() {
		op.OnBefore(func(
			ctx prestress.OperationContext,
			params prestress.OperationParams,
		) error {
			detailsMap := op.Details()
			maps.Copy(detailsMap, ctx.Details())
			maps.Copy(detailsMap, params.Details())

			detailsSlice := make([]any, len(detailsMap)*2)
			index := 0
			for key, val := range detailsMap {
				detailsSlice[index+0] = key
				detailsSlice[index+1] = val
				index += 2
			}

			logger.Info("Before"+op.Name(), detailsSlice...)
			return nil
		})

		op.OnAfter(func(
			ctx prestress.OperationContext,
			params prestress.OperationParams,
			res prestress.OperationResult,
		) error {
			detailsMap := op.Details()
			maps.Copy(detailsMap, ctx.Details())
			maps.Copy(detailsMap, params.Details())
			maps.Copy(detailsMap, res.Details())

			detailsSlice := make([]any, len(detailsMap)*2)
			index := 0
			for key, val := range detailsMap {
				detailsSlice[index+0] = key
				detailsSlice[index+1] = val
				index += 2
			}

			logger.Info("After"+op.Name(), detailsSlice...)
			return nil
		})
	}

	return
}
