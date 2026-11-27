// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package common_test

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

// withMockedAPI points the shared client at httpmock for the duration of a test.
func withMockedAPI(t *testing.T) {
	t.Helper()
	httpmock.Activate(t)

	origClient := httpLib.Client
	client, err := ovh.NewClient("ovh-eu", "app_key", "app_secret", "consumer_key")
	td.Require(t).CmpNoError(err)
	httpLib.Client = httpLib.NewAPIClient(client)

	t.Cleanup(func() { httpLib.Client = origClient })

	// go-ovh computes its clock delta before the first signed call.
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
		httpmock.NewStringResponder(200, "0"))
}

// The parameters of a creation do not always come from the caller's flag
// struct: --from-file, --editor and piped input fill the request without going
// through it. A caller that has to find the created object back needs what was
// actually sent, which is what this function is for — the response of an
// asynchronous endpoint is a task and says nothing about the object.
func TestCreateResourceWithParametersReturnsWhatWasSent(t *testing.T) {
	withMockedAPI(t)

	file := filepath.Join(t.TempDir(), "params.json")
	td.Require(t).CmpNoError(os.WriteFile(file,
		[]byte(`{"from": "alias@example.com", "to": "dest@example.com"}`), 0o600))

	flags.ParametersFile = file
	t.Cleanup(func() { flags.ParametersFile = "" })

	var sentBody string
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection",
		func(r *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(r.Body)
			td.Require(t).CmpNoError(err)
			sentBody = string(body)

			return httpmock.NewStringResponse(200, `{"id": "1234567", "action": "add"}`), nil
		})

	// Empty, as it would be on a run driven entirely by the file.
	var cliParams struct {
		From string `json:"from,omitempty"`
		To   string `json:"to,omitempty"`
	}

	created, sent, err := common.CreateResourceWithParameters(
		&cobra.Command{},
		"/email/domain/{serviceName}/redirection",
		"/email/domain/example.com/redirection",
		"{}",
		cliParams,
		nil,
		[]string{"from", "to"},
	)

	td.Require(t).CmpNoError(err)
	td.Cmp(t, json.RawMessage(sentBody), td.JSON(`{"from": "alias@example.com", "to": "dest@example.com"}`))
	td.Cmp(t, sent, map[string]any{"from": "alias@example.com", "to": "dest@example.com"},
		"the values that were sent come back, although no flag carried them")
	td.Cmp(t, created["action"], "add",
		"the response is the task, which names no address at all")
}

// A creation that never reached the API still returns what it was going to
// send, so the caller can say what it failed to create.
func TestCreateResourceWithParametersKeepsTheParametersOnFailure(t *testing.T) {
	withMockedAPI(t)

	file := filepath.Join(t.TempDir(), "params.json")
	td.Require(t).CmpNoError(os.WriteFile(file,
		[]byte(`{"from": "alias@example.com", "to": "dest@example.com"}`), 0o600))

	flags.ParametersFile = file
	t.Cleanup(func() { flags.ParametersFile = "" })

	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection",
		httpmock.NewStringResponder(400, `{"message": "This redirection already exists"}`))

	var cliParams struct {
		From string `json:"from,omitempty"`
		To   string `json:"to,omitempty"`
	}

	created, sent, err := common.CreateResourceWithParameters(
		&cobra.Command{},
		"/email/domain/{serviceName}/redirection",
		"/email/domain/example.com/redirection",
		"{}",
		cliParams,
		nil,
		[]string{"from", "to"},
	)

	td.Require(t).CmpError(err)
	td.Cmp(t, created, td.Nil())
	td.Cmp(t, sent, map[string]any{"from": "alias@example.com", "to": "dest@example.com"})
}
