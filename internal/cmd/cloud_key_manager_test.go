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

//
// Secret tests
//

func (ms *MockSuite) TestCloudKeyManagerSecretListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/secret",
		httpmock.NewStringResponder(200, `[
			{
				"id": "secret-1",
				"resourceStatus": "READY",
				"currentState": {
					"name": "my-secret",
					"secretType": "OPAQUE",
					"location": {"region": "GRA"}
				}
			}
		]`))

	out, err := cmd.Execute("cloud", "key-manager", "secret", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("secret-1"))
	assert.Cmp(out, td.Contains("my-secret"))
}

func (ms *MockSuite) TestCloudKeyManagerSecretGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/secret/secret-1",
		httpmock.NewStringResponder(200, `{
			"id": "secret-1",
			"resourceStatus": "READY",
			"currentState": {
				"name": "my-secret",
				"secretType": "OPAQUE",
				"location": {"region": "GRA"}
			}
		}`))

	out, err := cmd.Execute("cloud", "key-manager", "secret", "get", "secret-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("secret-1"))
	assert.Cmp(out, td.Contains("my-secret"))
}

func (ms *MockSuite) TestCloudKeyManagerSecretCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/secret",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "my-secret",
					"secretType": "OPAQUE",
					"location": {"region": "GRA"}
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "secret-123"}`),
	)

	out, err := cmd.Execute("cloud", "key-manager", "secret", "create", "--cloud-project", "fakeProjectID",
		"--name", "my-secret", "--secret-type", "OPAQUE", "--region", "GRA")

	require.CmpNoError(err)
	assert.String(out, `✅ Key Manager secret created successfully (id: secret-123)`)
}

func (ms *MockSuite) TestCloudKeyManagerSecretCreateErrorCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/secret",
		httpmock.NewStringResponder(400, `{"message": "invalid secret type"}`),
	)

	out, err := cmd.Execute("cloud", "key-manager", "secret", "create", "--cloud-project", "fakeProjectID",
		"--name", "my-secret", "--secret-type", "OPAQUE", "--region", "GRA")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("failed to create Key Manager secret"))
}

func (ms *MockSuite) TestCloudKeyManagerSecretDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/secret/secret-1",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "key-manager", "secret", "delete", "secret-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("deleted successfully"))
}

func (ms *MockSuite) TestCloudKeyManagerSecretPayloadCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/secret/secret-1/payload",
		httpmock.NewStringResponder(200, `{"payload": "super-secret-value"}`))

	out, err := cmd.Execute("cloud", "key-manager", "secret", "payload", "secret-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("super-secret-value"))
}

func (ms *MockSuite) TestCloudKeyManagerSecretConsumerRegisterCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/secret/secret-1/consumer",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"resourceId": "11111111-1111-1111-1111-111111111111",
				"resourceType": "INSTANCE",
				"service": "COMPUTE"
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "consumer-1"}`),
	)

	out, err := cmd.Execute("cloud", "key-manager", "secret", "consumer", "register", "secret-1", "--cloud-project", "fakeProjectID",
		"--resource-id", "11111111-1111-1111-1111-111111111111", "--resource-type", "INSTANCE", "--service", "COMPUTE")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Consumer registered successfully for secret secret-1 (id: consumer-1)"))
}

func (ms *MockSuite) TestCloudKeyManagerSecretConsumerListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/secret/secret-1/consumer",
		httpmock.NewStringResponder(200, `[
			{
				"id": "consumer-1",
				"resourceId": "11111111-1111-1111-1111-111111111111",
				"resourceType": "INSTANCE",
				"service": "COMPUTE"
			}
		]`))

	out, err := cmd.Execute("cloud", "key-manager", "secret", "consumer", "list", "secret-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("consumer-1"))
}

func (ms *MockSuite) TestCloudKeyManagerSecretConsumerDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/secret/secret-1/consumer/consumer-1",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "key-manager", "secret", "consumer", "delete", "secret-1", "consumer-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("deleted successfully"))
}

func (ms *MockSuite) TestCloudKeyManagerSecretEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/secret/secret-1",
		httpmock.NewStringResponder(200, `{
			"id": "secret-1",
			"checksum": "abc123",
			"resourceStatus": "READY",
			"targetSpec": {
				"name": "my-secret",
				"secretType": "OPAQUE",
				"location": {"region": "GRA"},
				"metadata": {"old": "value"}
			}
		}`))

	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/secret/secret-1",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"checksum": "abc123",
				"targetSpec": {
					"metadata": {"old": "value", "env": "prod"}
				}
			}`),
		),
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "key-manager", "secret", "edit", "secret-1", "--cloud-project", "fakeProjectID", "--metadata", "env=prod")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("updated successfully"))
}

func (ms *MockSuite) TestCloudKeyManagerContainerEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/container/container-1",
		httpmock.NewStringResponder(200, `{
			"id": "container-1",
			"checksum": "def456",
			"resourceStatus": "READY",
			"targetSpec": {
				"name": "my-container",
				"type": "GENERIC",
				"location": {"region": "GRA"},
				"secretRefs": []
			}
		}`))

	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/container/container-1",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"checksum": "def456",
				"targetSpec": {
					"secretRefs": [
						{"name": "private_key", "secret": {"id": "secret-9"}}
					]
				}
			}`),
		),
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "key-manager", "container", "edit", "container-1", "--cloud-project", "fakeProjectID", "--secret-ref", "private_key=secret-9")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("updated successfully"))
}

//
// Container tests
//

func (ms *MockSuite) TestCloudKeyManagerContainerListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/container",
		httpmock.NewStringResponder(200, `[
			{
				"id": "container-1",
				"resourceStatus": "READY",
				"currentState": {
					"name": "my-container",
					"type": "GENERIC",
					"location": {"region": "GRA"}
				}
			}
		]`))

	out, err := cmd.Execute("cloud", "key-manager", "container", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("container-1"))
	assert.Cmp(out, td.Contains("my-container"))
}

func (ms *MockSuite) TestCloudKeyManagerContainerGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/container/container-1",
		httpmock.NewStringResponder(200, `{
			"id": "container-1",
			"resourceStatus": "READY",
			"currentState": {
				"name": "my-container",
				"type": "GENERIC",
				"location": {"region": "GRA"}
			}
		}`))

	out, err := cmd.Execute("cloud", "key-manager", "container", "get", "container-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("container-1"))
	assert.Cmp(out, td.Contains("my-container"))
}

func (ms *MockSuite) TestCloudKeyManagerContainerCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/container",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "my-container",
					"type": "GENERIC",
					"location": {"region": "GRA"},
					"secretRefs": [
						{
							"name": "private_key",
							"secret": {"id": "secret-1"}
						}
					]
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "container-123"}`),
	)

	out, err := cmd.Execute("cloud", "key-manager", "container", "create", "--cloud-project", "fakeProjectID",
		"--name", "my-container", "--type", "GENERIC", "--region", "GRA", "--secret-ref", "private_key=secret-1")

	require.CmpNoError(err)
	assert.String(out, `✅ Key Manager container created successfully (id: container-123)`)
}

func (ms *MockSuite) TestCloudKeyManagerContainerCreateErrorCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/container",
		httpmock.NewStringResponder(400, `{"message": "invalid container type"}`),
	)

	out, err := cmd.Execute("cloud", "key-manager", "container", "create", "--cloud-project", "fakeProjectID",
		"--name", "my-container", "--type", "GENERIC", "--region", "GRA")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("failed to create Key Manager container"))
}

func (ms *MockSuite) TestCloudKeyManagerContainerDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/container/container-1",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "key-manager", "container", "delete", "container-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("deleted successfully"))
}

func (ms *MockSuite) TestCloudKeyManagerContainerConsumerRegisterCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/container/container-1/consumer",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"resourceId": "22222222-2222-2222-2222-222222222222",
				"resourceType": "LOADBALANCER",
				"service": "LOADBALANCER"
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "consumer-2"}`),
	)

	out, err := cmd.Execute("cloud", "key-manager", "container", "consumer", "register", "container-1", "--cloud-project", "fakeProjectID",
		"--resource-id", "22222222-2222-2222-2222-222222222222", "--resource-type", "LOADBALANCER", "--service", "LOADBALANCER")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Consumer registered successfully for container container-1 (id: consumer-2)"))
}

func (ms *MockSuite) TestCloudKeyManagerContainerConsumerDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/keyManager/container/container-1/consumer/consumer-2",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "key-manager", "container", "consumer", "delete", "container-1", "consumer-2", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("deleted successfully"))
}
