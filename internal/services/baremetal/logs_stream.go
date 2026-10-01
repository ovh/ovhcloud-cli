// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/cache"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// Subscribing a server to its logs means naming a Graylog stream, and a stream
// is named by a UUID. Nobody has one to hand: this account carries 25 Log Data
// Platform services and 59 streams between them, one service holding 21 on its
// own, and the streams are how a person thinks of them — "Prestashop nginx logs
// (filebeat)", "datastream_test". So a title is accepted and resolved here,
// exactly as `vrack attach` resolves a server to its interface and `baremetal
// traffic` resolves it to its network controllers.
//
// The sweep costs 85 requests. Measured on 20 August 2026 against the real
// account: 2.15 seconds with the parallelism below. That is cheap enough that
// none of it is cached — a subscription is a write, and a stale identifier
// resolved from a cache would send a machine's logs to the wrong place. The
// completion helper caches its suggestions, because a stale suggestion is only
// a suggestion.

// ldpStream is one Graylog stream and the Log Data Platform service holding it.
//
// The service travels with the stream because everything downstream needs it:
// the subscription is created against the server but its operation is followed
// on the LDP service, and a title alone does not say which service that is.
type ldpStream struct {
	ServiceName string
	StreamID    string
	Title       string
}

// listLdpStreams reads every stream of every LDP service on the account.
func listLdpStreams() ([]ldpStream, error) {
	var services []string
	if err := httpLib.Client.Get("/v1/dbaas/logs", &services); err != nil {
		return nil, fmt.Errorf("failed to list the Log Data Platform services: %w", err)
	}

	if len(services) == 0 {
		return nil, nil
	}

	// Two rounds, both bounded: the stream identifiers of each service, then
	// the object behind each identifier, which is the only place the title is.
	type pair struct{ service, stream string }

	var (
		mutex sync.Mutex
		pairs []pair
	)

	listing := new(errgroup.Group)
	listing.SetLimit(10)
	for _, service := range services {
		listing.Go(func() error {
			var ids []string
			path := fmt.Sprintf("/v1/dbaas/logs/%s/output/graylog/stream", url.PathEscape(service))
			if err := httpLib.Client.Get(path, &ids); err != nil {
				return fmt.Errorf("failed to list the streams of %s: %w", service, err)
			}

			mutex.Lock()
			defer mutex.Unlock()
			for _, id := range ids {
				pairs = append(pairs, pair{service: service, stream: id})
			}

			return nil
		})
	}
	if err := listing.Wait(); err != nil {
		return nil, err
	}

	streams := make([]ldpStream, len(pairs))
	reading := new(errgroup.Group)
	reading.SetLimit(10)
	for index, p := range pairs {
		reading.Go(func() error {
			var object struct {
				StreamID string `json:"streamId"`
				Title    string `json:"title"`
			}

			path := fmt.Sprintf("/v1/dbaas/logs/%s/output/graylog/stream/%s",
				url.PathEscape(p.service), url.PathEscape(p.stream))
			if err := httpLib.Client.Get(path, &object); err != nil {
				return fmt.Errorf("failed to read the stream %s of %s: %w", p.stream, p.service, err)
			}

			streams[index] = ldpStream{
				ServiceName: p.service,
				StreamID:    object.StreamID,
				Title:       object.Title,
			}

			return nil
		})
	}
	if err := reading.Wait(); err != nil {
		return nil, err
	}

	sort.Slice(streams, func(i, j int) bool {
		if streams[i].Title != streams[j].Title {
			return streams[i].Title < streams[j].Title
		}
		return streams[i].ServiceName < streams[j].ServiceName
	})

	return streams, nil
}

// resolveStream turns what the operator typed into a stream and its service.
//
// A UUID is taken as given and costs nothing: that is what -o json hands back,
// and re-reading 85 objects to confirm an identifier the caller already has
// would be work done to learn nothing. Anything else is a title, matched
// exactly — this account holds streams called "Test" and "test", so a
// case-insensitive match would silently pick one of two different streams.
func resolveStream(wanted string) (ldpStream, error) {
	if looksLikeUUID(wanted) {
		return ldpStream{StreamID: wanted}, nil
	}

	streams, err := listLdpStreams()
	if err != nil {
		return ldpStream{}, err
	}

	if len(streams) == 0 {
		return ldpStream{}, fmt.Errorf(
			"this account has no Log Data Platform service, so there is no stream to subscribe to.\n"+
				"   A subscription needs one; see `ovhcloud ldp list` once you have ordered it (looking for %q)", wanted)
	}

	var matches []ldpStream
	for _, stream := range streams {
		if stream.Title == wanted {
			matches = append(matches, stream)
		}
	}

	switch len(matches) {
	case 1:
		return matches[0], nil

	case 0:
		// Not `ovhcloud ldp list`: that lists Log Data Platform services, not
		// Graylog streams, and no command of this tree lists streams at all.
		// nearbyTitles already says how many there are and which ones are close,
		// so pointing at completion is the only thing here that works.
		return ldpStream{}, fmt.Errorf(
			"no stream is called %q. %s\n   Complete them with: ovhcloud baremetal logs subscribe <server> --stream <TAB>",
			wanted, nearbyTitles(streams, wanted))

	default:
		var lines []string
		for _, match := range matches {
			lines = append(lines, fmt.Sprintf("     %s  (service %s)", match.StreamID, match.ServiceName))
		}

		return ldpStream{}, fmt.Errorf(
			"%d streams are called %q, on different Log Data Platform services.\n"+
				"   Name the one you mean by its identifier:\n%s",
			len(matches), wanted, strings.Join(lines, "\n"))
	}
}

// nearbyTitles says what does exist, so a typo does not end in a bare refusal.
//
// It only offers titles that contain what was typed, or are contained by it: a
// list of 56 unrelated names would be noise, and an edit distance would suggest
// a stream that has nothing to do with the request.
func nearbyTitles(streams []ldpStream, wanted string) string {
	folded := strings.ToLower(wanted)

	var near []string
	for _, stream := range streams {
		title := strings.ToLower(stream.Title)
		if title == "" {
			continue
		}
		if strings.Contains(title, folded) || strings.Contains(folded, title) {
			near = append(near, fmt.Sprintf("%q (service %s)", stream.Title, stream.ServiceName))
		}
	}

	if len(near) == 0 {
		// A bare count names nothing, and the reviewer who hit this said so:
		// "no stream is called X. This account has 59 of them" leaves you with
		// no idea what a title even looks like here, and the completion hint
		// that follows needs a TAB key — which a script, a CI log or a web
		// console does not have. Three examples cost one line and show the
		// shape.
		return fmt.Sprintf("This account has %d of them, such as %s.",
			len(streams), strings.Join(someTitles(streams, 3), ", "))
	}

	// Five is enough to recognise a typo. The rest are counted rather than
	// dropped in silence: a list that stops without saying so reads as the
	// whole answer.
	const shown = 5
	if len(near) > shown {
		return fmt.Sprintf("Close to: %s, and %d more.", strings.Join(near[:shown], ", "), len(near)-shown)
	}

	return "Close to: " + strings.Join(near, ", ") + "."
}

// someTitles picks the first few titles that exist, to show what one looks
// like. Sorted, so the same account gives the same examples twice running: an
// error message that changes between two identical calls reads as instability.
func someTitles(streams []ldpStream, count int) []string {
	var titles []string
	for _, stream := range streams {
		if stream.Title != "" {
			titles = append(titles, fmt.Sprintf("%q", stream.Title))
		}
	}
	sort.Strings(titles)
	if len(titles) > count {
		titles = titles[:count]
	}
	return titles
}

// looksLikeUUID recognises the shape of a stream identifier.
//
// It is deliberately a shape test and not a parse: the point is to tell "the
// caller gave me an identifier" from "the caller gave me a title", and a title
// that happens to be 36 characters of hex and dashes is not a case worth a
// dependency.
func looksLikeUUID(candidate string) bool {
	if len(candidate) != 36 {
		return false
	}

	for index, char := range candidate {
		switch index {
		case 8, 13, 18, 23:
			if char != '-' {
				return false
			}
		default:
			isDigit := char >= '0' && char <= '9'
			isLower := char >= 'a' && char <= 'f'
			isUpper := char >= 'A' && char <= 'F'
			if !isDigit && !isLower && !isUpper {
				return false
			}
		}
	}

	return true
}

// CompleteLogStream suggests the titles of the streams on the account.
//
// This is the one place the sweep is cached. A suggestion that is ten minutes
// stale costs a <tab> that offers a stream somebody deleted; the resolution
// above is not cached, because there the same staleness would send a machine's
// logs to the wrong place.
func CompleteLogStream(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	const (
		namespace = "completion"
		key       = "baremetal-log-stream-titles"
		ttl       = 10 * time.Minute
	)

	if data, found := cache.Read(namespace, key, ttl); found {
		trimmed := strings.Trim(string(data), "\n")
		if trimmed == "" {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return strings.Split(trimmed, "\n"), cobra.ShellCompDirectiveNoFileComp
	}

	streams, err := listLdpStreams()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// A title with a space in it cannot be offered as a bare word, and a title
	// with a newline in it would corrupt the cache file. Both are dropped
	// rather than mangled: completion is a convenience, and a suggestion that
	// does not work when accepted is worse than no suggestion.
	var titles []string
	for _, stream := range streams {
		if stream.Title == "" || strings.ContainsAny(stream.Title, " \t\n") {
			continue
		}
		titles = append(titles, stream.Title)
	}

	cache.Write(namespace, key, []byte(strings.Join(titles, "\n")), ttl)

	return titles, cobra.ShellCompDirectiveNoFileComp
}

// CompleteLogKind suggests the kinds of log the server being named can send.
func CompleteLogKind(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	kinds, err := logKindsOf(args[0])
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return kinds, cobra.ShellCompDirectiveNoFileComp
}
