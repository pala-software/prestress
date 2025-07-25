package prestress

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
)

var ErrInvalidSchema = errors.New("invalid schema")
var ErrForbiddenSchema = errors.New("forbidden schema")

type BeginOperationHandler struct{}

func (BeginOperationHandler) Name() string {
	return "Begin"
}

func (op BeginOperationHandler) Execute(
	initCtx OperationContext,
	params EmptyOperationParams,
) (ctx OperationContext, err error) {
	ctx = initCtx

	if ctx.Schema == "" {
		err = ErrInvalidSchema
		return
	}

	if ctx.Schema == "pg_temp" {
		err = ErrForbiddenSchema
		return
	}

	_, err = ctx.Tx.Exec(
		ctx,
		fmt.Sprintf(
			// pg_temp is set to last in search_path so that we don't accidentally or
			// in any case query temporary tables implicitly.
			"SET LOCAL search_path TO %s, pg_temp",
			pgx.Identifier{ctx.Schema}.Sanitize(),
		),
	)
	if err != nil {
		ctx.Tx.Rollback(ctx)
		return
	}

	return
}

// Begin operation should never be called from HTTP.
func (BeginOperationHandler) Handle(
	writer http.ResponseWriter,
	request *http.Request,
	handle func(EmptyOperationParams) (OperationContext, error),
) {
	writer.WriteHeader(500)
}

type BeginOperation struct {
	*Operation[EmptyOperationParams, OperationContext]

	conn *pgx.Conn
}

func (op BeginOperation) Begin(
	initCtx context.Context,
	schema string,
) (ctx OperationContext, err error) {
	tx, err := op.conn.Begin(initCtx)
	if err != nil {
		return
	}

	ctx = OperationContext{
		Context: initCtx,
		Tx:      tx,
		Schema:  schema,
	}
	ctx, err = op.Execute(ctx, EmptyOperationParams{})
	return
}

func (op BeginOperation) BeginHTTP(
	request *http.Request,
) (ctx OperationContext, err error) {
	schema := request.PathValue("schema")
	ctx, err = op.Begin(request.Context(), schema)
	return
}

func NewBeginOperation(conn *pgx.Conn) *BeginOperation {
	return &BeginOperation{
		NewOperation(new(BeginOperationHandler), nil),
		conn,
	}
}
