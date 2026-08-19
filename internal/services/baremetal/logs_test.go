// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"strings"
	"testing"
	"time"
)

// A stream is named by a title or by a UUID, and the two are told apart by
// shape alone. Getting this wrong in either direction is silent: a UUID
// mistaken for a title triggers a sweep that finds nothing, and a title
// mistaken for a UUID is sent to the API as a stream identifier.
func TestAStreamIdentifierIsRecognisedByItsShape(t *testing.T) {
	if !looksLikeUUID("4459af12-6451-45de-808b-2b959c11a17e") {
		t.Fatal("a real stream identifier must be taken as one")
	}
	if !looksLikeUUID("4459AF12-6451-45DE-808B-2B959C11A17E") {
		t.Fatal("the API answers in lowercase but accepts either")
	}

	for _, notAnIdentifier := range []string{
		"TO REMOVE 1",
		"datastream_test",
		"",
		"4459af12-6451-45de-808b-2b959c11a17",           // one short
		"4459af12-6451-45de-808b-2b959c11a17ee",         // one long
		"4459af12x6451-45de-808b-2b959c11a17e",          // separator moved
		"zzzzzzzz-6451-45de-808b-2b959c11a17e",          // not hexadecimal
		"Prestashop nginx logs (filebeat) padded to 36", // right length, wrong everything
	} {
		if looksLikeUUID(notAnIdentifier) {
			t.Fatalf("%q must be treated as a title, not an identifier", notAnIdentifier)
		}
	}
}

// A UUID is taken as given: re-reading 85 objects to confirm an identifier the
// caller already holds is work done to learn nothing, and it is what `-o json`
// hands back.
func TestAnIdentifierIsUsedWithoutASweep(t *testing.T) {
	stream, err := resolveStream("4459af12-6451-45de-808b-2b959c11a17e")
	if err != nil {
		t.Fatalf("an identifier must resolve to itself: %s", err)
	}
	if stream.StreamID != "4459af12-6451-45de-808b-2b959c11a17e" {
		t.Fatalf("got %q", stream.StreamID)
	}
	if stream.Title != "" || stream.ServiceName != "" {
		t.Fatal("nothing was read, so nothing but the identifier may be claimed")
	}
}

// The account this was built against holds 59 streams. A refusal that listed
// them all would answer worse than one that names the near misses and counts
// the rest.
func TestNearbyTitlesNamesAFewAndCountsTheRest(t *testing.T) {
	var streams []ldpStream
	for _, title := range []string{
		"Stream1", "Stream 2", "Substream1a", "My first data stream",
		"Data-Stream-Pierrick-PCC", "datastream_test", "Test datastream",
	} {
		streams = append(streams, ldpStream{Title: title, ServiceName: "ldp-xx-1"})
	}

	near := nearbyTitles(streams, "Stream")
	if !strings.Contains(near, "and 2 more") {
		t.Fatalf("a truncated list has to say it was truncated, got %q", near)
	}
	if strings.Count(near, "ldp-xx-1") != 5 {
		t.Fatalf("five near misses were expected, got %q", near)
	}

	// Nothing close: say how many exist rather than pretend to suggest.
	if far := nearbyTitles(streams, "zzzz"); !strings.Contains(far, "7 of them") {
		t.Fatalf("got %q", far)
	}
}

// A stream is named to the operator the way they named it. When they typed a
// title, the identifier the API acted on is shown beside it; when they pasted
// an identifier, repeating it says nothing.
func TestAStreamIsNamedTheWayItWasAskedFor(t *testing.T) {
	titled := streamLabel(ldpStream{Title: "TO REMOVE 1", StreamID: "69813166-7b66-4a09-8fcd-c47781d23966", ServiceName: "ldp-nx-19421"})
	if !strings.Contains(titled, "TO REMOVE 1") || !strings.Contains(titled, "ldp-nx-19421") {
		t.Fatalf("got %q", titled)
	}

	bare := streamLabel(ldpStream{StreamID: "69813166-7b66-4a09-8fcd-c47781d23966"})
	if bare != "69813166-7b66-4a09-8fcd-c47781d23966" {
		t.Fatalf("got %q", bare)
	}
}

// The link is valid for thirty minutes. A command that printed it without the
// expiry would hand over something that stops working while it is on screen.
func TestExpiryIsSaidInMinutesLeft(t *testing.T) {
	soon := time.Now().Add(29*time.Minute + 40*time.Second).UTC().Format(time.RFC3339)
	if phrase := expiryPhrase(soon); !strings.Contains(phrase, "30m0s") {
		t.Fatalf("got %q", phrase)
	}

	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	if phrase := expiryPhrase(past); !strings.Contains(phrase, "already expired") {
		t.Fatalf("an expired link must say so, got %q", phrase)
	}

	// An unreadable date is still shown: the API said something, and dropping
	// it would leave a link with no expiry at all.
	if phrase := expiryPhrase("not a date"); !strings.Contains(phrase, "not a date") {
		t.Fatalf("got %q", phrase)
	}
	if phrase := expiryPhrase(""); !strings.Contains(phrase, "no stated expiry") {
		t.Fatalf("got %q", phrase)
	}
}

// The change is made on /v2 and followed on /v1, and the only thing joining the
// two is what the API answered. If it answered neither, the wait has nothing to
// poll and has to say so instead of looping on an empty path.
func TestAWaitWithNothingToFollowSaysSo(t *testing.T) {
	if _, err := waitForLogOperation("", "op-1"); err == nil {
		t.Fatal("a wait without a Log Data Platform service must refuse")
	}
	if _, err := waitForLogOperation("ldp-nx-19421", ""); err == nil {
		t.Fatal("a wait without an operation must refuse")
	}
}
