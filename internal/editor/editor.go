// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package editor

import (
	"fmt"
	"os"
	"os/exec"
)

const defaultEditor = "vi"

func EditValueWithEditor(value []byte) ([]byte, error) {
	editor := defaultEditor
	if s := os.Getenv("EDITOR"); s != "" {
		editor = s
	}

	// Create temp file
	f, err := os.CreateTemp("", "ovh-cli-edit")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(value); err != nil {
		return nil, fmt.Errorf("failed to write input file: %w", err)
	}

	// Open editor
	cmd := exec.Command("sh", "-c", editor+" "+f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to edit file: %w", err)
	}

	// Read updated file
	b, err := os.ReadFile(f.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read updated file: %w", err)
	}

	return b, nil
}
