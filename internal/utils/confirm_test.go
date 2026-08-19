// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package utils

import (
	"bytes"
	"strings"
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// withInteractiveInput runs f as if the operator were typing answer at an
// interactive terminal.
func withInteractiveInput(t *testing.T, answer string, f func()) {
	t.Helper()

	withInteractiveIO(t, answer, f)
}

// withInteractiveIO does the same and hands back everything the guard printed,
// so a test can read the question the way the operator does.
func withInteractiveIO(t *testing.T, answer string, f func()) string {
	t.Helper()

	var printed bytes.Buffer

	origInput, origOutput, origInteractive := confirmInput, confirmOutput, confirmInteractive
	confirmInput = strings.NewReader(answer)
	confirmOutput = &printed
	confirmInteractive = func() bool { return true }
	defer func() {
		confirmInput, confirmOutput, confirmInteractive = origInput, origOutput, origInteractive
	}()

	f()

	return printed.String()
}

// The exact name is the guard between a typo and an erased disk, so the
// accepting path has to be exercised, not only the refusals around it.
func TestConfirmByName_AcceptsTheExactName(t *testing.T) {
	var confirmed bool
	withInteractiveInput(t, "ns3168421.ip-51-77-12.eu\n", func() {
		confirmed = ConfirmByName("ns3168421.ip-51-77-12.eu", "this wipes both disks")
	})

	td.Cmp(t, confirmed, true)
}

// Surrounding whitespace is the operator's, not the server's.
func TestConfirmByName_AcceptsTheNameWithSurroundingSpaces(t *testing.T) {
	var confirmed bool
	withInteractiveInput(t, "  ns3168421.ip-51-77-12.eu  \n", func() {
		confirmed = ConfirmByName("ns3168421.ip-51-77-12.eu", "this wipes both disks")
	})

	td.Cmp(t, confirmed, true)
}

// Anything else is a refusal: a near miss must not pass.
func TestConfirmByName_RejectsAnythingElse(t *testing.T) {
	for _, answer := range []string{
		"ns3168421.ip-51-77-12.e\n",  // one character short
		"ns3168422.ip-51-77-12.eu\n", // the neighbouring server
		"yes\n",
		"\n",
		"", // stream closed before any answer
	} {
		var confirmed bool
		withInteractiveInput(t, answer, func() {
			confirmed = ConfirmByName("ns3168421.ip-51-77-12.eu", "this wipes both disks")
		})

		td.Cmp(t, confirmed, false, "answer %q must not confirm", answer)
	}
}

// Without a terminal there is no prompt at all: an unattended run has to opt
// in with --yes rather than be asked a question nobody will read.
func TestConfirmByName_RefusesWhenNotInteractive(t *testing.T) {
	origInput, origInteractive := confirmInput, confirmInteractive
	confirmInput = strings.NewReader("ns3168421.ip-51-77-12.eu\n")
	confirmInteractive = func() bool { return false }
	defer func() { confirmInput, confirmInteractive = origInput, origInteractive }()

	td.Cmp(t, ConfirmByName("ns3168421.ip-51-77-12.eu", "this wipes both disks"), false)
}

// A reboot asks for a yes, not for the server's name: the friction has to match
// the risk or it gets bypassed by habit.
func TestConfirmYesNo_AcceptsYes(t *testing.T) {
	for _, answer := range []string{"y\n", "Y\n", "yes\n", "  YES  \n"} {
		var confirmed bool
		withInteractiveInput(t, answer, func() {
			confirmed = ConfirmYesNo("this reboots the server")
		})
		td.Cmp(t, confirmed, true, "answer %q must confirm", answer)
	}
}

// Everything else declines, including the empty answer that a hurried Enter
// produces — the prompt reads [y/N] and must mean it.
func TestConfirmYesNo_RefusesAnythingElse(t *testing.T) {
	for _, answer := range []string{"\n", "n\n", "no\n", "ya\n", "sure\n", "yes please\n"} {
		var confirmed bool
		withInteractiveInput(t, answer, func() {
			confirmed = ConfirmYesNo("this reboots the server")
		})
		td.Cmp(t, confirmed, false, "answer %q must decline", answer)
	}
}

// Non-interactive means a pipeline, and a pipeline that did not say --yes did
// not consent.
func TestConfirmYesNo_RefusesWhenNotInteractive(t *testing.T) {
	origInteractive := confirmInteractive
	confirmInteractive = func() bool { return false }
	defer func() { confirmInteractive = origInteractive }()

	td.Cmp(t, ConfirmYesNo("this reboots the server"), false)
}

// Ctrl-D after typing is still an answer. bufio reports end of file alongside
// the text it did read, and treating that as a failure threw away an answer
// the operator had plainly given, then declined without saying why.
func TestConfirmYesNo_HonoursAnAnswerEndedByEOF(t *testing.T) {
	var confirmed bool
	withInteractiveInput(t, "y", func() {
		confirmed = ConfirmYesNo("this reboots the server")
	})

	td.Cmp(t, confirmed, true)
}

func TestConfirmByName_HonoursAnAnswerEndedByEOF(t *testing.T) {
	var confirmed bool
	withInteractiveInput(t, "ns3168421.ip-51-77-12.eu", func() {
		confirmed = ConfirmByName("ns3168421.ip-51-77-12.eu", "this wipes both disks")
	})

	td.Cmp(t, confirmed, true)
}

// Reaching the end of the stream with nothing typed is not a decline anybody
// made, and the operator saw a prompt they could not answer. Say so.
func TestConfirmYesNo_SaysSoWhenNothingCouldBeRead(t *testing.T) {
	var confirmed bool
	printed := withInteractiveIO(t, "", func() {
		confirmed = ConfirmYesNo("this reboots the server")
	})

	td.Cmp(t, confirmed, false)
	td.Cmp(t, strings.Contains(printed, "No answer could be read"), true,
		"the operator has to be told, got %q", printed)
}

// The wording is the guard. It names the risk before the question, and it is
// the one part a test could not reach while only the reader was injectable.
func TestConfirmYesNo_AsksBeforeItReads(t *testing.T) {
	printed := withInteractiveIO(t, "n\n", func() {
		ConfirmYesNo("this reboots ns3168421.ip-51-77-12.eu")
	})

	td.Cmp(t, strings.Contains(printed, "this reboots ns3168421.ip-51-77-12.eu"), true,
		"the warning has to reach the operator, got %q", printed)
	td.Cmp(t, strings.Contains(printed, "[y/N]"), true,
		"the default has to be visible, got %q", printed)
	td.Cmp(t, strings.HasSuffix(printed, "› "), true,
		"the prompt must not end with a newline, or the answer lands on the next line: %q", printed)
}

// ConfirmByName asks for the name itself, and shows which one.
func TestConfirmByName_AsksForTheExactName(t *testing.T) {
	printed := withInteractiveIO(t, "\n", func() {
		ConfirmByName("ns3168421.ip-51-77-12.eu", "this wipes both disks")
	})

	td.Cmp(t, strings.Contains(printed, `Type "ns3168421.ip-51-77-12.eu" to confirm`), true,
		"the operator has to be told what to type, got %q", printed)
}
