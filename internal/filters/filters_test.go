package filters

import (
	"encoding/json"
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

func TestFilterLines_NoFilters(t *testing.T) {
	values := []map[string]any{
		{"id": 1, "name": "foo"},
		{"id": 2, "name": "bar"},
	}
	result, err := FilterLines(values, nil)

	td.Require(t).CmpNoError(err)
	td.Cmp(t, result, values)
}

func TestFilterLines_SimpleEquality(t *testing.T) {
	values := []map[string]any{
		{"id": 1, "name": "foo"},
		{"id": 2, "name": "bar"},
	}
	result, err := FilterLines(values, []string{`name == "foo"`})

	td.Require(t).CmpNoError(err)
	td.Require(t).Len(result, 1)
	td.Cmp(t, "foo", result[0]["name"])
}

func TestFilterLines_MultipleFilters(t *testing.T) {
	values := []map[string]any{
		{"id": 1, "name": "foo", "active": true},
		{"id": 2, "name": "bar", "active": false},
		{"id": 3, "name": "foo", "active": false},
	}
	result, err := FilterLines(values, []string{`name == "foo"`, `active == true`})

	td.Require(t).CmpNoError(err)
	td.Require(t).Len(result, 1)
	td.Cmp(t, 1, result[0]["id"])
}

func TestFilterLines_JsonNumberComparison(t *testing.T) {
	values := []map[string]any{
		{"id": json.Number("1"), "score": json.Number("10.5")},
		{"id": json.Number("2"), "score": json.Number("5.2")},
	}
	result, err := FilterLines(values, []string{`score > 6`})

	td.Require(t).CmpNoError(err)
	td.Require(t).Len(result, 1)
	td.Cmp(t, json.Number("1"), result[0]["id"])
}

func TestFilterLines_OperatorError(t *testing.T) {
	values := []map[string]any{
		{"id": 1, "name": "foo"},
	}
	_, err := FilterLines(values, []string{`name >`})

	td.CmpError(t, err, "failed to parse filter `name >`: syntax error at position 5: unexpected end of input")
}

func TestFilterLines_IntAndStringComparison(t *testing.T) {
	values := []map[string]any{
		{"id": 1, "name": "foo"},
	}

	_, err := FilterLines(values, []string{`name > 1`})

	td.CmpNoError(t, err)
}

func TestFilterLines_EmptyValues(t *testing.T) {
	result, err := FilterLines([]map[string]any{}, []string{`id == 1`})

	td.Require(t).CmpNoError(err)
	td.CmpEmpty(t, result)
}
