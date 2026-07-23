// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestCloudSSHKeyListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/sshKey",
		httpmock.NewStringResponder(200, `[
			{
				"name": "my-key",
				"publicKey": "ssh-rsa AAAAB3Nza example-key",
				"createdAt": "2025-01-01T00:00:00Z",
				"updatedAt": "2025-01-02T00:00:00Z"
			}
		]`))

	out, err := cmd.Execute("cloud", "ssh-key", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("my-key"))
	assert.Cmp(out, td.Contains("2025-01-01T00:00:00Z"))
}

func (ms *MockSuite) TestCloudSSHKeyGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/sshKey/my-key",
		httpmock.NewStringResponder(200, `{
			"name": "my-key",
			"publicKey": "ssh-rsa AAAAB3Nza example-key",
			"createdAt": "2025-01-01T00:00:00Z",
			"updatedAt": "2025-01-02T00:00:00Z"
		}`))

	out, err := cmd.Execute("cloud", "ssh-key", "get", "my-key", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("my-key"))
	assert.Cmp(out, td.Contains("ssh-rsa AAAAB3Nza example-key"))
}

func (ms *MockSuite) TestCloudSSHKeyCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/sshKey",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"name": "my-key",
				"publicKey": "ssh-rsa AAAAB3Nza example-key"
			}`),
		),
		httpmock.NewStringResponder(200, `{
			"name": "my-key",
			"publicKey": "ssh-rsa AAAAB3Nza example-key"
		}`),
	)

	out, err := cmd.Execute("cloud", "ssh-key", "create", "--cloud-project", "fakeProjectID", "--name", "my-key", "--public-key", "ssh-rsa AAAAB3Nza example-key")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅ SSH key successfully created"))
}

func (ms *MockSuite) TestCloudSSHKeyDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/sshKey/my-key",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "ssh-key", "delete", "my-key", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅ SSH key successfully deleted"))
}

func (ms *MockSuite) TestCloudSSHKeyDeleteCmdError(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/sshKey/missing-key",
		httpmock.NewStringResponder(404, `{"class":"Client::NotFound::SSHKeyDoesNotExist","message":"SSH key not found"}`))

	_, err := cmd.Execute("cloud", "ssh-key", "delete", "missing-key", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("error deleting SSH key"))
}
