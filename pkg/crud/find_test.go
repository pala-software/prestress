package crud_test

import (
	"context"
	"testing"

	"gitlab.com/pala-software/prestress/pkg/crud"
)

func TestFindWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := expectItems(
		ctx,
		"find",
		crud.Where{},
		[]string{},
	)
	if err != context.Canceled {
		t.Errorf(
			"expected error to be '%v', got '%v'",
			context.Canceled,
			err,
		)
	}
}

func TestFindAll(t *testing.T) {
	err := expectItems(
		context.Background(),
		"find",
		crud.Where{},
		[]string{"1", "2"},
	)
	if err != nil {
		t.Error(err)
	}
}

func TestFindWithFilter(t *testing.T) {
	err := expectItems(
		context.Background(),
		"find",
		crud.Where{"value": "1"},
		[]string{"1"},
	)
	if err != nil {
		t.Error(err)
	}
}
