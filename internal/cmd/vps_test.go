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

func (ms *MockSuite) TestVpsListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/vps",
		httpmock.NewStringResponder(200, `["vps-12345","vps-67890"]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/vps/vps-12345",
		httpmock.NewStringResponder(200, `{"name": "vps-12345", "displayName": "VPS 12345", "state": "running", "zone": "Region OpenStack: os-waw2"}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/vps/vps-67890",
		httpmock.NewStringResponder(200, `{"name": "vps-67890", "displayName": "VPS 67890", "state": "stopped", "zone": "Region OpenStack: os-gra1"}`).Once())

	out, err := cmd.Execute("vps", "ls", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"displayName": "VPS 12345",
			"name": "vps-12345",
			"state": "running",
			"zone": "Region OpenStack: os-waw2"
		},
		{
			"displayName": "VPS 67890",
			"name": "vps-67890",
			"state": "stopped",
			"zone": "Region OpenStack: os-gra1"
		}
	]`))
}

func (ms *MockSuite) TestVpsGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/vps/vps-67890",
		httpmock.NewStringResponder(200, `{"name": "vps-67890", "displayName": "VPS 67890", "state": "stopped", "zone": "Region OpenStack: os-gra1"}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/vps/vps-67890/datacenter",
		httpmock.NewStringResponder(200, `{"country": "fr", "name": "os-gra1", "longName": "Region OpenStack: os-gra1"}`).Once())

	out, err := cmd.Execute("vps", "get", "vps-67890", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"displayName": "VPS 67890",
		"name": "vps-67890",
		"state": "stopped",
		"zone": "Region OpenStack: os-gra1",
		"datacenter": {
			"country": "fr",
			"name": "os-gra1",
			"longName": "Region OpenStack: os-gra1"
		}
	}`))
}

// registerVpsServiceInfos wires a service whose renewal is currently automatic,
// and captures whatever the CLI decides to write back.
func registerVpsServiceInfos(captured *map[string]any) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/vps/fakeVps/serviceInfos",
		httpmock.NewStringResponder(200, `{
			"serviceId": 1,
			"domain": "fakeVps",
			"renew": {"automatic": true, "deleteAtExpiration": false, "forced": false, "manualPayment": false, "period": 1}
		}`),
	)
	httpmock.RegisterResponder("PUT", "https://eu.api.ovh.com/v1/vps/fakeVps/serviceInfos",
		func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			var sent map[string]any
			if err := json.Unmarshal(body, &sent); err != nil {
				return nil, err
			}
			*captured = sent
			return httpmock.NewStringResponse(200, `null`), nil
		},
	)
}

// Editing the renewal period used to send every other renewal setting along
// with it, at its zero value: a service that renewed itself automatically for
// years stopped doing so, and nothing in the output said it had changed.
func (ms *MockSuite) TestVpsServiceInfoEditSendsOnlyWhatWasAsked(assert, require *td.T) {
	var sent map[string]any
	registerVpsServiceInfos(&sent)

	_, err := cmd.Execute("vps", "service-info", "edit", "fakeVps", "--renew-period", "12")

	require.CmpNoError(err)
	renew, _ := sent["renew"].(map[string]any)
	require.NotNil(renew, "the renewal block must be written")
	assert.Cmp(renew["period"], float64(12), "the period the operator asked for")
	assert.Cmp(renew["automatic"], true, "automatic renewal must survive untouched")
}

// The flag being absent and the flag being set to false are different
// intentions, and pflag can tell them apart: an explicit false must be sent.
func (ms *MockSuite) TestVpsServiceInfoEditSendsAnExplicitFalse(assert, require *td.T) {
	var sent map[string]any
	registerVpsServiceInfos(&sent)

	_, err := cmd.Execute("vps", "service-info", "edit", "fakeVps", "--renew-automatic=false")

	require.CmpNoError(err)
	renew, _ := sent["renew"].(map[string]any)
	require.NotNil(renew)
	assert.Cmp(renew["automatic"], false, "the operator asked for it, so it is sent")
}
