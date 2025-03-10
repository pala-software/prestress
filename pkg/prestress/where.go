package prestress

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Where map[string]string

func ParseWhere(query url.Values) Where {
	where := make(Where, len(query))
	for key, values := range query {
		var found bool

		key, found = strings.CutPrefix(key, "where[")
		if !found {
			continue
		}

		key, found = strings.CutSuffix(key, "]")
		if !found {
			continue
		}

		if len(values) == 0 {
			continue
		}

		where[key] = values[0]
	}
	return where
}

// TODO: Test
func (where Where) String(table string, paramStart int) string {
	if len(where) == 0 {
		return ""
	}

	conditions := make([]string, 0, len(where))
	n := paramStart
	for column := range where {
		conditions = append(
			conditions,
			fmt.Sprintf(
				"%s = %s",
				pgx.Identifier{table, column}.Sanitize(),
				"$"+strconv.Itoa(n),
			),
		)
		n++
	}
	return "WHERE " + strings.Join(conditions, " AND ")
}

// TODO: Test
func (where Where) Values() []any {
	values := make([]any, 0, len(where))
	for _, value := range where {
		values = append(values, value)
	}
	return values
}
