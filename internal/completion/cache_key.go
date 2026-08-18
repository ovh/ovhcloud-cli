// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package completion

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/flags"
)

// cacheKeyFor derives the cache file name of an endpoint's suggestions.
//
// The key is scoped to the active profile as well as to the endpoint: profiles
// point at different accounts, and a cache shared between them would suggest,
// for ten minutes, the identifiers of a fleet the user is no longer working on.
//
// The readable prefix is there so that the cache directory can be inspected by
// hand; the digest is what makes the key unambiguous, since the prefix drops
// every character a file name cannot carry.
func cacheKeyFor(endpoint string) string {
	digest := sha256.Sum256([]byte(flags.Profile + "\x00" + endpoint))

	return fmt.Sprintf("%s-%x", slugify(endpoint), digest[:8])
}

// slugify keeps the letters and digits of an endpoint and collapses everything
// else into single dashes, so the result is safe as a file name on every
// platform. It is truncated because a path is bounded and an endpoint is not.
func slugify(endpoint string) string {
	var out strings.Builder
	lastWasDash := true // avoids a leading dash

	for _, r := range strings.ToLower(endpoint) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			out.WriteRune(r)
			lastWasDash = false
		case !lastWasDash:
			out.WriteByte('-')
			lastWasDash = true
		}
	}

	slug := strings.Trim(out.String(), "-")
	if len(slug) > 60 {
		slug = strings.Trim(slug[:60], "-")
	}
	if slug == "" {
		slug = "endpoint"
	}

	return slug
}
