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

	powerServer = "ns0000001.ip-203-0-113.eu"
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

// The window this covers is not a race, it is the normal case: the PUT lands at
// once and the machine only goes dark 143 seconds later, so for two minutes the
// API answers powerState "poweron" sitting on the power-off entry. A second
// `power off` typed in that window used to overwrite the record with that very
// entry, and `power on` then restored a power-off — announcing it, wrongly, as
// the boot the server was on before.
func (ms *MockSuite) TestBaremetalPowerOffTwiceKeepsTheBootItWasReallyOn(assert, require *td.T) {
	registerPower(assert, "poweron", 230242)
	_, err := cmd.Execute("baremetal", "power", "off", powerServer, "--yes")
	require.CmpNoError(err)

	// Still shutting down: running, and already on the power-off entry.
	cmd.PostExecute()
	powerBodies = map[string][]map[string]any{}
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/"+powerServer,
		httpmock.NewStringResponder(200, `{"powerState": "poweron", "state": "ok", "bootId": 95083}`))
	_, err = cmd.Execute("baremetal", "power", "off", powerServer, "--yes")
	require.CmpNoError(err)

	cmd.PostExecute()
	powerBodies = map[string][]map[string]any{}
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/"+powerServer,
		httpmock.NewStringResponder(200, `{"powerState": "poweroff", "state": "ok", "bootId": 95083}`))

	_, err = cmd.Execute("baremetal", "power", "on", powerServer)

	require.CmpNoError(err)
	require.Cmp(len(powerBodies["boot"]), 1)
	assert.Cmp(powerBodies["boot"][0]["bootId"], td.Not(float64(95083)),
		"never the power-off entry: restoring it puts the server back where it shuts down")
	assert.Cmp(powerBodies["boot"][0]["bootId"], float64(230242), "the rescue boot it was really on")
}

// The same window, but nothing was ever recorded: someone else powered the
// server off. There is no previous boot to report, and reporting the power-off
// entry as one would be worse than reporting none.
func (ms *MockSuite) TestBaremetalPowerOffOnThePowerEntryReportsNoPreviousBoot(assert, require *td.T) {
	registerPower(assert, "poweron", 95083)

	out, err := cmd.Execute("baremetal", "power", "off", powerServer, "--yes", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Not(td.Contains("previousBootId")),
		"absent, so a caller can tell an unknown previous boot from a known one")
	assert.Cmp(out, td.Contains("95083"), "the entry it was put on is still reported")
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

// `power status` renders one object, so it has no rows to filter. A flag that
// can only be ignored is worse than an absent one: docgen documents it, cobra
// accepts it, and the operator believes the answer was narrowed.
func (ms *MockSuite) TestBaremetalPowerStatusOffersNoFilterFlag(assert, require *td.T) {
	_, err := cmd.Execute("baremetal", "power", "status", powerServer, "--filter", `x=="y"`)

	require.CmpError(err, "--filter must not be accepted on a single-object command")
	assert.Cmp(err.Error(), td.Contains("filter"))
}

// `power on` puts the boot back so a later reboot does not shut the machine
// down. When the machine is already on that boot, there is nothing to put back,
// and the PUT it used to send was a write to the boot configuration of a
// production server bought for nothing — announced, on top, as "its boot is now
// N", which was not true because the boot had not moved.
//
// `power off` already makes this exact comparison before deciding what to
// record. Leaving it out of `power on` was an omission, not a difference.
func (ms *MockSuite) TestBaremetalPowerOnDoesNotRewriteABootAlreadyInPlace(assert, require *td.T) {
	// Running on the disk entry, which is also what bootToRestore falls back to
	// when this machine has no record — so target and current state coincide.
	registerPower(assert, "poweron", 1)

	out, err := cmd.Execute("baremetal", "power", "on", powerServer)

	require.CmpNoError(err)
	assert.Cmp(powerCallCount("PUT", ""), 0, "no write to the boot configuration")
	assert.Cmp(powerCallCount("POST", "/reboot"), 0, "and no reboot")
	assert.Cmp(out, td.Contains("Nothing to do"))
}

// The same, reached the other way: the operator names the boot the server is
// already on. An explicit --boot is a statement of intent, not a reason to write
// a value that is already there.
func (ms *MockSuite) TestBaremetalPowerOnHonoursAnExplicitBootAlreadySet(assert, require *td.T) {
	registerPower(assert, "poweron", 230242)

	out, err := cmd.Execute("baremetal", "power", "on", powerServer, "--boot", "230242")

	require.CmpNoError(err)
	assert.Cmp(powerCallCount("PUT", ""), 0)
	assert.Cmp(out, td.Contains("already running on boot 230242"))
}

// And the preview of that no-op must not have a side effect: the record of the
// boot the machine came from is local state, so --dry-run leaves it alone. The
// second run proves it survived — it still names the remembered boot rather than
// falling back to the disk.
func (ms *MockSuite) TestBaremetalPowerOnDryRunKeepsTheRememberedBoot(assert, require *td.T) {
	registerPower(assert, "poweron", 230242)
	_, err := cmd.Execute("baremetal", "power", "off", powerServer, "--yes")
	require.CmpNoError(err)

	// Now off, and on the entry it was told to restore — so `power on` has
	// nothing to do but still holds a record worth keeping.
	cmd.PostExecute()
	powerBodies = map[string][]map[string]any{}
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/"+powerServer,
		httpmock.NewStringResponder(200, `{"powerState": "poweron", "state": "ok", "bootId": 230242}`))

	_, err = cmd.Execute("baremetal", "power", "on", powerServer, "--dry-run")
	require.CmpNoError(err)

	cmd.PostExecute()
	out, err := cmd.Execute("baremetal", "power", "on", powerServer)

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("was on before it was powered off"),
		"the dry run did not consume the record")
}
