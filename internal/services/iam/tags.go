// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

// The fine-grained routes exist because there is no way to remove a tag today,
// and because the way to add one is a read-modify-write.
//
// `iam resource edit --tag` does not destroy the other tags -- measured, and it
// was worth measuring, because the PUT it ends with does: posting two tags then
// PUTting one leaves exactly one. What saves it is that common.EditResource
// GETs the resource and merges before it PUTs. So the merge is done by the
// client, over two calls, and an edit landing between them is lost.
//
// What it cannot do at all is remove. Measured on a real resource:
//
//	edit --tag owner=x   with a shorter list -> nothing is removed, by design
//	edit --tag only=     with an empty value -> {"only": ""}, a tag that looks
//	                                            removed and is not
//
// So the two commands here are the only way to take a tag off a resource, and
// `set` adds one without reading the others first -- one POST per tag, no
// window for a concurrent write to fall into.
const resourcePath = "/v2/iam/resource"

// tagAssignment is one key=value pair given on the command line.
type tagAssignment struct {
	Key   string
	Value string
}

// parseTagAssignments reads key=value pairs. Only the first "=" separates, so
// a value may contain one.
func parseTagAssignments(given []string) ([]tagAssignment, error) {
	if len(given) == 0 {
		return nil, fmt.Errorf("no tag given; expected at least one key=value")
	}

	assignments := make([]tagAssignment, 0, len(given))
	seen := map[string]bool{}

	for _, raw := range given {
		key, value, found := strings.Cut(raw, "=")
		key = strings.TrimSpace(key)

		if !found {
			return nil, fmt.Errorf("%q is not a key=value pair", raw)
		}
		if key == "" {
			return nil, fmt.Errorf("%q has no key", raw)
		}
		if seen[key] {
			return nil, fmt.Errorf("%q is given twice, and the second would silently win", key)
		}

		seen[key] = true
		assignments = append(assignments, tagAssignment{Key: key, Value: value})
	}

	return assignments, nil
}

// SetResourceTags adds or updates tags without touching the others.
func SetResourceTags(_ *cobra.Command, args []string) {
	urn := args[0]

	assignments, err := parseTagAssignments(args[1:])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("%s/%s/tag", resourcePath, url.PathEscape(urn))

	if reportTagDryRun(endpoint, assignments) {
		return
	}

	// One call per tag: the API takes a single key/value pair, and posting
	// them one at a time is what keeps the others in place.
	for _, assignment := range assignments {
		body := map[string]any{"key": assignment.Key, "value": assignment.Value}
		if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
			display.OutputError(&flags.OutputFormatConfig,
				"failed to set tag %s on %s: %s", assignment.Key, urn, err)
			return
		}
	}

	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"urn": urn, "set": len(assignments)},
		"✅ %d tag(s) set on %s. The tags already there were left alone.", len(assignments), urn)
}

// RemoveResourceTags removes tags by key.
func RemoveResourceTags(_ *cobra.Command, args []string) {
	urn := args[0]
	keys := args[1:]

	if len(keys) == 0 {
		display.OutputError(&flags.OutputFormatConfig, "no key given; expected at least one tag key to remove")
		return
	}

	present, err := tagsOf(urn)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to read the tags of %s: %s\n   Check the URN with: ovhcloud iam resource list", urn, err)
		return
	}

	var missing []string
	for _, key := range keys {
		if _, set := present[key]; !set {
			missing = append(missing, key)
		}
	}

	// Removing a key that is not there is a no-op the API would answer 204 to,
	// which reads as success. Saying it is not there, and what is, is the
	// answer to the mistake actually made: a typo in the key.
	if len(missing) > 0 {
		display.OutputError(&flags.OutputFormatConfig,
			"%s carries no tag %s.\n   It carries: %s", urn, strings.Join(missing, ", "), knownKeys(present))
		return
	}

	calls := make([]common.Call, 0, len(keys))
	for _, key := range keys {
		calls = append(calls, common.Call{
			Method:   "DELETE",
			Endpoint: fmt.Sprintf("%s/%s/tag/%s", resourcePath, url.PathEscape(urn), url.PathEscape(key)),
		})
	}
	if common.ReportDryRun(calls...) {
		return
	}

	for _, call := range calls {
		if err := httpLib.Client.Delete(call.Endpoint, nil); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to remove a tag from %s: %s", urn, err)
			return
		}
	}

	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"urn": urn, "removed": len(keys)},
		"✅ %d tag(s) removed from %s. The others were left alone.", len(keys), urn)
}

// tagsOf reads what a resource currently carries.
func tagsOf(urn string) (map[string]any, error) {
	var resource struct {
		Tags map[string]any `json:"tags"`
	}

	if err := httpLib.Client.Get(fmt.Sprintf("%s/%s", resourcePath, url.PathEscape(urn)), &resource); err != nil {
		return nil, err
	}

	if resource.Tags == nil {
		return map[string]any{}, nil
	}

	return resource.Tags, nil
}

// knownKeys says what is there, so a typo is visible.
func knownKeys(tags map[string]any) string {
	if len(tags) == 0 {
		return "no tag at all"
	}

	keys := make([]string, 0, len(tags))
	for key := range tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return strings.Join(keys, ", ")
}

// reportTagDryRun previews the writes. common.Call carries a method and an
// endpoint and nothing else, so a plain ReportDryRun would print the same line
// once per tag and never say which tag. A preview that cannot be told apart
// from another preview is not a preview.
func reportTagDryRun(endpoint string, assignments []tagAssignment) bool {
	if !flags.DryRun {
		return false
	}

	message := "🔍 Dry run: nothing was sent. This would have been called:"
	details := make([]map[string]any, 0, len(assignments))

	for _, assignment := range assignments {
		message += fmt.Sprintf("\n  POST %s\n    with {%q: %q}", endpoint, assignment.Key, assignment.Value)
		details = append(details, map[string]any{
			"method":   "POST",
			"endpoint": endpoint,
			"payload":  map[string]any{"key": assignment.Key, "value": assignment.Value},
		})
	}

	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"calls": details}, "%s", message)

	return true
}
