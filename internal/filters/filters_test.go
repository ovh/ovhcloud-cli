// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

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

	td.CmpNoError(t, err)
	td.Cmp(t, result, values)
}

func TestFilterLines_SimpleEquality(t *testing.T) {
	values := []map[string]any{
		{"id": 1, "name": "foo"},
		{"id": 2, "name": "bar"},
	}
	result, err := FilterLines(values, []string{`name == "foo"`})

	td.CmpNoError(t, err)
	td.Cmp(t, result, values[0:1])
}

func TestFilterLines_MultipleFilters(t *testing.T) {
	values := []map[string]any{
		{"id": 1, "name": "foo", "active": true},
		{"id": 2, "name": "bar", "active": false},
		{"id": 3, "name": "foo", "active": false},
	}
	result, err := FilterLines(values, []string{`name == "foo"`, `active == true`})

	td.CmpNoError(t, err)
	td.Cmp(t, result, values[0:1])
}

func TestFilterLines_JsonNumberComparison(t *testing.T) {
	values := []map[string]any{
		{"id": json.Number("1"), "score": json.Number("10.5")},
		{"id": json.Number("2"), "score": json.Number("5.2")},
	}
	result, err := FilterLines(values, []string{`score > 6`})

	td.CmpNoError(t, err)
	td.Cmp(t, result, values[0:1])
}

func TestFilterLines_OperatorError(t *testing.T) {
	values := []map[string]any{
		{"id": 1, "name": "foo"},
	}
	result, err := FilterLines(values, []string{`name >`})

	td.CmpEmpty(t, result)
	td.CmpContains(t, err, `failed to parse filter "name >": parsing error`)
}

func TestFilterLines_IntAndStringComparison(t *testing.T) {
	values := []map[string]any{
		{"id": 1, "name": "foo"},
	}

	result, err := FilterLines(values, []string{`name > 1`})

	td.CmpNoError(t, err)
	td.CmpEmpty(t, result)
}

func TestFilterLines_EmptyValues(t *testing.T) {
	result, err := FilterLines([]map[string]any{}, []string{`id == 1`})

	td.CmpNoError(t, err)
	td.CmpEmpty(t, result)
}
