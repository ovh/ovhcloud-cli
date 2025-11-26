// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0
// Crafted with Codex

package cmd_test

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
	"github.com/ovh/ovhcloud-cli/internal/services/webhosting"
)

func (ms *MockSuite) TestWebhostingDomainAddCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/attachedDomain",
		tdhttpmock.JSONBody(td.JSON(`{
			"domain": "example.com",
			"path": "/www",
			"runtimeId": 42,
			"ssl": true,
			"bypassDNSConfiguration": true
		}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "domain", "add", "myservice", "--domain", "example.com", "--path", "/www", "--runtime-id", "42", "--enable-ssl", "--bypass-dns")

	require.CmpNoError(err)
	assert.String(out, "✅ Domain attached")
}

func (ms *MockSuite) TestWebhostingDomainDigStatusCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/attachedDomain/example.com/digStatus",
		httpmock.NewStringResponder(200, `{
			"domain": "example.com",
			"recommendedIps": {
				"recommendedIpV4": ["192.0.2.10"],
				"recommendedIpV6": ["2001:db8::1"]
			},
			"records": {
				"www": {
					"dnsConfigured": true,
					"isOvhIp": false,
					"type": "A"
				}
			}
		}`),
	)

	out, err := cmd.Execute("webhosting", "domain", "dig-status", "myservice", "example.com", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"domain": "example.com",
		"recommendedIps": {
			"recommendedIpV4": ["192.0.2.10"],
			"recommendedIpV6": ["2001:db8::1"]
		},
		"records": {
			"www": {
				"dnsConfigured": true,
				"isOvhIp": false,
				"type": "A"
			}
		}
	}`))
}

func (ms *MockSuite) TestWebhostingDomainFindCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/attachedDomain?domain=example.com",
		httpmock.NewStringResponder(200, `["hosting1"]`),
	)

	out, err := cmd.Execute("webhosting", "domain", "find", "example.com", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"domain":"example.com","serviceName":"hosting1"}]`))
}

func (ms *MockSuite) TestWebhostingDomainAvailableOfferCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/availableOffer?domain=example.com",
		httpmock.NewStringResponder(200, `["start","pro"]`),
	)

	out, err := cmd.Execute("webhosting", "domain", "available-offer", "example.com", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"domain":"example.com","offer":"start"},{"domain":"example.com","offer":"pro"}]`))
}

func (ms *MockSuite) TestWebhostingIncidentCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/incident",
		httpmock.NewStringResponder(200, `["incident1","incident2"]`),
	)

	out, err := cmd.Execute("webhosting", "incident", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"incident":"incident1"},{"incident":"incident2"}]`))
}

func (ms *MockSuite) TestWebhostingCronCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cron",
		tdhttpmock.JSONBody(td.JSON(`{
			"command": "echo ok",
			"frequency": "0 0 * * *",
			"language": "node10",
			"email": "admin@example.com",
			"description": "nightly"
		}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "cron", "create", "myservice",
		"--command", "echo ok",
		"--frequency", "0 0 * * *",
		"--language", "node10",
		"--email", "admin@example.com",
		"--description", "nightly",
	)

	require.CmpNoError(err)
	assert.String(out, "✅ Cron created")
}

func (ms *MockSuite) TestWebhostingCronAvailableLanguages(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cronAvailableLanguage",
		httpmock.NewStringResponder(200, `["php8.1","python3"]`),
	)

	out, err := cmd.Execute("webhosting", "cron", "available-languages", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"language":"php8.1"},{"language":"python3"}]`))
}

func (ms *MockSuite) TestWebhostingDatabaseDumpCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/database/db1/dump",
		tdhttpmock.JSONBody(td.JSON(`{
			"date": "yesterday",
			"sendEmail": true
		}`)),
		httpmock.NewStringResponder(200, `{"id":123}`),
	)

	out, err := cmd.Execute("webhosting", "db", "dump", "create", "myservice", "db1", "--date", "yesterday", "--send-email")

	require.CmpNoError(err)
	assert.String(out, "⚡️ Dump requested")
}

func (ms *MockSuite) TestWebhostingDatabaseDumpGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/database/db1/dump/123",
		httpmock.NewStringResponder(200, `{
			"id": 123,
			"status": "created",
			"type": "now",
			"creationDate": "2025-01-01T00:00:00Z",
			"deletionDate": "2025-02-01T00:00:00Z",
			"url": "https://example.com/dump",
			"taskId": 42
		}`),
	)

	out, err := cmd.Execute("webhosting", "db", "dump", "get", "myservice", "db1", "123", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"id":123,
		"status":"created",
		"type":"now",
		"creationDate":"2025-01-01T00:00:00Z",
		"deletionDate":"2025-02-01T00:00:00Z",
		"url":"https://example.com/dump",
		"taskId":42,
		"urlShort":"https://example.com/dump"
	}`))
}

func (ms *MockSuite) TestWebhostingDatabaseStatsCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/database/db1/statistics?period=daily&type=statement",
		httpmock.NewStringResponder(200, `[{"timestamp":1700000000,"value":12}]`),
	)

	out, err := cmd.Execute("webhosting", "db", "stats", "myservice", "db1", "--period", "daily", "--type", "statement", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"timestamp":"2023-11-14T22:13:20Z","value":12}]`))
}

func (ms *MockSuite) TestWebhostingAPICallCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/token",
		httpmock.NewStringResponder(200, `{"token":"abc123"}`),
	)

	out, err := cmd.Execute("webhosting", "api", "call", "GET", "/hosting/web/myservice/token", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"details":{"token":"abc123"}}`))
}

func (ms *MockSuite) TestWebhostingWebsiteCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/website",
		tdhttpmock.JSONBody(td.JSON(`{
			"path": "/www",
			"vcsBranch": "main",
			"vcsUrl": "https://example.com/repo.git"
		}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "website", "create", "myservice",
		"--path", "/www",
		"--vcs-url", "https://example.com/repo.git",
		"--branch", "main",
	)

	require.CmpNoError(err)
	assert.String(out, "✅ Website created")
}

func (ms *MockSuite) TestWebhostingUserCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/user",
		tdhttpmock.JSONBody(td.JSON(`{
			"home": "/home/site",
			"login": "user1",
			"password": "Secret123!",
			"sshState": "active"
		}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "user", "create", "myservice",
		"--home", "/home/site",
		"--login", "user1",
		"--password", "Secret123!",
		"--ssh-state", "active",
	)

	require.CmpNoError(err)
	assert.String(out, "✅ User created")
}

func (ms *MockSuite) TestWebhostingSSLCreationCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ssl",
		tdhttpmock.JSONBody(td.JSON(`{
			"certificate": "CERT",
			"chain": "CHAIN",
			"key": "KEY"
		}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "ssl", "create", "myservice",
		"--certificate", "CERT",
		"--chain", "CHAIN",
		"--key", "KEY",
	)

	require.CmpNoError(err)
	assert.String(out, "⚡️ SSL creation requested")
}

func (ms *MockSuite) TestWebhostingRuntimeCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/runtime",
		tdhttpmock.JSONBody(td.JSON(`{
			"name": "app",
			"type": "nodejs",
			"publicDir": "public",
			"appEnv": "production",
			"appBootstrap": "npm start",
			"isDefault": true,
			"attachedDomains": ["example.com"]
		}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "runtime", "create", "myservice",
		"--name", "app",
		"--type", "nodejs",
		"--public-dir", "public",
		"--app-env", "production",
		"--app-bootstrap", "npm start",
		"--runtime-default",
		"--domain", "example.com",
	)

	require.CmpNoError(err)
	assert.String(out, "✅ Runtime created")
}

func (ms *MockSuite) TestWebhostingDomainPurgeAndDelete(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/attachedDomain/example.com/purgeCache",
		httpmock.NewStringResponder(200, `{}`),
	)
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/attachedDomain/example.com",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "domain", "purge-cache", "myservice", "example.com")
	require.CmpNoError(err)
	assert.String(out, "⚡️ Purge triggered")

	out, err = cmd.Execute("webhosting", "domain", "delete", "myservice", "example.com")
	require.CmpNoError(err)
	assert.String(out, "✅ Domain deleted")
}

func (ms *MockSuite) TestWebhostingDomainUpdate(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/attachedDomain/example.com",
		httpmock.NewStringResponder(200, `{"domain":"example.com","path":"/old"}`),
	)
	httpmock.RegisterResponder(http.MethodPut,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/attachedDomain/example.com",
		func(req *http.Request) (*http.Response, error) {
			var body map[string]any
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				require.CmpNoError(err)
			}

			assert.Cmp(body["path"], "/new")
			_, hasBypass := body["bypassDNSConfiguration"]
			assert.False(hasBypass, "bypassDNSConfiguration should be omitted when not provided")
			_, hasOwnLog := body["ownLog"]
			assert.False(hasOwnLog, "ownLog should be omitted when not provided")

			return httpmock.NewStringResponse(200, `{}`), nil
		},
	)

	out, err := cmd.Execute("webhosting", "domain", "update", "myservice", "example.com", "--path", "/new")
	require.CmpNoError(err)
	assert.String(out, "✅ Resource updated successfully")
}

func (ms *MockSuite) TestWebhostingCronDelete(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cron/12",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "cron", "delete", "myservice", "12")
	require.CmpNoError(err)
	assert.String(out, "✅ Cron deleted")
}

func (ms *MockSuite) TestWebhostingModuleList(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/module",
		httpmock.NewStringResponder(200, `[123]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/module/123",
		httpmock.NewStringResponder(200, `{
			"id": 123,
			"moduleId": 2277,
			"status": "created",
			"targetUrl": "example.com",
			"path": "www",
			"language": "fr",
			"adminName": "admin"
		}`),
	)

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/moduleList",
		httpmock.NewStringResponder(200, `[2277]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/moduleList/2277",
		httpmock.NewStringResponder(200, `{"id":2277,"name":"WordPress","latest":true}`),
	)

	out, err := cmd.Execute("webhosting", "module", "list", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{
		"id": 123,
		"moduleId": 2277,
		"moduleName": "WordPress",
		"status": "created",
		"targetUrl": "example.com",
		"path": "www",
		"language": "fr",
		"adminName": "admin"
	}]`))
}

func (ms *MockSuite) TestWebhostingModuleInstallByName(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/moduleList",
		httpmock.NewStringResponder(200, `[10]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/moduleList/10",
		httpmock.NewStringResponder(200, `{"id":10,"name":"WordPress","latest":true}`),
	)

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/module",
		tdhttpmock.JSONBody(td.JSON(`{
			"moduleId": 10,
			"domain": "example.com",
			"path": "/www",
			"language": "en",
			"adminName": "admin",
			"adminPassword": "pwd"
		}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "module", "install", "myservice",
		"--module-name", "WordPress",
		"--domain", "example.com",
		"--path", "/www",
		"--language", "en",
		"--admin", "admin",
		"--admin-password", "pwd",
	)

	require.CmpNoError(err)
	assert.String(out, "⚡️ Module installation requested")
}

func (ms *MockSuite) TestWebhostingOvhConfigList(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ovhConfig",
		httpmock.NewStringResponder(200, `[1]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ovhConfig/1",
		httpmock.NewStringResponder(200, `{"id":1,"path":"/","engineName":"php","engineVersion":"8.1","environment":"production","status":"active"}`),
	)

	out, err := cmd.Execute("webhosting", "ovh-config", "list", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"engineName":"php","engineVersion":"8.1","environment":"production","id":1,"path":"/","status":"active"}]`))
}

func (ms *MockSuite) TestWebhostingOvhConfigGet(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ovhConfig/1",
		httpmock.NewStringResponder(200, `{"id":1,"path":"/","engineName":"php"}`),
	)

	out, err := cmd.Execute("webhosting", "ovh-config", "get", "myservice", "1", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"engineName":"php","id":1,"path":"/"}`))
}

func (ms *MockSuite) TestWebhostingOvhConfigChange(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ovhConfig/1/changeConfiguration",
		tdhttpmock.JSONBody(td.JSON(`{"engineName":"php","engineVersion":"8.1","environment":"production","httpFirewall":"security","container":"stable"}`)),
		httpmock.NewStringResponder(200, `{"id":123}`),
	)

	out, err := cmd.Execute("webhosting", "ovh-config", "change", "myservice", "1",
		"--engine-name", "php",
		"--engine-version", "8.1",
		"--environment", "production",
		"--http-firewall", "security",
		"--container", "stable",
		"--json",
	)

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"id":123}`))
}

func (ms *MockSuite) TestWebhostingOvhConfigChangeFromFile(assert, require *td.T) {
	tmp, err := os.CreateTemp("", "ovhconfig-change-*.json")
	require.CmpNoError(err)
	defer os.Remove(tmp.Name())

	_, err = tmp.WriteString(`{"engineName":"php","engineVersion":"8.2"}`)
	require.CmpNoError(err)
	require.CmpNoError(tmp.Close())

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ovhConfig/1/changeConfiguration",
		tdhttpmock.JSONBody(td.JSON(`{"engineName":"php","engineVersion":"8.2"}`)),
		httpmock.NewStringResponder(200, `{"id":555}`),
	)

	out, err := cmd.Execute("webhosting", "ovh-config", "change", "myservice", "1",
		"--from-file", tmp.Name(),
		"--json",
	)

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"id":555}`))
}

func (ms *MockSuite) TestWebhostingOvhConfigRollback(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ovhConfig/1/rollback",
		tdhttpmock.JSONBody(td.JSON(`{"rollbackId":5}`)),
		httpmock.NewStringResponder(200, `{"id":200}`),
	)

	out, err := cmd.Execute("webhosting", "ovh-config", "rollback", "myservice", "1",
		"--rollback-id", "5",
		"--json",
	)

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"id":200}`))
}

func (ms *MockSuite) TestWebhostingOvhConfigCapabilities(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ovhConfigCapabilities",
		httpmock.NewStringResponder(200, `[{"version":"8.2","containerImage":["stable","legacy"]}]`),
	)

	out, err := cmd.Execute("webhosting", "ovh-config", "capabilities", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"entries":[{"containerImage":["stable","legacy"],"version":"8.2"}]}`))
}

func (ms *MockSuite) TestWebhostingOvhConfigRecommended(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ovhConfigRecommendedValues",
		httpmock.NewStringResponder(200, `{"engineName":"php","engineVersion":"8.1","environment":"production","httpFirewall":"security","container":"stable"}`),
	)

	out, err := cmd.Execute("webhosting", "ovh-config", "recommended", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"container":"stable","engineName":"php","engineVersion":"8.1","environment":"production","httpFirewall":"security"}`))
}

func (ms *MockSuite) TestWebhostingOvhConfigRefresh(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ovhConfigRefresh",
		httpmock.NewStringResponder(200, `{"id":777}`),
	)

	out, err := cmd.Execute("webhosting", "ovh-config", "refresh", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"id":777}`))
}

func (ms *MockSuite) TestWebhostingOwnLogList(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ownLogs",
		httpmock.NewStringResponder(200, `[1,2]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ownLogs/1",
		httpmock.NewStringResponder(200, `{"id":1,"fqdn":"logs1","status":"ok","logs":"https://logs1","stats":"https://stats1"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ownLogs/2",
		httpmock.NewStringResponder(200, `{"id":2,"fqdn":"logs2","status":"ok","logs":"https://logs2","stats":"https://stats2"}`),
	)

	out, err := cmd.Execute("webhosting", "own-log", "list", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"fqdn":"logs1","id":1,"logs":"https://logs1","stats":"https://stats1","status":"ok"},{"fqdn":"logs2","id":2,"logs":"https://logs2","stats":"https://stats2","status":"ok"}]`))
}

func (ms *MockSuite) TestWebhostingOwnLogGet(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ownLogs/1",
		httpmock.NewStringResponder(200, `{"id":1,"fqdn":"logs1"}`),
	)

	out, err := cmd.Execute("webhosting", "own-log", "get", "myservice", "1", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"fqdn":"logs1","id":1}`))
}

func (ms *MockSuite) TestWebhostingOwnLogUserList(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ownLogs/1/userLogs",
		httpmock.NewStringResponder(200, `["user1"]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ownLogs/1/userLogs/user1",
		httpmock.NewStringResponder(200, `{"login":"user1","description":"desc","status":"ok","creationDate":"2025-01-01T00:00:00Z"}`),
	)

	out, err := cmd.Execute("webhosting", "own-log", "user", "list", "myservice", "1", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"creationDate":"2025-01-01T00:00:00Z","description":"desc","login":"user1","status":"ok"}]`))
}

func (ms *MockSuite) TestWebhostingOwnLogUserGet(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ownLogs/1/userLogs/user1",
		httpmock.NewStringResponder(200, `{"login":"user1","description":"desc"}`),
	)

	out, err := cmd.Execute("webhosting", "own-log", "user", "get", "myservice", "1", "user1", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"description":"desc","login":"user1"}`))
}

func (ms *MockSuite) TestWebhostingOwnLogUserCreate(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ownLogs/1/userLogs",
		tdhttpmock.JSONBody(td.JSON(`{"login":"user1","password":"Passw0rd!","description":"desc"}`)),
		httpmock.NewStringResponder(200, `{"login":"user1"}`),
	)

	out, err := cmd.Execute("webhosting", "own-log", "user", "create", "myservice", "1",
		"--login", "user1",
		"--password", "Passw0rd!",
		"--description", "desc",
		"--json",
	)

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"login":"user1"}`))
}

func (ms *MockSuite) TestWebhostingOwnLogUserUpdate(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ownLogs/1/userLogs/user1",
		tdhttpmock.JSONBody(td.JSON(`{"description":"new"}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "own-log", "user", "update", "myservice", "1", "user1",
		"--description", "new",
	)

	require.CmpNoError(err)
	assert.String(out, "✅ Resource updated successfully")
}

func (ms *MockSuite) TestWebhostingOwnLogUserDelete(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ownLogs/1/userLogs/user1",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "own-log", "user", "delete", "myservice", "1", "user1")

	require.CmpNoError(err)
	assert.String(out, "✅ User log deleted")
}

func (ms *MockSuite) TestWebhostingOwnLogUserChangePassword(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ownLogs/1/userLogs/user1/changePassword",
		tdhttpmock.JSONBody(td.JSON(`{"password":"NewPass!"}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "own-log", "user", "change-password", "myservice", "1", "user1",
		"--password", "NewPass!",
	)

	require.CmpNoError(err)
	assert.String(out, "⚡️ Password change requested")
}

func (ms *MockSuite) TestWebhostingRuntimeAvailableTypes(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/runtimeAvailableTypes",
		httpmock.NewStringResponder(200, `["nodejs-18","php-8.2"]`),
	)

	out, err := cmd.Execute("webhosting", "runtime", "available-types", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"types":[{"type":"nodejs-18"},{"type":"php-8.2"}]}`))
}

func (ms *MockSuite) TestWebhostingRequestAction(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/request",
		tdhttpmock.JSONBody(td.JSON(`{"action":"FLUSH_CACHE"}`)),
		httpmock.NewStringResponder(200, `{"id":555}`),
	)

	out, err := cmd.Execute("webhosting", "request-action", "myservice", "--action", "FLUSH_CACHE", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"id":555}`))
}

func (ms *MockSuite) TestWebhostingModuleInstall(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/module",
		tdhttpmock.JSONBody(td.JSON(`{
			"moduleId": 10,
			"domain": "example.com",
			"path": "/www",
			"language": "en",
			"adminName": "admin",
			"adminPassword": "pwd"
		}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "module", "install", "myservice",
		"--module-id", "10",
		"--domain", "example.com",
		"--path", "/www",
		"--language", "en",
		"--admin", "admin",
		"--admin-password", "pwd",
	)

	require.CmpNoError(err)
	assert.String(out, "⚡️ Module installation requested")
}

func (ms *MockSuite) TestWebhostingModuleCatalogList(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/moduleList?active=true",
		httpmock.NewStringResponder(200, `[1]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/moduleList/1",
		httpmock.NewStringResponder(200, `{"id":1,"name":"WordPress","branch":"stable","version":"6.5","active":true,"latest":true}`),
	)

	out, err := cmd.Execute("webhosting", "module", "catalog", "list", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"active":true,"branch":"stable","id":1,"latest":true,"name":"WordPress","version":"6.5"}]`))
}

func (ms *MockSuite) TestWebhostingModuleCatalogGet(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/moduleList/1",
		httpmock.NewStringResponder(200, `{"id":1,"name":"WordPress"}`),
	)

	out, err := cmd.Execute("webhosting", "module", "catalog", "get", "1", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"id":1,"name":"WordPress"}`))
}

func (ms *MockSuite) TestWebhostingOfferCapabilities(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/offerCapabilities?offer=START",
		httpmock.NewStringResponder(200, `{"support":"email","datacenter":"ovh"}`),
	)

	out, err := cmd.Execute("webhosting", "offer", "capabilities", "START", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"datacenter":"ovh","support":"email"}`))
}

func (ms *MockSuite) TestWebhostingVcsSupported(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/vcs/supported",
		httpmock.NewStringResponder(200, `["github"]`),
	)

	out, err := cmd.Execute("webhosting", "offer", "vcs-supported", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"platform":"github"}]`))
}

func (ms *MockSuite) TestWebhostingAbuseState(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/abuseState",
		httpmock.NewStringResponder(200, `{"status":"clean"}`),
	)

	out, err := cmd.Execute("webhosting", "abuse-state", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"status":"clean"}`))
}

func (ms *MockSuite) TestWebhostingRuntimeDelete(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/runtime/77",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "runtime", "delete", "myservice", "77")
	require.CmpNoError(err)
	assert.String(out, "✅ Runtime deleted")
}

func (ms *MockSuite) TestWebhostingWebsiteDeploy(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/website/5/deploy",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "website", "deploy", "myservice", "5")
	require.CmpNoError(err)
	assert.String(out, "⚡️ Deployment triggered")
}

func (ms *MockSuite) TestWebhostingWebsiteCreationCapabilities(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/websiteCreationCapabilities",
		httpmock.NewStringResponder(200, `{"allowedWebsites":5,"existingWebsites":2}`),
	)

	out, err := cmd.Execute("webhosting", "website", "creation-capabilities", "myservice", "--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"allowedWebsites":5,"existingWebsites":2}`))
}

func (ms *MockSuite) TestWebhostingWebsiteDelete(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/website/5?deleteFiles=false",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "website", "delete", "myservice", "5")
	require.CmpNoError(err)
	assert.String(out, "✅ Website deleted")
}

func (ms *MockSuite) TestWebhostingWebsiteDeleteWithFiles(assert, require *td.T) {
	defer func() { webhosting.WebsiteDeleteFiles = false }()

	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/website/5?deleteFiles=true",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "website", "delete", "myservice", "5", "--delete-files")
	require.CmpNoError(err)
	assert.String(out, "✅ Website deleted")
}

func (ms *MockSuite) TestWebhostingVcsWebhooks(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/vcs/webhooks?path=%2Fwww&vcs=github",
		httpmock.NewStringResponder(200, `{"push":"https://hook"}`),
	)

	out, err := cmd.Execute("webhosting", "vcs", "webhooks", "myservice",
		"--path", "/www",
		"--vcs", "github",
		"--json",
	)

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"path":"/www","push":"https://hook","vcs":"github"}`))
}

func (ms *MockSuite) TestWebhostingSSLRegenerateAndDelete(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ssl/regenerate",
		httpmock.NewStringResponder(200, `{}`),
	)
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/ssl",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "ssl", "regenerate", "myservice")
	require.CmpNoError(err)
	assert.String(out, "⚡️ Regeneration requested")

	out, err = cmd.Execute("webhosting", "ssl", "delete", "myservice")
	require.CmpNoError(err)
	assert.String(out, "✅ SSL deleted")
}

func (ms *MockSuite) TestWebhostingUserChangePassword(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/user/user1/changePassword",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "user", "change-password", "myservice", "user1", "--password", "newpwd")
	require.CmpNoError(err)
	assert.String(out, "✅ Password updated")
}

func (ms *MockSuite) TestWebhostingEnvVarUpdate(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/envVar/A",
		httpmock.NewStringResponder(200, `{"key":"A","type":"plain","value":"1"}`),
	)
	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/envVar/A",
		tdhttpmock.JSONBody(td.JSON(`{"value":"2"}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "env", "update", "myservice", "A", "--value", "2")
	require.CmpNoError(err)
	assert.String(out, "✅ Resource updated successfully")
}

func (ms *MockSuite) TestWebhostingBoostAndSnapshot(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/requestBoost",
		func(req *http.Request) (*http.Response, error) {
			var body map[string]any
			require.CmpNoError(json.NewDecoder(req.Body).Decode(&body))
			assert.Cmp(body, td.JSON(`{"offer":"PERFORMANCE_1"}`))
			return httpmock.NewStringResponse(200, `{}`), nil
		},
	)
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/restoreSnapshot",
		tdhttpmock.JSONBody(td.JSON(`{"backup":"daily"}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "request-boost", "myservice", "--offer", "PERFORMANCE_1")
	require.CmpNoError(err)
	assert.String(out, "⚡️ Boost PERFORMANCE_1 requested")

	out, err = cmd.Execute("webhosting", "restore-snapshot", "myservice", "--backup", "daily")
	require.CmpNoError(err)
	assert.String(out, "⚡️ Snapshot restore requested")
}

func (ms *MockSuite) TestWebhostingUnblockTCPOut(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/unblockTCPOut",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "unblock-tcp-out", "myservice")
	require.CmpNoError(err)
	assert.String(out, "⚡️ Outgoing TCP unblocked")
}

func (ms *MockSuite) TestWebhostingBoostHistoryCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/boostHistory",
		httpmock.NewStringResponder(200, `["2025-01-01T00:00:00Z"]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/boostHistory/2025-01-01T00:00:00Z",
		httpmock.NewStringResponder(200, `{
			"date": "2025-01-01T00:00:00Z",
			"offer": "perf1",
			"boostOffer": "boost1",
			"accountId": "1234"
		}`),
	)

	out, err := cmd.Execute("webhosting", "boost-history", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"entries": [
			{
				"accountId": "1234",
				"boostOffer": "boost1",
				"date": "2025-01-01T00:00:00Z",
				"offer": "perf1"
			}
		]
	}`))
}

func (ms *MockSuite) TestWebhostingCdnAvailableOptions(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cdn/availableOptions",
		httpmock.NewStringResponder(200, `[{"type":"cache","category":"rule","maxItems":5,"config":{"ttl":{"message":"time to live"}}}]`),
	)

	out, err := cmd.Execute("webhosting", "cdn", "available-options", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"options":[{"type":"cache","category":"rule","maxItems":5,"config":{"ttl":{"message":"time to live"}}}]}`))
}

func (ms *MockSuite) TestWebhostingCdnDomainStatistics(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cdn/domain/example.com/statistics?period=week",
		httpmock.NewStringResponder(200, `[{"name":"requests","unit":"hits","points":[{"timestamp":"2024-01-01T00:00:00Z","value":10}]}]`),
	)

	out, err := cmd.Execute("webhosting", "cdn", "domain", "statistics", "myservice", "example.com", "--period", "week", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"series":[{"name":"requests","points":[{"timestamp":"2024-01-01T00:00:00Z","value":10}],"unit":"hits"}]}`))
}

func (ms *MockSuite) TestWebhostingCdnServiceInfoGet(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cdn/serviceInfos",
		httpmock.NewStringResponder(200, `{"serviceId":123}`),
	)

	out, err := cmd.Execute("webhosting", "cdn", "service-info", "get", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"serviceId":123}`))
}

func (ms *MockSuite) TestWebhostingCdnServiceInfoUpdate(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cdn/serviceInfosUpdate",
		func(req *http.Request) (*http.Response, error) {
			var body map[string]any
			require.CmpNoError(json.NewDecoder(req.Body).Decode(&body))
			assert.Cmp(body, td.JSON(`{"renew":{"automatic":true}}`))
			return httpmock.NewStringResponse(200, `{}`), nil
		},
	)

	out, err := cmd.Execute("webhosting", "cdn", "service-info", "update", "myservice", "--renew-automatic")

	require.CmpNoError(err)
	assert.String(out, "✅ CDN service info updated")
}

func (ms *MockSuite) TestWebhostingCdnDomainList(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cdn/domain",
		httpmock.NewStringResponder(200, `[{"name":"cdn.example.com","status":"ok"}]`),
	)

	out, err := cmd.Execute("webhosting", "cdn", "domain", "list", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"name":"cdn.example.com","status":"ok"}]`))
}

func (ms *MockSuite) TestWebhostingCdnDomainOptionList(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cdn/domain/example.com/option",
		httpmock.NewStringResponder(200, `[{"name":"cacheRule","type":"cache","enabled":true}]`),
	)

	out, err := cmd.Execute("webhosting", "cdn", "domain", "option", "list", "myservice", "example.com", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"enabled":true,"name":"cacheRule","type":"cache"}]`))
}

func (ms *MockSuite) TestWebhostingCdnDomainOptionAdd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cdn/domain/example.com/option",
		func(req *http.Request) (*http.Response, error) {
			var body map[string]any
			require.CmpNoError(json.NewDecoder(req.Body).Decode(&body))
			assert.Cmp(body, td.JSON(`{
				"name":"cacheRule",
				"type":"cache",
				"enabled":true,
				"config":{"ttl":120}
			}`))
			return httpmock.NewStringResponse(200, `{"name":"cacheRule","type":"cache","enabled":true}`), nil
		},
	)

	out, err := cmd.Execute("webhosting", "cdn", "domain", "option", "add", "myservice", "example.com",
		"--name", "cacheRule",
		"--type", "cache",
		"--enabled",
		"--ttl", "120",
		"--json",
	)

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"enabled":true,"name":"cacheRule","type":"cache"}`))
}

func (ms *MockSuite) TestWebhostingCdnDomainOptionGet(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cdn/domain/example.com/option/cacheRule",
		httpmock.NewStringResponder(200, `{"name":"cacheRule","type":"cache","enabled":true}`),
	)

	out, err := cmd.Execute("webhosting", "cdn", "domain", "option", "get", "myservice", "example.com", "cacheRule", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"enabled":true,"name":"cacheRule","type":"cache"}`))
}

func (ms *MockSuite) TestWebhostingCdnDomainOptionUpdate(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPut,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cdn/domain/example.com/option/cacheRule",
		func(req *http.Request) (*http.Response, error) {
			var body map[string]any
			require.CmpNoError(json.NewDecoder(req.Body).Decode(&body))
			assert.Cmp(body, td.JSON(`{"enabled":false,"config":{"ttl":300}}`))
			return httpmock.NewStringResponse(200, `{}`), nil
		},
	)

	out, err := cmd.Execute("webhosting", "cdn", "domain", "option", "update", "myservice", "example.com", "cacheRule",
		"--enabled=false",
		"--ttl", "300",
	)

	require.CmpNoError(err)
	assert.String(out, "✅ Resource updated successfully")
}

func (ms *MockSuite) TestWebhostingCdnDomainOptionDelete(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cdn/domain/example.com/option/cacheRule",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "cdn", "domain", "option", "delete", "myservice", "example.com", "cacheRule")

	require.CmpNoError(err)
	assert.String(out, "✅ CDN option deleted")
}

func (ms *MockSuite) TestWebhostingCdnOperationList(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cdn/operation",
		httpmock.NewStringResponder(200, `[{"id":1,"function":"deploy","status":"todo"}]`),
	)

	out, err := cmd.Execute("webhosting", "cdn", "operation", "list", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"function":"deploy","id":1,"status":"todo"}]`))
}

func (ms *MockSuite) TestWebhostingCdnOperationGet(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cdn/operation/42",
		httpmock.NewStringResponder(200, `{"id":42,"function":"deploy","status":"todo"}`),
	)

	out, err := cmd.Execute("webhosting", "cdn", "operation", "get", "myservice", "42", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"function":"deploy","id":42,"status":"todo"}`))
}

func (ms *MockSuite) TestWebhostingModuleDelete(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/module/10",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "module", "delete", "myservice", "10")
	require.CmpNoError(err)
	assert.String(out, "✅ Module deletion requested")
}

func (ms *MockSuite) TestWebhostingDatabaseCopyRestore(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/database/db1/copyRestore",
		tdhttpmock.JSONBody(td.JSON(`{"copyId":"123","flushDatabase":true}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "db", "copy", "restore", "myservice", "db1", "--copy-id", "123", "--flush")
	require.CmpNoError(err)
	assert.String(out, "⚡️ Restore requested")
}

func (ms *MockSuite) TestWebhostingDatabaseCopyGet(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/database/db1/copy/abc",
		httpmock.NewStringResponder(200, `{
			"id": "abc",
			"status": "done",
			"creationDate": "2025-01-01T00:00:00Z",
			"lastUpdate": "2025-01-02T00:00:00Z",
			"expirationDate": "2025-02-01T00:00:00Z"
		}`),
	)

	out, err := cmd.Execute("webhosting", "db", "copy", "get", "myservice", "db1", "abc", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"id":"abc",
		"status":"done",
		"creationDate":"2025-01-01T00:00:00Z",
		"lastUpdate":"2025-01-02T00:00:00Z",
		"expirationDate":"2025-02-01T00:00:00Z"
	}`))
}

func (ms *MockSuite) TestWebhostingDatabaseAvailableTypes(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/databaseAvailableType",
		httpmock.NewStringResponder(200, `["mysql","postgresql"]`),
	)

	out, err := cmd.Execute("webhosting", "db", "available-type", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"type":"mysql"},{"type":"postgresql"}]`))
}

func (ms *MockSuite) TestWebhostingDatabaseCapabilities(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/database/db1/capabilities",
		httpmock.NewStringResponder(200, `{"changePassword":true,"delete":false}`),
	)

	out, err := cmd.Execute("webhosting", "db", "capabilities", "myservice", "db1", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"changePassword":true,"delete":false}`))
}

func (ms *MockSuite) TestWebhostingDatabaseAvailableVersions(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/databaseAvailableVersion?type=mysql",
		httpmock.NewStringResponder(200, `{"default":"8.0","list":["5.7","8.0"]}`),
	)

	out, err := cmd.Execute("webhosting", "db", "available-version", "myservice", "--type", "mysql", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"default":"8.0","list":["5.7","8.0"]}`))
}

func (ms *MockSuite) TestWebhostingDatabaseCreationCapabilities(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/databaseCreationCapabilities",
		httpmock.NewStringResponder(200, `[{
			"type": "sqlpro",
			"isolation": "shared",
			"available": 2,
			"quota": {"value": 100, "unit": "MB"},
			"engines": ["mysql","postgresql"]
		}]`),
	)

	out, err := cmd.Execute("webhosting", "db", "creation-capabilities", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{
		"type":"sqlpro",
		"isolation":"shared",
		"available":2,
		"quota":{"unit":"MB","value":100},
		"engines":["mysql","postgresql"],
		"quotaDisplay":"100.00 MB",
		"enginesDisplay":"mysql, postgresql"
	}]`))
}

func (ms *MockSuite) TestWebhostingEmailInfoCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/email",
		httpmock.NewStringResponder(200, `{"email":"alerts@example.com","state":"ok","bounce":1,"sentToday":5,"sent":100,"maxPerDay":200}`),
	)

	out, err := cmd.Execute("webhosting", "email", "info", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"email":"alerts@example.com","state":"ok","bounce":1,"sentToday":5,"sent":100,"maxPerDay":200}`))
}

func (ms *MockSuite) TestWebhostingEmailUpdateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/email",
		httpmock.NewStringResponder(200, `{"email":"old@example.com"}`),
	)
	httpmock.RegisterResponder(http.MethodPut,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/email",
		func(req *http.Request) (*http.Response, error) {
			var body map[string]any
			require.CmpNoError(json.NewDecoder(req.Body).Decode(&body))
			assert.Cmp(body, td.JSON(`{"email":"new@example.com"}`))
			return httpmock.NewStringResponse(200, `{}`), nil
		},
	)

	out, err := cmd.Execute("webhosting", "email", "update", "myservice", "--contact-email", "new@example.com")

	require.CmpNoError(err)
	assert.String(out, "✅ Resource updated successfully")
}

func (ms *MockSuite) TestWebhostingEmailBouncesCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/email/bounces?limit=2",
		httpmock.NewStringResponder(200, `[{"date":1700000000,"to":"bounce@example.com","message":"error"}]`),
	)

	out, err := cmd.Execute("webhosting", "email", "bounces", "myservice", "--limit", "2", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"date":"2023-11-14T22:13:20Z","message":"error","to":"bounce@example.com"}]`))
}

func (ms *MockSuite) TestWebhostingEmailRequestCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/email/request",
		tdhttpmock.JSONBody(td.JSON(`{"action":"BLOCK"}`)),
		httpmock.NewStringResponder(200, `"ok"`),
	)

	out, err := cmd.Execute("webhosting", "email", "request-action", "myservice", "--action", "BLOCK")

	require.CmpNoError(err)
	assert.String(out, "⚡️ Email action requested")
}

func (ms *MockSuite) TestWebhostingEmailVolumesCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/email/volumes",
		httpmock.NewStringResponder(200, `[{"date":1700000000,"volume":42}]`),
	)

	out, err := cmd.Execute("webhosting", "email", "volumes", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"date":"2023-11-14T22:13:20Z","volume":42}]`))
}

func (ms *MockSuite) TestWebhostingEmailOptionListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/emailOption",
		httpmock.NewStringResponder(200, `[123]`),
	)

	out, err := cmd.Execute("webhosting", "email-option", "list", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"id":"123"}]`))
}

func (ms *MockSuite) TestWebhostingEmailOptionGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/emailOption/123",
		httpmock.NewStringResponder(200, `{"id":123,"domain":"example.com","creationDate":"2025-01-01T00:00:00Z"}`),
	)

	out, err := cmd.Execute("webhosting", "email-option", "get", "myservice", "123", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"creationDate":"2025-01-01T00:00:00Z","domain":"example.com","id":123}`))
}

func (ms *MockSuite) TestWebhostingEmailOptionServiceInfoCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/emailOption/123/serviceInfos",
		httpmock.NewStringResponder(200, `{
			"serviceId": 42,
			"renew": {
				"automatic": null,
				"deleteAtExpiration": null,
				"forced": null,
				"manualPayment": null,
				"period": null
			}
		}`),
	)

	out, err := cmd.Execute("webhosting", "email-option", "service-info", "myservice", "123", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"serviceId":42,"renew":{"automatic":null,"deleteAtExpiration":null,"forced":null,"manualPayment":null,"period":null}}`))
}

func (ms *MockSuite) TestWebhostingEmailOptionTerminateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/emailOption/123/terminate",
		httpmock.NewStringResponder(200, `"done"`),
	)

	out, err := cmd.Execute("webhosting", "email-option", "terminate", "myservice", "123")

	require.CmpNoError(err)
	assert.String(out, "⚡️ Email option termination requested")
}

func (ms *MockSuite) TestWebhostingExtraSqlListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/extraSqlPerso",
		httpmock.NewStringResponder(200, `["sqlA","sqlB"]`),
	)

	out, err := cmd.Execute("webhosting", "extra-sql", "list", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"id":"sqlA"},{"id":"sqlB"}]`))
}

func (ms *MockSuite) TestWebhostingExtraSqlGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/extraSqlPerso/sqlA",
		httpmock.NewStringResponder(200, `{"name":"sqlA","status":"ok","database":4}`),
	)

	out, err := cmd.Execute("webhosting", "extra-sql", "get", "myservice", "sqlA", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"database":4,"name":"sqlA","status":"ok"}`))
}

func (ms *MockSuite) TestWebhostingExtraSqlDatabasesCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/extraSqlPerso/sqlA/databases",
		httpmock.NewStringResponder(200, `["db1","db2"]`),
	)

	out, err := cmd.Execute("webhosting", "extra-sql", "databases", "myservice", "sqlA", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"database":"db1"},{"database":"db2"}]`))
}

func (ms *MockSuite) TestWebhostingExtraSqlServiceInfoGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/extraSqlPerso/sqlA/serviceInfos",
		httpmock.NewStringResponder(200, `{"state":"ok","quantity":1}`),
	)

	out, err := cmd.Execute("webhosting", "extra-sql", "service-info", "get", "myservice", "sqlA", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"quantity":1,"state":"ok"}`))
}

func (ms *MockSuite) TestWebhostingExtraSqlServiceInfoUpdateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/extraSqlPerso/sqlA/serviceInfosUpdate",
		tdhttpmock.JSONBody(td.JSON(`{"renew":{"automatic":true,"period":12}}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "extra-sql", "service-info", "update", "myservice", "sqlA",
		"--renew-automatic",
		"--renew-period", "12",
	)

	require.CmpNoError(err)
	assert.String(out, "✅ Extra SQL service info updated")
}

func (ms *MockSuite) TestWebhostingExtraSqlTerminateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/extraSqlPerso/sqlA/terminate",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "extra-sql", "terminate", "myservice", "sqlA")

	require.CmpNoError(err)
	assert.String(out, "⚡️ Extra SQL option termination requested")
}

func (ms *MockSuite) TestWebhostingSSHKeyGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/key/ssh",
		httpmock.NewStringResponder(200, `{"publicKey":"ssh-rsa AAA"}`),
	)

	out, err := cmd.Execute("webhosting", "ssh-key", "get", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"publicKey":"ssh-rsa AAA"}`))
}

func (ms *MockSuite) TestWebhostingSSHKeyCreateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/key/ssh",
		httpmock.NewStringResponder(200, `{"publicKey":"ssh-rsa AAA"}`),
	)

	out, err := cmd.Execute("webhosting", "ssh-key", "create", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"publicKey":"ssh-rsa AAA"}`))
}

func (ms *MockSuite) TestWebhostingServiceInfoGet(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/serviceInfos",
		httpmock.NewStringResponder(200, `{"state":"ok"}`),
	)

	out, err := cmd.Execute("webhosting", "service-info", "get", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"state":"ok"}`))
}

func (ms *MockSuite) TestWebhostingServiceInfoUpdate(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/serviceInfos",
		tdhttpmock.JSONBody(td.JSON(`{"renew":{"automatic":true}}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "service-info", "update", "myservice", "--renew-automatic")

	require.CmpNoError(err)
	assert.String(out, "✅ Service info updated")
}

func (ms *MockSuite) TestWebhostingLocalSeoDirectories(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/localSeo/directoriesList?country=FR&offer=LOCAL",
		httpmock.NewStringResponder(200, `{"searchEngines":[{"code":"GOOGLE","displayName":"Google"}]}`),
	)

	out, err := cmd.Execute("webhosting", "local-seo", "directories", "--country", "FR", "--offer", "LOCAL", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"searchEngines":[{"code":"GOOGLE","displayName":"Google"}]}`))
}

func (ms *MockSuite) TestWebhostingLocalSeoEmailAvailability(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/localSeo/emailAvailability?email=test%40example.com",
		httpmock.NewStringResponder(200, `{"availability":"available"}`),
	)

	out, err := cmd.Execute("webhosting", "local-seo", "email-availability", "--email", "test@example.com", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"availability":"available"}`))
}

func (ms *MockSuite) TestWebhostingLocalSeoEmailAvailabilityService(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/localSeo/emailAvailability?email=test%40example.com",
		httpmock.NewStringResponder(200, `{"availability":"alreadyUsed","serviceName":"myservice"}`),
	)

	out, err := cmd.Execute("webhosting", "local-seo", "email-availability", "myservice", "--email", "test@example.com", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"availability":"alreadyUsed","serviceName":"myservice"}`))
}

func (ms *MockSuite) TestWebhostingLocalSeoVisibilityCheck(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/localSeo/visibilityCheck",
		tdhttpmock.JSONBody(td.JSON(`{"country":"FR","name":"My shop","street":"1 rue","zip":"75001"}`)),
		httpmock.NewStringResponder(200, `{"alreadyManaged":false,"searchData":{"id":1,"token":"abc","name":"My shop"}}`),
	)

	out, err := cmd.Execute("webhosting", "local-seo", "visibility-check",
		"--country", "FR",
		"--name", "My shop",
		"--street", "1 rue",
		"--zip", "75001",
		"--json",
	)

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"alreadyManaged":false,"searchData":{"id":1,"name":"My shop","token":"abc"}}`))
}

func (ms *MockSuite) TestWebhostingLocalSeoVisibilityCheckNotFound(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/localSeo/visibilityCheck",
		httpmock.NewStringResponder(400, `{"message":"Location not found","httpCode":400,"errorCode":"notFound"}`),
	)

	out, err := cmd.Execute("webhosting", "local-seo", "visibility-check",
		"--country", "FR",
		"--name", "Unknown",
		"--street", "1 rue",
		"--zip", "75001",
		"--json",
	)

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message":"Location not found","notFound":true}`))
}

func (ms *MockSuite) TestWebhostingLocalSeoVisibilityResult(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/localSeo/visibilityCheckResult?directory=GOOGLE&id=1&token=abc",
		httpmock.NewStringResponder(200, `[{"name":"My shop","city":"Paris","listingUrl":"https://example.com"}]`),
	)

	out, err := cmd.Execute("webhosting", "local-seo", "visibility-result", "1",
		"--directory", "GOOGLE",
		"--token", "abc",
		"--json",
	)

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"directory":"GOOGLE","results":[{"city":"Paris","listingUrl":"https://example.com","name":"My shop"}]}`))
}

func (ms *MockSuite) TestWebhostingLocalSeoAccountList(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/localSeo/account",
		httpmock.NewStringResponder(200, `[1,2]`),
	)

	out, err := cmd.Execute("webhosting", "local-seo", "account", "list", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"id":"1"},{"id":"2"}]`))
}

func (ms *MockSuite) TestWebhostingLocalSeoAccountGet(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/localSeo/account/1",
		httpmock.NewStringResponder(200, `{"id":1,"email":"user@example.com","status":"ok"}`),
	)

	out, err := cmd.Execute("webhosting", "local-seo", "account", "get", "myservice", "1", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"email":"user@example.com","id":1,"status":"ok"}`))
}

func (ms *MockSuite) TestWebhostingLocalSeoAccountLogin(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/localSeo/account/1/login",
		httpmock.NewStringResponder(200, `"https://sso.example.com"`),
	)

	out, err := cmd.Execute("webhosting", "local-seo", "account", "login", "myservice", "1")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("https://sso.example.com"))
}

func (ms *MockSuite) TestWebhostingLocalSeoLocationList(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/localSeo/location",
		httpmock.NewStringResponder(200, `[10]`),
	)

	out, err := cmd.Execute("webhosting", "local-seo", "location", "list", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"id":"10"}]`))
}

func (ms *MockSuite) TestWebhostingLocalSeoLocationGet(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/localSeo/location/10",
		httpmock.NewStringResponder(200, `{"id":10,"name":"Shop","status":"ok"}`),
	)

	out, err := cmd.Execute("webhosting", "local-seo", "location", "get", "myservice", "10", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"id":10,"name":"Shop","status":"ok"}`))
}

func (ms *MockSuite) TestWebhostingLocalSeoLocationServiceInfoUpdate(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/localSeo/location/10/serviceInfosUpdate",
		tdhttpmock.JSONBody(td.JSON(`{"renew":{"automatic":true}}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "local-seo", "location", "service-info", "update", "myservice", "10",
		"--renew-automatic",
	)

	require.CmpNoError(err)
	assert.String(out, "✅ Local SEO service info updated")
}

func (ms *MockSuite) TestWebhostingLocalSeoLocationTerminate(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/localSeo/location/10/terminate",
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "local-seo", "location", "terminate", "myservice", "10")

	require.CmpNoError(err)
	assert.String(out, "⚡️ Local SEO location termination requested")
}

func (ms *MockSuite) TestWebhostingCdnGet(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cdn",
		httpmock.NewStringResponder(200, `{
			"domain": "cdn.example.com",
			"type": "basic",
			"version": "v2",
			"status": "active",
			"free": true,
			"taskId": 42
		}`),
	)

	out, err := cmd.Execute("webhosting", "cdn", "get", "myservice", "--json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"domain":"cdn.example.com",
		"type":"basic",
		"version":"v2",
		"status":"active",
		"free":true,
		"taskId":42
	}`))
}

func (ms *MockSuite) TestWebhostingDatabaseImport(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/database/db1/import",
		tdhttpmock.JSONBody(td.JSON(`{"documentId":"doc123","flushDatabase":true,"sendEmail":true}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("webhosting", "db", "import", "myservice", "db1", "--document-id", "doc123", "--flush", "--send-email")
	require.CmpNoError(err)
	assert.String(out, "⚡️ Import requested")
}

func (ms *MockSuite) TestWebhostingCronUpdate(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPut,
		"https://eu.api.ovh.com/1.0/hosting/web/myservice/cron/12",
		func(req *http.Request) (*http.Response, error) {
			var body map[string]any
			require.CmpNoError(json.NewDecoder(req.Body).Decode(&body))
			assert.Cmp(body, td.JSON(`{"command":"echo new"}`))
			return httpmock.NewStringResponse(200, `{}`), nil
		},
	)

	out, err := cmd.Execute("webhosting", "cron", "update", "myservice", "12", "--command", "echo new")
	require.CmpNoError(err)
	assert.String(out, "✅ Resource updated successfully")
}
