package subscriber_test

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"

	"gitlab.com/pala-software/prestress/pkg/crud"
	"gitlab.com/pala-software/prestress/pkg/subscriber"
)

func TestSimpleInsertOnCreate(t *testing.T) {
	err := container.Invoke(func(
		subscribe *subscriber.SubscribeOperation,
		create *crud.CreateOperation,
	) (err error) {
		ctx, err := begin(context.Background())
		if err != nil {
			return
		}

		sub, err := subscribe.Execute(ctx, subscriber.SubscribeParams{
			Table: "document",
		})
		if err != nil {
			return
		}

		if len(sub.Change) > 0 {
			err = fmt.Errorf("got %d initial changes, 0 expected", len(sub.Change))
			return
		}

		_, err = create.Execute(ctx, crud.CreateParams{
			Table: "document",
			Data: map[string]any{
				"body": "3",
			},
		})
		if err != nil {
			return
		}

		if len(sub.Change) != 1 {
			err = fmt.Errorf("got %d changes, 1 expected", len(sub.Change))
			return
		}

		change := <-sub.Change
		if len(sub.Change) > 0 {
			err = fmt.Errorf("got %d more changes, 0 expected", len(sub.Change))
			return
		}
		if change.RowOperation != "INSERT" {
			err = fmt.Errorf("got %s operation, expected INSERT", change.RowOperation)
			return
		}

		// We don't want to persist the changes in this test so that we can expect
		// similar starting state in other tests.
		err = ctx.Rollback()
		if err != nil {
			return
		}

		return
	})

	if err != nil {
		t.Error(err)
	}
}

func TestLocking(t *testing.T) {
	err := container.Invoke(func(
		subscribe *subscriber.SubscribeOperation,
	) (err error) {
		var wg sync.WaitGroup
		for range runtime.GOMAXPROCS(0) {
			wg.Go(func() {
				initCtx, cancel := context.WithCancel(context.Background())
				defer cancel()

				ctx, err := begin(initCtx)
				if err != nil {
					return
				}

				_, err = subscribe.Execute(ctx, subscriber.SubscribeParams{
					Table: "document",
				})
				if err != nil {
					return
				}

				err = ctx.Commit()
				if err != nil {
					return
				}
			})
		}

		wg.Wait();
		return
	})

	if err != nil {
		t.Fatal(err)
	}
}
