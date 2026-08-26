// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
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

// TestVpsReinstallResolvesSSHKeyCmd checks that "--ssh-key <name>" is resolved
// against /me/sshKey and sent as publicSshKey (not as the name), so the key
// actually populates authorized_keys (issue #260).
func (ms *MockSuite) TestVpsReinstallResolvesSSHKeyCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/me/sshKey",
		httpmock.NewStringResponder(200, `["mykey"]`))
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/me/sshKey/mykey",
		httpmock.NewStringResponder(200, `{"keyName": "mykey", "key": "ssh-ed25519 AAAATEST test@host"}`))

	var body map[string]any
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/vps/vps-67890/rebuild",
		func(req *http.Request) (*http.Response, error) {
			_ = json.NewDecoder(req.Body).Decode(&body)
			return httpmock.NewStringResponse(200, `{"id": 123}`), nil
		})

	_, err := cmd.Execute("vps", "reinstall", "vps-67890",
		"--image-id", "img-1", "--ssh-key", "mykey", "--do-not-send-password")

	require.CmpNoError(err)
	assert.Cmp(body["publicSshKey"], "ssh-ed25519 AAAATEST test@host")
	assert.Cmp(body["sshKey"], nil) // the name must not be sent as-is
	assert.Cmp(body["imageId"], "img-1")
	assert.Cmp(body["doNotSendPassword"], true)
}

// TestVpsReinstallUnknownSSHKeyCmd checks that an unknown "--ssh-key" name fails
// fast with a clear error instead of silently leaving authorized_keys empty (#260).
func (ms *MockSuite) TestVpsReinstallUnknownSSHKeyCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/me/sshKey",
		httpmock.NewStringResponder(200, `["other"]`))
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/me/sshKey/other",
		httpmock.NewStringResponder(200, `{"keyName": "other", "key": "ssh-ed25519 AAAAOTHER"}`))

	_, err := cmd.Execute("vps", "reinstall", "vps-67890",
		"--image-id", "img-1", "--ssh-key", "missing")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("not found"))
}
