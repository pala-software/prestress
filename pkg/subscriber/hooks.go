package subscriber

import (
	"fmt"

	"gitlab.com/pala-software/prestress/pkg/prestress"
)

func (SubscribeOperationHandler) createChangeTable(
	ctx prestress.OperationContext,
	params prestress.EmptyOperationParams,
	initRes prestress.OperationContext,
) (res prestress.OperationContext, err error) {
	res = initRes
	_, err = ctx.Tx.Exec(
		ctx,
		`CREATE TEMPORARY TABLE IF NOT EXISTS pg_temp.prestress_change
				OF prestress.change
				ON COMMIT DELETE ROWS`,
	)
	return
}

func (op *SubscribeOperationHandler) collectChanges(
	ctx prestress.OperationContext,
) (err error) {
	rows, err := ctx.Tx.Query(
		ctx,
		`SELECT
			subscription_id,
			row_key,
			row_data,
			row_operation
		FROM
			pg_temp.prestress_change`,
	)
	if err != nil {
		return
	}

	defer rows.Close()
	var subId int
	var change Change
	for rows.Next() {
		err = rows.Scan(
			&subId,
			&change.RowKey,
			&change.RowData,
			&change.RowOperation,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		subscription, exists := op.subscriptions[subId]
		if !exists {
			fmt.Println("change collected for subscription that does not exist")
			continue
		}

		subscription.Change <- change
	}

	return
}
