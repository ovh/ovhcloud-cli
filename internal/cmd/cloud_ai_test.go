// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// TestCloudAIAuthorizationGetCmd tests the "cloud ai authorization get" command
func (ms *MockSuite) TestCloudAIAuthorizationGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/test-project/ai/authorization",
		httpmock.NewStringResponder(200, `{"authorized": true}`).Once())

	out, err := cmd.Execute("cloud", "ai", "authorization", "get",
		"--cloud-project", "test-project", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("authorized"))
	assert.Cmp(out, td.Contains("true"))
}

// TestCloudAIAuthorizationCreateCmd tests the "cloud ai authorization create" command
func (ms *MockSuite) TestCloudAIAuthorizationCreateCmd(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/cloud/project/test-project/ai/authorization",
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute("cloud", "ai", "authorization", "create",
		"--cloud-project", "test-project")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅"))
	assert.Cmp(out, td.Contains("AI Endpoints authorized successfully"))
}
