package crud_test

import (
	"net/url"
	"testing"

	"gitlab.com/pala-software/prestress/pkg/crud"
)

var whereQuery = url.Values{
	"where[one]": []string{"value-1"},
	"where[2]": []string{"2"},
	"where[null_column][null]": []string{"1"},
	"where[notNullColumn][null]": []string{"0"},
}

// Columns should be in sorted order (similar to alphabetical order)
func TestWhereColumns(t *testing.T) {
	where := crud.ParseWhere(whereQuery);
	expectedColumns := []string{"2", "notNullColumn", "null_column", "one"}
	actualColumns := where.Columns()
	if len(actualColumns) != len(expectedColumns) {
		t.Errorf(
			"expected %d columns, got %d",
			len(expectedColumns),
			len(actualColumns),
		)
		return
	}
	for index := range expectedColumns {
		if actualColumns[index] != expectedColumns[index] {
			t.Errorf(
				"expected column '%s', got '%s' at position %d",
				expectedColumns[index],
				actualColumns[index],
				index,
			)
		}
	}
}

// Values should be in the order of sorted columns excluding null conditions
func TestWhereValues(t *testing.T) {
	where := crud.ParseWhere(whereQuery)
	expectedValues := []any{"2", "value-1"}
	actualValues := where.Values()
	if len(actualValues) != len(expectedValues) {
		t.Errorf(
			"expected %d values, got %d",
			len(expectedValues),
			len(actualValues),
		)
		return
	}
	for index := range expectedValues {
		if actualValues[index] != expectedValues[index] {
			t.Errorf(
				"expected value '%v', got '%v' at position %d",
				expectedValues[index],
				actualValues[index],
				index,
			)
		}
	}
}

func TestWhereSQL(t *testing.T) {
	where := crud.ParseWhere(whereQuery)
	expectedSQL := `SELECT * FROM "public"."value" WHERE ` +
		`"value"."2" = $10 AND `+
		`"value"."notNullColumn" IS NOT NULL AND ` +
		`"value"."null_column" IS NULL AND ` + 
		`"value"."one" = $11`;
	actualSQL := `SELECT * FROM "public"."value" ` + where.String("value", 10);
	if actualSQL != expectedSQL {
		t.Errorf("generated SQL does not match expected SQL\nexpected: %s\nactual:   %s", expectedSQL, actualSQL)
	}
}
