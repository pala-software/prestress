package crud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

type FindParams struct {
	Table  string
	Where  Where
	Limit  int
	Offset int
}

func (params FindParams) Details() map[string]string {
	return map[string]string{
		"table":  params.Table,
		"limit":  strconv.Itoa(params.Limit),
		"offset": strconv.Itoa(params.Offset),
	}
}

type FindResult struct {
	pgx.Rows
}

func (FindResult) Details() map[string]string {
	// TODO: Implement
	return map[string]string{}
}

type FindOperationHandler struct{}

func (FindOperationHandler) Name() string {
	return "Find"
}

func (op FindOperationHandler) Execute(
	ctx prestress.OperationContext,
	params FindParams,
) (res FindResult, err error) {
	rows, err := ctx.Tx.Query(
		ctx,
		fmt.Sprintf(
			"SELECT to_json(t) FROM %s AS t %s LIMIT %d OFFSET %d",
			pgx.Identifier{ctx.Schema, params.Table}.Sanitize(),
			params.Where.String("t", 1),
			params.Limit,
			params.Offset,
		),
		params.Where.Values()...,
	)
	res = FindResult{rows}
	return
}

func (op FindOperationHandler) Handle(
	writer http.ResponseWriter,
	request *http.Request,
	handle func(FindParams) (FindResult, error),
) {
	var err error

	table := request.PathValue("table")
	query := request.URL.Query()

	where := ParseWhere(query)

	limit := 100
	if query.Has("limit") {
		limit, err = strconv.Atoi(query.Get("limit"))
		if err != nil {
			writer.WriteHeader(400)
			return
		}
	}

	offset := 0
	if query.Has("offset") {
		offset, err = strconv.Atoi(query.Get("offset"))
		if err != nil {
			writer.WriteHeader(400)
			return
		}
	}

	params := FindParams{
		Table:  table,
		Where:  where,
		Limit:  limit,
		Offset: offset,
	}
	rows, err := handle(params)
	if err != nil {
		prestress.HandleDatabaseError(writer, err)
		return
	}
	defer rows.Close()

	first := true
	row := json.RawMessage{}
	for rows.Next() {
		err := rows.Scan(&row)
		if err != nil {
			fmt.Println(err)
			return
		}

		encodedRow, err := json.Marshal(row)
		if err != nil {
			fmt.Println(err)
			return
		}

		if first {
			first = false
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(200)
			writer.Write([]byte("["))
			defer writer.Write([]byte("]"))
		} else {
			writer.Write([]byte(","))
		}

		_, err = writer.Write(encodedRow)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	err = rows.Err()
	if err != nil {
		prestress.HandleDatabaseError(writer, err)
		return
	}

	if first {
		writer.WriteHeader(200)
		writer.Write([]byte("[]"))
	}
}

type FindOperation struct {
	*prestress.Operation[FindParams, FindResult]
}

func NewFindOperation(begin *prestress.BeginOperation) *FindOperation {
	return &FindOperation{
		prestress.NewOperation(
			new(FindOperationHandler),
			begin,
		),
	}
}
