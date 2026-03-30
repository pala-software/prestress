package crud

import (
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Condition interface {
	SQL(variable pgx.Identifier, paramN *int) string
	ParamValues() []any
}

type Equals struct {
	Value any
}

func (Equals) SQL(variable pgx.Identifier, paramN *int) string {
	n := *paramN
	*paramN = n+1

	return fmt.Sprintf(
		"%s = $%d",
		variable.Sanitize(),
		n,
	)
}

func (condition Equals) ParamValues() []any {
	return []any{condition.Value}
}

type IsNull struct {}

func (IsNull) SQL(variable pgx.Identifier, paramN *int) string {
	return fmt.Sprintf(
		"%s IS NULL",
		variable.Sanitize(),
	)
}

func (IsNull) ParamValues() []any {
 	return nil 
}

type IsNotNull struct {}

func (IsNotNull) SQL(variable pgx.Identifier, paramN *int) string {
	return fmt.Sprintf(
		"%s IS NOT NULL",
		variable.Sanitize(),
	)
}

func (IsNotNull) ParamValues() []any {
 	return nil 
}

type Where map[string]Condition

func ParseWhere(query url.Values) Where {
	where := make(Where, len(query))
	for key, values := range query {
		var value any
		var found bool

		if len(values) == 0 {
			continue
		} else {
			value = values[0]
		}
		
		rest, found := strings.CutPrefix(key, "where[")
		if !found {
			continue
		}

		column, rest, found := strings.Cut(rest, "]")
		if !found {
			continue
		}

		switch rest {
			case "":
				where[column] = Equals{value} 
			
			case "[null]":
				switch value {
				case "1":
					where[column] = &IsNull{}
				case "0":
					where[column] = &IsNotNull{}
				}
		}
	}
	return where
}

func (where Where) Columns() []string {
	columns := slices.Collect(maps.Keys(where))
	slices.Sort(columns)
	return columns
}

func (where Where) String(table string, paramN int) string {
	if len(where) == 0 {
		return ""
	}

	conditions := make([]string, 0, len(where))
	for _, column := range where.Columns() {
		conditions = append(
			conditions,
			where[column].SQL(pgx.Identifier{table, column}, &paramN),
		)
	}
	return "WHERE " + strings.Join(conditions, " AND ")
}

func (where Where) Values() []any {
	values := []any{}
	for _, column := range where.Columns() {
		values = append(values, where[column].ParamValues()...)
	}
	return values
}
