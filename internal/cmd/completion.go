// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for ovhcloud CLI.

To load completions in your current shell session:

  bash:
    source <(ovhcloud completion bash)

  zsh:
    source <(ovhcloud completion zsh)

  fish:
    ovhcloud completion fish | source

To make completions permanent, run:

  ovhcloud completion install
`,
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}

	completionCmd.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Install shell completion permanently in your shell rc file",
		RunE:  runCompletionInstall,
	})

	rootCmd.AddCommand(completionCmd)
}

func runCompletionInstall(_ *cobra.Command, _ []string) error {
	shell := os.Getenv("SHELL")
	shellName := filepath.Base(shell)

	var rcFile string
	var completionLine string
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine user home directory: %w", err)
	}

	switch shellName {
	case "bash":
		rcFile = filepath.Join(home, ".bashrc")
		completionLine = `eval "$(ovhcloud completion bash)"`
	case "zsh":
		rcFile = filepath.Join(home, ".zshrc")
		completionLine = `eval "$(ovhcloud completion zsh)"`
	case "fish":
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(home, ".config")
		}
		fishCompDir := filepath.Join(configDir, "fish", "completions")
		if err := os.MkdirAll(fishCompDir, 0755); err != nil {
			return fmt.Errorf("failed to create fish completions dir: %w", err)
		}
		destFile := filepath.Join(fishCompDir, "ovhcloud.fish")
		f, err := os.Create(destFile)
		if err != nil {
			return fmt.Errorf("failed to create fish completion file: %w", err)
		}
		defer f.Close()
		if err := rootCmd.GenFishCompletion(f, true); err != nil {
			return fmt.Errorf("failed to write fish completion: %w", err)
		}
		fmt.Printf("✅ Fish completion installed to %s\n", destFile)
		return nil
	default:
		return fmt.Errorf("unsupported shell %q — please run 'ovhcloud completion bash|zsh|fish|powershell' manually", shellName)
	}

	// Check if already installed
	content, err := os.ReadFile(rcFile)
	if err == nil && strings.Contains(string(content), completionLine) {
		fmt.Printf("✅ Completion already installed in %s\n", rcFile)
		return nil
	}

	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", rcFile, err)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "\n# ovhcloud CLI shell completion\n%s\n", completionLine); err != nil {
		return fmt.Errorf("failed to write to %s: %w", rcFile, err)
	}

	fmt.Printf("✅ Completion installed in %s\n", rcFile)
	fmt.Printf("   Reload your shell or run: source %s\n", rcFile)
	return nil
}
