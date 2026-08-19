// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

const testBackupServer = "https://eu.api.ovh.com/v1/dedicated/server/ns1.example/features/backupFTP"

// Only seven of thirty-five servers on a real account had a Backup FTP space.
// Not having one is a state, not a failure, and the command that creates the
// included space is named.
func (ms *MockSuite) TestBaremetalBackupFtpSaysWhenThereIsNoSpace(assert, require *td.T) {
	httpmock.RegisterResponder("GET", testBackupServer,
		httpmock.NewStringResponder(404, `{"message":"The requested object (backupFTP) does not exist"}`))

	_, err := cmd.Execute("baremetal", "backup", "ftp", "show", "ns1.example")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("has no Backup FTP space"))
	assert.Cmp(err.Error(), td.Contains("backup ftp create"))
}

// A read that fails for another reason is not an absent space.
func (ms *MockSuite) TestBaremetalBackupFtpReportsAFailedReadAsItself(assert, require *td.T) {
	httpmock.RegisterResponder("GET", testBackupServer,
		httpmock.NewStringResponder(500, `{"class":"Server::InternalServerError","message":"Internal server error"}`))

	_, err := cmd.Execute("baremetal", "backup", "ftp", "show", "ns1.example")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("failed to read"))
	assert.Cmp(err.Error(), td.Not(td.Contains("backup ftp create")), "the wrong remedy must not be offered")
}

// The API accepts an access rule with all three protocols false: a rule that
// appears in the list, says an IP block is allowed, and lets it reach nothing.
func (ms *MockSuite) TestBaremetalBackupAclRefusesARuleThatOpensNothing(assert, require *td.T) {
	_, err := cmd.Execute("baremetal", "backup", "ftp", "acl", "add",
		"ns1.example", "1.2.3.4/32", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("--ftp"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "a rule opening nothing costs no request")
}

// The blocks a space accepts are read from the server, not guessed: the list is
// regional, and a block from another region is refused by the API several
// seconds later without naming the ones that would work.
func (ms *MockSuite) TestBaremetalBackupAclChecksTheBlockAgainstTheServer(assert, require *td.T) {
	httpmock.RegisterResponder("GET", testBackupServer+"/authorizableBlocks",
		httpmock.NewStringResponder(200, `["51.68.100.160/28","141.94.98.55/32"]`))

	_, err := cmd.Execute("baremetal", "backup", "ftp", "acl", "add",
		"ns1.example", "1.2.3.4/32", "--ftp", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("2 block(s)"))
	assert.Cmp(err.Error(), td.Contains("authorizable-blocks"))
	for call := range httpmock.GetCallCountInfo() {
		assert.Cmp(call, td.Not(td.HasPrefix("POST")))
	}
}

// A server without a space cannot have an access list, and the message says
// which of the two is missing.
func (ms *MockSuite) TestBaremetalBackupAclListSaysWhenThereIsNoSpace(assert, require *td.T) {
	httpmock.RegisterResponder("GET", testBackupServer+"/access",
		httpmock.NewStringResponder(404, `{"message":"The requested object (backupFTP) does not exist"}`))

	_, err := cmd.Execute("baremetal", "backup", "ftp", "acl", "list", "ns1.example")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("nothing can reach it"))
}

// Deleting the space loses everything on it — the API says so itself — so it
// takes the strongest guard the CLI has, the server's name typed by hand.
func (ms *MockSuite) TestBaremetalBackupFtpDeletePreviewsWithoutSending(assert, require *td.T) {
	out, err := cmd.Execute("baremetal", "backup", "ftp", "delete", "ns1.example", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("DELETE"))
	for call := range httpmock.GetCallCountInfo() {
		assert.Cmp(call, td.Not(td.HasPrefix("DELETE")))
	}
}

// Thirty-four of thirty-five servers answered 403 on the cloud backup offer,
// and the thirty-fifth answered its sizes. A 403 there is a business answer,
// not a permission problem.
func (ms *MockSuite) TestBaremetalBackupCloudOfferReadsA403AsAnAnswer(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/ns1.example/backupCloudOfferDetails",
		httpmock.NewStringResponder(403, `{"class":"Client::Forbidden","message":"not offered"}`))

	out, err := cmd.Execute("baremetal", "backup", "cloud", "offer", "ns1.example")

	require.CmpNoError(err, "a server the offer does not cover is not a failed command")
	assert.Cmp(out, td.Contains("No cloud backup offer"))
}

// The four cloud backup passwords come back in the response body, unlike the
// Backup FTP one which the API mails.
func (ms *MockSuite) TestBaremetalBackupCloudPasswordIsWithheldUnderJson(assert, require *td.T) {
	httpmock.RegisterResponder("POST",
		"https://eu.api.ovh.com/v1/dedicated/server/ns1.example/features/backupCloud/password",
		httpmock.NewStringResponder(200, `{"sftpArchive":"SftpArchiveSecret1","sftpStorage":"SftpStorageSecret1","swiftArchive":"SwiftArchiveSecret","swiftStorage":"SwiftStorageSecret"}`))

	out, err := cmd.Execute("baremetal", "backup", "cloud", "password", "ns1.example",
		"--yes", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Not(td.Contains("SftpArchiveSecret1")), "a password must not reach a log")
	assert.Cmp(out, td.Contains("characters"))
}

// A server with no capacity to order says so rather than printing an empty
// list: four of thirty-five measured were in that state.
func (ms *MockSuite) TestBaremetalBackupOrderableSaysWhenThereIsNothing(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/ns1.example/orderable/backupStorage",
		httpmock.NewStringResponder(200, `{"orderable":false,"capacities":[]}`))

	out, err := cmd.Execute("baremetal", "backup", "orderable", "ns1.example")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("No backup storage can be ordered"))
}

func (ms *MockSuite) TestBaremetalBackupOrderableSpeaksTheCatalogue(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/ns1.example/orderable/backupStorage",
		httpmock.NewStringResponder(200, `{"orderable":true,"capacities":[500,1000,5000,10000]}`))

	out, err := cmd.Execute("baremetal", "backup", "orderable", "ns1.example")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("500 GB"))
	assert.Cmp(out, td.Contains("10 TB"), "a catalogue says terabytes")
}
