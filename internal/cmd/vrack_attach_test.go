// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// Responses recorded from the live API on 19 August 2026.
//
// The two lists have the same shape and answer opposite questions:
// allowedServices is what may still be attached, dedicatedServerInterfaceDetails
// is what already is. Both carry the owning server, which is what lets an
// operator name a machine instead of a UUID.
const (
	vrackName  = "pn-1080348"
	vrackOther = "pn-1066983"

	// One server, one interface: the ordinary case.
	vrackAllowedOne = `{"dedicatedServerInterface": [{"dedicatedServerInterface": "1688f939-b93f-47a2-9648-b270f8150a53", "name": "9c:6b:00:22:03:32", "dedicatedServer": "ns0000001.ip-203-0-113.eu"}], "cloudProject": [], "dedicatedServer": []}`

	// One server, two interfaces — an aggregation and a plain one. They do not
	// carry the same traffic, so this is not a choice a CLI may make silently.
	vrackAllowedTwo = `{"dedicatedServerInterface": [{"dedicatedServerInterface": "aaaaaaaa-1111-2222-3333-444444444444", "name": "vrack_aggregation", "dedicatedServer": "ns0000003.ip-203-0-113.eu"}, {"dedicatedServerInterface": "bbbbbbbb-1111-2222-3333-444444444444", "name": "c4:70:bd:16:b2:f5", "dedicatedServer": "ns0000003.ip-203-0-113.eu"}]}`

	vrackAllowedNone = `{"dedicatedServerInterface": []}`

	vrackAttachedOne = `[{"dedicatedServerInterface": "f644bc65-4a61-45f4-9bed-90a0059d35e4", "dedicatedServer": "ns0000002.ip-203-0-113.eu", "name": "d0:50:99:d7:55:0b"}]`

	// 23 of the 35 servers on the account carry a name their owner chose. It is
	// the only string by which somebody recognises the machine they are about
	// to unplug.
	vrackIamServers = `[{"name": "ns0000002.ip-203-0-113.eu", "displayName": "Mail relay - Paris", "type": "dedicatedServer"}, {"name": "ns0000001.ip-203-0-113.eu", "displayName": "ns0000001.ip-203-0-113.eu", "type": "dedicatedServer"}]`

	vrackTask = `{"id": 559188894, "function": "vrack_dedicatedServerInterface", "status": "todo", "serviceName": "pn-1080348"}`
)

var vrackBodies map[string][]map[string]any

// registerVrack wires one vRack and gives the test its own cache directory:
// the display names are cached on disk, so a shared cache would let one test
// answer another one's question.
func registerVrack(t *td.T, allowed, attached string) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	vrackBodies = map[string][]map[string]any{}

	base := "https://eu.api.ovh.com/v1/vrack/" + vrackName

	httpmock.RegisterResponder("GET", base,
		httpmock.NewStringResponder(200, `{"serviceName": "`+vrackName+`", "name": "kuba", "description": "", "iam": {"urn": "urn:v1:eu:resource:vrack:`+vrackName+`"}}`))
	httpmock.RegisterResponder("GET", base+"/allowedServices",
		httpmock.NewStringResponder(200, allowed))
	httpmock.RegisterResponder("GET", base+"/dedicatedServerInterfaceDetails",
		httpmock.NewStringResponder(200, attached))
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/iam/resource?resourceType=dedicatedServer",
		httpmock.NewStringResponder(200, vrackIamServers))

	httpmock.RegisterResponder("POST", base+"/dedicatedServerInterface",
		recordVrackBody("attach"))
	httpmock.RegisterResponder("DELETE", `=~^https://eu\.api\.ovh\.com/v1/vrack/`+vrackName+`/dedicatedServerInterface/.+`,
		recordVrackBody("detach"))
}

func recordVrackBody(key string) httpmock.Responder {
	return func(req *http.Request) (*http.Response, error) {
		var body map[string]any
		if req.Body != nil {
			raw, _ := io.ReadAll(req.Body)
			_ = json.Unmarshal(raw, &body)
		}
		vrackBodies[key] = append(vrackBodies[key], body)
		return httpmock.NewStringResponse(200, vrackTask), nil
	}
}

// An operator names a machine. The API takes a UUID. Turning one into the other
// is the command; without it this is `vrack attach pn-1080348
// 1688f939-b93f-47a2-9648-b270f8150a53`, which nobody can type from memory.
func (ms *MockSuite) TestVrackAttachResolvesTheServerToItsInterface(assert, require *td.T) {
	registerVrack(assert, vrackAllowedOne, `[]`)

	_, err := cmd.Execute("vrack", "attach", vrackName, "ns0000001.ip-203-0-113.eu", "--yes")

	require.CmpNoError(err)
	require.Cmp(len(vrackBodies["attach"]), 1)
	assert.Cmp(vrackBodies["attach"][0]["dedicatedServerInterface"],
		"1688f939-b93f-47a2-9648-b270f8150a53")
}

// A server with an aggregation and a plain interface would work either way most
// of the time, and cut the wrong network the rest of it.
func (ms *MockSuite) TestVrackAttachRefusesToChooseBetweenTwoInterfaces(assert, require *td.T) {
	registerVrack(assert, vrackAllowedTwo, `[]`)

	out, err := cmd.Execute("vrack", "attach", vrackName, "ns0000003.ip-203-0-113.eu", "--yes")

	require.CmpError(err)
	assert.Cmp(len(vrackBodies["attach"]), 0, "nothing is attached while the choice is open")
	assert.Contains(out+err.Error(), "--interface")
	assert.Contains(out+err.Error(), "vrack_aggregation", "and both candidates are named")
}

// --interface is checked against the server's own interfaces rather than passed
// through: a UUID belonging to another machine would otherwise be accepted by
// the API, attaching something nobody named.
func (ms *MockSuite) TestVrackAttachRejectsAnInterfaceOfAnotherServer(assert, require *td.T) {
	registerVrack(assert, vrackAllowedTwo, `[]`)

	_, err := cmd.Execute("vrack", "attach", vrackName, "ns0000003.ip-203-0-113.eu",
		"--interface", "f644bc65-4a61-45f4-9bed-90a0059d35e4", "--yes")

	require.CmpError(err)
	assert.Cmp(len(vrackBodies["attach"]), 0)
}

func (ms *MockSuite) TestVrackAttachUsesTheNamedInterface(assert, require *td.T) {
	registerVrack(assert, vrackAllowedTwo, `[]`)

	_, err := cmd.Execute("vrack", "attach", vrackName, "ns0000003.ip-203-0-113.eu",
		"--interface", "bbbbbbbb-1111-2222-3333-444444444444", "--yes")

	require.CmpNoError(err)
	require.Cmp(len(vrackBodies["attach"]), 1)
	assert.Cmp(vrackBodies["attach"][0]["dedicatedServerInterface"],
		"bbbbbbbb-1111-2222-3333-444444444444")
}

// 7 of the 35 servers measured have no virtual interface at all. That absence
// looks exactly like "already attached elsewhere" in allowedServices, and the
// two call for opposite actions — so the command asks the server itself.
func (ms *MockSuite) TestVrackAttachSaysWhenTheServerHasNoInterfaceAtAll(assert, require *td.T) {
	registerVrack(assert, vrackAllowedNone, `[]`)
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/ns0000004.ip-203-0-113.eu/virtualNetworkInterface",
		httpmock.NewStringResponder(200, `[]`))

	out, err := cmd.Execute("vrack", "attach", vrackName, "ns0000004.ip-203-0-113.eu", "--yes")

	require.CmpError(err)
	assert.Contains(out+err.Error(), "no virtual network interface")
	assert.Cmp(len(vrackBodies["attach"]), 0)
}

// The same silence, the opposite cause: the server has an interface, so it is
// attached somewhere else. Telling an operator to check their hardware here
// would send them looking for a problem that does not exist.
func (ms *MockSuite) TestVrackAttachSaysWhenTheInterfaceIsTakenElsewhere(assert, require *td.T) {
	registerVrack(assert, vrackAllowedNone, `[]`)
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/ns0000002.ip-203-0-113.eu/virtualNetworkInterface",
		httpmock.NewStringResponder(200, `["f644bc65-4a61-45f4-9bed-90a0059d35e4"]`))

	out, err := cmd.Execute("vrack", "attach", vrackName, "ns0000002.ip-203-0-113.eu", "--yes")

	require.CmpError(err)
	assert.Contains(out+err.Error(), "one vRack at a time")
	assert.Not(out+err.Error(), td.Contains("no virtual network interface"))
}

// Detaching cuts a machine off its private network. Without --yes it must not
// happen, whatever else the command was told.
func (ms *MockSuite) TestVrackDetachRefusesWithoutConfirmation(assert, require *td.T) {
	registerVrack(assert, `{}`, vrackAttachedOne)

	_, err := cmd.Execute("vrack", "detach", vrackName, "ns0000002.ip-203-0-113.eu")

	require.CmpError(err)
	assert.Cmp(len(vrackBodies["detach"]), 0)
}

func (ms *MockSuite) TestVrackDetachSendsNothingOnDryRun(assert, require *td.T) {
	registerVrack(assert, `{}`, vrackAttachedOne)

	out, err := cmd.Execute("vrack", "detach", vrackName, "ns0000002.ip-203-0-113.eu", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(len(vrackBodies["detach"]), 0)
	assert.Contains(out, "DELETE")
}

// A vRack holding two servers and a Cloud Connect used to print exactly like an
// empty one: name, description, IAM URN.
func (ms *MockSuite) TestVrackGetShowsWhatIsInside(assert, require *td.T) {
	registerVrack(assert, `{}`, vrackAttachedOne)
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/vrack/"+vrackName+"/cloudProject",
		httpmock.NewStringResponder(200, `["a4c28bfb2d1142caae15610c2c443a3f"]`))

	out, err := cmd.Execute("vrack", "get", vrackName)

	require.CmpNoError(err)
	assert.Contains(out, "ns0000002.ip-203-0-113.eu")
	assert.Contains(out, "Mail relay - Paris")
	assert.Contains(out, "d0:50:99:d7:55:0b")
	assert.Contains(out, "Public Cloud projects", "types this command cannot attach are listed too")
}

// The display name makes the output readable; it is never what the command acts
// on. Losing the lookup must cost the column, not the listing.
func (ms *MockSuite) TestVrackGetStillListsWhenTheNameLookupFails(assert, require *td.T) {
	registerVrack(assert, `{}`, vrackAttachedOne)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/iam/resource?resourceType=dedicatedServer",
		httpmock.NewStringResponder(500, `{"message": "nope"}`))

	out, err := cmd.Execute("vrack", "get", vrackName)

	require.CmpNoError(err)
	assert.Contains(out, "ns0000002.ip-203-0-113.eu")
}

// Ten content lists are read to fill this page, and any of them can fail. The
// vRack itself is not one of them: it was already read, and it is what the
// operator asked for. Reporting a failed section through OutputWarning ended
// the process on the spot — ExitFunc(0) — so `vrack get` printed nothing at all
// and exited 0. Run under the real exit semantics, not the suite's no-op stub,
// because that stub is exactly what hid this.
func (ms *MockSuite) TestVrackGetStillPrintsTheVrackWhenAListFails(assert, require *td.T) {
	registerVrack(assert, `{}`, vrackAttachedOne)
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/vrack/"+vrackName+"/dedicatedServerInterfaceDetails",
		httpmock.NewStringResponder(500, `{"message": "nope"}`))

	out, _, exited := executeWithRealExit(assert, "vrack", "get", vrackName)

	assert.False(exited, "a section that could not be read must not end the command")
	assert.Contains(out, "urn:v1:eu:resource:vrack:"+vrackName, "the vRack is still printed")
	assert.Contains(out, "status code 500", "and the failed section is named, not swallowed")
	assert.Cmp(out, td.Not(td.Contains("This vRack is empty")),
		"never empty on a list that failed: the servers may well be there")
}

// Same work, other door: somebody holding a server should not have to know the
// vRack domain exists. The API is asked which vRack the machine is in.
func (ms *MockSuite) TestBaremetalVrackDetachFindsTheVrackItself(assert, require *td.T) {
	registerVrack(assert, `{}`, vrackAttachedOne)
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/ns0000002.ip-203-0-113.eu/vrack",
		httpmock.NewStringResponder(200, `["`+vrackName+`"]`))

	_, err := cmd.Execute("baremetal", "vrack", "detach", "ns0000002.ip-203-0-113.eu", "--yes")

	require.CmpNoError(err)
	assert.Cmp(len(vrackBodies["detach"]), 1, "the vRack was resolved and the interface detached")
}

func (ms *MockSuite) TestBaremetalVrackShowSaysWhenThereIsNoVrack(assert, require *td.T) {
	registerVrack(assert, `{}`, `[]`)
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/ns0000004.ip-203-0-113.eu/vrack",
		httpmock.NewStringResponder(200, `[]`))

	out, err := cmd.Execute("baremetal", "vrack", "show", "ns0000004.ip-203-0-113.eu")

	require.CmpNoError(err)
	assert.Contains(out, "not in any vRack")
}
