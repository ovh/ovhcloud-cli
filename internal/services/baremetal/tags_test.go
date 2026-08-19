// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
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
	filters, err := parseTagFilters([]string{"owner:NEQ=Denis", "owner:NEQ=Yaniv"})
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
