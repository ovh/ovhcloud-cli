// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

// Package cache stores small results on disk between runs of the CLI.
//
// Every invocation is a brand new process, so an in-memory cache buys nothing:
// shell completion runs one process per <tab>, and the catalogue is fetched
// again by every command that needs a price. What both need is for the answer
// to outlive the process that fetched it.
//
// Nothing here ever fails a caller. A cache that returns an error is worse than
// no cache: the command it serves would start failing for a reason that has
// nothing to do with what the operator asked.
package cache

import (
	"os"
	"path/filepath"
	"time"
)

// Dir returns the directory holding one namespace, honouring XDG_CACHE_HOME.
// It returns "" when no location can be determined, which every caller reads
// as "no cache available".
func Dir(namespace string) string {
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		base = filepath.Join(home, ".cache")
	}
	return filepath.Join(base, "ovhcloud", namespace)
}

// Read returns the entry stored under key if it exists and is younger than ttl.
// An expired entry is removed on the way out, so a cache that stops being read
// does not grow for ever.
func Read(namespace, key string, ttl time.Duration) ([]byte, bool) {
	dir := Dir(namespace)
	if dir == "" {
		return nil, false
	}

	path := filepath.Join(dir, key)
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	if time.Since(info.ModTime()) > ttl {
		_ = os.Remove(path)
		return nil, false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return data, true
}

// Write stores data under key. The write goes through a temporary file and a
// rename, so a reader never sees a half-written entry — and a second process
// writing the same key at the same time leaves one whole entry rather than two
// interleaved halves.
//
// It is best-effort: no cache directory, a full disk or a failed write are all
// silently ignored.
func Write(namespace, key string, data []byte, ttl time.Duration) {
	dir := Dir(namespace)
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	PurgeExpired(namespace, ttl)

	tmp, err := os.CreateTemp(dir, key+".tmp-*")
	if err != nil {
		return
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return
	}
	if err := tmp.Close(); err != nil {
		return
	}

	_ = os.Rename(tmp.Name(), filepath.Join(dir, key))
}

// PurgeExpired removes every entry of a namespace older than ttl.
func PurgeExpired(namespace string, ttl time.Duration) {
	dir := Dir(namespace)
	if dir == "" {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if time.Since(info.ModTime()) > ttl {
			_ = os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
}
