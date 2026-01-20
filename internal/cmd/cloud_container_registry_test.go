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

func (ms *MockSuite) TestCloudContainerRegistryListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry",
		httpmock.NewStringResponder(200, `[
			{
				"createdAt": "2025-08-22T09:24:18.953364Z",
				"deliveredAt": "2025-08-22T09:26:54.540629Z",
				"iamEnabled": false,
				"id": "0b1b2dc2-952b-11f0-afd9-0050568ce122",
				"name": "ZuperRegistry",
				"region": "EU-WEST-PAR",
				"size": 0,
				"status": "READY",
				"updatedAt": "2025-08-22T09:28:41.468178Z",
				"url": "https://fake.url.bhs5.container-registry.ovh.net",
				"version": "2.12.2"
			}
		]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region",
		httpmock.NewStringResponder(200, `["GRA", "EU-WEST-PAR"]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA",
		httpmock.NewStringResponder(200, `{
			"name": "GRA",
			"type": "region",
			"status": "UP",
			"services": [],
			"countryCode": "fr",
			"ipCountries": [],
			"continentCode": "NA",
			"availabilityZones": [],
			"datacenterLocation": "GRA"
		}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/EU-WEST-PAR",
		httpmock.NewStringResponder(200, `{
			"name": "EU-WEST-PAR",
			"type": "region-3-az",
			"status": "UP",
			"services": [],
			"countryCode": "fr",
			"ipCountries": [],
			"continentCode": "EU",
			"availabilityZones": [],
			"datacenterLocation": "EU-WEST-PAR"
		}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/0b1b2dc2-952b-11f0-afd9-0050568ce122/plan",
		httpmock.NewStringResponder(200, `{
			"code": "registry.s-plan-equivalent.hour.consumption",
			"createdAt": "2019-09-13T15:53:33.599585Z",
			"updatedAt": "2021-03-29T10:09:03.960847Z",
			"name": "SMALL",
			"id": "9f728ba5-998b-4401-ab0f-497cd8bc6a89",
			"registryLimits": {
				"imageStorage": 214748364800,
				"parallelRequest": 15
			},
			"features": {
				"vulnerability": false
			}
		}`).Once())

	out, err := cmd.Execute("cloud", "container-registry", "ls", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌──────────────────────────────────────┬───────────────┬─────────────┬───────┬────────────────┬─────────┬────────┐
│                  id                  │     name      │   region    │ plan  │ deploymentMode │ version │ status │
├──────────────────────────────────────┼───────────────┼─────────────┼───────┼────────────────┼─────────┼────────┤
│ 0b1b2dc2-952b-11f0-afd9-0050568ce122 │ ZuperRegistry │ EU-WEST-PAR │ SMALL │ 3-AZ           │ 2.12.2  │ READY  │
└──────────────────────────────────────┴───────────────┴─────────────┴───────┴────────────────┴─────────┴────────┘
💡 Use option --json or --yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestCloudContainerRegistryGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000",
		httpmock.NewStringResponder(200, `{
			"createdAt": "2025-08-22T09:24:18.953364Z",
			"deliveredAt": "2025-08-22T09:26:54.540629Z",
			"iamEnabled": false,
			"id": "550e8400-e29b-41d4-a716-446655440000",
			"name": "MyRegistry",
			"region": "GRA",
			"size": 1073741824,
			"status": "READY",
			"updatedAt": "2025-08-22T09:28:41.468178Z",
			"url": "https://registry123.gra.container-registry.ovh.net",
			"version": "2.12.2"
		}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/plan",
		httpmock.NewStringResponder(200, `{
			"code": "registry.m-plan-equivalent.hour.consumption",
			"createdAt": "2019-09-13T15:53:33.599585Z",
			"updatedAt": "2021-03-29T10:09:03.960847Z",
			"name": "MEDIUM",
			"id": "9f728ba5-998b-4401-ab0f-497cd8bc6a89",
			"registryLimits": {
				"imageStorage": 644245094400,
				"parallelRequest": 30
			},
			"features": {
				"vulnerability": true
			}
		}`).Once())

	out, err := cmd.Execute("cloud", "container-registry", "get", "550e8400-e29b-41d4-a716-446655440000", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 🚀 Managed Private Registry 550e8400-e29b-41d4-a716-446655440000

  *MyRegistry*

  ## General information

  **Region**:        GRA
  **Status**:        READY
  **Creation date**: 2025-08-22T09:24:18.953364Z
  **Delivery date**: 2025-08-22T09:26:54.540629Z
  **Update date**:   2025-08-22T09:28:41.468178Z

  ## Registry state

  **Version**:     2.12.2
  **Plan**:        MEDIUM
  **Usage**:       1.00 GiB / 600 GiB
  **IAM enabled**: false

  Registry URL https://registry123.gra.container-registry.ovh.net

  💡 Use option --json or --yaml to get the raw output with all information

`)
}

func (ms *MockSuite) TestCloudContainerRegistryCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry",
		tdhttpmock.JSONBody(td.JSON(`{
			"name": "NewRegistry",
			"region": "GRA",
			"planID": "plan-id-123"
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "7f8e9d0c-1a2b-3c4d-5e6f-7a8b9c0d1e2f",
			"name": "NewRegistry",
			"region": "GRA",
			"status": "INSTALLING"
		}`).Once())

	out, err := cmd.Execute("cloud", "container-registry", "create", "--name", "NewRegistry", "--region", "GRA", "--plan-id", "plan-id-123", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Container registry '7f8e9d0c-1a2b-3c4d-5e6f-7a8b9c0d1e2f' created successfully`)
}

func (ms *MockSuite) TestCloudContainerRegistryEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000",
		httpmock.NewStringResponder(200, `{
			"id": "550e8400-e29b-41d4-a716-446655440000",
			"name": "OldName",
			"region": "GRA",
			"status": "READY"
		}`).Once())

	httpmock.RegisterMatcherResponder(
		http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000",
		tdhttpmock.JSONBody(td.JSON(`{
			"name": "NewName"
		}`)),
		httpmock.NewStringResponder(200, `{"id": "550e8400-e29b-41d4-a716-446655440000"}`).Once())

	out, err := cmd.Execute("cloud", "container-registry", "edit", "550e8400-e29b-41d4-a716-446655440000", "--name", "NewName", "--cloud-project", "fakeProjectID", "--yaml")

	require.CmpNoError(err)
	assert.String(out, `message: ✅ Resource updated successfully
`)
}

func (ms *MockSuite) TestCloudContainerRegistryDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000",
		httpmock.NewStringResponder(204, ``).Once())

	out, err := cmd.Execute("cloud", "container-registry", "delete", "550e8400-e29b-41d4-a716-446655440000", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Container registry deleted successfully`)
}

func (ms *MockSuite) TestCloudContainerRegistryUsersListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/users",
		httpmock.NewStringResponder(200, `[
			{
				"id": 1,
				"user": "user1",
				"email": "user1@example.com"
			},
			{
				"id": 2,
				"user": "admin-user",
				"email": "admin@example.com"
			}
		]`).Once())

	out, err := cmd.Execute("cloud", "container-registry", "users", "list", "550e8400-e29b-41d4-a716-446655440000", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌────┬────────────┬───────────────────┐
│ id │    user    │       email       │
├────┼────────────┼───────────────────┤
│ 1  │ user1      │ user1@example.com │
│ 2  │ admin-user │ admin@example.com │
└────┴────────────┴───────────────────┘
💡 Use option --json or --yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestCloudContainerRegistryUsersGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/users/42",
		httpmock.NewStringResponder(200, `{
			"id": 42,
			"user": "testuser",
			"email": "testuser@example.com"
		}`).Once())

	out, err := cmd.Execute("cloud", "container-registry", "users", "get", "550e8400-e29b-41d4-a716-446655440000", "42", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 🚀 Managed Private Registry user

  ## General information

  **ID**:       42
  **Username**: testuser
  **Email**:    testuser@example.com mailto:testuser@example.com

  💡 Use option --json or --yaml to get the raw output with all information

`)
}

func (ms *MockSuite) TestCloudContainerRegistryUsersCreateCmdWithAllOptionalParameters(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/users",
		tdhttpmock.JSONBody(td.JSON(`{
			"email": "newuser@example.com",
			"login": "newuser"
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": 99,
			"user": "newuser",
			"email": "newuser@example.com",
			"password": "generatedPassword123"
		}`).Once())

	out, err := cmd.Execute("cloud", "container-registry", "users", "create", "550e8400-e29b-41d4-a716-446655440000", "--email", "newuser@example.com", "--login", "newuser", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Container registry user 'newuser' created successfully with password 'generatedPassword123'`)
}

func (ms *MockSuite) TestCloudContainerRegistryUsersCreateCmdWithoutOptionalParameters(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/users",
		// Expect an empty JSON body when optional fields are omitted
		tdhttpmock.JSONBody(td.JSON(`{}`)),
		httpmock.NewStringResponder(200, `{
			"id": 100,
			"user": "auto-generated-user",
			"email": "auto@example.com",
			"password": "autoPass123"
		}`).Once())

	out, err := cmd.Execute("cloud", "container-registry", "users", "create", "550e8400-e29b-41d4-a716-446655440000", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Container registry user 'auto-generated-user' created successfully with password 'autoPass123'`)
}

func (ms *MockSuite) TestCloudContainerRegistryUsersSetAsAdminCmd(assert, require *td.T) {
	httpmock.RegisterResponder("PUT", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/users/42/setAsAdmin",
		httpmock.NewStringResponder(200, ``).Once())

	out, err := cmd.Execute("cloud", "container-registry", "users", "set-as-admin", "550e8400-e29b-41d4-a716-446655440000", "42", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Container registry user successfully set as admin`)
}

func (ms *MockSuite) TestCloudContainerRegistryUsersDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/users/42",
		httpmock.NewStringResponder(204, ``).Once())

	out, err := cmd.Execute("cloud", "container-registry", "users", "delete", "550e8400-e29b-41d4-a716-446655440000", "42", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Container registry user deleted successfully`)
}
