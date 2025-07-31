package crud

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

type DeleteParams struct {
	Table string
	Where Where
}

func (params DeleteParams) Details() map[string]string {
	return map[string]string{
		"table": params.Table,
	}
}

type DeleteOperationHandler struct{}

func (DeleteOperationHandler) Name() string {
	return "Delete"
}

func (op DeleteOperationHandler) Execute(
	ctx prestress.OperationContext,
	params DeleteParams,
) (res prestress.EmptyOperationResult, err error) {
	_, err = ctx.Tx.Exec(
		ctx,
		fmt.Sprintf(
			"DELETE FROM %s AS t %s",
			pgx.Identifier{ctx.Schema, params.Table}.Sanitize(),
			params.Where.String("t", 1),
		),
		params.Where.Values()...,
	)
	return
}

func (op DeleteOperationHandler) Handle(
	writer http.ResponseWriter,
	request *http.Request,
	handle func(DeleteParams) (
		prestress.EmptyOperationResult,
		prestress.OperationContext,
		error,
	),
) {
	var err error

	table := request.PathValue("table")
	query := request.URL.Query()
	where := ParseWhere(query)

	params := DeleteParams{
		Table: table,
		Where: where,
	}
	_, ctx, err := handle(params)
	if err != nil {
		prestress.HandleDatabaseError(writer, err)
		return
	}

	err = ctx.Commit()
	if err != nil {
		prestress.HandleDatabaseError(writer, err)
		return
	}

	writer.WriteHeader(204)
}

type DeleteOperation struct {
	*prestress.Operation[
		DeleteParams,
		prestress.EmptyOperationResult,
	]
}

func NewDeleteOperation(begin *prestress.BeginOperation) *DeleteOperation {
	return &DeleteOperation{
		prestress.NewOperation(
			new(DeleteOperationHandler),
			begin,
		),
	}
}
