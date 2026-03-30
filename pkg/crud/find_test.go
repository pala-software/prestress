package crud_test

import (
	"context"
	"errors"
	"testing"

	"gitlab.com/pala-software/prestress/pkg/crud"
)

func TestFindWithCancelledContext(t *testing.T) {
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
			[]*string{},
		)
		if err != nil {
			ctx.Rollback()
			return
		}

		err = ctx.Commit()
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
			[]*string{nil, strPtr("1"), strPtr("2")},
		)
		if err != nil {
			ctx.Rollback()
			return
		}

		err = ctx.Commit()
		return
	}()

	if err != nil {
		t.Error(err)
	}
}

func TestFindWithEquals(t *testing.T) {
	err := func() (err error) {
		ctx, err := begin(context.Background())
		if err != nil {
			return
		}

		err = expectItems(
			ctx,
			"find",
			crud.Where{"value": crud.Equals{"1"}},
			[]*string{strPtr("1")},
		)
		if err != nil {
			ctx.Rollback()
			return
		}

		err = ctx.Commit()
		return
	}()

	if err != nil {
		t.Error(err)
	}
}

func TestFindNull(t *testing.T) {
	err := func() (err error) {
		ctx, err := begin(context.Background())
		if err != nil {
			return
		}

		err = expectItems(
			ctx,
			"find",
			crud.Where{"value": crud.IsNull{}},
			[]*string{nil},
		)
		if err != nil {
			ctx.Rollback()
			return
		}

		err = ctx.Commit()
		return
	}()

	if err != nil {
		t.Error(err)
	}
}

func TestFindNotNull(t *testing.T) {
	err := func() (err error) {
		ctx, err := begin(context.Background())
		if err != nil {
			return
		}

		err = expectItems(
			ctx,
			"find",
			crud.Where{"value": crud.IsNotNull{}},
			[]*string{strPtr("1"), strPtr("2")},
		)
		if err != nil {
			ctx.Rollback()
			return
		}

		err = ctx.Commit()
		return
	}()

	if err != nil {
		t.Error(err)
	}
}
