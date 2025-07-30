package crud_test

import (
	"context"
	"errors"
	"testing"

	"gitlab.com/pala-software/prestress/pkg/crud"
)

func TestUpdateWithCancelledContext(t *testing.T) {
	err := container.Invoke(func(update *crud.UpdateOperation) (err error) {
		initCtx, cancel := context.WithCancel(context.Background())
		ctx, err := begin(initCtx)
		if err != nil {
			cancel()
			return
		}

		oldValue := "1"
		newValue := "1-2"

		table := "update"
		where := crud.Where{"value": oldValue}
		data := map[string]any{"value": newValue}

		err = expectItems(ctx, table, where, []string{oldValue})
		if err != nil {
			ctx.Rollback()
			cancel()
			return
		}

		params := crud.UpdateParams{
			Table: table,
			Where: where,
			Data:  data,
		}
		cancel()
		_, err = update.Execute(ctx, params)
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

		err = expectItems(checkCtx, table, where, []string{oldValue})
		if err != nil {
			checkCtx.Rollback()
			return
		}

		err = checkCtx.Commit()
		return
	})

	if err != nil {
		t.Error(err)
	}
}

func TestUpdate(t *testing.T) {
	err := container.Invoke(func(update *crud.UpdateOperation) (err error) {
		ctx, err := begin(context.Background())
		if err != nil {
			return
		}

		oldValue := "2"
		newValue := "2-2"

		table := "update"
		where := crud.Where{"value": oldValue}
		data := map[string]any{"value": newValue}

		err = expectItems(ctx, table, where, []string{oldValue})
		if err != nil {
			ctx.Rollback()
			return
		}

		params := crud.UpdateParams{
			Table: table,
			Where: where,
			Data:  data,
		}
		_, err = update.Execute(ctx, params)
		if err != nil {
			ctx.Rollback()
			return
		}

		err = expectItems(ctx, table, where, []string{})
		if err != nil {
			ctx.Rollback()
			return
		}

		ctx.Commit()
		return
	})

	if err != nil {
		t.Error(err)
	}
}
