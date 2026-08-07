// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestCloudExportTerraformCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/network/private",
		httpmock.NewStringResponder(200, `[
			{ "id": "net-123", "name": "my-network", "region": "GRA9" }
		]`))
	// User ids are numbers in the API: exercise the numeric-id normalisation.
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/user",
		httpmock.NewStringResponder(200, `[
			{ "id": 12345, "username": "admin", "description": "ci user" }
		]`))

	outDir := require.TempDir()

	out, err := cmd.Execute("cloud", "export-terraform",
		"--cloud-project", "fakeProjectID",
		"--output-dir", outDir)

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Exported 2 resource"))

	content, readErr := os.ReadFile(filepath.Join(outDir, "imports.tf"))
	require.CmpNoError(readErr)

	generated := string(content)
	// Network import block (import id = service_name/network_id).
	assert.Cmp(generated, td.Contains("to = ovh_cloud_project_network_private.my-network"))
	assert.Cmp(generated, td.Contains(`id = "fakeProjectID/net-123"`))
	// User import block (import id = service_name/id, numeric id stringified).
	assert.Cmp(generated, td.Contains("to = ovh_cloud_project_user.admin"))
	assert.Cmp(generated, td.Contains(`id = "fakeProjectID/12345"`))
}

func (ms *MockSuite) TestCloudExportTerraformResourceFilterCmd(assert, require *td.T) {
	// With --resources user, only the user endpoint must be queried; registering
	// no network responder ensures the network endpoint is never called.
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/user",
		httpmock.NewStringResponder(200, `[ { "id": 42, "username": "solo" } ]`))

	outDir := require.TempDir()

	out, err := cmd.Execute("cloud", "export-terraform",
		"--cloud-project", "fakeProjectID",
		"--resources", "user",
		"--output-dir", outDir)

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Exported 1 resource"))

	content, readErr := os.ReadFile(filepath.Join(outDir, "imports.tf"))
	require.CmpNoError(readErr)
	assert.Cmp(string(content), td.Contains("ovh_cloud_project_user.solo"))
	assert.Cmp(string(content), td.Not(td.Contains("network_private")))
}
