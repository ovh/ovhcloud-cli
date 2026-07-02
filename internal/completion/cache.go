// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package completion

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// completionCacheTTL is how long cached completion results remain valid.
//
// Shell completion runs as a brand new, short-lived process on every <tab>, so
// an in-memory cache would be useless: results are cached on disk to avoid
// hitting the API on every keystroke.
const completionCacheTTL = 10 * time.Minute

// completionCacheDir returns the directory where completion results are cached,
// honouring XDG_CACHE_HOME. It returns "" if it cannot be determined.
func completionCacheDir() string {
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		base = filepath.Join(home, ".cache")
	}
	return filepath.Join(base, "ovhcloud", "completion")
}

// readCachedSuggestions returns the cached suggestions for the given key when a
// cache file exists and is younger than completionCacheTTL.
func readCachedSuggestions(key string) ([]string, bool) {
	dir := completionCacheDir()
	if dir == "" {
		return nil, false
	}

	path := filepath.Join(dir, key)
	info, err := os.Stat(path)
	if err != nil || time.Since(info.ModTime()) > completionCacheTTL {
		return nil, false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	trimmed := strings.Trim(string(data), "\n")
	if trimmed == "" {
		return []string{}, true
	}
	return strings.Split(trimmed, "\n"), true
}

// writeCachedSuggestions stores suggestions for the given key. It is best-effort:
// any error (no cache dir, write failure...) is silently ignored so completion
// never fails because of the cache. The write is atomic (temp file + rename).
func writeCachedSuggestions(key string, suggestions []string) {
	dir := completionCacheDir()
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}

	tmp, err := os.CreateTemp(dir, key+".tmp-*")
	if err != nil {
		return
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(strings.Join(suggestions, "\n")); err != nil {
		tmp.Close()
		return
	}
	if err := tmp.Close(); err != nil {
		return
	}

	_ = os.Rename(tmp.Name(), filepath.Join(dir, key))
}
