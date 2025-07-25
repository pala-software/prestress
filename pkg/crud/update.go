package crud

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

type UpdateParams struct {
	Table string
	Where Where
	Data  map[string]any
}

func (params UpdateParams) Details() map[string]string {
	return map[string]string{
		"table": params.Table,
	}
}

type UpdateOperationHandler struct{}

func (UpdateOperationHandler) Name() string {
	return "Update"
}

func (op UpdateOperationHandler) Execute(
	ctx prestress.OperationContext,
	params UpdateParams,
) (res prestress.EmptyOperationResult, err error) {
	if len(params.Data) == 0 {
		return
	}

	patch := make([]string, 0, len(params.Data))
	values := make([]any, 0, len(patch)+len(params.Where))
	n := 1
	for column, value := range params.Data {
		patch = append(
			patch,
			fmt.Sprintf(
				"%s = %s",
				pgx.Identifier{column}.Sanitize(),
				"$"+strconv.Itoa(n),
			),
		)
		values = append(values, value)
		n++
	}
	values = append(values, params.Where.Values()...)

	_, err = ctx.Tx.Exec(
		ctx,
		fmt.Sprintf(
			"UPDATE %s AS t SET %s %s",
			pgx.Identifier{ctx.Schema, params.Table}.Sanitize(),
			strings.Join(patch, ", "),
			params.Where.String("t", len(patch)+1),
		),
		values...,
	)
	return
}

func (op UpdateOperationHandler) Handle(
	writer http.ResponseWriter,
	request *http.Request,
	handle func(UpdateParams) (prestress.EmptyOperationResult, error),
) {
	var err error

	table := request.PathValue("table")
	query := request.URL.Query()

	where := ParseWhere(query)

	if request.Body == nil {
		writer.WriteHeader(400)
		return
	}

	msg, err := io.ReadAll(request.Body)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(500)
		return
	}

	var data map[string]any
	err = json.Unmarshal(msg, &data)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(400)
		return
	}
	params := UpdateParams{
		Table: table,
		Where: where,
		Data:  data,
	}
	_, err = handle(params)
	if err != nil {
		prestress.HandleDatabaseError(writer, err)
		return
	}

	writer.WriteHeader(204)
}

type UpdateOperation struct {
	*prestress.Operation[
		UpdateParams,
		prestress.EmptyOperationResult,
	]
}

func NewUpdateOperation(begin *prestress.BeginOperation) *UpdateOperation {
	return &UpdateOperation{
		prestress.NewOperation(
			new(UpdateOperationHandler),
			begin,
		),
	}
}
