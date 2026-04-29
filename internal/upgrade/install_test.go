// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package upgrade

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func buildTarGz(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0o755,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestSelfReplace(t *testing.T) {
	tarball := buildTarGz(t, map[string][]byte{
		"ovhcloud": []byte("NEW_BINARY_CONTENTS"),
	})

	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/gzip")
		w.Write(tarball)
	}))
	defer server.Close()

	target := filepath.Join(t.TempDir(), "ovhcloud")
	if err := os.WriteFile(target, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := selfReplace(context.Background(), server.Client(), server.URL, "v1.4.2", target, "linux", "amd64")
	if err != nil {
		t.Fatalf("selfReplace: %v", err)
	}

	if !strings.HasSuffix(requestedPath, "/v1.4.2/ovhcloud-cli_Linux_x86_64.tar.gz") {
		t.Errorf("requested %q, want suffix /v1.4.2/ovhcloud-cli_Linux_x86_64.tar.gz", requestedPath)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "NEW_BINARY_CONTENTS" {
		t.Errorf("target contents = %q, want NEW_BINARY_CONTENTS", got)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("mode = %v, want 0755", info.Mode().Perm())
	}
}

func TestAssetName(t *testing.T) {
	cases := []struct {
		goos, goarch string
		want         string
	}{
		{"linux", "amd64", "ovhcloud-cli_Linux_x86_64.tar.gz"},
		{"linux", "arm64", "ovhcloud-cli_Linux_arm64.tar.gz"},
		{"linux", "386", "ovhcloud-cli_Linux_i386.tar.gz"},
		{"darwin", "amd64", "ovhcloud-cli_Darwin_x86_64.tar.gz"},
		{"darwin", "arm64", "ovhcloud-cli_Darwin_arm64.tar.gz"},
	}
	for _, tc := range cases {
		t.Run(tc.goos+"_"+tc.goarch, func(t *testing.T) {
			got := assetName(tc.goos, tc.goarch)
			if got != tc.want {
				t.Errorf("assetName(%s,%s) = %q, want %q", tc.goos, tc.goarch, got, tc.want)
			}
		})
	}
}

func TestCheckWritable(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "ovhcloud")
	if err := os.WriteFile(target, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := CheckWritable(target); err != nil {
		t.Fatalf("writable target: got err %v, want nil", err)
	}
}

func TestSelfReplaceHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	target := filepath.Join(t.TempDir(), "ovhcloud")
	os.WriteFile(target, []byte("OLD"), 0o755)

	err := selfReplace(context.Background(), server.Client(), server.URL, "v1.4.2", target, "linux", "amd64")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	got, _ := os.ReadFile(target)
	if string(got) != "OLD" {
		t.Errorf("target clobbered on error: %q", got)
	}
}
