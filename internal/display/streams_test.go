// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package display

import (
	"io"
	"os"
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// captureStreams runs f with both standard streams replaced by pipes and
// returns what was written to each of them.
func captureStreams(t *testing.T, f func()) (stdout, stderr string) {
	t.Helper()

	outR, outW, err := os.Pipe()
	td.Require(t).CmpNoError(err)
	errR, errW, err := os.Pipe()
	td.Require(t).CmpNoError(err)

	origOut, origErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outW, errW
	defer func() {
		os.Stdout, os.Stderr = origOut, origErr
	}()

	f()

	outW.Close()
	errW.Close()

	outBytes, err := io.ReadAll(outR)
	td.Require(t).CmpNoError(err)
	errBytes, err := io.ReadAll(errR)
	td.Require(t).CmpNoError(err)

	return string(outBytes), string(errBytes)
}

// An error must never reach stdout: a redirected stdout has to stay parsable,
// so `cmd -o json > out.json` leaves out.json empty when the call failed.
func TestOutputError_GoesToStderrOnly(t *testing.T) {
	origExit := ExitFunc
	ExitFunc = func(int) {}
	defer func() { ExitFunc = origExit }()

	// "interactive" is in the list on purpose: its viewer takes over stdout,
	// so a diagnostic must be rendered plainly instead of being handed to it.
	for _, format := range []string{"", "json", "yaml", "interactive"} {
		out, errOut := captureStreams(t, func() {
			OutputError(&OutputFormat{Output: format}, "something went wrong")
		})

		td.Cmp(t, out, "", "stdout stays empty with output format %q", format)
		td.Cmp(t, errOut, td.Contains("something went wrong"),
			"stderr carries the error with output format %q", format)
	}
}

// Success messages keep going to stdout, which is what scripts consume.
func TestOutputInfo_GoesToStdout(t *testing.T) {
	out, errOut := captureStreams(t, func() {
		OutputInfo(&OutputFormat{}, nil, "task started")
	})

	td.Cmp(t, out, td.Contains("task started"))
	td.Cmp(t, errOut, "")
}

// A warning follows the same rule as an error, in every format.
func TestOutputWarning_GoesToStderrOnly(t *testing.T) {
	origExit := ExitFunc
	ExitFunc = func(int) {}
	defer func() { ExitFunc = origExit }()

	for _, format := range []string{"", "json", "yaml", "interactive"} {
		out, errOut := captureStreams(t, func() {
			OutputWarning(&OutputFormat{Output: format}, "careful")
		})

		td.Cmp(t, out, "", "stdout stays empty with output format %q", format)
		td.Cmp(t, errOut, td.Contains("careful"),
			"stderr carries the warning with output format %q", format)
	}
}
