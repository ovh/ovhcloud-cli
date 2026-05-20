// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package upgrade

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckWritableReadOnlyDir(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root, permission checks bypassed")
	}

	readOnlyDir := filepath.Join(t.TempDir(), "ro")
	if err := os.Mkdir(readOnlyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	roTarget := filepath.Join(readOnlyDir, "ovhcloud")
	if err := os.WriteFile(roTarget, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(readOnlyDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(readOnlyDir, 0o755) })

	if err := CheckWritable(roTarget); err == nil {
		t.Fatal("read-only dir: got nil, want error")
	}
}
