// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// The three boot entries a dedicated server carries, as the API returned them
// on 18 August 2026. All 35 servers on the account had a `power` entry; its
// identifier is 95083 in Europe and 95644 in Canada, which is why this command
// asks the API which entry does the job instead of carrying a table.
const (
	powerBootEntries = `{"1": {"bootId": 1, "bootType": "harddisk", "kernel": "hd", "description": "Boot to disk"}, "95083": {"bootId": 95083, "bootType": "power", "kernel": "poweroff", "description": "Power-off server"}, "230242": {"bootId": 230242, "bootType": "rescue", "kernel": "rescue12-customer", "description": "Customer rescue system (Debian-12-based)"}}`

	powerServer = "ns3070493.ip-57-129-37.eu"
)

var powerBodies map[string][]map[string]any

func recordPowerBody(key string) httpmock.Responder {
	return func(req *http.Request) (*http.Response, error) {
		var body map[string]any
		if req.Body != nil {
			raw, _ := io.ReadAll(req.Body)
			_ = json.Unmarshal(raw, &body)
		}
		powerBodies[key] = append(powerBodies[key], body)
		return httpmock.NewStringResponse(200, `null`), nil
	}
}

// registerPower wires a server in the given power state, sitting on the given
// boot, and gives the test a cache directory of its own — the command
// remembers the previous boot on disk, so a shared cache would let one test
// answer another one's question.
func registerPower(t *td.T, power string, bootID int) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	powerBodies = map[string][]map[string]any{}

	base := "https://eu.api.ovh.com/v1/dedicated/server/" + powerServer

	httpmock.RegisterResponder("GET", base,
		httpmock.NewStringResponder(200,
			`{"powerState": "`+power+`", "state": "ok", "bootId": `+itoa(bootID)+`}`))
	httpmock.RegisterResponder("PUT", base, recordPowerBody("boot"))
	httpmock.RegisterResponder("POST", base+"/reboot", recordPowerBody("reboot"))

	var entries map[string]map[string]any
	_ = json.Unmarshal([]byte(powerBootEntries), &entries)
	for id, entry := range entries {
		payload, _ := json.Marshal(entry)
		httpmock.RegisterResponder("GET", base+"/boot/"+id,
			httpmock.NewStringResponder(200, string(payload)))
	}
	httpmock.RegisterResponder("GET", base+"/boot?bootType=power",
		httpmock.NewStringResponder(200, `[95083]`))
	httpmock.RegisterResponder("GET", base+"/boot?bootType=harddisk",
		httpmock.NewStringResponder(200, `[1]`))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}

func powerCallCount(method, suffix string) int {
	return httpmock.GetCallCountInfo()[method+" https://eu.api.ovh.com/v1/dedicated/server/"+powerServer+suffix]
}

// The identifier is read from the API, not carried in Go: it is 95083 in
// Europe and 95644 in Canada, and a table copied into the binary stops being
// true without saying so.
func (ms *MockSuite) TestBaremetalPowerOffFindsTheEntryInsteadOfHardcodingIt(assert, require *td.T) {
	registerPower(assert, "poweron", 230242)

	_, err := cmd.Execute("baremetal", "power", "off", powerServer, "--yes")

	require.CmpNoError(err)
	assert.Cmp(powerCallCount("GET", "/boot?bootType=power"), 1, "the API is asked which entry powers off")
	require.Cmp(len(powerBodies["boot"]), 1)
	assert.Cmp(powerBodies["boot"][0]["bootId"], float64(95083))
	assert.Cmp(powerCallCount("POST", "/reboot"), 1, "and the machine is rebooted into it")
}

// Powering off rewrites the boot configuration, and the previous value is gone
// from the API the moment it does. Remembering it is what makes `power on` put
// the server back where it was rather than somewhere plausible.
func (ms *MockSuite) TestBaremetalPowerOnRestoresTheBootItWasOn(assert, require *td.T) {
	registerPower(assert, "poweron", 230242)
	_, err := cmd.Execute("baremetal", "power", "off", powerServer, "--yes")
	require.CmpNoError(err)

	// The same cache directory, as the same operator on the same machine would
	// have. Only the server's state changes: it is now off.
	cmd.PostExecute()
	powerBodies = map[string][]map[string]any{}
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/"+powerServer,
		httpmock.NewStringResponder(200, `{"powerState": "poweroff", "state": "ok", "bootId": 95083}`))

	out, err := cmd.Execute("baremetal", "power", "on", powerServer)

	require.CmpNoError(err)
	require.Cmp(len(powerBodies["boot"]), 1)
	assert.Cmp(powerBodies["boot"][0]["bootId"], float64(230242),
		"the rescue boot it was on, not the disk")
	assert.Cmp(out, td.Contains("was on before it was powered off"), "and the source is stated")
}

// A `power on` run from another machine than the `power off` has no record.
// Falling back to the disk is right; doing it silently is not — the server may
// have been in rescue, and this one was.
func (ms *MockSuite) TestBaremetalPowerOnSaysWhenItIsGuessing(assert, require *td.T) {
	registerPower(assert, "poweroff", 95083)

	out, err := cmd.Execute("baremetal", "power", "on", powerServer)

	require.CmpNoError(err)
	require.Cmp(len(powerBodies["boot"]), 1)
	assert.Cmp(powerBodies["boot"][0]["bootId"], float64(1), "the disk entry, from the API")
	assert.Cmp(out, td.Contains("no record of the previous boot"),
		"and the operator is told this is a fallback")
}

func (ms *MockSuite) TestBaremetalPowerOnHonoursAnExplicitBoot(assert, require *td.T) {
	registerPower(assert, "poweron", 230242)
	_, err := cmd.Execute("baremetal", "power", "off", powerServer, "--yes")
	require.CmpNoError(err)

	cmd.PostExecute()
	powerBodies = map[string][]map[string]any{}
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/"+powerServer,
		httpmock.NewStringResponder(200, `{"powerState": "poweroff", "state": "ok", "bootId": 95083}`))

	out, err := cmd.Execute("baremetal", "power", "on", powerServer, "--boot", "1")

	require.CmpNoError(err)
	assert.Cmp(powerBodies["boot"][0]["bootId"], float64(1), "--boot wins over the remembered value")
	assert.Cmp(out, td.Contains("--boot"))
}

// Powering off interrupts everything running on the machine: an unattended run
// that did not say --yes must not do it.
func (ms *MockSuite) TestBaremetalPowerOffRefusesWithoutConfirmation(assert, require *td.T) {
	registerPower(assert, "poweron", 230242)

	_, err := cmd.Execute("baremetal", "power", "off", powerServer)

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("cancelled"))
	assert.Cmp(powerCallCount("PUT", ""), 0, "the boot is not changed")
	assert.Cmp(powerCallCount("POST", "/reboot"), 0, "and nothing is rebooted")
}

// The PUT is the call that outlives the reboot: it rewrites the boot
// configuration, and every later reboot obeys it. A preview showing the reboot
// alone would hide exactly that.
func (ms *MockSuite) TestBaremetalPowerOffDryRunDescribesBothCalls(assert, require *td.T) {
	registerPower(assert, "poweron", 230242)

	out, err := cmd.Execute("baremetal", "power", "off", powerServer, "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("PUT /v1/dedicated/server/"+powerServer))
	assert.Cmp(out, td.Contains("POST /v1/dedicated/server/"+powerServer+"/reboot"))
	assert.Cmp(powerCallCount("PUT", ""), 0)
	assert.Cmp(powerCallCount("POST", "/reboot"), 0)
}

// A server left on the power-off entry shuts itself down at every reboot,
// including one asked for from the manager. That is the trap this whole
// mechanism creates, and `status` is where it gets said.
func (ms *MockSuite) TestBaremetalPowerStatusWarnsAboutThePowerOffBoot(assert, require *td.T) {
	registerPower(assert, "poweroff", 95083)

	out, err := cmd.Execute("baremetal", "power", "status", powerServer)

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("will shut it down again"))
}

func (ms *MockSuite) TestBaremetalPowerStatusIsQuietOnANormalBoot(assert, require *td.T) {
	registerPower(assert, "poweron", 230242)

	out, err := cmd.Execute("baremetal", "power", "status", powerServer)

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Customer rescue system"), "the boot is named, not just numbered")
	assert.Cmp(out, td.Not(td.Contains("will shut it down again")), "and no warning is invented")
}

func (ms *MockSuite) TestBaremetalPowerOffRefusesAServerAlreadyOff(assert, require *td.T) {
	registerPower(assert, "poweroff", 95083)

	_, err := cmd.Execute("baremetal", "power", "off", powerServer, "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("already off"))
	assert.Cmp(powerCallCount("PUT", ""), 0)
}

// `power on` on a running server is not a no-op: it takes the power-off entry
// away, so that the next reboot does not shut the machine down. What it must
// not do is reboot a server nobody asked to restart.
func (ms *MockSuite) TestBaremetalPowerOnDoesNotRebootARunningServer(assert, require *td.T) {
	registerPower(assert, "poweron", 95083)

	out, err := cmd.Execute("baremetal", "power", "on", powerServer)

	require.CmpNoError(err)
	assert.Cmp(len(powerBodies["boot"]), 1, "the boot is put back")
	assert.Cmp(powerCallCount("POST", "/reboot"), 0, "and the server is left running")
	assert.Cmp(out, td.Contains("already running"))
}
