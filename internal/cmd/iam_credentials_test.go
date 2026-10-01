// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"net/http"
	"strings"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

const (
	credentialsURL  = "https://eu.api.ovh.com/v1/me/api/credential"
	applicationsURL = "https://eu.api.ovh.com/v1/me/api/application"
	checkURL        = "https://eu.api.ovh.com/v2/iam/authorization/check"
	actionsURL      = "https://eu.api.ovh.com/v2/iam/reference/action"
	typesURL        = "https://eu.api.ovh.com/v2/iam/reference/resource/type"
	testURN         = "urn:v1:eu:resource:dedicatedServer:ns1.example"
	resourceURL     = "https://eu.api.ovh.com/v2/iam/resource/" + testURN
)

// registerCredentials answers the collection and each key, and records the
// query every call was made with.
func registerCredentials(queries *[]string) {
	httpmock.RegisterResponder(http.MethodGet, credentialsURL,
		func(req *http.Request) (*http.Response, error) {
			*queries = append(*queries, req.URL.RawQuery)
			return httpmock.NewStringResponse(200, `[1,2]`), nil
		})
	httpmock.RegisterResponder(http.MethodGet, credentialsURL+"/1",
		func(req *http.Request) (*http.Response, error) {
			*queries = append(*queries, req.URL.RawQuery)
			return httpmock.NewStringResponse(200, `{"credentialId":1,"applicationId":42,"status":"validated",
				"lastUse":"2024-12-19T17:13:01+01:00","expiration":null,"allowedIPs":null,
				"rules":[{"method":"GET","path":"/*"},{"method":"POST","path":"/*"}]}`), nil
		})
	httpmock.RegisterResponder(http.MethodGet, credentialsURL+"/2",
		func(req *http.Request) (*http.Response, error) {
			*queries = append(*queries, req.URL.RawQuery)
			return httpmock.NewStringResponse(200, `{"credentialId":2,"applicationId":42,"status":"validated",
				"lastUse":null,"expiration":null,"allowedIPs":["203.0.113.7/32"],
				"rules":[{"method":"GET","path":"/me"}]}`), nil
		})
}

// A key with a rule on "/*" can call anything with that verb. 139 of the rules
// on the account measured are exactly that, so a rule count would have hidden it.
func (ms *MockSuite) TestCredentialListSaysHowFarAKeyReaches(assert, require *td.T) {
	registerCredentials(&[]string{})

	out, err := cmd.Execute("iam", "credential", "list")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("whole API: GET,POST"))
	assert.Cmp(out, td.Contains("1 path(s)"))
	assert.Cmp(out, td.Contains("never used"), "a key that was never used says so")
	assert.Cmp(out, td.Contains("never expires"))
	assert.Cmp(out, td.Contains("anywhere"), "an unrestricted key says so rather than showing a blank")
}

func (ms *MockSuite) TestCredentialListKeepsOnlyTheUnusedOnesWhenAsked(assert, require *td.T) {
	registerCredentials(&[]string{})

	out, err := cmd.Execute("iam", "credential", "list", "--unused")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("never used"))
	assert.Cmp(out, td.Not(td.Contains("2024-12-19")), "the key that was used is gone")
}

// FetchObjectsParallel builds each object URL as path + "/%s", so a query
// carried over would ask for ".../credential?status=expired/1". Same trap as
// the ticket filters of #261 and the log subscriptions of #263.
func (ms *MockSuite) TestTheStatusFilterStaysOnTheCollectionCall(assert, require *td.T) {
	var queries []string
	registerCredentials(&queries)

	_, err := cmd.Execute("iam", "credential", "list", "--status", "validated")

	require.CmpNoError(err)
	require.Cmp(len(queries) >= 3, true)
	assert.Cmp(queries[0], td.Contains("status=validated"), "the collection is filtered")
	for _, query := range queries[1:] {
		assert.Cmp(query, "", "no per-object call carries the filter")
	}
}

// The accepted values are read from the embedded schema.
func (ms *MockSuite) TestCredentialListRefusesAStatusTheApiDoesNotKnow(assert, require *td.T) {
	registerCredentials(&[]string{})

	_, err := cmd.Execute("iam", "credential", "list", "--status", "revoked")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("revoked"))
	assert.Cmp(err.Error(), td.Contains("validated"), "the refusal lists what is accepted")
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "nothing is asked before the flag is checked")
}

// Revoking a key breaks whatever holds it, at once and without a way back.
// The dry-run test below proves the preview; this one proves the guard, which
// is the half that actually stops a mistake.
func (ms *MockSuite) TestRevokingACredentialAsksFirst(assert, require *td.T) {
	_, err := cmd.Execute("iam", "credential", "delete", "1")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("aborted"))
	assert.Cmp(httpmock.GetCallCountInfo()["DELETE "+credentialsURL+"/1"], 0, "nothing is revoked without --yes")
}

func (ms *MockSuite) TestRevokingACredentialShowsWhatItWouldDoUnderDryRun(assert, require *td.T) {
	out, err := cmd.Execute("iam", "credential", "delete", "1", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("DELETE"))
	assert.Cmp(out, td.Contains("/v1/me/api/credential/1"))
	assert.Cmp(httpmock.GetCallCountInfo()["DELETE "+credentialsURL+"/1"], 0, "nothing is revoked")
}

// Deleting an application takes every key issued against it down with it, so
// the prompt says how many rather than saying "some".
func (ms *MockSuite) TestDeletingAnApplicationCountsTheKeysItWouldBreak(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, credentialsURL,
		httpmock.NewStringResponder(200, `[1,2,3]`))

	_, err := cmd.Execute("iam", "application", "delete", "42")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("aborted"), "no --yes means no deletion")
	assert.Cmp(httpmock.GetCallCountInfo()["DELETE "+applicationsURL+"/42"], 0)
}

// --- iam check ---

func (ms *MockSuite) TestCheckReportsWhatIsAllowedAndWhatIsNot(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost, checkURL,
		httpmock.NewStringResponder(200, `[{"resourceURN":"`+testURN+`",
			"authorizedActions":["dedicatedServer:apiovh:get"],
			"unauthorizedActions":["vps:apiovh:reboot"]}]`))

	out, err := cmd.Execute("iam", "check",
		"dedicatedServer:apiovh:get", "vps:apiovh:reboot", "--on", testURN)

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("dedicatedServer:apiovh:get"))
	assert.Cmp(out, td.Contains("yes"))
	assert.Cmp(out, td.Contains("vps:apiovh:reboot"))
	assert.Cmp(out, td.Contains("no"))
	assert.Cmp(out, td.Contains("dedicatedServer:ns1.example"), "the URN is shortened to what identifies it")
	assert.Cmp(out, td.Not(td.Contains("urn:v1:eu:resource")))
}

func (ms *MockSuite) TestCheckSaysWhereToFindAResourceWhenNoneIsGiven(assert, require *td.T) {
	_, err := cmd.Execute("iam", "check", "dedicatedServer:apiovh:get")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("--on"))
	assert.Cmp(err.Error(), td.Contains("iam resource list"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

// One unknown action fails the batch, so the others go unanswered. Reporting
// the raw 400 would read as "the check does not work".
func (ms *MockSuite) TestCheckExplainsThatOneBadActionSankTheRest(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost, checkURL,
		httpmock.NewStringResponder(400, `{"class":"Client::BadRequest","message":"Unknown action \"a:b:c\""}`))

	_, err := cmd.Execute("iam", "check", "a:b:c", "dedicatedServer:apiovh:get", "--on", testURN)

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("One unknown action fails the whole check"))
	assert.Cmp(err.Error(), td.Contains("iam reference actions"))
}

// --- iam reference ---

// The reference holds over nine thousand actions. Printing them is not an
// answer, so it says how to narrow.
func (ms *MockSuite) TestTheActionReferenceRefusesToPrintItselfWhole(assert, require *td.T) {
	var actions []string
	for i := range 401 {
		actions = append(actions, `{"action":"account:apiovh:a`+string(rune('a'+i%26))+
			string(rune('a'+i/26))+`","categories":["READ"],"resourceType":"account"}`)
	}
	httpmock.RegisterResponder(http.MethodGet, actionsURL,
		httpmock.NewStringResponder(200, "["+strings.Join(actions, ",")+"]"))

	_, err := cmd.Execute("iam", "reference", "actions")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("401 actions match"))
	assert.Cmp(err.Error(), td.Contains("--type"))
	assert.Cmp(err.Error(), td.Contains("iam reference resource-types"))
}

func (ms *MockSuite) TestTheActionReferenceFiltersOnTheServerAndOnTheCategory(assert, require *td.T) {
	var query string
	httpmock.RegisterResponder(http.MethodGet, actionsURL,
		func(req *http.Request) (*http.Response, error) {
			query = req.URL.RawQuery
			return httpmock.NewStringResponse(200, `[
				{"action":"dedicatedServer:apiovh:get","categories":["READ"],"resourceType":"dedicatedServer","description":"Get a server"},
				{"action":"dedicatedServer:apiovh:networking/delete","categories":["DELETE"],"resourceType":"dedicatedServer","description":"Reset networking"}]`), nil
		})

	out, err := cmd.Execute("iam", "reference", "actions", "--type", "dedicatedServer", "--category", "DELETE")

	require.CmpNoError(err)
	assert.Cmp(query, td.Contains("resourceType=dedicatedServer"), "the product filter is sent to the API")
	assert.Cmp(out, td.Contains("networking/delete"))
	assert.Cmp(out, td.Not(td.Contains("apiovh:get")), "the category is applied to what came back")
}

func (ms *MockSuite) TestResourceTypesAreListed(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, typesURL,
		httpmock.NewStringResponder(200, `["vps","account","dedicatedServer"]`))

	out, err := cmd.Execute("iam", "reference", "resource-types")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("dedicatedServer"))
	assert.Cmp(strings.Index(out, "account") < strings.Index(out, "vps"), true, "sorted")
}

// --- fine-grained tagging ---

// One POST per tag, and no read of the others first: that is what keeps them
// and what leaves no window for a concurrent write.
func (ms *MockSuite) TestSettingTagsPostsOnePerTagAndReadsNothingFirst(assert, require *td.T) {
	var bodies []string
	httpmock.RegisterResponder(http.MethodPost, resourceURL+"/tag",
		func(req *http.Request) (*http.Response, error) {
			body := make([]byte, req.ContentLength)
			req.Body.Read(body)
			bodies = append(bodies, string(body))
			return httpmock.NewStringResponse(200, `{}`), nil
		})

	out, err := cmd.Execute("iam", "resource", "tag", "set", testURN, "env=prod", "team=infra")

	require.CmpNoError(err)
	require.Cmp(len(bodies), 2)
	assert.Cmp(bodies[0], td.Contains(`"key":"env"`))
	assert.Cmp(bodies[1], td.Contains(`"key":"team"`))
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+resourceURL], 0, "the other tags are not read")
	assert.Cmp(httpmock.GetCallCountInfo()["PUT "+resourceURL], 0, "and never PUT back")
	assert.Cmp(out, td.Contains("left alone"))
}

// The preview prints one line per tag, and each line has to say which tag:
// common.Call carries no body, so the plain preview would repeat the same
// endpoint and never name what would be written.
func (ms *MockSuite) TestSettingTagsUnderDryRunSaysWhichTag(assert, require *td.T) {
	out, err := cmd.Execute("iam", "resource", "tag", "set", testURN, "env=prod", "team=infra", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains(`"env": "prod"`))
	assert.Cmp(out, td.Contains(`"team": "infra"`))
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "nothing is written")
}

func (ms *MockSuite) TestSettingTheSameKeyTwiceIsRefusedBeforeAnyCall(assert, require *td.T) {
	_, err := cmd.Execute("iam", "resource", "tag", "set", testURN, "env=prod", "env=staging")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("given twice"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

// Removing a key that is not there answers 204 and reads as success. The
// mistake behind it is a typo, so the answer is what the resource does carry.
func (ms *MockSuite) TestRemovingAKeyThatIsNotThereNamesTheOnesThatAre(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, resourceURL,
		httpmock.NewStringResponder(200, `{"urn":"`+testURN+`","tags":{"env":"prod","team":"infra"}}`))

	_, err := cmd.Execute("iam", "resource", "tag", "remove", testURN, "onwer")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("carries no tag onwer"))
	assert.Cmp(err.Error(), td.Contains("env, team"))
	assert.Cmp(httpmock.GetCallCountInfo()["DELETE "+resourceURL+"/tag/onwer"], 0)
}

func (ms *MockSuite) TestRemovingTagsDeletesEachKeyAndLeavesTheRest(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, resourceURL,
		httpmock.NewStringResponder(200, `{"tags":{"env":"prod","team":"infra","owner":"x"}}`))
	httpmock.RegisterResponder(http.MethodDelete, resourceURL+"/tag/env",
		httpmock.NewStringResponder(204, ""))
	httpmock.RegisterResponder(http.MethodDelete, resourceURL+"/tag/team",
		httpmock.NewStringResponder(204, ""))

	out, err := cmd.Execute("iam", "resource", "tag", "remove", testURN, "env", "team")

	require.CmpNoError(err)
	assert.Cmp(httpmock.GetCallCountInfo()["DELETE "+resourceURL+"/tag/env"], 1)
	assert.Cmp(httpmock.GetCallCountInfo()["DELETE "+resourceURL+"/tag/team"], 1)
	assert.Cmp(httpmock.GetCallCountInfo()["PUT "+resourceURL], 0, "never a whole-map write")
	assert.Cmp(out, td.Contains("left alone"))
}

func (ms *MockSuite) TestRemovingTagsUnderDryRunTouchesNothing(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, resourceURL,
		httpmock.NewStringResponder(200, `{"tags":{"env":"prod"}}`))

	out, err := cmd.Execute("iam", "resource", "tag", "remove", testURN, "env", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("DELETE"))
	assert.Cmp(out, td.Contains("/tag/env"))
	assert.Cmp(httpmock.GetCallCountInfo()["DELETE "+resourceURL+"/tag/env"], 0)
}

// Disabling a user cuts their access. It interrupts, it does not destroy.
func (ms *MockSuite) TestDisablingAUserAsksFirst(assert, require *td.T) {
	_, err := cmd.Execute("iam", "user", "disable", "alice")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("aborted"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

// Enabling one does not interrupt anything, so it does not ask.
func (ms *MockSuite) TestEnablingAUserDoesNotAsk(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/me/identity/user/alice/enable",
		httpmock.NewStringResponder(200, `null`))

	out, err := cmd.Execute("iam", "user", "enable", "alice")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("enabled"))
}

// --filter is registered on `iam credential list`, so it has to reach the
// rows. The assertion that carries this test is the absence of key 1: checking
// only that key 2 is present would pass just as well with no filtering.
func (ms *MockSuite) TestIamCredentialListIsFiltered(assert, require *td.T) {
	var queries []string
	registerCredentials(&queries)

	out, err := cmd.Execute("iam", "credential", "list", "--filter", `credentialId==2`)

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("203.0.113.7/32"), "the key the filter keeps")
	assert.Cmp(out, td.Not(td.Contains("2024-12-19")), "the key the filter excludes must not be printed")
}
