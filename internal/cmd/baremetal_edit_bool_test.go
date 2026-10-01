// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
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

const editServer = "https://eu.api.ovh.com/v1/dedicated/server/ns1.example"

// registerEditableServer answers the read EditResource does before its write,
// with both booleans on, and records the body of the PUT.
func registerEditableServer(body *map[string]any) {
	httpmock.RegisterResponder("GET", editServer,
		httpmock.NewStringResponder(200,
			`{"name":"ns1.example","monitoring":true,"noIntervention":true,"state":"ok","rootDevice":null}`))
	httpmock.RegisterResponder("PUT", editServer, func(req *http.Request) (*http.Response, error) {
		raw, _ := io.ReadAll(req.Body)
		_ = json.Unmarshal(raw, body)
		return httpmock.NewStringResponse(200, `null`), nil
	})
}

// `omitempty` on a bool drops false, so turning a flag off sent nothing —
// and EditResource reads the object, merges the empty command-line map into it
// and PUTs the result, so the current true came straight back while the command
// printed "✅ Resource updated successfully". The doctor's remedy for a server
// refusing hardware intervention was exactly this call.
func (ms *MockSuite) TestBaremetalEditSendsAnExplicitFalse(assert, require *td.T) {
	var body map[string]any
	registerEditableServer(&body)

	_, err := cmd.Execute("baremetal", "edit", "ns1.example", "--no-intervention=false")

	require.CmpNoError(err)
	require.NotNil(body["noIntervention"], "the field must be in the body at all")
	assert.Cmp(body["noIntervention"], false, "and it must carry the false that was asked for")
}

// Monitoring had the same defect and looked healthy, because the only caller
// that names it happens to pass true.
func (ms *MockSuite) TestBaremetalEditSendsAnExplicitFalseForMonitoring(assert, require *td.T) {
	var body map[string]any
	registerEditableServer(&body)

	_, err := cmd.Execute("baremetal", "edit", "ns1.example", "--monitoring=false")

	require.CmpNoError(err)
	assert.Cmp(body["monitoring"], false)
}

// The other half: a flag not typed must not appear, or every edit would rewrite
// fields it was never asked about.
func (ms *MockSuite) TestBaremetalEditLeavesUntypedBooleansAlone(assert, require *td.T) {
	var body map[string]any
	registerEditableServer(&body)

	_, err := cmd.Execute("baremetal", "edit", "ns1.example", "--boot-script", "hello")

	require.CmpNoError(err)
	assert.Cmp(body["monitoring"], true, "the value read back, not one this command invented")
	assert.Cmp(body["noIntervention"], true)
}

// The wasm build keeps one process across invocations, so a pointer set by an
// earlier command is still set for the next one — and the flag's Changed bit is
// reset in between, so the "was it typed" question answers no while the pointer
// says yes. Two commands in one process, which is the shape that exposes it.
func (ms *MockSuite) TestBaremetalEditDoesNotCarryABooleanIntoTheNextRun(assert, require *td.T) {
	var first map[string]any
	registerEditableServer(&first)

	_, err := cmd.Execute("baremetal", "edit", "ns1.example", "--no-intervention=false")
	require.CmpNoError(err)
	require.Cmp(first["noIntervention"], false, "the first command asked for it")

	cmd.PostExecute()

	var second map[string]any
	registerEditableServer(&second)

	_, err = cmd.Execute("baremetal", "edit", "ns1.example", "--boot-script", "hello")

	require.CmpNoError(err)
	assert.Cmp(second["noIntervention"], true,
		"the second command did not ask, so it must send back what it read")
}
