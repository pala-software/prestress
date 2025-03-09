package prestress_test

import (
	"database/sql"
	"fmt"
)

func expectValues(rows *sql.Rows, expectedValues []string) error {
	var err error
	count := 0
	for _, expectedValue := range expectedValues {
		if !rows.Next() {
			return fmt.Errorf(
				"expected %d rows, got %d",
				len(expectedValues),
				count,
			)
		}

		var actualValue string
		err = rows.Scan(&actualValue)
		if err != nil {
			return err
		}

		if actualValue != expectedValue {
			return fmt.Errorf(
				"expected value '%s', got '%s'",
				expectedValue,
				actualValue,
			)
		}

		count++
	}

	if rows.Next() {
		return fmt.Errorf(
			"expected %d rows, got too many",
			len(expectedValues),
		)
	}

	return nil
}
