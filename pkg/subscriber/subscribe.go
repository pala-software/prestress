package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	ctx           context.Context
	conn          *pgxpool.Conn
	mutex         *sync.Mutex
	subscriptions map[int]*Subscription
}

func (SubscribeOperationHandler) Name() string {
	return "Subscribe"
}

func (op *SubscribeOperationHandler) Execute(
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

	// Invoke permission check
	_, err = ctx.Tx.Exec(
		ctx,
		fmt.Sprintf(
			"SELECT FROM %s LIMIT 0",
			pgx.Identifier{ctx.Schema, params.Table}.Sanitize(),
		),
	)
	if err != nil {
		return
	}

	op.mutex.Lock()
	defer op.mutex.Unlock()

	var subId int
	err = op.conn.QueryRow(
		op.ctx,
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
		op.mutex.Lock()
		defer op.mutex.Unlock()

		_, err := op.conn.Exec(
			op.ctx,
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
	handle func(SubscribeParams) (
		*Subscription,
		prestress.OperationContext,
		error,
	),
) {
	var err error

	table := request.PathValue("table")
	params := SubscribeParams{
		Table: table,
	}
	sub, ctx, err := handle(params)
	if err != nil {
		prestress.HandleDatabaseError(writer, err)
		return
	}

	err = ctx.Commit()
	if err != nil {
		prestress.HandleDatabaseError(writer, err)
		return
	}

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("X-Accel-Buffering", "no")
	responseController := http.NewResponseController(writer)

	// Send empty comment to make sure connection is open
	writer.Write([]byte(":\n\n"))
	err = responseController.Flush()
	if err != nil {
		fmt.Println(err)
		return
	}

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
	pool *pgxpool.Pool,
	lifecycle *prestress.Lifecycle,
	begin *prestress.BeginOperation,
	create *crud.CreateOperation,
	update *crud.UpdateOperation,
	delete *crud.DeleteOperation,
) *SubscribeOperation {
	handler := &SubscribeOperationHandler{
		subscriptions: make(map[int]*Subscription),
	}

	var cancel context.CancelFunc
	handler.ctx, cancel = context.WithCancel(context.Background())
	lifecycle.Start.Register(func() (err error) {
		handler.conn, err = pool.Acquire(handler.ctx)
		handler.mutex = &sync.Mutex{}
		return
	})
	lifecycle.Shutdown.Register(func() (err error) {
		defer cancel()
		handler.conn.Release()
		return
	})

	create.Before().Register(func(
		initCtx prestress.OperationContext,
		initParams crud.CreateParams,
	) (
		ctx prestress.OperationContext,
		params crud.CreateParams,
		err error,
	) {
		ctx = initCtx
		params = initParams
		err = handler.createChangeTable(ctx)
		return
	})

	create.After().Register(func(
		ctx prestress.OperationContext,
		params crud.CreateParams,
		initRes prestress.EmptyOperationResult,
	) (res prestress.EmptyOperationResult, err error) {
		res = initRes
		err = handler.collectChanges(ctx)
		return
	})

	update.Before().Register(func(
		initCtx prestress.OperationContext,
		initParams crud.UpdateParams,
	) (
		ctx prestress.OperationContext,
		params crud.UpdateParams,
		err error,
	) {
		ctx = initCtx
		params = initParams
		err = handler.createChangeTable(ctx)
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

	delete.Before().Register(func(
		initCtx prestress.OperationContext,
		initParams crud.DeleteParams,
	) (
		ctx prestress.OperationContext,
		params crud.DeleteParams,
		err error,
	) {
		ctx = initCtx
		params = initParams
		err = handler.createChangeTable(ctx)
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
