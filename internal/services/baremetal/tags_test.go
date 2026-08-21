// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/ovh/go-ovh/ovh"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// The common case is the one nobody should have to look up.
func TestATagWithoutAnOperatorMeansEquals(t *testing.T) {
	filters, err := parseTagFilters([]string{"owner=Denis"})
	if err != nil {
		t.Fatalf("unexpected refusal: %s", err)
	}

	got := filters["owner"]
	if len(got) != 1 || got[0].Operator != "EQ" || got[0].Value != "Denis" {
		t.Fatalf("got %+v", got)
	}
}

// The operator is spelled with the name the API uses, so a refusal can list the
// ones that exist and completion can offer them.
func TestAnOperatorIsSpelledWhereItIsRead(t *testing.T) {
	filters, err := parseTagFilters([]string{"owner:NEQ=Denis", "Project:like=Proof%", "Team:EXISTS"})
	if err != nil {
		t.Fatalf("unexpected refusal: %s", err)
	}

	if filters["owner"][0].Operator != "NEQ" {
		t.Fatalf("got %+v", filters["owner"])
	}
	if filters["Project"][0].Operator != "LIKE" || filters["Project"][0].Value != "Proof%" {
		t.Fatalf("a lowercase operator must be accepted, and the pattern passed through: %+v", filters["Project"])
	}
	if filters["Team"][0].Operator != "EXISTS" || filters["Team"][0].Value != "" {
		t.Fatalf("got %+v", filters["Team"])
	}
}

// An unknown operator is refused with the list, not sent to the API to be
// refused there as a malformed parameter.
func TestAnUnknownOperatorIsRefusedWithTheList(t *testing.T) {
	_, err := parseTagFilters([]string{"owner:CONTAINS=Denis"})
	if err == nil {
		t.Fatal("an operator the API does not have must be refused")
	}
	for _, expected := range []string{"EQ", "NEQ", "LIKE", "ILIKE", "EXISTS", "NEXISTS"} {
		if !strings.Contains(err.Error(), expected) {
			t.Fatalf("the refusal must name %s: %s", expected, err)
		}
	}
}

// EXISTS asks whether a key is set at all. A value beside it has nothing to
// compare against, and accepting it would read as a filter that was applied.
func TestAValuelessOperatorRefusesAValue(t *testing.T) {
	if _, err := parseTagFilters([]string{"owner:EXISTS=Denis"}); err == nil {
		t.Fatal("EXISTS takes no value")
	}
	if _, err := parseTagFilters([]string{"owner:NEXISTS=Denis"}); err == nil {
		t.Fatal("NEXISTS takes no value")
	}
}

// And the other way round: a comparison with nothing to compare against is not
// quietly turned into EXISTS, which would answer a different question.
func TestAComparisonWithoutAValueIsRefused(t *testing.T) {
	_, err := parseTagFilters([]string{"owner"})
	if err == nil {
		t.Fatal("EQ with no value must be refused")
	}
	if !strings.Contains(err.Error(), "EXISTS") {
		t.Fatalf("the refusal must point at the operator that does answer it: %s", err)
	}
}

func TestATagFilterWithoutAKeyIsRefused(t *testing.T) {
	for _, raw := range []string{"=Denis", "", ":EQ=Denis", "   =x"} {
		if _, err := parseTagFilters([]string{raw}); err == nil {
			t.Fatalf("%q names no tag and must be refused", raw)
		}
	}
}

// A tag value may contain anything, including another "=" and the wildcards
// LIKE uses. Only the first "=" separates.
func TestAValueIsTakenWhole(t *testing.T) {
	filters, err := parseTagFilters([]string{"kernel=vmlinuz=6.1", "empty="})
	if err != nil {
		t.Fatalf("unexpected refusal: %s", err)
	}
	if filters["kernel"][0].Value != "vmlinuz=6.1" {
		t.Fatalf("got %q", filters["kernel"][0].Value)
	}
	if len(filters["empty"]) != 1 || filters["empty"][0].Value != "" {
		t.Fatalf("an empty value is a filter; got %+v", filters["empty"])
	}
}

// Several comparisons on one key are what the API's own shape allows: the
// parameter maps a key to a list of filters, not to one.
func TestSeveralComparisonsOnOneKeyAccumulate(t *testing.T) {
	filters, err := parseTagFilters([]string{"owner:NEQ=Denis", "owner:NEQ=bob"})
	if err != nil {
		t.Fatalf("unexpected refusal: %s", err)
	}
	if len(filters["owner"]) != 2 {
		t.Fatalf("got %+v", filters["owner"])
	}
}

// The query is JSON in a query string, and both halves have to survive: a bare
// brace or an unescaped quote produces a 400 the operator cannot read.
func TestTheQueryIsEncodedJson(t *testing.T) {
	if query, err := tagQuery(nil); err != nil || query != "" {
		t.Fatalf("no filter means no query, got %q / %v", query, err)
	}

	filters, err := parseTagFilters([]string{"Compliance=PCI-DSS"})
	if err != nil {
		t.Fatal(err)
	}

	query, err := tagQuery(filters)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(query, "?iamTags=") {
		t.Fatalf("got %q", query)
	}

	// The braces and quotes of the JSON have no business travelling raw in a
	// query string. QueryUnescape happily returns an unescaped string
	// unchanged, so a round trip alone would pass on a query that was never
	// escaped at all.
	for _, raw := range []string{"{", "}", "\"", "[", "]"} {
		if strings.Contains(query, raw) {
			t.Fatalf("%q must not travel raw in the query: %s", raw, query)
		}
	}

	decoded, err := url.QueryUnescape(strings.TrimPrefix(query, "?iamTags="))
	if err != nil {
		t.Fatalf("the query must be escaped: %s", err)
	}

	var back map[string][]tagFilter
	if err := json.Unmarshal([]byte(decoded), &back); err != nil {
		t.Fatalf("the API decodes this as JSON: %s", err)
	}
	if back["Compliance"][0].Value != "PCI-DSS" {
		t.Fatalf("got %+v", back)
	}
}

// A tag value may contain a space, and QueryEscape renders one as "+". The
// round trip has to survive it, because a value that comes back as "a+b"
// instead of "a b" is a filter that silently matches nothing.
func TestASpaceInAValueSurvivesTheQuery(t *testing.T) {
	filters, err := parseTagFilters([]string{"Project:LIKE=Proof of Concept%"})
	if err != nil {
		t.Fatal(err)
	}

	query, err := tagQuery(filters)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := url.QueryUnescape(strings.TrimPrefix(query, "?iamTags="))
	if err != nil {
		t.Fatal(err)
	}

	var back map[string][]tagFilter
	if err := json.Unmarshal([]byte(decoded), &back); err != nil {
		t.Fatalf("the API decodes this as JSON: %s", err)
	}
	if back["Project"][0].Value != "Proof of Concept%" {
		t.Fatalf("got %q", back["Project"][0].Value)
	}
}

// The schema documents "ovh:" as the prefix of every tag OVHcloud computes
// itself — ovh:default on a billing account, ovh:whoisOwner on a domain. Reading
// the operator from the FIRST colon made that whole namespace unreachable:
// "ovh:default=x" parsed as key "ovh" with operator "DEFAULT" and came back
// refused for an operator nobody typed.
func TestATagKeyMayContainAColon(t *testing.T) {
	filters, err := parseTagFilters([]string{"ovh:default:EQ=true"})
	if err != nil {
		t.Fatalf("unexpected refusal: %s", err)
	}
	if len(filters["ovh:default"]) != 1 {
		t.Fatalf("the key is everything before the operator: %+v", filters)
	}
	if filters["ovh:default"][0].Operator != "EQ" || filters["ovh:default"][0].Value != "true" {
		t.Fatalf("got %+v", filters["ovh:default"][0])
	}

	// And with a valueless operator, which has no "=" to anchor on.
	filters, err = parseTagFilters([]string{"ovh:whoisOwner:EXISTS"})
	if err != nil {
		t.Fatalf("unexpected refusal: %s", err)
	}
	if len(filters["ovh:whoisOwner"]) != 1 || filters["ovh:whoisOwner"][0].Operator != "EXISTS" {
		t.Fatalf("got %+v", filters)
	}
}

// What follows the last colon and is not an operator is genuinely ambiguous, and
// the CLI must not choose. Reading it as a key would send "owner:CONTAINS" to the
// API, come back with no servers, and print "nothing matches" — a wrong answer
// wearing the shape of an answer.
func TestAnAmbiguousColonIsRefusedWithBothReadings(t *testing.T) {
	_, err := parseTagFilters([]string{"owner:CONTAINS=Denis"})
	if err == nil {
		t.Fatal("the CLI must not silently pick one of the two readings")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("the refusal has to say so: %s", err)
	}
	if !strings.Contains(err.Error(), "owner:CONTAINS") {
		t.Fatalf("the key reading has to be named: %s", err)
	}
	if !strings.Contains(err.Error(), "EQ") {
		t.Fatalf("and the operators listed: %s", err)
	}
}

// The completer must only ever offer something the parser accepts. On a key
// holding a colon it offered "ovh:EQ" — the operator EQ appended to the first
// segment — which no longer names the key at all. A completion that cannot be
// accepted is worse than none: it is the CLI telling the operator to type
// something it will then reject.
//
// Every suggestion is fed straight back through the parser here, which is the
// only assertion that cannot drift from the parser's actual rules.
func TestTheCompleterOnlyOffersWhatTheParserAccepts(t *testing.T) {
	withTagAPI(t, `[{"id": "srv-1", "iam": {"tags": {"ovh:default": "true", "owner": "Denis"}}}]`)

	operators, err := tagOperators()
	if err != nil {
		t.Fatalf("could not read the operators: %s", err)
	}

	for _, toComplete := range []string{"", "ovh:", "ovh:default:", "owner:"} {
		suggestions, _ := CompleteBaremetalTag(nil, nil, toComplete)
		if len(suggestions) == 0 {
			t.Fatalf("nothing offered for %q, so this test would prove nothing", toComplete)
		}
		for _, suggestion := range suggestions {
			candidate := suggestion
			if !strings.HasSuffix(candidate, "EXISTS") {
				candidate += "=x"
			}
			if _, _, _, err := splitTagFilter(candidate, operators); err != nil {
				t.Fatalf("completing %q offered %q, which the parser refuses: %s",
					toComplete, suggestion, err)
			}
		}
	}
}

// And it has to reach the key, not stop at its first segment: typing "ovh:" and
// pressing tab used to offer six operators and never the key "ovh:default".
func TestTheCompleterReachesAKeyWithAColon(t *testing.T) {
	withTagAPI(t, `[{"id": "srv-1", "iam": {"tags": {"ovh:default": "true"}}}]`)

	suggestions, _ := CompleteBaremetalTag(nil, nil, "ovh:")

	var found bool
	for _, suggestion := range suggestions {
		if strings.HasPrefix(suggestion, "ovh:default:") {
			found = true
		}
	}
	if !found {
		t.Fatalf("the key has to be reached, got %v", suggestions)
	}
}

// withTagAPI points the shared client at httpmock for the v2 collection the
// completer reads its keys from.
func withTagAPI(t *testing.T, servers string) {
	t.Helper()
	httpmock.Activate(t)

	origClient := httpLib.Client
	client, err := ovh.NewClient("ovh-eu", "app_key", "app_secret", "consumer_key")
	if err != nil {
		t.Fatalf("could not build a client: %s", err)
	}
	httpLib.Client = client
	t.Cleanup(func() { httpLib.Client = origClient })

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
		httpmock.NewStringResponder(200, "0"))
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/dedicated/server",
		httpmock.NewStringResponder(200, servers))
}
