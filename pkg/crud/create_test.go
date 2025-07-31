package crud_test

import (
	"context"
	"errors"
	"testing"

	"gitlab.com/pala-software/prestress/pkg/crud"
)

func TestCreateWithCancelledContext(t *testing.T) {
	err := container.Invoke(func(create *crud.CreateOperation) (err error) {
		initCtx, cancel := context.WithCancel(context.Background())
		ctx, err := begin(initCtx)
		if err != nil {
			cancel()
			return
		}

		table := "create"
		data := map[string]any{
			"value": "withCancelledContext",
		}

		params := crud.CreateParams{
			Table: table,
			Data:  data,
		}
		cancel()
		_, err = create.Execute(ctx, params)
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

		err = expectItems(
			checkCtx,
			table,
			crud.Where{"value": "withCancelledContext"},
			[]string{},
		)
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

func TestCreate(t *testing.T) {
	err := container.Invoke(func(create *crud.CreateOperation) (err error) {
		ctx, err := begin(context.Background())
		if err != nil {
			return
		}

		table := "create"
		data := map[string]any{
			"value": "3",
		}

		params := crud.CreateParams{
			Table: table,
			Data:  data,
		}
		_, err = create.Execute(ctx, params)
		if err != nil {
			ctx.Rollback()
			return
		}

		err = expectItems(
			ctx,
			table,
			crud.Where{"value": "3"},
			[]string{"3"},
		)
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
