// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/go-ovh/ovh"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// The API takes three booleans and accepts all three false, which creates a
// rule that appears in the access list, says an IP block is allowed, and lets
// it reach nothing. That is worse than no rule: it reads as protection.
func TestAnAccessRuleHasToOpenSomething(t *testing.T) {
	BackupAclFtp, BackupAclNfs, BackupAclCifs = false, false, false
	if err := checkAclProtocols(); err == nil {
		t.Fatal("a rule opening no protocol must be refused")
	}

	BackupAclNfs = true
	defer func() { BackupAclNfs = false }()
	if err := checkAclProtocols(); err != nil {
		t.Fatalf("one protocol is enough: %s", err)
	}
}

// The table says what a rule does rather than showing three booleans to
// cross-read.
func TestAclProtocolsReadsAsAList(t *testing.T) {
	both := aclProtocols(map[string]any{"ftp": true, "nfs": false, "cifs": true})
	if both != "FTP, CIFS" {
		t.Fatalf("got %q", both)
	}

	if none := aclProtocols(map[string]any{"ftp": false, "nfs": false, "cifs": false}); none != "none" {
		t.Fatalf("a rule opening nothing should say so, got %q", none)
	}
}

// The list of blocks a Backup FTP space accepts is per server and regional:
// measured on a real account, the European servers offer around thirty and the
// Canadian ones five — their own.
func TestPickBlockIgnoresCaseAndAnswersTheApiSpelling(t *testing.T) {
	block, ok := pickBlock([]string{"2001:DB8::/64", "203.0.113.160/28"}, "2001:db8::/64")
	if !ok {
		t.Fatal("a block the server accepts must be accepted")
	}
	if block != "2001:DB8::/64" {
		t.Fatalf("the API spelling must be sent, got %q", block)
	}

	if _, ok := pickBlock([]string{"203.0.113.160/28"}, "1.2.3.4/32"); ok {
		t.Fatal("a block the server does not accept must be refused")
	}
}

// A European server measured for this offered thirty blocks. Thirty lines in an
// error message answer worse than a count and the command that prints them.
func TestUnauthorizableBlockCountsRatherThanLists(t *testing.T) {
	err := unauthorizableBlock("ns1.example", "1.2.3.4/32",
		[]string{"a/32", "b/32", "c/32"})

	if !strings.Contains(err.Error(), "3 block(s)") {
		t.Fatalf("the refusal should count them, got %q", err)
	}
	if strings.Contains(err.Error(), "a/32") {
		t.Fatalf("the refusal must not print them, got %q", err)
	}
	if !strings.Contains(err.Error(), "authorizable-blocks") {
		t.Fatalf("the way out should be named, got %q", err)
	}
}

// Unlike the Backup FTP password, which the API mails, this response carries
// four credentials. A password reset is exactly the command run inside a
// pipeline whose output is kept.
func TestCloudBackupPasswordsAreWithheldByDefault(t *testing.T) {
	RevealBackupPassword = false
	view := backupPasswordView(map[string]any{
		"sftpArchive":  "SftpArchiveSecret1",
		"sftpStorage":  "SftpStorageSecret1",
		"swiftArchive": "SwiftArchiveSecret",
		"swiftStorage": "SwiftStorageSecret",
	})

	for _, field := range []string{"sftpArchive", "sftpStorage", "swiftArchive", "swiftStorage"} {
		value, _ := view[field].(string)
		if !strings.Contains(value, "characters") {
			t.Errorf("%s should be a fingerprint, got %q", field, value)
		}
	}
	if view["hidden"] != true {
		t.Fatal("the output should say the passwords were withheld")
	}
}

func TestCloudBackupPasswordsArePrintedOnDemand(t *testing.T) {
	RevealBackupPassword = true
	defer func() { RevealBackupPassword = false }()

	view := backupPasswordView(map[string]any{"sftpArchive": "SftpArchiveSecret1"})
	if view["sftpArchive"] != "SftpArchiveSecret1" {
		t.Fatalf("--reveal must print the password, got %v", view["sftpArchive"])
	}
	if _, hidden := view["hidden"]; hidden {
		t.Fatal("nothing is hidden when the passwords are revealed")
	}
}

// The API answers capacities in gigabytes; a catalogue says terabytes.
func TestReadableCapacitySpeaksTheCatalogue(t *testing.T) {
	for gigabytes, want := range map[int64]string{
		500:   "500 GB",
		1000:  "1 TB",
		5000:  "5 TB",
		10000: "10 TB",
	} {
		if got := readableCapacity(gigabytes); got != want {
			t.Errorf("readableCapacity(%d) = %q, want %q", gigabytes, got, want)
		}
	}
}

// A creation measured on a real server produced a space that answered 200 after
// seven minutes with an empty ftpBackupName, while its task was still running —
// and the API refused every write against it until the task finished. Waiting
// for the object to exist handed back a space nothing could be done with.
func TestABackupFtpSpaceWithoutANameIsNotReady(t *testing.T) {
	if backupFtpIsUsable(map[string]any{"ftpBackupName": "", "type": "included"}) {
		t.Fatal("a space with no name is still being placed on a storage server")
	}
	if !backupFtpIsUsable(map[string]any{"ftpBackupName": "ftpback-rbx7-847.ovh.net"}) {
		t.Fatal("a named space is usable")
	}
	if backupFtpIsUsable(nil) {
		t.Fatal("no space is not a ready space")
	}
}

// The two directions of the wait ask different questions. A space halfway
// through its creation is not ready, but it is very much still there — reusing
// the readiness predicate for the deletion would have reported "gone" about a
// space being provisioned.
func TestWaitingForADeletionIsNotWaitingForUnreadiness(t *testing.T) {
	halfway := map[string]any{"ftpBackupName": "", "type": "included"}
	ready := map[string]any{"ftpBackupName": "ftpback-rbx2-222.ovh.net"}

	if backupFtpSettled(halfway, true, true) {
		t.Fatal("a space still being provisioned has not finished being created")
	}
	if backupFtpSettled(halfway, true, false) {
		t.Fatal("a space still being provisioned is not gone either")
	}

	if !backupFtpSettled(ready, true, true) {
		t.Fatal("a named space has finished being created")
	}
	if !backupFtpSettled(nil, false, false) {
		t.Fatal("no space at all is what a deletion waits for")
	}
	if backupFtpSettled(nil, false, true) {
		t.Fatal("no space is not a created space")
	}
}

// withBackupAPI points the shared client at httpmock and shortens the poll, so
// the wait can be exercised without the sixteen minutes a real creation took.
func withBackupAPI(t *testing.T, attempts int) {
	t.Helper()
	httpmock.Activate(t)

	origClient := httpLib.Client
	origInterval, origAttempts := backupPollInterval, backupPollAttempts
	client, err := ovh.NewClient("ovh-eu", "app_key", "app_secret", "consumer_key")
	td.Require(t).CmpNoError(err)
	httpLib.Client = client
	backupPollInterval, backupPollAttempts = time.Millisecond, attempts

	t.Cleanup(func() {
		httpLib.Client = origClient
		backupPollInterval, backupPollAttempts = origInterval, origAttempts
	})

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
		httpmock.NewStringResponder(200, "0"))
}

// The deletion waits for the space to stop being there, and every failure to
// read used to count as that: the loop decided on `err == nil`, so a 500, a rate
// limit or a dropped connection during the poll ended the wait and printed
// "✅ deleted" over a space that was still there. This is the worst shape the
// class takes — the operator is told the thing is gone.
//
// A read that did not happen is now a third outcome, distinct from the API
// saying the space is absent.
func TestDeleteWaitDoesNotReadAFailureAsAGoneSpace(t *testing.T) {
	withBackupAPI(t, 2)
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/srv/features/backupFTP",
		httpmock.NewStringResponder(500, `{"message": "gateway is having a day"}`))

	err := waitForBackupFtp("srv", false)

	td.Require(t).CmpError(err, "a deletion is never concluded from a failed read")
	td.Cmp(t, err.Error(), td.Contains("stopped waiting"))
	td.Cmp(t, err.Error(), td.Contains("not deleted yet"))
}

// And the deletion still completes on the answer that means it: a 404 is the API
// saying the space is not there, which is exactly what was asked for.
func TestDeleteWaitFinishesOnA404(t *testing.T) {
	withBackupAPI(t, 5)
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/srv/features/backupFTP",
		httpmock.NewStringResponder(404, `{"message": "The requested object (backupFTP) does not exist"}`))

	td.CmpNoError(t, waitForBackupFtp("srv", false))
}

// The same distinction the other way: waiting for a creation must not be ended
// by a 404 either, and that is the answer the route gives for most of the
// sixteen minutes a creation takes.
func TestCreateWaitIsNotEndedByAnAbsentSpace(t *testing.T) {
	withBackupAPI(t, 2)
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/srv/features/backupFTP",
		httpmock.NewStringResponder(404, `{"message": "The requested object (backupFTP) does not exist"}`))

	err := waitForBackupFtp("srv", true)

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains("not created yet"))
}

// The sentinel is what makes the three outcomes tellable apart, and it has to
// survive being wrapped — the message the operator reads is built around it.
func TestNoSpaceSurvivesWrapping(t *testing.T) {
	_, err := func() (map[string]any, error) {
		withBackupAPI(t, 1)
		httpmock.RegisterResponder("GET",
			"https://eu.api.ovh.com/v1/dedicated/server/srv/features/backupFTP",
			httpmock.NewStringResponder(404, `{"message": "The requested object (backupFTP) does not exist"}`))
		return backupFtpSpace("srv")
	}()

	td.Require(t).CmpError(err)
	td.Cmp(t, isNoSpace(err), true, "a 404 is recognisable through the wrapping")
	td.Cmp(t, err.Error(), td.Contains("srv has no Backup FTP space"), "and still reads as before")
	td.Cmp(t, err.Error(), td.Contains("backup ftp create srv"), "with the way out named")

	td.Cmp(t, isNoSpace(errors.New("failed to read the Backup FTP space of srv: 500")), false,
		"positive control: an ordinary failure is not an absent space")
}
