package prestress_test

import (
	"context"
	"testing"

	"gitlab.com/pala-software/prestress/pkg/prestress"
)

func TestUpdateWithCancelledContext(t *testing.T) {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	oldValue := "1"
	newValue := "1-2"

	auth := prestress.AuthenticationResult{
		Role:      "anonymous",
		Variables: map[string]any{},
	}
	schema := "test"
	table := "update"
	where := prestress.Where{"value": oldValue}
	data := map[string]any{"value": newValue}

	err = expectItems(context.Background(), table, where, []string{oldValue})
	if err != nil {
		t.Error(err)
		return
	}

	err = server.Update(ctx, auth, schema, table, where, data)
	if err != context.Canceled {
		t.Errorf(
			"expected error to be '%v', got '%v'",
			context.Canceled,
			err,
		)
		return
	}

	err = expectItems(context.Background(), table, where, []string{oldValue})
	if err != nil {
		t.Error(err)
		return
	}
}

func TestUpdate(t *testing.T) {
	var err error

	oldValue := "2"
	newValue := "2-2"

	ctx := context.Background()
	auth := prestress.AuthenticationResult{
		Role:      "anonymous",
		Variables: map[string]any{},
	}
	schema := "test"
	table := "update"
	where := prestress.Where{"value": oldValue}
	data := map[string]any{"value": newValue}

	err = expectItems(ctx, table, where, []string{oldValue})
	if err != nil {
		t.Error(err)
		return
	}

	err = server.Update(ctx, auth, schema, table, where, data)
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
