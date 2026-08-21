// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package ip

import (
	"errors"
	"strings"
	"testing"
)

// The route is summarised by the API as "Reverse delegation on IPv6 subnet" and
// answers HTTP 500 for everything else — measured on 492 of 537 blocks, every
// IPv4 mask and every IPv6 /128, while the 45 IPv6 /56 and /64 answered 200.
// The note explains the failure; it does not stop the call, because the scope
// is the route's own and not a rule this CLI enforces on the API's behalf.
func TestDelegationErrorExplainsAnIPv4Failure(t *testing.T) {
	err := delegationError("read", "203.0.113.80/32", errors.New("Internal server error"))
	if !strings.Contains(err.Error(), "IPv6 subnets") {
		t.Fatalf("an IPv4 block should be told what this route covers, got %q", err)
	}
	if !strings.Contains(err.Error(), "Internal server error") {
		t.Fatalf("the API's own failure must survive, got %q", err)
	}
}

func TestDelegationErrorAddsNothingForAnIPv6Subnet(t *testing.T) {
	err := delegationError("read", "2001:db8::/64", errors.New("boom"))
	if strings.Contains(err.Error(), "IPv6 subnets") {
		t.Fatalf("an IPv6 subnet is already in scope, got %q", err)
	}
}

// The token lets whoever holds it move the address to the named account. The
// substitution happens on the object, so -o json is covered by the same
// decision as the table: a masking that only applies to human-readable output
// protects nothing, because the pipeline is where the value would be logged.
func TestMigrationTokenIsWithheldByDefault(t *testing.T) {
	RevealMigrationToken = false
	view := migrationTokenView(map[string]any{
		"customerId": "ab12345-ovh",
		"token":      "ExampleTokenValue123",
	})

	if view["token"] == "ExampleTokenValue123" {
		t.Fatal("the token must not be printed by default")
	}
	if !strings.Contains(view["token"].(string), "20 characters") {
		t.Fatalf("the fingerprint should say how long the token is, got %v", view["token"])
	}
	if view["customerId"] != "ab12345-ovh" {
		t.Fatalf("the customer is the answer to the usual question, got %v", view["customerId"])
	}
}

func TestMigrationTokenIsPrintedOnDemand(t *testing.T) {
	RevealMigrationToken = true
	defer func() { RevealMigrationToken = false }()

	view := migrationTokenView(map[string]any{"token": "ExampleTokenValue123"})
	if view["token"] != "ExampleTokenValue123" {
		t.Fatalf("--reveal must print the token, got %v", view["token"])
	}
	if _, hidden := view["hidden"]; hidden {
		t.Fatal("nothing is hidden when the token is revealed")
	}
}

// The API has one licence route per product and no index of them, so this list
// is the index. A product dropped from it becomes a licence the command cannot
// see, and the command answers "no licence attached".
func TestEveryLicenceProductWithARouteIsQueried(t *testing.T) {
	for _, product := range []string{
		"cloudLinux", "cpanel", "directadmin", "plesk",
		"sqlserver", "virtuozzo", "windows", "worklight",
	} {
		found := false
		for _, queried := range licenseProducts {
			if queried == product {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s has a licence route and is never queried", product)
		}
	}

	if len(licenseProducts) != 8 {
		t.Fatalf("the API has eight licence routes, this queries %d", len(licenseProducts))
	}
}

// The measurement the note is built on says every IPv4 mask AND every IPv6
// /128 answers 500. A note attached to IPv4 alone leaves the /128 with a bare
// 500 — the case the comment names and the code missed.
func TestDelegationErrorExplainsASingleIPv6Address(t *testing.T) {
	err := delegationError("read", "2001:db8::1/128", errors.New("Internal server error"))
	if !strings.Contains(err.Error(), "IPv6 subnets") {
		t.Fatalf("a /128 is not a subnet and should be told so, got %q", err)
	}
	if !strings.Contains(err.Error(), "single IPv6 address") {
		t.Fatalf("the note should say what a /128 is, got %q", err)
	}
}

func TestIPv6SubnetRecognisesWhatTheRouteServes(t *testing.T) {
	for block, want := range map[string]bool{
		"2001:db8::/56":   true,
		"2001:db8::/64":   true,
		"2001:db8::1/128": false,
		"203.0.113.80/32": false,
		"203.0.113.32/30": false,
	} {
		if got := isIPv6Subnet(block); got != want {
			t.Errorf("isIPv6Subnet(%q) = %v, want %v", block, got, want)
		}
	}
}

// The preview routes read and the two others write; one prefix cannot describe
// both. A slicing refused by the API used to report that a read had failed.
func TestByoipErrorNamesWhatFailed(t *testing.T) {
	read := byoipError("read the bring-your-own-IP configuration of", "1.2.3.0/24", errors.New("nope"))
	if !strings.Contains(read.Error(), "read") {
		t.Fatalf("got %q", read)
	}

	write := byoipError("slice", "1.2.3.0/24", errors.New("nope"))
	if strings.Contains(write.Error(), "read") {
		t.Fatalf("a failed slicing is not a failed read, got %q", write)
	}
	if !strings.Contains(write.Error(), "slice") {
		t.Fatalf("got %q", write)
	}
}
