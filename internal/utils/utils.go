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
// data only.
// Both indirections exist so that the confirmation path itself can be
// exercised by a test: it is the guard standing between a typo and an erased
// disk, and a guard nobody executes is a guard nobody has checked.
var (
	confirmInput       io.Reader = os.Stdin
	confirmInteractive           = func() bool {
		return !IsInputFromPipe() && term.IsTerminal(os.Stdin.Fd())
	}
)

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
		fmt.Fprintf(os.Stderr,
			"🛑 %s\n   Refusing to continue without a confirmation. Re-run with --yes to confirm, or --dry-run to preview.\n",
			warning)
		return false
	}

	fmt.Fprintf(os.Stderr, "⚠️  %s\n   Continue? [y/N] › ", warning)

	reader := bufio.NewReader(confirmInput)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

func ConfirmByName(expected, warning string) bool {
	if !confirmInteractive() {
		fmt.Fprintf(os.Stderr,
			"🛑 %s\n   Refusing to continue without a confirmation. Re-run with --yes to confirm, or --dry-run to preview.\n",
			warning)
		return false
	}

	fmt.Fprintf(os.Stderr, "⚠️  %s\n   Type %q to confirm › ", warning, expected)

	reader := bufio.NewReader(confirmInput)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	return strings.TrimSpace(answer) == expected
}
