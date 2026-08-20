// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"fmt"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
	"github.com/ovh/ovhcloud-cli/internal/display"
)

const doctorServer = "https://eu.api.ovh.com/v1/dedicated/server/ns1.example"

// registerHealthyServer answers every route doctor reads with a server that has
// nothing wrong with it.
func registerHealthyServer() {
	httpmock.RegisterResponder("GET", doctorServer,
		httpmock.NewStringResponder(200, `{"state":"ok","powerState":"poweron","monitoring":true,"noIntervention":false,"bootId":1}`))
	httpmock.RegisterResponder("GET", doctorServer+"/boot/1",
		httpmock.NewStringResponder(200, `{"bootId":1,"bootType":"harddisk","kernel":"hd","description":"Boot to disk"}`))
	httpmock.RegisterResponder("GET", doctorServer+"/serviceInfos",
		httpmock.NewStringResponder(200, `{"expiration":"2099-01-01","renew":{"automatic":true,"deleteAtExpiration":false}}`))
	httpmock.RegisterResponder("GET", doctorServer+"/task?status=doing",
		httpmock.NewStringResponder(200, `[]`))
	httpmock.RegisterResponder("GET", doctorServer+"/plannedIntervention",
		httpmock.NewStringResponder(200, `[]`))
}

func (ms *MockSuite) TestBaremetalDoctorSaysWhenNothingIsWrong(assert, require *td.T) {
	registerHealthyServer()

	out, err := cmd.Execute("baremetal", "doctor", "ns1.example")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Nothing to report"))
}

// Six servers of the fleet measured were booted on a rescue system, and three
// of them carried bootType "internal" — the obvious check would have called
// them healthy.
func (ms *MockSuite) TestBaremetalDoctorCatchesTheRescueTypedInternal(assert, require *td.T) {
	registerHealthyServer()
	httpmock.RegisterResponder("GET", doctorServer,
		httpmock.NewStringResponder(200, `{"state":"ok","powerState":"poweron","monitoring":true,"bootId":46371}`))
	httpmock.RegisterResponder("GET", doctorServer+"/boot/46371",
		httpmock.NewStringResponder(200, `{"bootId":46371,"bootType":"internal","kernel":"rescue-customer","description":"Customer rescue system (Debian-10-based)[REMOVAL ON 2025-06-23]"}`))

	out, err := cmd.Execute("baremetal", "doctor", "ns1.example")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("rescue"))
	assert.Cmp(out, td.Contains("boot-image"), "the retired image is its own finding")
}

// A server that could not be read is not a healthy server, and counting it as
// checked is the one mistake this command cannot afford.
func (ms *MockSuite) TestBaremetalDoctorRefusesToGiveACleanBillForAServerItCouldNotRead(assert, require *td.T) {
	httpmock.RegisterResponder("GET", doctorServer,
		httpmock.NewStringResponder(500, `{"class":"Server::InternalServerError","message":"Internal server error"}`))

	_, err := cmd.Execute("baremetal", "doctor", "ns1.example")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("not a clean bill of health"))
}

// `renew.automatic` comes back differently between consecutive reads of the
// same object. The finding says so rather than picking one of the answers.
func (ms *MockSuite) TestBaremetalDoctorReportsTheApiDisagreeingWithItself(assert, require *td.T) {
	registerHealthyServer()

	answers := []string{
		`{"expiration":"2099-01-01","renew":{"automatic":true}}`,
		`{"expiration":"2099-01-01","renew":{"automatic":false}}`,
	}
	var call int
	httpmock.RegisterResponder("GET", doctorServer+"/serviceInfos",
		func(*http.Request) (*http.Response, error) {
			body := answers[call%len(answers)]
			call++
			return httpmock.NewStringResponse(200, body), nil
		})

	out, err := cmd.Execute("baremetal", "doctor", "ns1.example")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("different answers"))
	assert.Cmp(call >= 2, true, fmt.Sprintf("the field has to be read more than once, was read %d time(s)", call))
}

// The exit code stays 0 when findings are reported, because the command ran and
// answered. --strict is for the pipeline gating on it, and has to be asked for.
func (ms *MockSuite) TestBaremetalDoctorOnlyFailsWhenAskedTo(assert, require *td.T) {
	registerHealthyServer()
	httpmock.RegisterResponder("GET", doctorServer,
		httpmock.NewStringResponder(200, `{"state":"ok","powerState":"poweron","monitoring":false,"bootId":1}`))

	out, err := cmd.Execute("baremetal", "doctor", "ns1.example")
	require.CmpNoError(err, "reporting a finding is not a failed command")
	assert.Cmp(out, td.Contains("monitoring"))

	// --strict changes the exit code and nothing else. A second document after
	// the table would be the only thing left under -o json, so the pipeline
	// that asked for --strict would get the error instead of the findings.
	var code *int
	previous := display.ExitFunc
	display.ExitFunc = func(c int) { code = &c }
	defer func() { display.ExitFunc = previous }()

	strictOut, strictErr := cmd.Execute("baremetal", "doctor", "ns1.example", "--strict")

	require.CmpNoError(strictErr, "--strict must not emit a second document")
	require.NotNil(code, "--strict has to set a non-zero exit code")
	assert.Cmp(*code, 1)
	assert.Cmp(strictOut, td.Contains("monitoring"), "the findings are still the output")
	assert.Cmp(strictOut, td.Not(td.Contains("🛑")), "no error document is printed")
}

// A clean fleet under --strict exits 0. The no-finding case returns before the
// --strict block, so this pins the ORDER: moving the exit above that return
// would fail a green fleet for having been asked to check.
func (ms *MockSuite) TestBaremetalDoctorStrictStaysGreenWithNoFinding(assert, require *td.T) {
	registerHealthyServer()
	httpmock.RegisterResponder("GET", doctorServer,
		httpmock.NewStringResponder(200, `{"state":"ok","powerState":"poweron","monitoring":true,"bootId":1}`))

	exited := false
	previous := display.ExitFunc
	display.ExitFunc = func(int) { exited = true }
	defer func() { display.ExitFunc = previous }()

	_, err := cmd.Execute("baremetal", "doctor", "ns1.example", "--strict")

	require.CmpNoError(err)
	assert.Cmp(exited, false, "nothing to report means nothing to fail on")
}

// With no argument it checks every server of the account.
func (ms *MockSuite) TestBaremetalDoctorChecksTheWholeFleetByDefault(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server",
		httpmock.NewStringResponder(200, `["ns1.example"]`))
	registerHealthyServer()

	out, err := cmd.Execute("baremetal", "doctor")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("1 server"))
}
