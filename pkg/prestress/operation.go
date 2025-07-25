package prestress

import (
	"context"
	"errors"
	"maps"
	"net/http"

	"github.com/jackc/pgx/v5"
)

var ErrNoBegin = errors.New("no begin operation provided")

type OperationContext struct {
	context.Context

	// Should be only set when execution originates from HTTP.
	Request *http.Request

	// Database transaction for this operation
	Tx pgx.Tx

	// Target schema
	Schema string

	// Extra variables for other packages to use
	Variables map[string]Loggable
}

func (ctx OperationContext) Details() (details map[string]string) {
	details = map[string]string{
		"schema": ctx.Schema,
	}
	for _, val := range ctx.Variables {
		maps.Copy(details, val.Details())
	}
	return
}

func (ctx OperationContext) Commit() error {
	return ctx.Tx.Commit(ctx)
}

func (ctx OperationContext) Rollback() error {
	return ctx.Tx.Rollback(ctx)
}

type OperationParams interface {
	Loggable
}

type OperationResult interface {
	Loggable
}

type EmptyOperationParams struct{}

func (EmptyOperationParams) Details() map[string]string {
	return map[string]string{}
}

type EmptyOperationResult struct{}

func (EmptyOperationResult) Details() map[string]string {
	return map[string]string{}
}

type OperationHandler[Params OperationParams, Res OperationResult] interface {
	// Name used for identifying the operation
	Name() string

	// Execute operation.
	Execute(OperationContext, Params) (Res, error)

	// Handle HTTP request.
	Handle(
		http.ResponseWriter,
		*http.Request,

		// Function to use instead of calling OperationHandler.handle directly
		func(Params) (Res, error),
	)
}

type BeforeEventHandler func(
	OperationContext,
	OperationParams,
) error

type AfterEventHandler func(
	OperationContext,
	OperationParams,
	OperationResult,
) error

type AnyOperation interface {
	Loggable
	http.Handler
	Name() string
	OnBefore(BeforeEventHandler)
	OnAfter(AfterEventHandler)
}

type BeforeOperationHook[Params OperationParams, Res OperationResult] func(
	OperationContext,
	Params,
) (OperationContext, Params, error)

type AfterOperationHook[Params OperationParams, Res OperationResult] func(
	OperationContext,
	Params,
	Res,
) (Res, error)

type Operation[Params OperationParams, Res OperationResult] struct {
	handler OperationHandler[Params, Res]
	begin   *BeginOperation
	before  Registry[BeforeOperationHook[Params, Res]]
	after   Registry[AfterOperationHook[Params, Res]]
}

func (op Operation[Params, Res]) Name() string {
	return op.handler.Name()
}

func (op Operation[Params, Res]) Before() *Registry[BeforeOperationHook[Params, Res]] {
	return &op.before
}

func (op Operation[Params, Res]) After() *Registry[AfterOperationHook[Params, Res]] {
	return &op.after
}

func (op Operation[Params, Res]) OnBefore(handler BeforeEventHandler) {
	op.before.Register(func(ctx OperationContext, params Params) (OperationContext, Params, error) {
		err := handler(ctx, params)
		return ctx, params, err
	})
}

func (op Operation[Params, Res]) OnAfter(handler AfterEventHandler) {
	op.after.Register(func(ctx OperationContext, params Params, res Res) (Res, error) {
		err := handler(ctx, params, res)
		return res, err
	})
}

func (op Operation[Params, Res]) Details() map[string]string {
	return map[string]string{
		"operation": op.handler.Name(),
	}
}

func (op Operation[Params, Res]) Execute(
	ctx OperationContext,
	params Params,
) (res Res, err error) {
	for _, hook := range op.Before().Value() {
		ctx, params, err = hook(ctx, params)
		if err != nil {
			return
		}
	}

	res, err = op.handler.Execute(ctx, params)
	if err != nil {
		return
	}

	for _, hook := range op.After().Value() {
		res, err = hook(ctx, params, res)
		if err != nil {
			return
		}
	}

	return
}

func (op Operation[Params, Res]) ServeHTTP(
	writer http.ResponseWriter,
	request *http.Request,
) {
	op.handler.Handle(
		writer,
		request,
		func(params Params) (res Res, err error) {
			if op.begin == nil {
				err = ErrNoBegin
				return
			}

			ctx, err := op.begin.BeginHTTP(request)
			if err != nil {
				ctx.Rollback()
				return
			}

			res, err = op.Execute(ctx, params)
			if err != nil {
				ctx.Rollback()
				return
			}

			err = ctx.Commit()
			return
		},
	)
}

func NewOperation[Params OperationParams, Res OperationResult](
	handler OperationHandler[Params, Res],
	begin *BeginOperation,
) *Operation[Params, Res] {
	return &Operation[Params, Res]{
		handler: handler,
		begin:   begin,
	}
}
