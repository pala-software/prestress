package crud_test

import (
	"context"
	"errors"
	"testing"

	"gitlab.com/pala-software/prestress/pkg/crud"
)

func TestDeleteWithCancelledContext(t *testing.T) {
	t.Skip("skipping: trying to re-acquire connection from the pool with done context hangs in CI")

	err := container.Invoke(func(delete *crud.DeleteOperation) (err error) {
		initCtx, cancel := context.WithCancel(context.Background())
		ctx, err := begin(initCtx)
		if err != nil {
			cancel()
			return
		}

		table := "delete"
		where := crud.Where{"value": "1"}

		err = expectItems(ctx, table, where, []string{"1"})
		if err != nil {
			cancel()
			return
		}

		params := crud.DeleteParams{
			Table: table,
			Where: where,
		}
		cancel()
		_, err = delete.Execute(ctx, params)
		if !errors.Is(err, context.Canceled) {
			t.Errorf(
				"expected error to be '%v', got '%v'",
				context.Canceled,
				err,
			)
		}

		checkCtx, err := begin(context.Background())
		if err != nil {
			return
		}

		err = expectItems(checkCtx, table, where, []string{"1"})
		if err != nil {
			return
		}

		err = checkCtx.Commit()
		if err != nil {
			return
		}

		return
	})

	if err != nil {
		t.Error(err)
	}
}

func TestDelete(t *testing.T) {
	err := container.Invoke(func(delete *crud.DeleteOperation) (err error) {
		ctx, err := begin(context.Background())
		if err != nil {
			return
		}

		table := "delete"
		where := crud.Where{"value": "2"}

		err = expectItems(ctx, table, where, []string{"2"})
		if err != nil {
			return
		}

		params := crud.DeleteParams{
			Table: table,
			Where: where,
		}
		_, err = delete.Execute(ctx, params)
		if err != nil {
			return
		}

		err = expectItems(ctx, table, where, []string{})
		if err != nil {
			return
		}

		return
	})

	if err != nil {
		t.Error(err)
	}
}
