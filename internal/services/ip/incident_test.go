// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package ip

import (
	"encoding/json"
	"strings"
	"testing"
)

// The API calls the same field `time` on all three blocking mechanisms and
// means the opposite thing on one of them: a cooldown before the release is
// accepted on anti-hack and ARP, the length of the block itself on spam. A
// table printing both under one header would be wrong on a third of its rows
// and look right.
func TestBlockNoteSaysWhatTheNumberMeansPerMechanism(t *testing.T) {
	cooldown := blockNote(blockedAddress{Mechanism: antihackMechanism, Seconds: 720})
	if !strings.Contains(cooldown, "can be unblocked in") {
		t.Fatalf("anti-hack should announce a cooldown, got %q", cooldown)
	}

	sentence := blockNote(blockedAddress{Mechanism: spamMechanism, Seconds: 720})
	if strings.Contains(sentence, "can be unblocked") {
		t.Fatalf("spam has no cooldown, got %q", sentence)
	}
	if !strings.Contains(sentence, "blocked for") {
		t.Fatalf("spam should announce the length of the block, got %q", sentence)
	}
}

func TestBlockNoteSaysWhenTheReleaseIsAlreadyPossible(t *testing.T) {
	note := blockNote(blockedAddress{Mechanism: arpMechanism, Seconds: 0})
	if !strings.Contains(note, "now") {
		t.Fatalf("a zero cooldown means the release is possible now, got %q", note)
	}
}

func TestChooseMechanismRefusesWhenNothingHoldsTheAddress(t *testing.T) {
	_, _, err := chooseMechanism("1.2.3.0/24", "1.2.3.4", nil)
	if err == nil {
		t.Fatal("releasing an address nothing blocks must be refused")
	}
	if !strings.Contains(err.Error(), "ovhcloud ip blocked") {
		t.Fatalf("the refusal should name the command that lists blocks, got %q", err)
	}
}

func TestChooseMechanismCarriesTheCooldownOfTheMatchingEntry(t *testing.T) {
	mechanism, cooldown, err := chooseMechanism("1.2.3.0/24", "1.2.3.4", []blockedAddress{
		{IP: "1.2.3.5", Mechanism: antihackMechanism, Seconds: 999},
		{IP: "1.2.3.4", Mechanism: arpMechanism, Seconds: 42},
	})
	if err != nil {
		t.Fatalf("unexpected refusal: %s", err)
	}
	if mechanism != arpMechanism {
		t.Fatalf("expected the arp entry, got %q", mechanism)
	}
	if cooldown != 42 {
		t.Fatalf("expected the cooldown of the matching entry, got %d", cooldown)
	}
}

// An address held by two mechanisms is released twice, once per mechanism.
// Picking one silently would report a success while the traffic stayed blocked.
func TestChooseMechanismRefusesToPickWhenSeveralHold(t *testing.T) {
	_, _, err := chooseMechanism("1.2.3.0/24", "1.2.3.4", []blockedAddress{
		{IP: "1.2.3.4", Mechanism: antihackMechanism},
		{IP: "1.2.3.4", Mechanism: spamMechanism},
	})
	if err == nil {
		t.Fatal("an address blocked twice must not be released at random")
	}
	for _, mechanism := range []string{antihackMechanism, spamMechanism, "--reason"} {
		if !strings.Contains(err.Error(), mechanism) {
			t.Fatalf("the refusal should name %q, got %q", mechanism, err)
		}
	}
}

func TestFormatDelayReadsAsAWait(t *testing.T) {
	for _, tc := range []struct {
		seconds int64
		want    string
	}{
		{0, "0s"},
		{45, "45s"},
		{720, "12m"},
		{5400, "1h 30m"},
		{90000, "1d 1h"},
		{172800, "2d"},
	} {
		if got := formatDelay(tc.seconds); got != tc.want {
			t.Errorf("formatDelay(%d) = %q, want %q", tc.seconds, got, tc.want)
		}
	}
}

// go-ovh decodes with UseNumber, so every number reaches these helpers as a
// json.Number. A type switch that only knew float64 already turned a vRack
// branch into dead code once in this CLI.
func TestIntFieldReadsTheShapeGoOvhActuallyProduces(t *testing.T) {
	var decoded map[string]any
	decoder := json.NewDecoder(strings.NewReader(`{"time": 720}`))
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("failed to decode: %s", err)
	}

	if got := intField(decoded, "time"); got != 720 {
		t.Fatalf("intField read %d from a json.Number, want 720", got)
	}
}

func TestSpamWindowKeepsAnExplicitWindow(t *testing.T) {
	SpamStatsFrom, SpamStatsTo = "2026-01-01T00:00:00Z", "2026-01-08T00:00:00Z"
	defer func() { SpamStatsFrom, SpamStatsTo = "", "" }()

	from, to, source, err := spamWindow("1.2.3.0/24", "1.2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if from != "2026-01-01T00:00:00Z" || to != "2026-01-08T00:00:00Z" {
		t.Fatalf("the window given on the command line must be used verbatim, got %s → %s", from, to)
	}
	if !strings.Contains(source, "command line") {
		t.Fatalf("the report should say where its window came from, got %q", source)
	}
}
