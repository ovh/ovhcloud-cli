// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package ip

import (
	"strings"
	"testing"
)

// The protocols an address accepts are read from the address, not from the
// enum shipped with the CLI. The two disagree in production: an account
// measured while writing this carries a rule using `arkSurvivalAscended`,
// which the API reports in supportedProtocols and the embedded schema does not
// list. Checking against the schema would refuse a protocol already in use.
func TestPickProtocolAcceptsWhatTheAddressReportsEvenOutsideTheEmbeddedEnum(t *testing.T) {
	supported := []string{"arkSurvivalAscended", "hl2Source", "teamspeak3"}

	protocol, err := pickProtocol("1.2.3.4", supported, "arkSurvivalAscended")
	if err != nil {
		t.Fatalf("a protocol the address reports must be accepted: %s", err)
	}
	if protocol != "arkSurvivalAscended" {
		t.Fatalf("got %q", protocol)
	}
}

func TestPickProtocolIgnoresCaseAndAnswersTheAPISpelling(t *testing.T) {
	protocol, err := pickProtocol("1.2.3.4", []string{"minecraftJava"}, "MINECRAFTJAVA")
	if err != nil {
		t.Fatalf("unexpected refusal: %s", err)
	}
	if protocol != "minecraftJava" {
		t.Fatalf("the API spelling must be sent, got %q", protocol)
	}
}

// The two addresses measured supported 14 and 20 protocols respectively, so
// there is no single list to name in a refusal: it has to be the address's own.
func TestPickProtocolListsWhatThisAddressSupports(t *testing.T) {
	_, err := pickProtocol("1.2.3.4", []string{"rust", "arma"}, "counterStrike2")
	if err == nil {
		t.Fatal("an unsupported protocol must be refused")
	}
	for _, expected := range []string{"arma", "rust"} {
		if !strings.Contains(err.Error(), expected) {
			t.Fatalf("the refusal should list %q, got %q", expected, err)
		}
	}
	if strings.Contains(err.Error(), "valheim") {
		t.Fatalf("the refusal must not name protocols this address does not support, got %q", err)
	}
}

func TestParsePortsReadsBothFormsFoundInProduction(t *testing.T) {
	for _, tc := range []struct {
		spec       string
		from, to   int
		shouldFail bool
	}{
		{spec: "7777", from: 7777, to: 7777},
		{spec: "7777-7778", from: 7777, to: 7778},
		{spec: " 27015 - 27020 ", from: 27015, to: 27020},
		{spec: "", shouldFail: true},
		{spec: "http", shouldFail: true},
		{spec: "7778-7777", shouldFail: true},
		{spec: "0", shouldFail: true},
		{spec: "1-70000", shouldFail: true},
	} {
		from, to, err := parsePorts(tc.spec)
		if tc.shouldFail {
			if err == nil {
				t.Errorf("parsePorts(%q) should have been refused", tc.spec)
			}
			continue
		}
		if err != nil {
			t.Errorf("parsePorts(%q): %s", tc.spec, err)
			continue
		}
		if from != tc.from || to != tc.to {
			t.Errorf("parsePorts(%q) = %d-%d, want %d-%d", tc.spec, from, to, tc.from, tc.to)
		}
	}
}

// Firewall mode only lets UDP traffic through when it matches a rule, so
// turning it on with no rule drops every UDP packet. One address of the
// account measured was exactly one flag away from that: firewall mode off,
// zero rules.
func TestFirewallModeWithoutAnyRuleIsRefused(t *testing.T) {
	if !firewallModeWouldBlackhole(true, 0) {
		t.Fatal("enabling firewall mode with no rule drops all UDP traffic")
	}
	if firewallModeWouldBlackhole(true, 1) {
		t.Fatal("one rule is enough to let something through")
	}
	if firewallModeWouldBlackhole(false, 0) {
		t.Fatal("disabling firewall mode blocks nothing")
	}
}

func TestGameFirewallWarningSaysWhichWayTheTrafficGoes(t *testing.T) {
	enabling := gameFirewallWarning("1.2.3.4", true, 3)
	if !strings.Contains(enabling, "drops") {
		t.Fatalf("enabling should warn about dropped traffic, got %q", enabling)
	}

	disabling := gameFirewallWarning("1.2.3.4", false, 3)
	if !strings.Contains(disabling, "lets") {
		t.Fatalf("disabling should warn about traffic let through, got %q", disabling)
	}
	if !strings.Contains(disabling, "1.2.3.4") {
		t.Fatalf("the warning should name the address, got %q", disabling)
	}
}

func TestPortsLabelRendersASinglePortAsOneNumber(t *testing.T) {
	single := portsLabel(map[string]any{"from": float64(123), "to": float64(123)})
	if single != "123" {
		t.Fatalf("a one-port range should read as one number, got %q", single)
	}

	span := portsLabel(map[string]any{"from": float64(7777), "to": float64(7778)})
	if span != "7777-7778" {
		t.Fatalf("got %q", span)
	}

	if empty := portsLabel(nil); empty != "" {
		t.Fatalf("an absent range should render as nothing, got %q", empty)
	}
}

// 0 is one of the five values the API accepts and it means "no delay", not
// "unset". Rendering it as 0h would read as a duration nobody chose.
func TestMitigationTimeoutLabelNamesTheZero(t *testing.T) {
	if label := mitigationTimeoutLabel(0); label != "no delay" {
		t.Fatalf("got %q", label)
	}
	if label := mitigationTimeoutLabel(360); label != "6h" {
		t.Fatalf("got %q", label)
	}
	if label := mitigationTimeoutLabel(15); label != "15min" {
		t.Fatalf("got %q", label)
	}
	if label := mitigationTimeoutLabel(1560); label != "26h" {
		t.Fatalf("got %q", label)
	}
}

func TestOnlyTheFiveAcceptedTimeoutsPass(t *testing.T) {
	for _, accepted := range mitigationTimeouts {
		if !slicesContainInt(mitigationTimeouts, accepted) {
			t.Fatalf("%d should be accepted", accepted)
		}
	}
	if slicesContainInt(mitigationTimeouts, 30) {
		t.Fatal("30 is not one of the values the API takes")
	}
}
