// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package display

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// A v2 resource has no currentState while it is CREATING: the API omits it, and
// some clients may get it as null. A column such as "currentState.status" must
// then render as an empty cell for that row, not abort the whole table: gval
// fails on a path through an absent or null object, which used to be fatal.
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

	td.Cmp(t, exitCalled, false, "the table must not exit on an absent or null parent")
	td.Cmp(t, out, td.Re(`ready\s+│\s+ACTIVE\s+│\s+READY`))
	td.Cmp(t, out, td.Re(`creating\s+│\s+│\s+CREATING`))
	td.Cmp(t, out, td.Re(`absent\s+│\s+│\s+CREATING`))
}

// A column whose parent exists but is not an object is a wrong column
// definition: it must keep failing instead of rendering an empty column.
func TestRenderTable_ScalarParentStillFails(t *testing.T) {
	values := []map[string]any{{"id": "ready", "name": "my-name"}}

	exitCalled := false
	orig := ExitFunc
	ExitFunc = func(int) { exitCalled = true }
	defer func() { ExitFunc = orig }()

	out := captureStdout(t, func() {
		RenderTable(values, []string{"id", "name.sub"}, &OutputFormat{})
	})

	td.Cmp(t, exitCalled, true, "a path through a scalar must fail")
	td.Cmp(t, out, td.Contains("failed to select row field"))
}

func TestColumnParentIsMissing(t *testing.T) {
	row := map[string]any{
		"name":         "my-name",
		"currentState": nil,
		"targetSpec":   map[string]any{"name": "my-name", "location": nil},
	}

	for column, expected := range map[string]bool{
		"currentState.status":         true,  // null parent
		"absent.status":               true,  // absent parent
		"targetSpec.location.region":  true,  // null parent, deeper
		"targetSpec.absent.region":    true,  // absent parent, deeper
		"name.sub":                    false, // scalar parent
		"targetSpec.name.sub":         false, // scalar parent, deeper
		"targetSpec.name":             false, // the parent is there
		"name":                        false, // no parent
		"currentState.status + \"x\"": false, // not a plain path
	} {
		td.Cmp(t, columnParentIsMissing(row, column), expected, column)
	}
}
