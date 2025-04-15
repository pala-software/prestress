package crud_test

import (
	"context"
	"testing"

	"gitlab.com/pala-software/prestress/pkg/crud"
	"gitlab.com/pala-software/prestress/pkg/prestress"
)

func TestCreateWithCancelledContext(t *testing.T) {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	auth := prestress.AuthenticationResult{
		Role:      prestress.AnonymousRole,
		Variables: map[string]any{},
	}
	schema := "test"
	table := "create"
	data := map[string]any{
		"value": "withCancelledContext",
	}

	err = feature.Create(ctx, auth, schema, table, data)
	if err != context.Canceled {
		t.Errorf(
			"expected error to be '%v', got '%v'",
			context.Canceled,
			err,
		)
		return
	}

	err = expectItems(
		context.Background(),
		"create",
		crud.Where{"value": "withCancelledContext"},
		[]string{},
	)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestCreate(t *testing.T) {
	var err error

	ctx := context.Background()
	auth := prestress.AuthenticationResult{
		Role:      "anonymous",
		Variables: map[string]any{},
	}
	schema := "test"
	table := "create"
	data := map[string]any{
		"value": "3",
	}

	err = feature.Create(ctx, auth, schema, table, data)
	if err != nil {
		t.Error(err)
		return
	}

	err = expectItems(
		context.Background(),
		"create",
		crud.Where{"value": "3"},
		[]string{"3"},
	)
	if err != nil {
		t.Error(err)
		return
	}
}
