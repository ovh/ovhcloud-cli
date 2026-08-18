// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package completion

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// cacheKeyFor derives the cache file name of an endpoint's suggestions.
//
// The key is scoped to the account the client actually resolved to, not merely
// to --profile: OVH_ENDPOINT and OVH_CONSUMER_KEY override the profile, and in
// legacy mode there is no profile at all. Keying on the profile alone would let
// two accounts, or two regions, share one cache entry and suggest for ten
// minutes the identifiers of a fleet the user is not working on.
//
// Credentials only ever enter the SHA-256 digest, so nothing readable is
// written to disk. The readable prefix is the endpoint alone, so that the cache
// directory can still be inspected by hand.
func cacheKeyFor(endpoint string) string {
	digest := sha256.Sum256([]byte(strings.Join(
		append([]string{flags.Profile, endpoint}, clientIdentity()...), "\x00")))

	return fmt.Sprintf("%s-%x", slugify(endpoint), digest[:8])
}

// clientIdentity returns what pins the API client to one account and one
// region. The access token is deliberately left out: it rotates, and including
// it would throw the cache away on every refresh for no gain in isolation.
func clientIdentity() []string {
	if httpLib.Client == nil {
		return nil
	}

	return []string{
		httpLib.Client.Endpoint(),
		httpLib.Client.AppKey,
		httpLib.Client.ConsumerKey,
		httpLib.Client.ClientID,
	}
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
