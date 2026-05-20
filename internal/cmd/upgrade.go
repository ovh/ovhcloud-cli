// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ovh/ovhcloud-cli/internal/upgrade"
	"github.com/ovh/ovhcloud-cli/internal/version"
)

const installDocURL = "https://github.com/ovh/ovhcloud-cli#installation"

// wrapPermissionError augments permission-denied errors with a suggestion to
// retry with elevated privileges or reinstall following the documentation.
func wrapPermissionError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, os.ErrPermission) {
		return fmt.Errorf("%w\n\nRetry with sudo, or reinstall following the documentation: %s", err, installDocURL)
	}
	return err
}

var upgradeAssumeYes bool

func init() {
	if runtime.GOARCH == "wasm" && runtime.GOOS == "js" {
		return
	}

	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade OVHcloud CLI to the latest version",
		RunE:  runUpgrade,
	}
	upgradeCmd.Flags().BoolVarP(&upgradeAssumeYes, "yes", "y", false, "Skip confirmation prompt")
	// Skip parent PersistentPreRun (no API client needed).
	upgradeCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {}

	rootCmd.AddCommand(upgradeCmd)
}

func runUpgrade(cmd *cobra.Command, _ []string) error {
	if version.Version == "undefined" {
		return errors.New("upgrade is not available in development builds")
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()

	tag, err := upgrade.LatestTag(ctx)
	if err != nil {
		return err
	}

	if tag == version.Version {
		fmt.Fprintf(cmd.OutOrStdout(), "Already on latest (%s)\n", tag)
		return nil
	}

	method := upgrade.DetectInstallMethod()
	switch method {
	case upgrade.MethodBrew:
		fmt.Fprintf(cmd.OutOrStdout(), "Upgrade via Homebrew:\n\n    brew upgrade --cask ovh/tap/ovhcloud-cli\n\n")
		return nil
	case upgrade.MethodGoInstall:
		fmt.Fprintf(cmd.OutOrStdout(), "Upgrade via go install:\n\n    go install github.com/ovh/ovhcloud-cli/cmd/ovhcloud@latest\n\n")
		return nil
	}

	if runtime.GOOS == "windows" {
		fmt.Fprintf(cmd.OutOrStdout(), "Automatic upgrade on Windows is not supported. Download the latest release from:\n\n    https://github.com/ovh/ovhcloud-cli/releases/tag/%s\n\n", tag)
		return nil
	}

	exe, err := upgrade.ResolveExecutable()
	if err != nil {
		return err
	}

	if err := upgrade.CheckWritable(exe); err != nil {
		return wrapPermissionError(err)
	}

	if !upgradeAssumeYes {
		fmt.Fprintf(cmd.OutOrStdout(), "Replace %s (%s) with %s? [y/N] ", exe, version.Version, tag)
		reader := bufio.NewReader(cmd.InOrStdin())
		line, _ := reader.ReadString('\n')
		answer := strings.TrimSpace(strings.ToLower(line))
		if answer != "y" && answer != "yes" {
			fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
			return nil
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Downloading %s...\n", tag)
	if err := upgrade.SelfReplace(ctx, tag, exe); err != nil {
		return wrapPermissionError(err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Installed %s at %s\n", tag, exe)
	return nil
}
