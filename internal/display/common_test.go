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
		raw      bool
		expected string
	}{
		// Without --raw: values are JSON-encoded (current behavior).
		{"string quoted by default", "my-token", false, `"my-token"`},
		{"number by default", float64(123), false, "123"},

		// With --raw: scalars are printed without JSON quoting.
		{"string raw", "my-token", true, "my-token"},
		{"number raw", float64(123), true, "123"},
		{"bool raw", true, true, "true"},
		{"nil raw", nil, true, ""},

		// With --raw: complex values still fall back to JSON.
		{"map raw falls back to JSON", map[string]any{"a": "b"}, true, `{"a":"b"}`},
		{"slice raw falls back to JSON", []any{"a", "b"}, true, `["a","b"]`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := formatCustomValue(tc.value, tc.raw)
			td.Require(t).CmpNoError(err)
			td.Cmp(t, got, tc.expected)
		})
	}
}
