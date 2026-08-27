// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package display

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

func TestFormatCustomValue(t *testing.T) {
	testCases := []struct {
		name     string
		value    any
		expected string
	}{
		// Scalars are always printed without JSON quoting.
		{"string unquoted", "my-token", "my-token"},
		{"number", float64(123), "123"},
		{"bool", true, "true"},
		{"nil", nil, ""},

		// Complex values still fall back to JSON.
		{"map falls back to JSON", map[string]any{"a": "b"}, `{"a":"b"}`},
		{"slice falls back to JSON", []any{"a", "b"}, `["a","b"]`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := formatCustomValue(tc.value)
			td.Require(t).CmpNoError(err)
			td.Cmp(t, got, tc.expected)
		})
	}
}
