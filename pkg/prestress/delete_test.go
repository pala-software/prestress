package prestress_test

import (
	"context"
	"testing"

	"gitlab.com/pala-software/prestress/pkg/prestress"
)

func TestDeleteWithCancelledContext(t *testing.T) {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	auth := prestress.AuthenticationResult{
		Role:      "anonymous",
		Variables: map[string]any{},
	}
	schema := "test"
	table := "delete"
	where := prestress.Where{"value": "1"}

	err = expectItems(context.Background(), table, where, []string{"1"})
	if err != nil {
		t.Error(err)
		return
	}

	err = server.Delete(ctx, auth, schema, table, where)
	if err != context.Canceled {
		t.Errorf(
			"expected error to be '%v', got '%v'",
			context.Canceled,
			err,
		)
		return
	}

	err = expectItems(context.Background(), table, where, []string{"1"})
	if err != nil {
		t.Error(err)
		return
	}
}

func TestDelete(t *testing.T) {
	var err error

	ctx := context.Background()
	auth := prestress.AuthenticationResult{
		Role:      "anonymous",
		Variables: map[string]any{},
	}
	schema := "test"
	table := "delete"
	where := prestress.Where{"value": "2"}

	err = expectItems(ctx, table, where, []string{"2"})
	if err != nil {
		t.Error(err)
		return
	}

	err = server.Delete(ctx, auth, schema, table, where)
	if err != nil {
		t.Error(err)
		return
	}

	err = expectItems(ctx, table, where, []string{})
	if err != nil {
		t.Error(err)
		return
	}
}
