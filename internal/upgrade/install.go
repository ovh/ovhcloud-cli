// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package upgrade

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const downloadBaseURL = "https://github.com/ovh/ovhcloud-cli/releases/download"

// CheckWritable verifies that the running process can replace the binary at
// targetPath by probing the parent directory. os.Rename only requires write
// permission on the parent directory, not on the target file itself.
func CheckWritable(targetPath string) error {
	dir := filepath.Dir(targetPath)
	probe, err := os.CreateTemp(dir, ".ovhcloud-upgrade-check-*")
	if err != nil {
		return fmt.Errorf("cannot write to %s: %w", dir, err)
	}
	probePath := probe.Name()
	probe.Close()
	os.Remove(probePath)
	return nil
}

// SelfReplace downloads the release asset for the given tag and replaces the
// binary at targetPath in place. Only supported on linux and darwin.
// Callers should resolve symlinks before invoking so the rename targets the
// real binary.
func SelfReplace(ctx context.Context, tag, targetPath string) error {
	if runtime.GOOS == "windows" {
		return errors.New("self-replace not supported on windows")
	}
	client := &http.Client{Timeout: 60 * time.Second}
	return selfReplace(ctx, client, downloadBaseURL, tag, targetPath, runtime.GOOS, runtime.GOARCH)
}

// ResolveExecutable returns the path of the current binary with symlinks
// resolved. Callers use the resolved path consistently for permission checks,
// user-facing messages, and the rename target.
func ResolveExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate current binary: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return exe, nil
}

func assetName(goos, goarch string) string {
	osName := strings.ToUpper(goos[:1]) + goos[1:]
	var arch string
	switch goarch {
	case "amd64":
		arch = "x86_64"
	case "386":
		arch = "i386"
	default:
		arch = goarch
	}
	return fmt.Sprintf("ovhcloud-cli_%s_%s.tar.gz", osName, arch)
}

func selfReplace(ctx context.Context, client *http.Client, baseURL, tag, targetPath, goos, goarch string) error {
	asset := assetName(goos, goarch)
	url := fmt.Sprintf("%s/%s/%s", strings.TrimRight(baseURL, "/"), tag, asset)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", asset, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: HTTP %d", asset, resp.StatusCode)
	}

	// Extract binary in the same directory as the target so os.Rename stays on
	// one filesystem.
	targetDir := filepath.Dir(targetPath)
	tmp, err := os.CreateTemp(targetDir, ".ovhcloud-upgrade-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if err := extractBinary(resp.Body, tmp); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o755); err != nil {
		tmp.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return fmt.Errorf("replace %s: %w", targetPath, err)
	}
	return nil
}

func extractBinary(r io.Reader, out io.Writer) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return errors.New("ovhcloud binary not found in archive")
			}
			return fmt.Errorf("tar next: %w", err)
		}
		if filepath.Base(hdr.Name) != "ovhcloud" {
			continue
		}
		if _, err := io.Copy(out, tr); err != nil {
			return fmt.Errorf("extract binary: %w", err)
		}
		return nil
	}
}
