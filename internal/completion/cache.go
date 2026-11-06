// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package completion

import (
	"strings"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/cache"
)

// completionCacheTTL is how long cached completion results remain valid.
//
// Shell completion runs as a brand new, short-lived process on every <tab>, so
// an in-memory cache would be useless: results are cached on disk to avoid
// hitting the API on every keystroke.
const completionCacheTTL = 10 * time.Minute

// completionCacheNamespace is the subdirectory holding these entries. Keeping
// each kind of cached result in its own namespace means one of them can be
// wiped, or given a different lifetime, without touching the others.
const completionCacheNamespace = "completion"

// completionCacheDir returns the directory where completion results are cached,
// honouring XDG_CACHE_HOME. It returns "" if it cannot be determined.
func completionCacheDir() string {
	return cache.Dir(completionCacheNamespace)
}

func readCachedSuggestions(key string) ([]string, bool) {
	data, found := cache.Read(completionCacheNamespace, key, completionCacheTTL)
	if !found {
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
	cache.Write(completionCacheNamespace, key, []byte(strings.Join(suggestions, "\n")), completionCacheTTL)
}
