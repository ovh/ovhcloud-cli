// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

const (
	logsServer        = "https://eu.api.ovh.com/v2/dedicated/server/ns1.example"
	logsKinds         = logsServer + "/log/kind"
	logsSubscriptions = logsServer + "/log/subscription"
	logsURL           = logsServer + "/log/url"
)

// registerOneLogKind gives the server the single kind the API offers today.
func registerOneLogKind() {
	httpmock.RegisterResponder(http.MethodGet, logsKinds,
		httpmock.NewStringResponder(200, `["install"]`))
	httpmock.RegisterResponder(http.MethodGet, logsKinds+"/install",
		httpmock.NewStringResponder(200, `{"kindId":"75fa0fec-812c-46af-a41b-76d1e3dc2843","name":"install","displayName":"Operating system installation logs","additionalReturnedFields":["level","os","serviceName","status"]}`))
}

// registerStreams answers the Log Data Platform sweep with two services, one of
// which carries a title the other one carries too — the case this account
// really has three times over.
func registerStreams() {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/dbaas/logs",
		httpmock.NewStringResponder(200, `["ldp-aa-1","ldp-bb-2"]`))
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/dbaas/logs/ldp-aa-1/output/graylog/stream",
		httpmock.NewStringResponder(200, `["11111111-1111-1111-1111-111111111111","33333333-3333-3333-3333-333333333333"]`))
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/dbaas/logs/ldp-bb-2/output/graylog/stream",
		httpmock.NewStringResponder(200, `["22222222-2222-2222-2222-222222222222"]`))
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/dbaas/logs/ldp-aa-1/output/graylog/stream/11111111-1111-1111-1111-111111111111",
		httpmock.NewStringResponder(200, `{"streamId":"11111111-1111-1111-1111-111111111111","title":"shared"}`))
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/dbaas/logs/ldp-aa-1/output/graylog/stream/33333333-3333-3333-3333-333333333333",
		httpmock.NewStringResponder(200, `{"streamId":"33333333-3333-3333-3333-333333333333","title":"TO REMOVE 1"}`))
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/dbaas/logs/ldp-bb-2/output/graylog/stream/22222222-2222-2222-2222-222222222222",
		httpmock.NewStringResponder(200, `{"streamId":"22222222-2222-2222-2222-222222222222","title":"shared"}`))
}

func (ms *MockSuite) TestBaremetalLogKindsSaysWhatAKindHolds(assert, require *td.T) {
	registerOneLogKind()

	out, err := cmd.Execute("baremetal", "logs", "kinds", "ns1.example")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Operating system installation logs"),
		"the name alone would send everybody to the API documentation")
	assert.Cmp(out, td.Contains("status"), "and the extra fields say what a search will return")
}

// A server with no kind is an answer, not an empty table with headers.
func (ms *MockSuite) TestBaremetalLogKindsSaysWhenThereAreNone(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, logsKinds, httpmock.NewStringResponder(200, `[]`))

	out, err := cmd.Execute("baremetal", "logs", "kinds", "ns1.example")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("no log kind"))
}

// The kind filter belongs to the collection call. Carried into the expansion it
// would build ".../log/subscription?kind=install/<id>", a route that does not
// exist, for a filter that would have looked like it worked.
func (ms *MockSuite) TestBaremetalLogSubscriptionsFilterStaysOnTheCollection(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, logsSubscriptions,
		httpmock.NewStringResponder(200, `["sub-1"]`))
	httpmock.RegisterResponder(http.MethodGet, logsSubscriptions+"/sub-1",
		httpmock.NewStringResponder(200, `{"subscriptionId":"sub-1","kind":"install","streamId":"11111111-1111-1111-1111-111111111111","serviceName":"ldp-aa-1"}`))

	out, err := cmd.Execute("baremetal", "logs", "subscription", "list", "ns1.example", "--kind", "install")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("sub-1"))

	for url := range httpmock.GetCallCountInfo() {
		assert.Cmp(url, td.Not(td.Contains("kind=install/")),
			"the query must not be carried into the per-object URL")
	}
}

// One kind means there is nothing to choose. Making somebody type the only
// possible value is friction with no purpose.
func (ms *MockSuite) TestBaremetalLogSubscribeDoesNotAskForTheOnlyKind(assert, require *td.T) {
	registerOneLogKind()
	var sent map[string]any
	httpmock.RegisterResponder(http.MethodPost, logsSubscriptions,
		func(req *http.Request) (*http.Response, error) {
			if err := json.NewDecoder(req.Body).Decode(&sent); err != nil {
				return nil, err
			}
			return httpmock.NewStringResponse(200, `{"operationId":"op-1","serviceName":"ldp-aa-1"}`), nil
		})

	_, err := cmd.Execute("baremetal", "logs", "subscribe", "ns1.example",
		"--stream", "11111111-1111-1111-1111-111111111111", "--yes")

	require.CmpNoError(err)
	assert.Cmp(sent["kind"], "install", "the server was asked, and it had one answer")
	assert.Cmp(sent["streamId"], "11111111-1111-1111-1111-111111111111")
}

// Two kinds means the command must not pick one. The day a second kind appears,
// this is the difference between a refusal and silently subscribing the wrong
// logs.
func (ms *MockSuite) TestBaremetalLogSubscribeRefusesToChooseBetweenKinds(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, logsKinds,
		httpmock.NewStringResponder(200, `["install","syslog"]`))

	_, err := cmd.Execute("baremetal", "logs", "subscribe", "ns1.example",
		"--stream", "11111111-1111-1111-1111-111111111111", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("--kind"))
	assert.Cmp(err.Error(), td.Contains("install, syslog"), "and it names them, so the answer is in the refusal")
	assert.Cmp(httpmock.GetCallCountInfo()["POST "+logsSubscriptions], 0, "nothing was sent")
}

// Three titles on this account are carried by two streams each. Picking one
// would send a machine's logs somewhere nobody asked for.
func (ms *MockSuite) TestBaremetalLogSubscribeRefusesAnAmbiguousTitle(assert, require *td.T) {
	registerOneLogKind()
	registerStreams()

	_, err := cmd.Execute("baremetal", "logs", "subscribe", "ns1.example", "--stream", "shared", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("11111111-1111-1111-1111-111111111111"))
	assert.Cmp(err.Error(), td.Contains("22222222-2222-2222-2222-222222222222"))
	assert.Cmp(err.Error(), td.Contains("ldp-bb-2"), "the service is what tells the two apart")
	assert.Cmp(httpmock.GetCallCountInfo()["POST "+logsSubscriptions], 0)
}

// A title resolves to the identifier the API is given, and the operator never
// has to know it.
func (ms *MockSuite) TestBaremetalLogSubscribeResolvesATitle(assert, require *td.T) {
	registerOneLogKind()
	registerStreams()
	var sent map[string]any
	httpmock.RegisterResponder(http.MethodPost, logsSubscriptions,
		func(req *http.Request) (*http.Response, error) {
			if err := json.NewDecoder(req.Body).Decode(&sent); err != nil {
				return nil, err
			}
			return httpmock.NewStringResponse(200, `{"operationId":"op-1","serviceName":"ldp-aa-1"}`), nil
		})

	_, err := cmd.Execute("baremetal", "logs", "subscribe", "ns1.example", "--stream", "TO REMOVE 1", "--yes")

	require.CmpNoError(err)
	assert.Cmp(sent["streamId"], "33333333-3333-3333-3333-333333333333")
}

// The operation says the platform finished. The subscription says the logs
// actually go somewhere. Every wait in this CLI reads the second.
func (ms *MockSuite) TestBaremetalLogSubscribeWaitReadsTheSubscriptionBack(assert, require *td.T) {
	registerOneLogKind()
	httpmock.RegisterResponder(http.MethodPost, logsSubscriptions,
		httpmock.NewStringResponder(200, `{"operationId":"op-1","serviceName":"ldp-aa-1"}`))
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/dbaas/logs/ldp-aa-1/operation/op-1",
		httpmock.NewStringResponder(200, `{"operationId":"op-1","state":"SUCCESS","subscriptionId":"sub-9"}`))
	httpmock.RegisterResponder(http.MethodGet, logsSubscriptions+"/sub-9",
		httpmock.NewStringResponder(200, `{"subscriptionId":"sub-9","kind":"install","streamId":"11111111-1111-1111-1111-111111111111","serviceName":"ldp-aa-1"}`))

	out, err := cmd.Execute("baremetal", "logs", "subscribe", "ns1.example",
		"--stream", "11111111-1111-1111-1111-111111111111", "--wait", "--yes")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("sub-9"))
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+logsSubscriptions+"/sub-9"], 1,
		"the state is read, not assumed from the operation")
}

// FAILURE and REVOKED both end the operation without doing the work.
func (ms *MockSuite) TestBaremetalLogSubscribeWaitReportsAFailedOperation(assert, require *td.T) {
	registerOneLogKind()
	httpmock.RegisterResponder(http.MethodPost, logsSubscriptions,
		httpmock.NewStringResponder(200, `{"operationId":"op-1","serviceName":"ldp-aa-1"}`))
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/dbaas/logs/ldp-aa-1/operation/op-1",
		httpmock.NewStringResponder(200, `{"operationId":"op-1","state":"FAILURE"}`))

	_, err := cmd.Execute("baremetal", "logs", "subscribe", "ns1.example",
		"--stream", "11111111-1111-1111-1111-111111111111", "--wait", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("FAILURE"))
	assert.Cmp(err.Error(), td.Not(td.Contains("now go to")), "a failed operation is not a subscription")
}

// The removal is the same rule the other way round: the operation finishing is
// not the subscription being gone.
func (ms *MockSuite) TestBaremetalLogUnsubscribeWaitChecksItIsActuallyGone(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, logsSubscriptions+"/sub-9",
		httpmock.NewStringResponder(200, `{"subscriptionId":"sub-9","kind":"install","streamId":"11111111-1111-1111-1111-111111111111","serviceName":"ldp-aa-1"}`))
	httpmock.RegisterResponder(http.MethodDelete, logsSubscriptions+"/sub-9",
		httpmock.NewStringResponder(200, `{"operationId":"op-2","serviceName":"ldp-aa-1"}`))
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/dbaas/logs/ldp-aa-1/operation/op-2",
		httpmock.NewStringResponder(200, `{"operationId":"op-2","state":"SUCCESS"}`))

	_, err := cmd.Execute("baremetal", "logs", "unsubscribe", "ns1.example", "sub-9", "--wait", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("still there"),
		"the subscription still answers, so the command must not report it gone")
}

// An identifier that is not there is not a failure to report as one: it is the
// list somebody needs.
func (ms *MockSuite) TestBaremetalLogUnsubscribeSaysWhatToListWhenTheIdIsWrong(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, logsSubscriptions+"/nope",
		httpmock.NewStringResponder(404, `{"message":"not found"}`))

	_, err := cmd.Execute("baremetal", "logs", "unsubscribe", "ns1.example", "nope", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("logs subscription list"))
	assert.Cmp(httpmock.GetCallCountInfo()["DELETE "+logsSubscriptions+"/nope"], 0)
}

// The prompt names what stops, read from the subscription, rather than quoting
// back an identifier the operator just pasted.
func (ms *MockSuite) TestBaremetalLogUnsubscribeNamesWhatItStops(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, logsSubscriptions+"/sub-9",
		httpmock.NewStringResponder(200, `{"subscriptionId":"sub-9","kind":"install","streamId":"11111111-1111-1111-1111-111111111111","serviceName":"ldp-aa-1"}`))

	out, err := cmd.Execute("baremetal", "logs", "unsubscribe", "ns1.example", "sub-9", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("DELETE"))
	assert.Cmp(httpmock.GetCallCountInfo()["DELETE "+logsSubscriptions+"/sub-9"], 0, "a dry run sends nothing")
}

// The link is the answer, so it is printed; what it costs is said beside it.
func (ms *MockSuite) TestBaremetalLogUrlPrintsTheLinkAndItsExpiry(assert, require *td.T) {
	registerOneLogKind()
	httpmock.RegisterResponder(http.MethodPost, logsURL,
		httpmock.NewStringResponder(200, `{"url":"https://get.logs.ovh.com/search?plq=abc","expirationDate":"2099-01-01T00:30:00+00:00"}`))

	out, err := cmd.Execute("baremetal", "logs", "url", "ns1.example")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("get.logs.ovh.com"))
	assert.Cmp(out, td.Contains("2099-01-01T00:30:00+00:00"), "a link that expires in silence is a trap")
	assert.Cmp(out, td.Contains("anyone holding it"), "and it carries its own authorisation")
}

// What an operator checks before agreeing to a subscription is the stream and
// the kind, and neither is in the URL.
func (ms *MockSuite) TestBaremetalLogSubscribeDryRunShowsTheBody(assert, require *td.T) {
	registerOneLogKind()

	out, err := cmd.Execute("baremetal", "logs", "subscribe", "ns1.example",
		"--stream", "11111111-1111-1111-1111-111111111111", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("streamId"))
	assert.Cmp(out, td.Contains("install"))
	assert.Cmp(httpmock.GetCallCountInfo()["POST "+logsSubscriptions], 0)
}
