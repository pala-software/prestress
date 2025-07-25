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

type CreateParams struct {
	Table string
	Data  map[string]any
}

func (params CreateParams) Details() map[string]string {
	return map[string]string{
		"table": params.Table,
	}
}

type CreateOperationHandler struct{}

func (CreateOperationHandler) Name() string {
	return "Create"
}

func (op CreateOperationHandler) Execute(
	ctx prestress.OperationContext,
	params CreateParams,
) (res prestress.EmptyOperationResult, err error) {
	if len(params.Data) == 0 {
		_, err = ctx.Tx.Exec(
			ctx,
			fmt.Sprintf(
				"INSERT INTO %s DEFAULT VALUES",
				pgx.Identifier{ctx.Schema, params.Table}.Sanitize(),
			),
		)
		if err != nil {
			return
		}
	} else {
		columns := make([]string, 0, len(params.Data))
		values := make([]any, 0, len(params.Data))
		placeholders := make([]string, 0, len(params.Data))
		n := 1
		for key, value := range params.Data {
			columns = append(columns, pgx.Identifier{key}.Sanitize())
			values = append(values, value)
			placeholders = append(placeholders, "$"+strconv.Itoa(n))
			n++
		}

		_, err = ctx.Tx.Exec(
			ctx,
			fmt.Sprintf(
				"INSERT INTO %s (%s) VALUES (%s)",
				pgx.Identifier{ctx.Schema, params.Table}.Sanitize(),
				strings.Join(columns, ", "),
				strings.Join(placeholders, ", "),
			),
			values...,
		)
		if err != nil {
			return
		}
	}

	return
}

func (op CreateOperationHandler) Handle(
	writer http.ResponseWriter,
	request *http.Request,
	handle func(CreateParams) (prestress.EmptyOperationResult, error),
) {
	var err error

	table := request.PathValue("table")

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

	params := CreateParams{
		Table: table,
		Data:  data,
	}
	_, err = handle(params)
	if err != nil {
		prestress.HandleDatabaseError(writer, err)
		return
	}

	writer.WriteHeader(204)
}

type CreateOperation struct {
	*prestress.Operation[
		CreateParams,
		prestress.EmptyOperationResult,
	]
}

func NewCreateOperation(begin *prestress.BeginOperation) *CreateOperation {
	return &CreateOperation{
		prestress.NewOperation(
			new(CreateOperationHandler),
			begin,
		),
	}
}
