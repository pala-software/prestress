package prestress

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Where map[string]string

// TODO: Test
func (where Where) String(table string, paramStart int) string {
	if len(where) == 0 {
		return ""
	}

	filters := make([]string, 0, len(where))
	n := paramStart
	for column := range where {
		filters = append(
			filters,
			fmt.Sprintf(
				"%s = %s",
				pgx.Identifier{table, column}.Sanitize(),
				"$"+strconv.Itoa(n),
			),
		)
		n++
	}
	return "WHERE " + strings.Join(filters, " AND ")
}

// TODO: Test
func (filters Where) Values() []any {
	values := make([]any, 0, len(filters))
	for _, value := range filters {
		values = append(values, value)
	}
	return values
}
