package crud_test

import (
	"context"
	"errors"
	"testing"

	"gitlab.com/pala-software/prestress/pkg/crud"
)

func TestFindWithCancelledContext(t *testing.T) {
	t.Skip("skipping: acquiring connection from the pool with done context hangs in CI")

	err := func() (err error) {
		initCtx, cancel := context.WithCancel(context.Background())
		ctx, err := begin(initCtx)
		if err != nil {
			cancel()
			return
		}

		cancel()
		err = expectItems(
			ctx,
			"find",
			crud.Where{},
			[]string{},
		)
		if err != nil {
			return
		}

		return
	}()

	if !errors.Is(err, context.Canceled) {
		t.Errorf(
			"expected error to be '%v', got '%v'",
			context.Canceled,
			err,
		)
	}
}

func TestFindAll(t *testing.T) {
	err := func() (err error) {
		ctx, err := begin(context.Background())
		if err != nil {
			return
		}

		err = expectItems(
			ctx,
			"find",
			crud.Where{},
			[]string{"1", "2"},
		)
		if err != nil {
			return
		}

		return
	}()

	if err != nil {
		t.Error(err)
	}
}

func TestFindWithFilter(t *testing.T) {
	err := func() (err error) {
		ctx, err := begin(context.Background())
		if err != nil {
			return
		}

		err = expectItems(
			ctx,
			"find",
			crud.Where{"value": "1"},
			[]string{"1"},
		)
		if err != nil {
			return
		}

		return
	}()

	if err != nil {
		t.Error(err)
	}
}
