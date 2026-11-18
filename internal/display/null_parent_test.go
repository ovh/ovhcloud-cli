// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package display

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// A v2 resource has a null currentState while it is CREATING. A column such as
// "currentState.status" must then render as an empty cell for that row, not
// abort the whole table: gval reports a path through a null (or absent) object
// as an "unknown parameter" error, which used to be fatal.
func TestRenderTable_NullParentIsAnEmptyCell(t *testing.T) {
	values := []map[string]any{
		{"id": "ready", "resourceStatus": "READY", "currentState": map[string]any{"status": "ACTIVE"}},
		{"id": "creating", "resourceStatus": "CREATING", "currentState": nil},
		{"id": "absent", "resourceStatus": "CREATING"},
	}

	exitCalled := false
	orig := ExitFunc
	ExitFunc = func(int) { exitCalled = true }
	defer func() { ExitFunc = orig }()

	out := captureStdout(t, func() {
		RenderTable(values, []string{"id", "currentState.status status", "resourceStatus"}, &OutputFormat{})
	})

	td.Cmp(t, exitCalled, false, "the table must not exit on a null parent")
	td.Cmp(t, out, td.Re(`ready\s+│\s+ACTIVE\s+│\s+READY`))
	td.Cmp(t, out, td.Re(`creating\s+│\s+│\s+CREATING`))
	td.Cmp(t, out, td.Re(`absent\s+│\s+│\s+CREATING`))
}
