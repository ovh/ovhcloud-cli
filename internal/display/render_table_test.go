// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package display

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// An empty list must render as an empty array: with a nil slice the json
// output was the literal null, so a consumer's type changed with the data.
func TestRenderTable_EmptyListIsEmptyArray(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		values []map[string]any
	}{
		{"nil slice", nil},
		{"empty slice", []map[string]any{}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			out := captureStdout(t, func() {
				RenderTable(testCase.values, []string{"name"}, &OutputFormat{Output: "json"})
			})

			td.Cmp(t, out, "[]\n")
		})
	}
}
