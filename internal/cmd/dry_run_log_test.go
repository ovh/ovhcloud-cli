// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

// Ces deux tests vivent dans leur propre fichier, et non dans
// baremetal_test.go, pour une raison mecanique : 23 branches en aval ajoutent
// elles aussi des tests juste apres TestBaremetalReinstallDryRun et dans le
// meme bloc d'imports. Les y placer a fait echouer les 23 merges d'un coup, sur
// une adjacence et non sur un desaccord. Un fichier neuf ne peut entrer en
// collision qu'avec un fichier de meme nom.

package cmd_test

import (
	"bytes"
	"log"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// A dry run prints the payload once. It used to print it twice: the message
// carries it, and a log.Println just above the --dry-run branch repeated the
// same JSON behind a Go timestamp no other command in this CLI emits.
//
//	🔍 Dry run: nothing was sent. This would have been posted to …
//	{ "operatingSystem": "debian12_64" }
//	2026/08/23 23:47:22 Final parameters:
//	{ "operatingSystem": "debian12_64" }
//
// The log line goes to stderr, so no assertion on stdout could ever have seen
// it — which is why it survived every green run. This one redirects the logger.
func (ms *MockSuite) TestBaremetalReinstallDryRunDoesNotLogTheParametersTwice(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/reinstall",
		httpmock.NewStringResponder(200, `{"taskId": 123}`),
	)
	var logged bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&logged)
	defer log.SetOutput(previous)

	out, err := cmd.Execute("baremetal", "reinstall", "fakeBaremetal", "--os", "debian12_64", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains(`"operatingSystem": "debian12_64"`),
		"the payload is still printed, once, as the message")
	assert.Cmp(logged.String(), td.Not(td.Contains("Final parameters")),
		"a dry run logs nothing: it sends nothing")
}

// The positive control of the test above: a REAL run still logs what it is
// about to send. Without it, deleting the log line altogether would pass.
func (ms *MockSuite) TestBaremetalReinstallStillLogsTheParametersItSends(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/dedicated/server/fakeBaremetal/reinstall",
		httpmock.NewStringResponder(200, `{"taskId": 123}`),
	)
	var logged bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&logged)
	defer log.SetOutput(previous)

	_, err := cmd.Execute("baremetal", "reinstall", "fakeBaremetal", "--os", "debian12_64", "--yes")

	require.CmpNoError(err)
	assert.Cmp(logged.String(), td.Contains("Final parameters"))
	assert.Cmp(logged.String(), td.Contains(`"operatingSystem": "debian12_64"`))
}
