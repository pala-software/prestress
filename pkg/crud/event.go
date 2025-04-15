package crud

import (
	"context"

	"github.com/jackc/pgx/v5"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

type BeforeBeginOperationEvent struct {
	OperationName string
	Auth          *prestress.AuthenticationResult
	Schema        *string
	Table         *string
	Context       context.Context
}

type AfterBeginOperationEvent struct {
	OperationName string
	Auth          prestress.AuthenticationResult
	Schema        string
	Table         string
	Transaction   pgx.Tx
	Context       context.Context
}

type BeforeCommitOperationEvent struct {
	OperationName string
	Auth          prestress.AuthenticationResult
	Schema        string
	Table         string
	Transaction   pgx.Tx
	Context       context.Context
}

type AfterCommitOperationEvent struct {
	OperationName string
	Auth          prestress.AuthenticationResult
	Schema        string
	Table         string
	Context       context.Context
}

func (event BeforeBeginOperationEvent) Event() string {
	return "BeforeBegin" + event.OperationName + "OperationEvent"
}

func (event AfterBeginOperationEvent) Event() string {
	return "AfterBegin" + event.OperationName + "OperationEvent"
}

func (event BeforeCommitOperationEvent) Event() string {
	return "BeforeCommit" + event.OperationName + "OperationEvent"
}

func (event AfterCommitOperationEvent) Event() string {
	return "AfterCommit" + event.OperationName + "OperationEvent"
}
