// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package display

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// captureStdout runs f with stdout replaced by a pipe and returns what was
// written to it.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	td.Require(t).CmpNoError(err)

	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	f()
	w.Close()

	out, err := io.ReadAll(r)
	td.Require(t).CmpNoError(err)

	return string(out)
}

// Values decoded from an API response are json.Number, because the client
// decodes with UseNumber, and were already printed verbatim. A float64
// reaches the table only when a value is built on the client side, and that
// path used to be rounded to the unit by %.0f. Both must now print the same
// thing, so that the rendering no longer depends on where the value came from.
func TestRenderTable_KeepsDecimalsExact(t *testing.T) {
	var decoded []map[string]any
	decoder := json.NewDecoder(strings.NewReader(
		`[{"name":"price","value":12.99},{"name":"bandwidth","value":2.5}]`))
	decoder.UseNumber()
	td.Require(t).CmpNoError(decoder.Decode(&decoded))

	fromClient := []map[string]any{
		{"name": "price", "value": 12.99},
		{"name": "bandwidth", "value": 2.5},
		{"name": "count", "value": float64(3)},
	}

	for _, values := range [][]map[string]any{decoded, fromClient} {
		out := captureStdout(t, func() {
			RenderTable(values, []string{"name", "value"}, &OutputFormat{})
		})

		td.Cmp(t, out, td.Contains("12.99"))
		td.Cmp(t, out, td.Contains("2.5"))
		td.Cmp(t, out, td.Not(td.Contains("13")), "12.99 must not be rounded up")
	}
}

// A whole number must not grow a decimal part on its way to the table.
func TestRenderTable_WholeFloatsStayIntegral(t *testing.T) {
	out := captureStdout(t, func() {
		RenderTable(
			[]map[string]any{{"name": "cores", "value": float64(64)}},
			[]string{"name", "value"},
			&OutputFormat{},
		)
	})

	td.Cmp(t, out, td.Contains("64"))
	td.Cmp(t, out, td.Not(td.Contains("64.0")))
}
