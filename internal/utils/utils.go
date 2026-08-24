// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"dario.cat/mergo"
	"github.com/charmbracelet/x/term"
)

func MergeMaps(left, right map[string]any) error {
	if err := mergo.Merge(&left, right, mergo.WithOverride, mergo.WithAppendSlice); err != nil {
		return fmt.Errorf("error merging maps: %w", err)
	}

	return nil
}

func IsInputFromPipe() bool {
	if runtime.GOARCH == "wasm" && runtime.GOOS == "js" {
		return false
	}

	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}

// ConfirmByName asks the operator to type the given name before a destructive
// action is carried out. It returns false — without prompting — when the input
// is not an interactive terminal, so an unattended run never destroys anything
// by default: such runs must opt in explicitly.
//
// The prompt is written to stderr so that a redirected stdout keeps holding
// data only. It deliberately does not go through internal/display: that
// package renders a *result* in the format the caller asked for, and every
// path through it ends in fmt.Println on stdout, JSON-encodes the message
// under -o json, overwrites the ResultString global, and — for a warning or an
// error — calls ExitFunc. A question is none of those things, and one that
// ended the process before reading the answer would not be a question.
//
// These indirections exist so that the confirmation path itself can be
// exercised by a test: it is the guard standing between a typo and an erased
// disk, and a guard nobody executes is a guard nobody has checked. The writer
// is injectable for the same reason as the reader — the wording an operator
// reads before destroying something is the part worth pinning, and while only
// the reader could be swapped that wording was the one thing a test could not
// see.
var (
	confirmInput       io.Reader = os.Stdin
	confirmOutput      io.Writer = os.Stderr
	confirmInteractive           = IsInteractiveTerminal
)

// readAnswer reads one typed line.
//
// A terminal can reach end of file with a line already typed — somebody who
// answers and then presses Ctrl-D rather than Enter — and bufio reports that
// as an error alongside the text. Discarding the text on any error threw such
// an answer away and refused in silence, which reads to the operator as the
// command ignoring them. Anything actually typed is honoured; only an empty
// read is a failure to answer, and that one is now said out loud instead of
// looking like a decline nobody made.
func readAnswer() (string, bool) {
	line, err := bufio.NewReader(confirmInput).ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		fmt.Fprintf(confirmOutput, "\n🛑 No answer could be read, so nothing was done.\n")
		return "", false
	}

	return strings.TrimSpace(line), true
}

// IsInteractiveTerminal reports whether there is somebody at a keyboard.
//
// Everything that asks a question has to agree on the answer, so there is one
// definition of it rather than one per caller: a confirmation that refuses and
// a picker that crashes were two readings of the same situation.
func IsInteractiveTerminal() bool {
	return !IsInputFromPipe() && term.IsTerminal(os.Stdin.Fd())
}

// ConfirmYesNo asks for a plain yes before an action that interrupts a running
// service without losing anything. Typing the resource name, as ConfirmByName
// requires, is the right friction for an irreversible act and the wrong one for
// a reboot: a guardrail that is tiresome out of proportion to its risk is a
// guardrail people learn to bypass.
//
// Anything other than "y" or "yes" declines, and a non-interactive session
// declines rather than guessing.
func ConfirmYesNo(warning string) bool {
	if !confirmInteractive() {
		fmt.Fprintf(confirmOutput,
			"🛑 %s\n   Refusing to continue without a confirmation. Re-run with --yes to confirm, or --dry-run to preview.\n",
			warning)
		return false
	}

	fmt.Fprintf(confirmOutput, "⚠️  %s\n   Continue? [y/N] › ", warning)

	answer, ok := readAnswer()
	if !ok {
		return false
	}

	switch strings.ToLower(answer) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

func ConfirmByName(expected, warning string) bool {
	if !confirmInteractive() {
		fmt.Fprintf(confirmOutput,
			"🛑 %s\n   Refusing to continue without a confirmation. Re-run with --yes to confirm, or --dry-run to preview.\n",
			warning)
		return false
	}

	fmt.Fprintf(confirmOutput, "⚠️  %s\n   Type %q to confirm › ", warning, expected)

	answer, ok := readAnswer()
	if !ok {
		return false
	}

	return answer == expected
}
