// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestCloudContainerRegistryListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry",
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

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region",
		httpmock.NewStringResponder(200, `["GRA", "EU-WEST-PAR"]`).Once())

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA",
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

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/EU-WEST-PAR",
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

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/0b1b2dc2-952b-11f0-afd9-0050568ce122/plan",
		httpmock.NewStringResponder(200, `{
			"code": "registry.s-plan-equivalent.hour.consumption",
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
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000",
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

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/plan",
		httpmock.NewStringResponder(200, `{
			"code": "registry.m-plan-equivalent.hour.consumption",
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
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000",
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

func (ms *MockSuite) TestCloudContainerRegistryIAMEnableCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/iam",
		tdhttpmock.JSONBody(td.JSON(`{
			"deleteUsers": false
		}`)),
		httpmock.NewStringResponder(200, ``).Once())

	out, err := cmd.Execute("cloud", "container-registry", "iam", "enable", "550e8400-e29b-41d4-a716-446655440000", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Container registry IAM enabled successfully`)
}

func (ms *MockSuite) TestCloudContainerRegistryIAMEnableCmdWithDeleteUsers(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/iam",
		tdhttpmock.JSONBody(td.JSON(`{
			"deleteUsers": true
		}`)),
		httpmock.NewStringResponder(200, ``).Once())

	out, err := cmd.Execute("cloud", "container-registry", "iam", "enable", "550e8400-e29b-41d4-a716-446655440000", "--delete-users", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Container registry IAM enabled successfully`)
}

func (ms *MockSuite) TestCloudContainerRegistryIAMDisableCmd(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/iam",
		httpmock.NewStringResponder(204, ``).Once())

	out, err := cmd.Execute("cloud", "container-registry", "iam", "disable", "550e8400-e29b-41d4-a716-446655440000", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Container registry IAM disabled successfully`)
}

func (ms *MockSuite) TestCloudContainerRegistryUsersListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/users",
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
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/users/42",
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

func (ms *MockSuite) TestCloudContainerRegistryOIDCGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/openIdConnect",
		httpmock.NewStringResponder(200, `{
			"adminGroup": "admins",
			"autoOnboard": true,
			"clientId": "client-id",
			"clientSecret": "client-secret",
			"createdAt": "2026-01-23T11:00:19.797Z",
			"endpoint": "https://oidc.example.com",
			"groupFilter": ".*",
			"groupsClaim": "groups",
			"id": "oidc-config-id",
			"name": "Example OIDC",
			"scope": "openid profile",
			"status": "READY",
			"updatedAt": "2026-01-23T11:00:19.797Z",
			"userClaim": "email",
			"verifyCert": true
		}`).Once())

	out, err := cmd.Execute("cloud", "container-registry", "oidc", "get", "550e8400-e29b-41d4-a716-446655440000", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	var configuration map[string]any
	require.CmpNoError(json.Unmarshal([]byte(out), &configuration))
	assert.Cmp(configuration, td.JSON(`{
		"adminGroup": "admins",
		"autoOnboard": true,
		"clientId": "client-id",
		"clientSecret": "client-secret",
		"createdAt": "2026-01-23T11:00:19.797Z",
		"endpoint": "https://oidc.example.com",
		"groupFilter": ".*",
		"groupsClaim": "groups",
		"id": "oidc-config-id",
		"name": "Example OIDC",
		"scope": "openid profile",
		"status": "READY",
		"updatedAt": "2026-01-23T11:00:19.797Z",
		"userClaim": "email",
		"verifyCert": true
	}`))
}

func (ms *MockSuite) TestCloudContainerRegistryOIDCCreateCmdWithAllOptionalParameters(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/openIdConnect",
		tdhttpmock.JSONBody(td.JSON(`{
			"deleteUsers": true,
			"provider": {
				"adminGroup": "admins",
				"autoOnboard": true,
				"clientId": "client-id",
				"clientSecret": "client-secret",
				"endpoint": "https://oidc.example.com",
				"groupFilter": ".*",
				"groupsClaim": "groups",
				"name": "Example OIDC",
				"scope": "openid,profile",
				"userClaim": "name",
				"verifyCert": true
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "oidc-config-id"}`).Once())

	out, err := cmd.Execute(
		"cloud", "container-registry", "oidc", "create", "550e8400-e29b-41d4-a716-446655440000",
		"--name", "Example OIDC",
		"--endpoint", "https://oidc.example.com",
		"--client-id", "client-id",
		"--client-secret", "client-secret",
		"--scope", "openid,profile",
		"--delete-users",
		"--auto-onboard",
		"--verify-cert",
		"--admin-group", "admins",
		"--group-filter", ".*",
		"--groups-claim", "groups",
		"--user-claim", "name",
		"--cloud-project", "fakeProjectID",
	)

	require.CmpNoError(err)
	assert.String(out, "✅ Container registry OIDC configuration created successfully")
}

func (ms *MockSuite) TestCloudContainerRegistryOIDCCreateCmdWithoutOptionalParameter(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/openIdConnect",
		tdhttpmock.JSONBody(td.JSON(`{
			"provider": {
				"clientId": "client-id",
				"clientSecret": "client-secret",
				"endpoint": "https://oidc.example.com",
				"name": "Example OIDC",
				"scope": "openid,profile"
			}
		}`)),
		httpmock.NewStringResponder(201, "").Once())

	out, err := cmd.Execute(
		"cloud", "container-registry", "oidc", "create", "550e8400-e29b-41d4-a716-446655440000",
		"--name", "Example OIDC",
		"--endpoint", "https://oidc.example.com",
		"--client-id", "client-id",
		"--client-secret", "client-secret",
		"--scope", "openid,profile",
		"--cloud-project", "fakeProjectID",
	)

	require.CmpNoError(err)
	assert.String(out, "✅ Container registry OIDC configuration created successfully")
}

func (ms *MockSuite) TestCloudContainerRegistryOIDCDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/openIdConnect",
		httpmock.NewStringResponder(204, ``).Once())

	out, err := cmd.Execute("cloud", "container-registry", "oidc", "delete", "550e8400-e29b-41d4-a716-446655440000", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, "✅ Container registry OIDC configuration deleted successfully")
}

func (ms *MockSuite) TestCloudContainerRegistryOIDCEditCmdWithAllOptionalParameters(assert, require *td.T) {
	httpmock.RegisterResponder(
		http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/openIdConnect",
		httpmock.NewStringResponder(200, `{
			"adminGroup": "admins",
			"autoOnboard": false,
			"clientId": "client-id",
			"clientSecret": "client-secret",
			"endpoint": "https://oidc.example.com",
			"groupFilter": ".*",
			"groupsClaim": "groups",
			"name": "Example OIDC",
			"scope": "openid,profile",
			"userClaim": "email",
			"verifyCert": false
		}`).Once())

	httpmock.RegisterMatcherResponder(
		http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/openIdConnect",
		tdhttpmock.JSONBody(td.JSON(`{
			"adminGroup":"new-admins",
			"autoOnboard":true,
			"clientId":"new-client-id",
			"clientSecret":"new-client-secret",
			"endpoint":"https://oidc-new.example.com",
			"groupsClaim":"new-groups",
			"name":"new Example OIDC",
			"scope":"openid,email",
			"userClaim":"name",
			"verifyCert":true
		}`)),
		httpmock.NewStringResponder(204, ``).Once(),
	)

	out, err := cmd.Execute(
		"cloud", "container-registry", "oidc", "edit", "550e8400-e29b-41d4-a716-446655440000",
		"--name", "new Example OIDC",
		"--endpoint", "https://oidc-new.example.com",
		"--client-id", "new-client-id",
		"--client-secret", "new-client-secret",
		"--scope", "openid,email",
		"--admin-group", "new-admins",
		"--group-filter", ".*",
		"--groups-claim", "new-groups",
		"--user-claim", "name",
		"--auto-onboard",
		"--verify-cert",
		"--cloud-project", "fakeProjectID",
	)

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), "✅ Resource updated successfully")
}

func (ms *MockSuite) TestCloudContainerRegistryOIDCEditCmdWithSingleParameter(assert, require *td.T) {
	httpmock.RegisterResponder(
		http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/openIdConnect",
		httpmock.NewStringResponder(200, `{
			"adminGroup": "admins",
			"autoOnboard": false,
			"clientId": "client-id",
			"clientSecret": "client-secret",
			"endpoint": "https://oidc.example.com",
			"groupFilter": ".*",
			"groupsClaim": "groups",
			"name": "Example OIDC",
			"scope": "openid,profile",
			"userClaim": "email",
			"verifyCert": false		
		}`).Once(),
	)

	httpmock.RegisterMatcherResponder(
		http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/openIdConnect",
		tdhttpmock.JSONBody(td.JSON(`{
			"adminGroup":"admins",
			"autoOnboard":false,
			"clientId":"client-id",
			"clientSecret":"client-secret",
			"endpoint":"https://oidc.example.com",
			"groupsClaim":"groups",
			"name":"Updated OIDC",
			"scope":"openid,profile",
			"userClaim":"email",
			"verifyCert":false
		}`)),
		httpmock.NewStringResponder(200, `✅ Resource updated successfully`).Once())

	out, err := cmd.Execute(
		"cloud", "container-registry", "oidc", "edit", "550e8400-e29b-41d4-a716-446655440000",
		"--name", "Updated OIDC",
		"--cloud-project", "fakeProjectID",
	)

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), "✅ Resource updated successfully")
}

func (ms *MockSuite) TestCloudContainerRegistryPlanListCapabilitiesCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/capabilities/plan",
		httpmock.NewStringResponder(200, `[
		  {
		    "code": "registry.m-plan-equivalent.hour.consumption",
		    "name": "MEDIUM",
		    "id": "c5ddc763-be75-48f7-b7ec-e923ca040bee",
		    "registryLimits": {
		      "imageStorage": 644245094400,
		      "parallelRequest": 45
		    },
		    "features": {
		      "vulnerability": true
		    }
		  },
		  {
		    "code": "registry.l-plan-equivalent.hour.consumption",
		    "name": "LARGE",
		    "id": "0dae73df-6c49-47bf-a9d5-6b866c74ac54",
		    "registryLimits": {
		      "imageStorage": 5497558138880,
		      "parallelRequest": 90
		    },
		    "features": {
		      "vulnerability": true
		    }
		  }
		]`).Once())

	out, err := cmd.Execute("cloud", "container-registry", "plan", "list-capabilities", "550e8400-e29b-41d4-a716-446655440000", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌──────────────────────────────────────┬────────┬───────────────┬──────────────┬─────────────────┐
│                  id                  │  name  │ vulnerability │ imageStorage │ parallelRequest │
├──────────────────────────────────────┼────────┼───────────────┼──────────────┼─────────────────┤
│ c5ddc763-be75-48f7-b7ec-e923ca040bee │ MEDIUM │ true          │ 600G         │ 45              │
│ 0dae73df-6c49-47bf-a9d5-6b866c74ac54 │ LARGE  │ true          │ 5T           │ 90              │
└──────────────────────────────────────┴────────┴───────────────┴──────────────┴─────────────────┘
💡 Use option --json or --yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestCloudContainerRegistryPlanUpgradeCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/plan",
		tdhttpmock.JSONBody(td.JSON(`{
		  "planID": "c5ddc763-be75-48f7-b7ec-e923ca040bee"
		}`)),
		httpmock.NewStringResponder(204, ""),
	)

	out, err := cmd.Execute("cloud", "container-registry", "plan", "upgrade", "550e8400-e29b-41d4-a716-446655440000", "--plan-id", "c5ddc763-be75-48f7-b7ec-e923ca040bee", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, "✅ Container registry 550e8400-e29b-41d4-a716-446655440000 plan upgraded to c5ddc763-be75-48f7-b7ec-e923ca040bee")
}

func (ms *MockSuite) TestCloudContainerRegistryIPRestrictionsManagementListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/ipRestrictions/management",
		httpmock.NewStringResponder(200, `[
			{
				"createdAt": "2026-01-23T10:00:00.000Z",
				"description": "Office network",
				"ipBlock": "192.0.2.0/24",
				"updatedAt": "2026-01-23T10:00:00.000Z"
			},
			{
				"createdAt": "2026-01-24T10:00:00.000Z",
				"description": "VPN network",
				"ipBlock": "10.0.0.0/8",
				"updatedAt": "2026-01-24T10:00:00.000Z"
			}
		]`).Once())

	out, err := cmd.Execute("cloud", "container-registry", "ip-restrictions", "management", "list", "550e8400-e29b-41d4-a716-446655440000", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌──────────────┬────────────────┬──────────────────────────┬──────────────────────────┐
│   ipBlock    │  description   │        createdAt         │        updatedAt         │
├──────────────┼────────────────┼──────────────────────────┼──────────────────────────┤
│ 192.0.2.0/24 │ Office network │ 2026-01-23T10:00:00.000Z │ 2026-01-23T10:00:00.000Z │
│ 10.0.0.0/8   │ VPN network    │ 2026-01-24T10:00:00.000Z │ 2026-01-24T10:00:00.000Z │
└──────────────┴────────────────┴──────────────────────────┴──────────────────────────┘
💡 Use option --json or --yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestCloudContainerRegistryIPRestrictionsRegistryListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/ipRestrictions/registry",
		httpmock.NewStringResponder(200, `[
			{
				"createdAt": "2026-01-23T11:00:00.000Z",
				"description": "Docker push",
				"ipBlock": "203.0.113.0/24",
				"updatedAt": "2026-01-23T11:00:00.000Z"
			}
		]`).Once())

	out, err := cmd.Execute("cloud", "container-registry", "ip-restrictions", "registry", "list", "550e8400-e29b-41d4-a716-446655440000", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌────────────────┬─────────────┬──────────────────────────┬──────────────────────────┐
│    ipBlock     │ description │        createdAt         │        updatedAt         │
├────────────────┼─────────────┼──────────────────────────┼──────────────────────────┤
│ 203.0.113.0/24 │ Docker push │ 2026-01-23T11:00:00.000Z │ 2026-01-23T11:00:00.000Z │
└────────────────┴─────────────┴──────────────────────────┴──────────────────────────┘
💡 Use option --json or --yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestCloudContainerRegistryIPRestrictionsManagementAddCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/ipRestrictions/management",
		httpmock.NewStringResponder(200, `[
			{
				"createdAt": "2026-01-23T10:00:00.000Z",
				"description": "Office network",
				"ipBlock": "192.0.2.0/24",
				"updatedAt": "2026-01-23T10:00:00.000Z"
			}
		]`).Once())

	httpmock.RegisterMatcherResponder(
		http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/ipRestrictions/management",
		tdhttpmock.JSONBody(td.JSON(`[
			{
				"description": "Office network",
				"ipBlock": "192.0.2.0/24"
			},
			{
				"description": "VPN network",
				"ipBlock": "10.0.0.0/8"
			}
		]`)),
		httpmock.NewStringResponder(204, "").Once())

	out, err := cmd.Execute("cloud", "container-registry", "ip-restrictions", "management", "add", "550e8400-e29b-41d4-a716-446655440000", "--ip-block", "10.0.0.0/8", "--description", "VPN network", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, "✅ IP restriction 10.0.0.0/8 added to management")
}

func (ms *MockSuite) TestCloudContainerRegistryIPRestrictionsManagementAddCmdWithoutDescription(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/ipRestrictions/management",
		httpmock.NewStringResponder(200, `[]`).Once())

	httpmock.RegisterMatcherResponder(
		http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/ipRestrictions/management",
		tdhttpmock.JSONBody(td.JSON(`[
			{
				"ipBlock": "192.0.2.0/24"
			}
		]`)),
		httpmock.NewStringResponder(204, "").Once())

	out, err := cmd.Execute("cloud", "container-registry", "ip-restrictions", "management", "add", "550e8400-e29b-41d4-a716-446655440000", "--ip-block", "192.0.2.0/24", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, "✅ IP restriction 192.0.2.0/24 added to management")
}

func (ms *MockSuite) TestCloudContainerRegistryIPRestrictionsRegistryAddCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/ipRestrictions/registry",
		httpmock.NewStringResponder(200, `[]`).Once())

	httpmock.RegisterMatcherResponder(
		http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/ipRestrictions/registry",
		tdhttpmock.JSONBody(td.JSON(`[
			{
				"description": "Docker pull",
				"ipBlock": "203.0.113.0/24"
			}
		]`)),
		httpmock.NewStringResponder(204, "").Once())

	out, err := cmd.Execute("cloud", "container-registry", "ip-restrictions", "registry", "add", "550e8400-e29b-41d4-a716-446655440000", "--ip-block", "203.0.113.0/24", "--description", "Docker pull", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, "✅ IP restriction 203.0.113.0/24 added to registry")
}

func (ms *MockSuite) TestCloudContainerRegistryIPRestrictionsManagementDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/ipRestrictions/management",
		httpmock.NewStringResponder(200, `[
			{
				"createdAt": "2026-01-23T10:00:00.000Z",
				"description": "Office network",
				"ipBlock": "192.0.2.0/24",
				"updatedAt": "2026-01-23T10:00:00.000Z"
			},
			{
				"createdAt": "2026-01-24T10:00:00.000Z",
				"description": "VPN network",
				"ipBlock": "10.0.0.0/8",
				"updatedAt": "2026-01-24T10:00:00.000Z"
			}
		]`).Once())

	httpmock.RegisterMatcherResponder(
		http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/ipRestrictions/management",
		tdhttpmock.JSONBody(td.JSON(`[
			{
				"description": "Office network",
				"ipBlock": "192.0.2.0/24"
			}
		]`)),
		httpmock.NewStringResponder(204, "").Once())

	out, err := cmd.Execute("cloud", "container-registry", "ip-restrictions", "management", "delete", "550e8400-e29b-41d4-a716-446655440000", "--ip-block", "10.0.0.0/8", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, "✅ IP restriction 10.0.0.0/8 deleted from management")
}

func (ms *MockSuite) TestCloudContainerRegistryIPRestrictionsRegistryDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/ipRestrictions/registry",
		httpmock.NewStringResponder(200, `[
			{
				"createdAt": "2026-01-23T11:00:00.000Z",
				"description": "Docker push",
				"ipBlock": "203.0.113.0/24",
				"updatedAt": "2026-01-23T11:00:00.000Z"
			}
		]`).Once())

	httpmock.RegisterMatcherResponder(
		http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/550e8400-e29b-41d4-a716-446655440000/ipRestrictions/registry",
		tdhttpmock.JSONBody(td.JSON(`[]`)),
		httpmock.NewStringResponder(204, "").Once())

	out, err := cmd.Execute("cloud", "container-registry", "ip-restrictions", "registry", "delete", "550e8400-e29b-41d4-a716-446655440000", "--ip-block", "203.0.113.0/24", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, "✅ IP restriction 203.0.113.0/24 deleted from registry")
}
