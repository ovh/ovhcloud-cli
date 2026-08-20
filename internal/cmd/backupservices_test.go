// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

const (
	backupTenants = "https://eu.api.ovh.com/v2/backupServices/tenant"
	backupTenant  = backupTenants + "/t-1"
	backupVspc    = backupTenant + "/vspc/s-1"
	backupAgents  = backupVspc + "/backupAgent"
)

// registerOneTenant gives the account the shape it has in reality: one backup
// tenant holding one VSPC tenant, which is why neither has to be named.
func registerOneTenant() {
	httpmock.RegisterResponder(http.MethodGet, backupTenants,
		httpmock.NewStringResponder(200, `[{"id":"t-1","resourceStatus":"READY","targetSpec":{"name":"t-1"},
			"currentState":{"name":"t-1","vaults":["v-1"],"vspcTenants":["s-1"]},"currentTasks":[]}]`))
	httpmock.RegisterResponder(http.MethodGet, backupTenant+"/vspc",
		httpmock.NewStringResponder(200, `[{"id":"s-1","resourceStatus":"READY","targetSpec":{"name":"vspc-tenant-1"},
			"currentState":{"name":"vspc-tenant-1","vspcType":"BASIC","region":"eu-west-rbx",
			"enabledAddons":["BACKUP_AGENT"],"backupAgents":[]},"currentTasks":[]}]`))
	httpmock.RegisterResponder(http.MethodGet, backupVspc+"/backupPolicies",
		httpmock.NewStringResponder(200, `["30d_retention","14d_retention"]`))
}

// registerServer answers the v1 read a creation derives its spec from.
func registerServerForAgent() {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/dedicated/server/ns1.example",
		httpmock.NewStringResponder(200, `{"name":"ns1.example","ip":"203.0.113.7","region":"eu-west-rbx"}`))
}

func registerAgents(body string) {
	httpmock.RegisterResponder(http.MethodGet, backupAgents, httpmock.NewStringResponder(200, body))
}

// "ERROR" answers "error doing what?" with nothing.
func (ms *MockSuite) TestBackupVaultListNamesTheFailedOperation(assert, require *td.T) {
	registerOneTenant()
	httpmock.RegisterResponder(http.MethodGet, backupTenant+"/vault",
		httpmock.NewStringResponder(200, `[{"id":"v-1","resourceStatus":"READY","targetSpec":{"name":"vault-sbg"},
			"currentState":{"name":"vault-sbg","type":"PAYGO","vspcTenants":["s-1"],
			"buckets":[{"region":"eu-west-sbg"}]},
			"currentTasks":[{"id":"k-1","type":"BACKUP_VAULT_CREATE","status":"ERROR"}]}]`))

	out, err := cmd.Execute("backup-services", "vault", "list")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("BACKUP_VAULT_CREATE ERROR"))
	assert.Cmp(out, td.Contains("eu-west-sbg"), "where a vault stores is what tells three of them apart")
	assert.Cmp(out, td.Not(td.Contains("allowedIps")),
		"the addresses allowed on a vault are not a field of a vault, and a column answering 0 would be wrong")
}

// The account has one tenant, so nothing has to be named to reach what hangs
// off it.
func (ms *MockSuite) TestBackupPoliciesResolveTheHierarchyAlone(assert, require *td.T) {
	registerOneTenant()

	out, err := cmd.Execute("backup-services", "policies")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("14d_retention"))
	assert.Cmp(out, td.Contains("30d_retention"))
}

// Two tenants means the command must not pick one, and the refusal has to be
// actionable: an identifier to paste and a name to recognise.
func (ms *MockSuite) TestBackupRefusesToChooseBetweenTenants(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, backupTenants,
		httpmock.NewStringResponder(200, `[{"id":"t-1","targetSpec":{"name":"first"}},{"id":"t-2","targetSpec":{"name":"second"}}]`))

	_, err := cmd.Execute("backup-services", "policies")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("--tenant"))
	assert.Cmp(err.Error(), td.Contains("t-2"))
	assert.Cmp(err.Error(), td.Contains("second"))
}

// An account with no backup tenant is a state, and saying so beats a stack of
// failed lookups underneath.
func (ms *MockSuite) TestBackupSaysWhenThereIsNoTenant(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, backupTenants, httpmock.NewStringResponder(200, `[]`))

	_, err := cmd.Execute("backup-services", "policies")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("no backup tenant"))
}

// No licence on the tenant is an answer, not an empty table.
func (ms *MockSuite) TestBackupLicensesSaysWhenThereAreNone(assert, require *td.T) {
	registerOneTenant()
	httpmock.RegisterResponder(http.MethodGet, backupVspc+"/backupLicenses",
		httpmock.NewStringResponder(200, `[]`))

	out, err := cmd.Execute("backup-services", "licenses", "list")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("no Veeam licence"))
}

// The deploy script is what puts the agent on a machine, and the links carry
// their own authorisation.
func (ms *MockSuite) TestBackupDeployScriptSaysWhatTheLinksAre(assert, require *td.T) {
	registerOneTenant()
	httpmock.RegisterResponder(http.MethodGet, backupVspc+"/managementAgent",
		httpmock.NewStringResponder(200, `{"linuxDeployScript":"curl -sSL https://example/install.sh | sudo sh",
			"linuxUrl":"https://example/linux","macUrl":"https://example/mac","windowsUrl":"https://example/win"}`))

	out, err := cmd.Execute("backup-services", "deploy-script")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("curl -sSL"))
	assert.Cmp(out, td.Contains("their own authorisation"))
}

// The whole point of the shell: an operator types a server name and never sees
// a UUID.
func (ms *MockSuite) TestBaremetalBackupAgentShowsWhatProtectsTheServer(assert, require *td.T) {
	registerOneTenant()
	registerAgents(`[{"id":"a-1","status":"NOT_INSTALLED","targetSpec":{"displayName":"agent-ns1.example","policy":""},
		"currentState":{"productResourceName":"ns1.example","ips":["203.0.113.7/32"],"type":"OVHCLOUD_BAREMETAL","policy":""}},
		{"id":"a-2","status":"ENABLED","targetSpec":{"displayName":"agent-other"},
		"currentState":{"productResourceName":"other.example"}}]`)

	out, err := cmd.Execute("baremetal", "backup-agent", "show", "ns1.example")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("a-1"))
	assert.Cmp(out, td.Not(td.Contains("a-2")), "another server's agent is not this server's")
	assert.Cmp(out, td.Contains("none"), "an agent on no policy retains nothing, and the table says so")
}

// A server with no agent gets the command that makes one, not an empty table.
func (ms *MockSuite) TestBaremetalBackupAgentSaysWhenThereIsNone(assert, require *td.T) {
	registerOneTenant()
	registerAgents(`[]`)

	out, err := cmd.Execute("baremetal", "backup-agent", "show", "ns1.example")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("backup-agent create ns1.example"))
}

// Nine agents out of nine on this account are named agent-<server>, carry the
// server's address in a /32 and sit in the server's region. Nothing but the
// server name should have to be typed.
func (ms *MockSuite) TestBaremetalBackupAgentCreateDerivesEverythingFromTheServer(assert, require *td.T) {
	registerOneTenant()
	registerServerForAgent()
	registerAgents(`[]`)

	var sent map[string]any
	httpmock.RegisterResponder(http.MethodPost, backupAgents,
		func(req *http.Request) (*http.Response, error) {
			if err := json.NewDecoder(req.Body).Decode(&sent); err != nil {
				return nil, err
			}
			return httpmock.NewStringResponse(200, ``), nil
		})

	_, err := cmd.Execute("baremetal", "backup-agent", "create", "ns1.example", "--yes")

	require.CmpNoError(err)
	assert.Cmp(sent["displayName"], "agent-ns1.example")
	assert.Cmp(sent["productResourceName"], "ns1.example")
	assert.Cmp(sent["region"], "eu-west-rbx")
	assert.Cmp(sent["ips"], []any{"203.0.113.7/32"})
}

// A second agent for the same server is not something to create quietly.
func (ms *MockSuite) TestBaremetalBackupAgentCreateRefusesADuplicate(assert, require *td.T) {
	registerOneTenant()
	registerServerForAgent()
	registerAgents(`[{"id":"a-1","status":"NOT_INSTALLED","currentState":{"productResourceName":"ns1.example"}}]`)

	_, err := cmd.Execute("baremetal", "backup-agent", "create", "ns1.example", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("already has a backup agent"))
	assert.Cmp(httpmock.GetCallCountInfo()["POST "+backupAgents], 0)
}

// A region the backup service does not have is refused here rather than as a
// 400 the operator cannot read.
func (ms *MockSuite) TestBaremetalBackupAgentCreateRefusesAnUnknownRegion(assert, require *td.T) {
	registerOneTenant()
	registerServerForAgent()
	registerAgents(`[]`)

	_, err := cmd.Execute("baremetal", "backup-agent", "create", "ns1.example", "--region", "eu-west-nowhere", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("eu-west-rbx"), "the refusal lists what does exist")
	assert.Cmp(httpmock.GetCallCountInfo()["POST "+backupAgents], 0)
}

// The PUT replaces the target spec, so what is not being changed has to travel
// with what is. Sending only the policy would blank the name and the addresses.
func (ms *MockSuite) TestBaremetalBackupAgentEditCarriesTheRestOver(assert, require *td.T) {
	registerOneTenant()
	registerAgents(`[{"id":"a-1","status":"NOT_INSTALLED",
		"targetSpec":{"displayName":"agent-ns1.example","ips":["203.0.113.7/32"],"policy":""},
		"currentState":{"productResourceName":"ns1.example"}}]`)

	var sent map[string]any
	httpmock.RegisterResponder(http.MethodPut, backupAgents+"/a-1",
		func(req *http.Request) (*http.Response, error) {
			if err := json.NewDecoder(req.Body).Decode(&sent); err != nil {
				return nil, err
			}
			return httpmock.NewStringResponse(200, ``), nil
		})

	_, err := cmd.Execute("baremetal", "backup-agent", "edit", "ns1.example", "--policy", "14d_retention", "--yes")

	require.CmpNoError(err)
	assert.Cmp(sent["policy"], "14d_retention")
	assert.Cmp(sent["displayName"], "agent-ns1.example", "the name was not being changed and must survive")
	assert.Cmp(sent["ips"], []any{"203.0.113.7/32"}, "nor were the addresses")
}

// A policy the tenant does not define is refused with the ones it does.
func (ms *MockSuite) TestBaremetalBackupAgentEditRefusesAnUnknownPolicy(assert, require *td.T) {
	registerOneTenant()
	registerAgents(`[{"id":"a-1","status":"NOT_INSTALLED","targetSpec":{"displayName":"agent-ns1.example"},
		"currentState":{"productResourceName":"ns1.example"}}]`)

	_, err := cmd.Execute("baremetal", "backup-agent", "edit", "ns1.example", "--policy", "99d_retention", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("14d_retention"))
	assert.Cmp(httpmock.GetCallCountInfo()["PUT "+backupAgents+"/a-1"], 0)
}

// An empty policy is how an agent is taken off retention, and it is the state
// all nine agents of this account are in — so it is not an unknown policy.
func (ms *MockSuite) TestBaremetalBackupAgentEditAcceptsNoPolicyAtAll(assert, require *td.T) {
	registerOneTenant()
	registerAgents(`[{"id":"a-1","status":"ENABLED","targetSpec":{"displayName":"agent-ns1.example","policy":"14d_retention"},
		"currentState":{"productResourceName":"ns1.example"}}]`)

	var sent map[string]any
	httpmock.RegisterResponder(http.MethodPut, backupAgents+"/a-1",
		func(req *http.Request) (*http.Response, error) {
			if err := json.NewDecoder(req.Body).Decode(&sent); err != nil {
				return nil, err
			}
			return httpmock.NewStringResponse(200, ``), nil
		})

	_, err := cmd.Execute("baremetal", "backup-agent", "edit", "ns1.example", "--policy", "", "--yes")

	require.CmpNoError(err)
	assert.Cmp(sent["policy"], "")
}

// The removal takes the strongest guard the CLI has, and a dry run sends
// nothing while still naming what would go.
func (ms *MockSuite) TestBaremetalBackupAgentDeleteIsPreviewedWithoutSending(assert, require *td.T) {
	registerOneTenant()
	registerAgents(`[{"id":"a-1","status":"ENABLED","targetSpec":{"displayName":"agent-ns1.example"},
		"currentState":{"productResourceName":"ns1.example","policy":"14d_retention"}}]`)

	out, err := cmd.Execute("baremetal", "backup-agent", "delete", "ns1.example", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("DELETE"))
	assert.Cmp(out, td.Contains("a-1"))
	assert.Cmp(httpmock.GetCallCountInfo()["DELETE "+backupAgents+"/a-1"], 0)
}

// The estate-wide view is the other question, and it is the one that made the
// posture of this account visible: nine agents provisioned, none deployed.
func (ms *MockSuite) TestBackupAgentsListsWhatEachOneProtects(assert, require *td.T) {
	registerOneTenant()
	httpmock.RegisterResponder(http.MethodGet, backupAgents,
		httpmock.NewStringResponder(200, `[
		{"id":"a-2","status":"NOT_INSTALLED","targetSpec":{"displayName":"agent-zeta"},
		 "currentState":{"productResourceName":"zeta.example","type":"OVHCLOUD_BAREMETAL","policy":""}},
		{"id":"a-1","status":"ENABLED","targetSpec":{"displayName":"agent-alpha"},
		 "currentState":{"productResourceName":"alpha.example","type":"OVHCLOUD_BAREMETAL","policy":"14d_retention"}}]`))

	out, err := cmd.Execute("backup-services", "agents")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("alpha.example"))
	assert.Cmp(out, td.Contains("zeta.example"))
	assert.Cmp(strings.Index(out, "alpha.example") < strings.Index(out, "zeta.example"), true,
		"the estate reads in the order of the machines it protects")
	assert.Cmp(out, td.Contains("none"), "an agent retaining nothing says so")
}

// No agent at all is an answer with the command that makes one.
func (ms *MockSuite) TestBackupAgentsSaysWhenThereAreNone(assert, require *td.T) {
	registerOneTenant()
	httpmock.RegisterResponder(http.MethodGet, backupAgents, httpmock.NewStringResponder(200, `[]`))

	out, err := cmd.Execute("backup-services", "agents")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("backup-agent create"))
}

// The backup API carries no price. What each part costs is joined from the
// account's service router by the resource's own identifier.
func (ms *MockSuite) TestBackupBillingJoinsThePriceToEachResource(assert, require *td.T) {
	registerOneTenant()
	httpmock.RegisterResponder(http.MethodGet, backupTenant+"/vault",
		httpmock.NewStringResponder(200, `[{"id":"v-1","resourceStatus":"READY","targetSpec":{"name":"vault-sbg"},
			"currentState":{"name":"vault-sbg","buckets":[]},"currentTasks":[]}]`))
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/me/consumption/usage/current",
		httpmock.NewStringResponder(200, `[]`))

	for name, id := range map[string]string{"t-1": "111", "s-1": "222", "v-1": "333"} {
		httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/services?resourceName="+name,
			httpmock.NewStringResponder(200, "["+id+"]"))
	}
	httpmock.RegisterResponder(http.MethodGet, `=~^https://eu\.api\.ovh\.com/v1/services/\d+$`,
		httpmock.NewStringResponder(200, `{"serviceId":333,"billing":{"nextBillingDate":"2026-09-01T00:00:00Z",
			"plan":{"code":"backup-vault-paygo","invoiceName":"Backup vault"},
			"pricing":{"duration":"P1M","price":{"currencyCode":"EUR","text":"0.00 €","value":0}},
			"renew":{"current":{"mode":"automatic"}},"lifecycle":{"current":{"state":"active"}}}}`))

	out, err := cmd.Execute("backup-services", "billing")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("backup-vault-paygo"))
	assert.Cmp(out, td.Contains("P1M"))
	assert.Cmp(out, td.Contains("none yet"), "pay-as-you-go with nothing stored reports no line")
	assert.Cmp(out, td.Contains("vault"), "and each kind of resource is named")
}

// A backup resource with no billable service behind it is what an included
// component looks like, not a failure of the command.
func (ms *MockSuite) TestBackupBillingKeepsGoingWhenAResourceHasNoService(assert, require *td.T) {
	registerOneTenant()
	httpmock.RegisterResponder(http.MethodGet, backupTenant+"/vault",
		httpmock.NewStringResponder(200, `[]`))
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/me/consumption/usage/current",
		httpmock.NewStringResponder(200, `[]`))
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/services?resourceName=t-1",
		httpmock.NewStringResponder(200, `[]`))
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/services?resourceName=s-1",
		httpmock.NewStringResponder(200, `[]`))

	out, err := cmd.Execute("backup-services", "billing")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("tenant"), "the table is still printed")
	assert.Cmp(out, td.Contains("—"), "and the missing price reads as missing")
}
