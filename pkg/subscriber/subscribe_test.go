package subscriber_test

import (
	"context"
	"fmt"
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

func FuzzLocking(f *testing.F) {
	err := container.Invoke(func(
		subscribe *subscriber.SubscribeOperation,
		create *crud.CreateOperation,
	) {
		for i := range 10 {
			f.Add(i)
		}

		f.Fuzz(func(t *testing.T, i int) {
			ctx, err := begin(context.Background())
			if err != nil {
				return
			}

			sub, err := subscribe.Execute(ctx, subscriber.SubscribeParams{
				Table: "document",
			})
			if err != nil {
				t.Fatal(err)
			}

			if len(sub.Change) > 0 {
				err = fmt.Errorf("got %d initial changes, 0 expected", len(sub.Change))
				t.Fatal(err)
			}

			err = ctx.Commit()
			if err != nil {
				t.Fatal(err)
			}

			ctx, err = begin(context.Background())
			if err != nil {
				return
			}

			_, err = create.Execute(ctx, crud.CreateParams{
				Table: "document",
				Data: map[string]any{
					"body": "3",
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			if len(sub.Change) != 1 {
				t.Fatalf("got %d changes, 1 expected", len(sub.Change))
			}

			change := <-sub.Change
			if len(sub.Change) > 0 {
				t.Fatalf("got %d more changes, 0 expected", len(sub.Change))
				return
			}
			if change.RowOperation != "INSERT" {
				t.Fatalf("got %s operation, expected INSERT", change.RowOperation)
			}

			// We don't want to persist the changes in this test so that we can expect
			// similar starting state in other tests.
			err = ctx.Rollback()
			if err != nil {
				t.Fatal(err)
			}
		})
	})

	if err != nil {
		f.Fatal(err)
	}
}
