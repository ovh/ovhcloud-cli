// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package completion

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCache_WriteThenRead(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	want := []string{"proj-1\tProd", "proj-2\tPreprod"}
	writeCachedSuggestions("cloud-projects", want)

	got, ok := readCachedSuggestions("cloud-projects")
	if !ok {
		t.Fatal("expected cache hit right after write")
	}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCache_MissWhenAbsent(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	if _, ok := readCachedSuggestions("cloud-projects"); ok {
		t.Error("expected a cache miss when no cache file exists")
	}
}

func TestCache_ExpiredByTTL(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	writeCachedSuggestions("cloud-projects", []string{"proj-1"})

	// Make the cache file older than the TTL.
	path := filepath.Join(completionCacheDir(), "cloud-projects")
	old := time.Now().Add(-completionCacheTTL - time.Minute)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatalf("failed to age cache file: %v", err)
	}

	if _, ok := readCachedSuggestions("cloud-projects"); ok {
		t.Error("expected a cache miss for an expired cache file")
	}
}

func TestCache_EmptyResultIsCached(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	writeCachedSuggestions("cloud-projects", []string{})

	got, ok := readCachedSuggestions("cloud-projects")
	if !ok {
		t.Fatal("expected a cache hit for an empty (but valid) cached result")
	}
	if len(got) != 0 {
		t.Errorf("expected no suggestions, got %q", got)
	}
}
