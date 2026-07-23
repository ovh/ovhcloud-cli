// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// TestCloudProjectListCmd tests the "cloud project list" command
func (ms *MockSuite) TestCloudProjectListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/publicCloud/project",
		httpmock.NewStringResponder(200, `[
			{
				"id": "project-1",
				"resourceStatus": "READY",
				"currentState": {"name": "Test Project 1", "mode": "standard"}
			},
			{
				"id": "project-2",
				"resourceStatus": "READY",
				"currentState": {"name": "Test Project 2", "mode": "discovery"}
			}
		]`).Once())

	out, err := cmd.Execute("cloud", "project", "list")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("project-1"))
	assert.Cmp(out, td.Contains("Test Project 1"))
	assert.Cmp(out, td.Contains("project-2"))
	assert.Cmp(out, td.Contains("discovery"))
}

// TestCloudProjectListCmdAlias tests the "cloud project ls" command alias
func (ms *MockSuite) TestCloudProjectListCmdAlias(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/publicCloud/project",
		httpmock.NewStringResponder(200, `[
			{
				"id": "project-1",
				"resourceStatus": "READY",
				"currentState": {"name": "Test Project", "mode": "standard"}
			}
		]`).Once())

	out, err := cmd.Execute("cloud", "project", "ls")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("project-1"))
}

// TestCloudProjectGetCmd tests the "cloud project get" command
func (ms *MockSuite) TestCloudProjectGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/publicCloud/project/project-123",
		httpmock.NewStringResponder(200, `{
			"id": "project-123",
			"resourceStatus": "READY",
			"createdAt": "2023-01-15T10:30:00Z",
			"currentState": {
				"name": "My Cloud Project",
				"mode": "standard"
			},
			"iam": {
				"urn": "urn:v1:eu:resource:cloudProject:project-123"
			}
		}`).Once())

	out, err := cmd.Execute("cloud", "project", "get", "project-123")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("project-123"))
	assert.Cmp(out, td.Contains("My Cloud Project"))
	assert.Cmp(out, td.Contains("standard"))
}

// TestCloudProjectEditCmd tests the "cloud project edit" command
func (ms *MockSuite) TestCloudProjectEditCmd(assert, require *td.T) {
	// Mock GET to retrieve current project state
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/project-456",
		httpmock.NewStringResponder(200, `{
			"project_id": "project-456",
			"projectName": "Original Name",
			"status": "ok",
			"description": "Original description",
			"manualQuota": false
		}`).Once())

	// Mock PUT to update the project
	httpmock.RegisterResponder("PUT", "https://eu.api.ovh.com/v1/cloud/project/project-456",
		httpmock.NewStringResponder(200, `{
			"project_id": "project-456",
			"projectName": "Original Name",
			"status": "ok",
			"description": "Updated description",
			"manualQuota": false
		}`).Once())

	out, err := cmd.Execute("cloud", "project", "edit", "project-456", "--description", "Updated description")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅"))
}

// TestCloudProjectEditCmdWithManualQuota tests the edit command with manual quota flag
func (ms *MockSuite) TestCloudProjectEditCmdWithManualQuota(assert, require *td.T) {
	// Mock GET to retrieve current project state
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/project-789",
		httpmock.NewStringResponder(200, `{
			"project_id": "project-789",
			"projectName": "Test Project",
			"status": "ok",
			"description": "Test description",
			"manualQuota": false
		}`).Once())

	// Mock PUT to update the project
	httpmock.RegisterResponder("PUT", "https://eu.api.ovh.com/v1/cloud/project/project-789",
		httpmock.NewStringResponder(200, `{
			"project_id": "project-789",
			"projectName": "Test Project",
			"status": "ok",
			"description": "Test description",
			"manualQuota": true
		}`).Once())

	out, err := cmd.Execute("cloud", "project", "edit", "project-789", "--manual-quota")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅"))
}

// TestCloudProjectServiceInfoCmd tests the "cloud project service-info" command
func (ms *MockSuite) TestCloudProjectServiceInfoCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/test-project/serviceInfos",
		httpmock.NewStringResponder(200, `{
			"renew": {
				"forced": true,
				"manualPayment": false,
				"deleteAtExpiration": false,
				"automatic": true,
				"period": 1
			},
			"serviceId": 12345,
			"status": "active",
			"creation": "2023-01-15T10:30:00Z",
			"expiration": "2024-01-15T10:30:00Z",
			"domain": "test.ovhcloud.com",
			"contactAdmin": "admin-nic",
			"contactTech": "tech-nic",
			"contactBilling": "billing-nic"
		}`).Once())

	out, err := cmd.Execute("cloud", "project", "service-info", "--cloud-project", "test-project", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("serviceId"))
	assert.Cmp(out, td.Contains("active"))
}

// TestCloudProjectChangeContactCmd tests the "cloud project change-contact" command
func (ms *MockSuite) TestCloudProjectChangeContactCmd(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/cloud/project/test-project/changeContact",
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute("cloud", "project", "change-contact", "--cloud-project", "test-project",
		"--contact-admin", "new-admin-nic")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅"))
	assert.Cmp(out, td.Contains("Contact change request submitted successfully"))
}

// TestCloudProjectChangeContactCmdMultipleContacts tests changing multiple contacts
func (ms *MockSuite) TestCloudProjectChangeContactCmdMultipleContacts(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/cloud/project/test-project/changeContact",
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute("cloud", "project", "change-contact", "--cloud-project", "test-project",
		"--contact-admin", "new-admin-nic",
		"--contact-billing", "new-billing-nic",
		"--contact-tech", "new-tech-nic")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅"))
	assert.Cmp(out, td.Contains("Contact change request submitted successfully"))
}

// TestCloudProjectChangeContactCmdNoParams tests change-contact with no parameters
func (ms *MockSuite) TestCloudProjectChangeContactCmdNoParams(assert, require *td.T) {
	out, err := cmd.Execute("cloud", "project", "change-contact", "--cloud-project", "test-project")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("🟠"))
	assert.Cmp(out, td.Contains("No parameters given, nothing to change"))
}

// TestCloudProjectTerminationInitCmd tests the "cloud project termination init" command
func (ms *MockSuite) TestCloudProjectTerminationInitCmd(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/cloud/project/test-project/terminate",
		httpmock.NewStringResponder(200, `{
			"token": "termination-token-12345"
		}`).Once())

	out, err := cmd.Execute("cloud", "project", "termination", "init", "--cloud-project", "test-project")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅"))
	assert.Cmp(out, td.Contains("Termination initiated"))
	assert.Cmp(out, td.Contains("termination-token-12345"))
}

// TestCloudProjectTerminationConfirmCmd tests the "cloud project termination confirm" command
func (ms *MockSuite) TestCloudProjectTerminationConfirmCmd(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/cloud/project/test-project/confirmTermination",
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute("cloud", "project", "termination", "confirm",
		"--cloud-project", "test-project",
		"--token", "termination-token-12345")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅"))
	assert.Cmp(out, td.Contains("Project termination confirmed successfully"))
}

// TestCloudProjectTerminationCancelCmd tests the "cloud project termination cancel" command
func (ms *MockSuite) TestCloudProjectTerminationCancelCmd(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/cloud/project/test-project/retain",
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute("cloud", "project", "termination", "cancel", "--cloud-project", "test-project")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅"))
	assert.Cmp(out, td.Contains("Project retained successfully"))
}

// TestCloudProjectUnleashCmd tests the "cloud project unleash" command
func (ms *MockSuite) TestCloudProjectUnleashCmd(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/cloud/project/test-project/unleash",
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute("cloud", "project", "unleash", "--cloud-project", "test-project")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅"))
	assert.Cmp(out, td.Contains("Project unleashed successfully"))
}
