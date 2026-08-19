// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package display

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"gopkg.in/ini.v1"
)

// The plain format exists so that a redirected table can be split by cut or
// awk: no borders, no colours, no trailing hint, and columns padded with
// spaces.
func TestRenderTable_PlainHasNoBordersNorHint(t *testing.T) {
	values := []map[string]any{
		{"name": "ns3168421.ip-51-77-12.eu", "state": "ok"},
		{"name": "ns5012993.ip-141-95-4.eu", "state": "ok"},
	}

	RenderTable(values, []string{"name", "state"}, &OutputFormat{Output: "plain"})

	td.Cmp(t, ResultString, td.Not(td.Contains("┌")))
	td.Cmp(t, ResultString, td.Not(td.Contains("│")))
	td.Cmp(t, ResultString, td.Not(td.Contains("Use option -o json")))
	td.Cmp(t, ResultString, td.Contains("NAME"), "headers are upper-cased")
	td.Cmp(t, ResultString, td.Contains("ns3168421.ip-51-77-12.eu  ok"),
		"columns are padded with spaces so cut and awk can split them")
}

// The default format must keep its borders: the plain shape is opt-in.
func TestRenderTable_DefaultKeepsBorders(t *testing.T) {
	values := []map[string]any{{"name": "ns3168421.ip-51-77-12.eu", "state": "ok"}}

	RenderTable(values, []string{"name", "state"}, &OutputFormat{})

	td.Cmp(t, ResultString, td.Contains("│"))
}

// A plain output must stay a plain output: "plain" is a format name, not a
// gval expression to evaluate.
func TestOutputFormat_PlainIsNotACustomFormat(t *testing.T) {
	format := OutputFormat{Output: "plain"}

	td.Cmp(t, format.IsPlain(), true)
	td.Cmp(t, format.CustomFormat(), "")
}

// The config table is rendered by a separate function, which used to ignore
// -o plain entirely and print its bordered form regardless.
func TestRenderConfigTable_HonoursPlain(t *testing.T) {
	cfg := ini.Empty()
	section, err := cfg.NewSection("default")
	td.Require(t).CmpNoError(err)
	_, err = section.NewKey("application_key", "abc123")
	td.Require(t).CmpNoError(err)

	RenderConfigTable(cfg, &OutputFormat{Output: "plain"})

	td.Cmp(t, ResultString, td.Not(td.Contains("│")))
	td.Cmp(t, ResultString, td.Contains("application_key"))
	td.Cmp(t, ResultString, td.Contains("SECTION"), "headers are upper-cased")
}
