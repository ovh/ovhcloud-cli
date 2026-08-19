// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
	"github.com/spf13/cobra"
)

// IAM tags are how a fleet is organised — owner, Compliance, Team, LandingZone
// — and until now nothing in this CLI could ask for them. The generic --filter
// cannot: it runs over the columns of the table, after every server has been
// fetched, and the tags are not among them.
//
// The v2 collection takes an iamTags query parameter that the v1 one does not.
// So this filters where the filtering belongs, on the server, and the servers
// that come back are then read on v1 — the only route that carries a machine
// rather than {id, iam}. Measured on 20 August 2026: both catalogues list
// exactly the same 35 servers on this account, so nothing is lost by asking
// one and reading the other.

// BaremetalTags are the tag filters given on the command line.
var BaremetalTags []string

// tagOperators are the comparisons the API accepts, read from the embedded
// schema rather than transcribed: a list copied into Go stops being true in
// silence, which is what the game protocols of #256 cost.
var tagOperators = sync.OnceValues(func() ([]string, error) {
	return openapi.GetComponentEnum(assets.BaremetalV2OpenapiSchema, "iam.resource.TagFilter.OperatorEnum")
})

// tagFilter is one comparison on one tag key, in the shape the API takes.
type tagFilter struct {
	Operator string `json:"operator"`
	Value    string `json:"value,omitempty"`
}

// valuelessOperators ask whether a key is set at all, so a value would have
// nothing to compare against. Passing one anyway is a mistake worth naming: it
// reads as a filter that was applied.
var valuelessOperators = map[string]bool{"EXISTS": true, "NEXISTS": true}

// parseTagFilters turns what was typed into the query the API takes.
//
// The shape is key[:OPERATOR][=value], one rule with no invented punctuation:
//
//	--tag owner=Denis          the common case, EQ, which is the API's own default
//	--tag owner:EXISTS         set to anything
//	--tag owner:NEQ=Denis      set to something else
//	--tag Project:LIKE=Proof%  the API's pattern syntax, passed through untouched
//
// The operator is spelled with the name the API uses, so a refusal can list the
// ones that exist and completion can offer them.
func parseTagFilters(given []string) (map[string][]tagFilter, error) {
	if len(given) == 0 {
		return nil, nil
	}

	operators, err := tagOperators()
	if err != nil {
		return nil, fmt.Errorf("failed to read the tag operators from the embedded schema: %w", err)
	}

	filters := make(map[string][]tagFilter, len(given))
	for _, raw := range given {
		key, operator, value, err := splitTagFilter(raw, operators)
		if err != nil {
			return nil, err
		}

		filters[key] = append(filters[key], tagFilter{Operator: operator, Value: value})
	}

	return filters, nil
}

func splitTagFilter(raw string, operators []string) (key, operator, value string, err error) {
	// The value is whatever follows the first "=", untouched: a tag value may
	// contain anything, including another "=" and the LIKE wildcards.
	head := raw
	hasValue := false
	if at := strings.Index(raw, "="); at >= 0 {
		head, value, hasValue = raw[:at], raw[at+1:], true
	}

	operator = "EQ"
	if at := strings.Index(head, ":"); at >= 0 {
		head, operator = head[:at], strings.ToUpper(head[at+1:])
	}

	key = strings.TrimSpace(head)
	if key == "" {
		return "", "", "", fmt.Errorf("%q names no tag; write it as key=value, or key:OPERATOR=value", raw)
	}

	if !slicesContain(operators, operator) {
		return "", "", "", fmt.Errorf("unknown tag operator %q in %q; use one of %s",
			operator, raw, strings.Join(operators, ", "))
	}

	switch {
	case valuelessOperators[operator] && hasValue:
		return "", "", "", fmt.Errorf("%s asks whether %q is set at all, so it takes no value (%q)", operator, key, raw)

	case !valuelessOperators[operator] && !hasValue:
		// An empty value is a legitimate filter; no value at all is not, and
		// silently turning it into EXISTS would answer a question nobody asked.
		return "", "", "", fmt.Errorf("%s needs something to compare %q against; write %s=<value>, or use %s to ask whether it is set",
			operator, key, key, strings.Join(valuelessNames(operators), " or "))
	}

	return key, operator, value, nil
}

func valuelessNames(operators []string) []string {
	var names []string
	for _, operator := range operators {
		if valuelessOperators[operator] {
			names = append(names, operator)
		}
	}
	sort.Strings(names)

	return names
}

// tagQuery renders the filters as the query string the collection takes.
func tagQuery(filters map[string][]tagFilter) (string, error) {
	if len(filters) == 0 {
		return "", nil
	}

	encoded, err := json.Marshal(filters)
	if err != nil {
		return "", fmt.Errorf("failed to render the tag filter: %w", err)
	}

	return "?iamTags=" + url.QueryEscape(string(encoded)), nil
}

// CompleteBaremetalTag offers the tag keys in use on the account, and the
// operators once a key has been typed.
//
// The keys come from the same collection the filter runs against, so what is
// offered is what exists rather than what somebody documented once.
func CompleteBaremetalTag(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if at := strings.Index(toComplete, ":"); at >= 0 {
		operators, err := tagOperators()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		prefix := toComplete[:at+1]
		suggestions := make([]string, 0, len(operators))
		for _, operator := range operators {
			suggestions = append(suggestions, prefix+operator)
		}

		return suggestions, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
	}

	if strings.Contains(toComplete, "=") {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	keys, err := tagKeysInUse()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return keys, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
}

// tagKeysInUse lists the tag keys actually set on the servers of the account.
//
// The v2 collection is the one that carries them: an object there is {id, iam}
// and iam.tags is the map. That is also why it cannot replace the v1 list —
// there is no machine in it, only its name and its identity.
func tagKeysInUse() ([]string, error) {
	var servers []struct {
		IAM struct {
			Tags map[string]string `json:"tags"`
		} `json:"iam"`
	}

	if err := httpLib.Client.Get("/v2/dedicated/server", &servers); err != nil {
		return nil, fmt.Errorf("failed to list the servers: %w", err)
	}

	seen := make(map[string]bool)
	for _, server := range servers {
		for key := range server.IAM.Tags {
			seen[key] = true
		}
	}

	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys, nil
}
