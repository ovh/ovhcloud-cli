// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupHome isolates HOME (and clears XDG_CONFIG_HOME) so completion install
// writes into a throwaway directory instead of the real user home.
func setupHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	return home
}

func TestCompletionInstall_Bash(t *testing.T) {
	home := setupHome(t)
	t.Setenv("SHELL", "/usr/bin/bash")

	if err := runCompletionInstall(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(home, ".bashrc"))
	if err != nil {
		t.Fatalf(".bashrc was not created: %v", err)
	}
	if !strings.Contains(string(content), `eval "$(ovhcloud completion bash)"`) {
		t.Errorf(".bashrc does not contain the bash completion line:\n%s", content)
	}
}

func TestCompletionInstall_Zsh(t *testing.T) {
	home := setupHome(t)
	t.Setenv("SHELL", "/bin/zsh")

	if err := runCompletionInstall(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(home, ".zshrc"))
	if err != nil {
		t.Fatalf(".zshrc was not created: %v", err)
	}
	if !strings.Contains(string(content), `eval "$(ovhcloud completion zsh)"`) {
		t.Errorf(".zshrc does not contain the zsh completion line:\n%s", content)
	}
}

func TestCompletionInstall_Idempotent(t *testing.T) {
	home := setupHome(t)
	t.Setenv("SHELL", "/bin/bash")

	for i := 0; i < 2; i++ {
		if err := runCompletionInstall(nil, nil); err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
	}

	content, err := os.ReadFile(filepath.Join(home, ".bashrc"))
	if err != nil {
		t.Fatalf(".bashrc was not created: %v", err)
	}
	if n := strings.Count(string(content), `eval "$(ovhcloud completion bash)"`); n != 1 {
		t.Errorf("expected the completion line exactly once, got %d:\n%s", n, content)
	}
}

func TestCompletionInstall_PreservesExistingContent(t *testing.T) {
	home := setupHome(t)
	t.Setenv("SHELL", "/bin/zsh")

	rcFile := filepath.Join(home, ".zshrc")
	existing := "export EDITOR=vim\nalias ll='ls -la'\n"
	if err := os.WriteFile(rcFile, []byte(existing), 0o644); err != nil {
		t.Fatalf("failed to seed .zshrc: %v", err)
	}

	if err := runCompletionInstall(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(rcFile)
	if err != nil {
		t.Fatalf("failed to read .zshrc: %v", err)
	}
	if !strings.Contains(string(content), existing) {
		t.Errorf("existing .zshrc content was lost:\n%s", content)
	}
	if !strings.Contains(string(content), `eval "$(ovhcloud completion zsh)"`) {
		t.Errorf("completion line was not appended:\n%s", content)
	}
}

func TestCompletionInstall_FishWithXDG(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SHELL", "/usr/bin/fish")

	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	if err := runCompletionInstall(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	destFile := filepath.Join(xdg, "fish", "completions", "ovhcloud.fish")
	info, err := os.Stat(destFile)
	if err != nil {
		t.Fatalf("fish completion file was not created at %s: %v", destFile, err)
	}
	if info.Size() == 0 {
		t.Errorf("fish completion file is empty")
	}
}

func TestCompletionInstall_FishDefaultConfigDir(t *testing.T) {
	home := setupHome(t) // XDG_CONFIG_HOME cleared → falls back to ~/.config
	t.Setenv("SHELL", "/usr/bin/fish")

	if err := runCompletionInstall(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	destFile := filepath.Join(home, ".config", "fish", "completions", "ovhcloud.fish")
	if _, err := os.Stat(destFile); err != nil {
		t.Fatalf("fish completion file was not created at %s: %v", destFile, err)
	}
}

func TestCompletionInstall_UnsupportedShell(t *testing.T) {
	setupHome(t)
	t.Setenv("SHELL", "/usr/bin/tcsh")

	err := runCompletionInstall(nil, nil)
	if err == nil {
		t.Fatal("expected an error for an unsupported shell, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("unexpected error message: %v", err)
	}
}
