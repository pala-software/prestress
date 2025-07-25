package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.com/pala-software/prestress/pkg/auth"
	"gitlab.com/pala-software/prestress/pkg/crud"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

type SubscribeParams struct {
	Table string
}

func (params SubscribeParams) Details() map[string]string {
	return map[string]string{
		"table": params.Table,
	}
}

type SubscribeOperationHandler struct {
	subscriptions map[int]*Subscription
}

func (SubscribeOperationHandler) Name() string {
	return "Subscribe"
}

func (op SubscribeOperationHandler) Execute(
	ctx prestress.OperationContext,
	params SubscribeParams,
) (sub *Subscription, err error) {
	authRes, ok := ctx.Variables["auth"].(*auth.AuthenticationResult)
	if !ok {
		err = auth.ErrAuthenticationRequired
		return
	}

	variables := map[string]any{}
	if authRes.Variables != nil {
		variables = authRes.Variables
	}

	encodedVariables, err := json.Marshal(variables)
	if err != nil {
		return
	}

	var subId int
	err = ctx.Tx.QueryRow(
		ctx,
		"SELECT prestress.setup_subscription($1, $2, $3, $4)",
		authRes.Role,
		ctx.Schema,
		params.Table,
		encodedVariables,
	).Scan(&subId)
	if err != nil {
		return
	}

	sub = &Subscription{
		Change: make(chan Change),
	}
	op.subscriptions[subId] = sub

	context.AfterFunc(ctx, func() {
		_, err := ctx.Tx.Exec(
			ctx,
			"SELECT prestress.teardown_subscription($1)",
			subId,
		)
		if err != nil {
			fmt.Println(err)
		}
		delete(op.subscriptions, subId)
	})

	return
}

func (op SubscribeOperationHandler) Handle(
	writer http.ResponseWriter,
	request *http.Request,
	handle func(SubscribeParams) (*Subscription, error),
) {
	var err error

	table := request.PathValue("table")
	params := SubscribeParams{
		Table: table,
	}
	sub, err := handle(params)
	if err != nil {
		prestress.HandleDatabaseError(writer, err)
		return
	}

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	responseController := http.NewResponseController(writer)

	for {
		select {
		case <-request.Context().Done():
			return

		case change := <-sub.Change:
			encodedChange, err := json.Marshal(change)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Prepend message with a "data: " prefix
			encodedChange = append([]byte("event: change\ndata: "), encodedChange...)

			// End the message with two newline characters
			encodedChange = append(encodedChange, 0x0A, 0x0A)

			_, err = writer.Write(encodedChange)
			if err != nil {
				fmt.Println(err)
				return
			}

			err = responseController.Flush()
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

type SubscribeOperation struct {
	*prestress.Operation[
		SubscribeParams,
		*Subscription,
	]
}

func NewSubscribeOperation(
	begin *prestress.BeginOperation,
	create *crud.CreateOperation,
	update *crud.UpdateOperation,
	delete *crud.DeleteOperation,
) *SubscribeOperation {
	handler := new(SubscribeOperationHandler)

	begin.After().Register(handler.createChangeTable)
	create.After().Register(func(
		ctx prestress.OperationContext,
		params crud.CreateParams,
		initRes prestress.EmptyOperationResult,
	) (res prestress.EmptyOperationResult, err error) {
		res = initRes
		err = handler.collectChanges(ctx)
		return
	})
	update.After().Register(func(
		ctx prestress.OperationContext,
		params crud.UpdateParams,
		initRes prestress.EmptyOperationResult,
	) (res prestress.EmptyOperationResult, err error) {
		res = initRes
		err = handler.collectChanges(ctx)
		return
	})
	delete.After().Register(func(
		ctx prestress.OperationContext,
		params crud.DeleteParams,
		initRes prestress.EmptyOperationResult,
	) (res prestress.EmptyOperationResult, err error) {
		res = initRes
		err = handler.collectChanges(ctx)
		return
	})

	return &SubscribeOperation{
		prestress.NewOperation(
			handler,
			begin,
		),
	}
}
