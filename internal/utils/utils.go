// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bufio"
	"fmt"
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
func ConfirmByName(expected, warning string) bool {
	if IsInputFromPipe() || !term.IsTerminal(os.Stdin.Fd()) {
		fmt.Fprintf(os.Stderr,
			"🛑 %s\n   Refusing to continue without a confirmation. Re-run with --yes to confirm, or --dry-run to preview.\n",
			warning)
		return false
	}

	fmt.Fprintf(os.Stderr, "⚠️  %s\n   Type %q to confirm › ", warning, expected)

	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	return strings.TrimSpace(answer) == expected
}
