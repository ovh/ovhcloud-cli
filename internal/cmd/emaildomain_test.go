// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestEmailDomainList(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/email/domain",
		httpmock.NewStringResponder(200, `["example.com", "test.com"]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/email/domain/example.com",
		httpmock.NewStringResponder(200, `{
			"domain": "example.com",
			"status": "ok",
			"offer": "email-domain-pack"
		}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/email/domain/test.com",
		httpmock.NewStringResponder(200, `{
			"domain": "test.com",
			"status": "ok",
			"offer": "email-domain-pack"
		}`).Once())

	out, err := cmd.Execute("email-domain", "list", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{"domain": "example.com", "status": "ok", "offer": "email-domain-pack"},
		{"domain": "test.com", "status": "ok", "offer": "email-domain-pack"}
	]`))
}

func (ms *MockSuite) TestEmailDomainRedirectionList(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection",
		httpmock.NewStringResponder(200, `["12345", "67890"]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection/12345",
		httpmock.NewStringResponder(200, `{
			"id": "12345",
			"from": "alias@example.com",
			"to": "destination@example.com",
			"localCopy": false
		}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection/67890",
		httpmock.NewStringResponder(200, `{
			"id": "67890",
			"from": "test@example.com",
			"to": "admin@example.com",
			"localCopy": true
		}`).Once())

	out, err := cmd.Execute("email-domain", "redirection", "list", "example.com", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{"id": "12345", "from": "alias@example.com", "to": "destination@example.com", "localCopy": false},
		{"id": "67890", "from": "test@example.com", "to": "admin@example.com", "localCopy": true}
	]`))
}

func (ms *MockSuite) TestEmailDomainRedirectionGet(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection/12345",
		httpmock.NewStringResponder(200, `{
			"id": "12345",
			"from": "alias@example.com",
			"to": "destination@example.com",
			"localCopy": false
		}`).Once())

	out, err := cmd.Execute("email-domain", "redirection", "get", "example.com", "12345", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"id": "12345",
		"from": "alias@example.com",
		"to": "destination@example.com",
		"localCopy": false
	}`))
}

// The creation endpoint is asynchronous: it answers with the task that will
// create the redirection, never with the redirection. The identifier of that
// answer used to be printed as the redirection's own, which sent people to an
// ID that "redirection get" answers 404 for.
func (ms *MockSuite) TestEmailDomainRedirectionCreate(assert, require *td.T) {
	httpmock.RegisterMatcherResponder("POST", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection",
		tdhttpmock.JSONBody(td.JSON(`{
			"from": "new-alias@example.com",
			"to": "admin@example.com",
			"localCopy": false
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "1234567",
			"domain": "example.com",
			"account": "new-alias@example.com",
			"action": "add",
			"type": "forward"
		}`),
	)

	out, err := cmd.Execute("email-domain", "redirection", "create", "example.com",
		"--from", "new-alias@example.com",
		"--to", "admin@example.com")

	require.CmpNoError(err)
	assert.String(out, `⏳ Email redirection creation requested (task ID: 1234567). Use --wait to wait for it to be applied.`)
}

func (ms *MockSuite) TestEmailDomainRedirectionCreateWithLocalCopy(assert, require *td.T) {
	httpmock.RegisterMatcherResponder("POST", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection",
		tdhttpmock.JSONBody(td.JSON(`{
			"from": "backup@example.com",
			"to": "external@otherdomain.com",
			"localCopy": true
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "7654321",
			"domain": "example.com",
			"account": "backup@example.com",
			"action": "add",
			"type": "forward"
		}`),
	)

	out, err := cmd.Execute("email-domain", "redirection", "create", "example.com",
		"--from", "backup@example.com",
		"--to", "external@otherdomain.com",
		"--local-copy")

	require.CmpNoError(err)
	assert.String(out, `⏳ Email redirection creation requested (task ID: 7654321). Use --wait to wait for it to be applied.`)
}

// With --wait the command follows the task, then goes and gets the redirection
// the task created, so what it prints is the identifier the user can use next.
func (ms *MockSuite) TestEmailDomainRedirectionCreateWait(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection",
		httpmock.NewStringResponder(200, `{
			"id": "1234567",
			"domain": "example.com",
			"account": "new-alias@example.com",
			"action": "add",
			"type": "forward"
		}`))

	// A redirection task carries no status: it is over once it has left the
	// queue, and the API then answers 404.
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/task/redirection/1234567",
		httpmock.NewStringResponder(404, `{"message": "This task does not exist"}`))

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection?from=new-alias%40example.com&to=admin%40example.com",
		httpmock.NewStringResponder(200, `["99999"]`))

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection/99999",
		httpmock.NewStringResponder(200, `{
			"id": "99999",
			"from": "new-alias@example.com",
			"to": "admin@example.com",
			"localCopy": false
		}`))

	out, err := cmd.Execute("email-domain", "redirection", "create", "example.com",
		"--from", "new-alias@example.com",
		"--to", "admin@example.com",
		"--wait")

	require.CmpNoError(err)
	assert.String(out, `✅ Email redirection created successfully (ID: 99999)`)
}

// --wait with -o json gives back the redirection itself, not the task: the
// whole point of waiting is to hand over the created object.
func (ms *MockSuite) TestEmailDomainRedirectionCreateWaitJSON(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection",
		httpmock.NewStringResponder(200, `{"id": "1234567", "action": "add", "type": "forward"}`))

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/task/redirection/1234567",
		httpmock.NewStringResponder(404, `{"message": "This task does not exist"}`))

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection?from=new-alias%40example.com&to=admin%40example.com",
		httpmock.NewStringResponder(200, `["99999"]`))

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection/99999",
		httpmock.NewStringResponder(200, `{
			"id": "99999",
			"from": "new-alias@example.com",
			"to": "admin@example.com",
			"localCopy": false
		}`))

	out, err := cmd.Execute("email-domain", "redirection", "create", "example.com",
		"--from", "new-alias@example.com",
		"--to", "admin@example.com",
		"--wait", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Email redirection created successfully (ID: 99999)",
		"details": {
			"id": "99999",
			"from": "new-alias@example.com",
			"to": "admin@example.com",
			"localCopy": false
		}
	}`))
}

// Nothing forbids several redirections between the same two addresses. Printing
// one of their IDs would hand the user an identifier that may belong to a
// redirection they did not just create, so the command says it cannot tell.
func (ms *MockSuite) TestEmailDomainRedirectionCreateWaitAmbiguous(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection",
		httpmock.NewStringResponder(200, `{"id": "1234567", "action": "add", "type": "forward"}`))

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/task/redirection/1234567",
		httpmock.NewStringResponder(404, `{"message": "This task does not exist"}`))

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection?from=new-alias%40example.com&to=admin%40example.com",
		httpmock.NewStringResponder(200, `["99999", "99998"]`))

	_, err := cmd.Execute("email-domain", "redirection", "create", "example.com",
		"--from", "new-alias@example.com",
		"--to", "admin@example.com",
		"--wait")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("2 redirections go from new-alias@example.com to admin@example.com"))
	assert.Cmp(err.Error(), td.Contains("task 1234567 is over"),
		"the redirection was created: only the lookup could not single it out")
}

// A task that ends badly is the creation failing, and the command has to say so
// rather than report a success it did not witness.
func (ms *MockSuite) TestEmailDomainRedirectionCreateWaitTaskFailed(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection",
		httpmock.NewStringResponder(200, `{"id": "1234567", "action": "add", "type": "forward"}`))

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/task/redirection/1234567",
		httpmock.NewStringResponder(200, `{"id": "1234567", "status": "customerError"}`))

	_, err := cmd.Execute("email-domain", "redirection", "create", "example.com",
		"--from", "new-alias@example.com",
		"--to", "admin@example.com",
		"--wait")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains(`task 1234567 ended with status "customerError"`))
}

// Deleting is asynchronous as well. Without --wait the command reports the
// request, and names the task as a task.
func (ms *MockSuite) TestEmailDomainRedirectionDelete(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection/12345",
		httpmock.NewStringResponder(200, `{
			"id": "7654321",
			"domain": "example.com",
			"account": "alias@example.com",
			"action": "delete",
			"type": "forward"
		}`))

	out, err := cmd.Execute("email-domain", "redirection", "delete", "example.com", "12345")

	require.CmpNoError(err)
	assert.String(out, `⏳ Email redirection 12345 deletion requested (task ID: 7654321). Use --wait to wait for it to be applied.`)
}

// An empty body is a valid answer: there is then no task to point at, and the
// message must not pretend otherwise.
func (ms *MockSuite) TestEmailDomainRedirectionDeleteWithoutTask(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection/12345",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("email-domain", "redirection", "delete", "example.com", "12345")

	require.CmpNoError(err)
	assert.String(out, `⏳ Email redirection 12345 deletion requested`)
}

// With --wait the command follows the task and then checks the redirection
// itself: a 404 on the task alone would also be what a wrong path answers.
func (ms *MockSuite) TestEmailDomainRedirectionDeleteWait(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection/12345",
		httpmock.NewStringResponder(200, `{"id": "7654321", "action": "delete", "type": "forward"}`))

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/task/redirection/7654321",
		httpmock.NewStringResponder(404, `{"message": "This task does not exist"}`))

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection/12345",
		httpmock.NewStringResponder(404, `{"message": "This redirection does not exist"}`))

	out, err := cmd.Execute("email-domain", "redirection", "delete", "example.com", "12345", "--wait")

	require.CmpNoError(err)
	assert.String(out, `✅ Email redirection 12345 deleted successfully`)
}
