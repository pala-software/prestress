package crud_test

import (
	"context"
	"errors"
	"testing"

	"gitlab.com/pala-software/prestress/pkg/crud"
)

func TestDeleteWithCancelledContext(t *testing.T) {
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
			ctx.Rollback()
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

		if err != nil {
			ctx.Rollback()
		} else {
			ctx.Commit()
		}

		checkCtx, err := begin(context.Background())
		if err != nil {
			return
		}

		err = expectItems(checkCtx, table, where, []string{"1"})
		if err != nil {
			checkCtx.Rollback()
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
			ctx.Rollback()
			return
		}

		params := crud.DeleteParams{
			Table: table,
			Where: where,
		}
		_, err = delete.Execute(ctx, params)
		if err != nil {
			ctx.Rollback()
			return
		}

		err = expectItems(ctx, table, where, []string{})
		if err != nil {
			ctx.Rollback()
			return
		}

		err = ctx.Commit()
		return
	})

	if err != nil {
		t.Error(err)
	}
}
