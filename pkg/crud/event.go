package crud

import (
	"context"
	"maps"

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

func (event BeforeBeginOperationEvent) Details() map[string]string {
	details := map[string]string{
		"operation": event.OperationName,
		"schema":    *event.Schema,
		"table":     *event.Table,
	}
	maps.Copy(details, event.Auth.Details())
	return details
}

func (event AfterBeginOperationEvent) Event() string {
	return "AfterBegin" + event.OperationName + "OperationEvent"
}

func (event AfterBeginOperationEvent) Details() map[string]string {
	details := map[string]string{
		"operation": event.OperationName,
		"schema":    event.Schema,
		"table":     event.Table,
	}
	maps.Copy(details, event.Auth.Details())
	return details
}

func (event BeforeCommitOperationEvent) Event() string {
	return "BeforeCommit" + event.OperationName + "OperationEvent"
}

func (event BeforeCommitOperationEvent) Details() map[string]string {
	details := map[string]string{
		"operation": event.OperationName,
		"schema":    event.Schema,
		"table":     event.Table,
	}
	maps.Copy(details, event.Auth.Details())
	return details
}

func (event AfterCommitOperationEvent) Event() string {
	return "AfterCommit" + event.OperationName + "OperationEvent"
}

func (event AfterCommitOperationEvent) Details() map[string]string {
	details := map[string]string{
		"operation": event.OperationName,
		"schema":    event.Schema,
		"table":     event.Table,
	}
	maps.Copy(details, event.Auth.Details())
	return details
}
