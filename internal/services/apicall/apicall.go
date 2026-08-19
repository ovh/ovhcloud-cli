// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

// Package apicall exposes the CLI's signed HTTP client as a command, so that
// an endpoint the CLI has no command for is still reachable with the
// credentials, the signature and the output formats already configured.
package apicall

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/editor"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
)

// DryRun reports the request instead of sending it. It is bound by the command.
var DryRun bool

// supportedMethods is the set of verbs the underlying client can sign. PATCH is
// absent because go-ovh does not implement it.
var supportedMethods = []string{"GET", "POST", "PUT", "DELETE"}

// normalizePath accepts the shapes a user actually types. "dedicated/server",
// "/dedicated/server" and "/v1/dedicated/server" are the same endpoint, and
// only the last is what the client expects.
func normalizePath(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// A path already carrying its API version is left alone: /v2 exists and
	// prefixing it with /v1 would silently call something else.
	if !strings.HasPrefix(path, "/v1/") && !strings.HasPrefix(path, "/v2/") {
		path = "/v1" + path
	}

	return path
}

// readBody builds the request payload from --from-file or from the editor.
// A write with neither is a write with no body, which several endpoints accept.
//
// The bytes the operator wrote are kept and forwarded, not decoded and
// re-encoded. Decoding into `any` turns every JSON number into a float64, whose
// 53-bit mantissa silently rewrites 9007199254740993 as 9007199254740992 — and
// this command, which knows nothing about the payload and exists precisely to
// send what it was handed, has no way to notice. The parse still happens: it is
// what refuses malformed JSON before anything is signed. Only its output is
// discarded.
func readBody() (any, error) {
	var raw []byte

	switch {
	case flags.ParametersFile != "":
		content, err := readFile(flags.ParametersFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read parameters file: %w", err)
		}
		if !json.Valid(content) {
			return nil, fmt.Errorf("failed to parse parameters file: %s is not valid JSON", flags.ParametersFile)
		}
		raw = content

	case flags.ParametersViaEditor:
		edited, err := editor.EditValueWithEditor([]byte("{}"))
		if err != nil {
			return nil, fmt.Errorf("failed to edit payload: %w", err)
		}
		if !json.Valid(edited) {
			return nil, errors.New("failed to parse edited payload: it is not valid JSON")
		}
		raw = edited
	}

	if raw == nil {
		return nil, nil
	}
	return json.RawMessage(raw), nil
}

// Call runs one signed request against the API and renders the answer.
func Call(cmd *cobra.Command, args []string) {
	call(cmd, args, false)
}

// CallWrappingResult is Call with the older output shape, where the answer sits
// under a "details" key.
//
// It exists for `ovhcloud webhosting api call`, which shipped that shape and
// whose users have jq expressions written against it. The wrapper is a defect —
// a caller reads .details.datacenter for a field the API documents as
// .datacenter — but it is a defect somebody's script already accommodates, and
// quietly changing it would break those scripts to fix a command that is being
// superseded anyway. The new name gets the right shape; the old one keeps its
// word.
func CallWrappingResult(cmd *cobra.Command, args []string) {
	call(cmd, args, true)
}

func call(_ *cobra.Command, args []string, wrapResult bool) {
	method := strings.ToUpper(args[0])
	path := normalizePath(args[1])

	if !contains(supportedMethods, method) {
		display.OutputError(&flags.OutputFormatConfig,
			"unsupported method %q, expected one of %s", method, strings.Join(supportedMethods, ", "))
		return
	}

	var body any
	if method == "POST" || method == "PUT" {
		var err error
		if body, err = readBody(); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}
	}

	if DryRun {
		reportDryRun(method, path, body)
		return
	}

	var result any
	var err error

	switch method {
	case "GET":
		err = httpLib.Client.Get(path, &result)
	case "POST":
		err = httpLib.Client.Post(path, body, &result)
	case "PUT":
		err = httpLib.Client.Put(path, body, &result)
	case "DELETE":
		err = httpLib.Client.Delete(path, &result)
	}

	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "request failed: %s", err)
		return
	}

	// A 204 leaves nothing to render, and printing an empty object would read
	// as "the API answered nothing" rather than "it answered nothing to say".
	if result == nil {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ %s %s completed", method, path)
		return
	}

	if wrapResult {
		display.OutputWithFormat(&display.OutputMessage{Details: result}, &flags.OutputFormatConfig)
		return
	}

	// The answer is relayed, not composed: it is rendered as the endpoint
	// worded it, so a script reads the fields the API documents.
	display.OutputRaw(result, &flags.OutputFormatConfig)
}

// reportDryRun prints the request that would have been signed and sent. It
// exists because this command has none of the confirmations the product
// commands carry: seeing the exact path before sending is the only guard.
func reportDryRun(method, path string, body any) {
	details := map[string]any{"method": method, "path": path}
	message := fmt.Sprintf("🔍 Dry run: nothing was sent. This would have been called:\n  %s %s", method, path)

	if body != nil {
		payload, err := json.MarshalIndent(body, "  ", "  ")
		if err == nil {
			details["body"] = body
			message += fmt.Sprintf("\n  with body:\n  %s", payload)
		}
	}

	display.OutputInfo(&flags.OutputFormatConfig, details, "%s", message)
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}

	return false
}
