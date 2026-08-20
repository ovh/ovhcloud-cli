// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

// A dedicated server can send its logs to a Log Data Platform stream, and none
// of it was reachable from here. The seven routes live in the v2 catalogue,
// which this CLI could not even fetch a schema for until now, so the surface
// was invisible rather than skipped.
//
// The v2 routes are a thin proxy over LDP rather than a resource of their own:
// they answer with dbaas.logs.* types, and a subscription that is being created
// or removed is followed on the v1 LDP service, not on anything under /v2. So a
// single --wait crosses both catalogues, and the object it needs to do that —
// the LDP service carrying the stream — is handed to it by the API itself.
//
// Every one of these operations is badged "Alpha version" upstream.

const (
	logKindsPath         = "/v2/dedicated/server/%s/log/kind"
	logSubscriptionsPath = "/v2/dedicated/server/%s/log/subscription"
	logURLPath           = "/v2/dedicated/server/%s/log/url"
)

var (
	// LogKind is the kind of log a command works on.
	//
	// It is optional. Only one kind exists today — "install", the operating
	// system installation logs — and requiring an operator to name the only
	// possible value is friction with no purpose. Defaulting to the string
	// "install" in Go would be worse: it would keep working, and keep being
	// the only kind anyone could reach, on the day a second one appears. So
	// the server is asked what it has, and the answer decides.
	LogKind string

	// LogStream is the stream to subscribe to, by title or by identifier.
	LogStream string

	// LogWait keeps the command running until the operation has finished.
	LogWait bool
)

// Polling settings for the LDP operation behind a subscription change.
//
// Variables rather than constants so a test can exercise the polling itself
// without waiting five seconds a round. Measured against the real API on
// 20 August 2026, a subscription settles in about seven seconds and its removal
// in six, so the five-minute ceiling is wide.
var (
	logPollInterval = 5 * time.Second
	logPollAttempts = 60
)

// logOperationStates: dbaas.logs.OperationStateEnum, which is a sixth status
// vocabulary in this repository and shares nothing with the five already here.
// FAILURE and REVOKED both end the operation without doing the work; RETRY is
// not an ending, it is the platform trying again.
var (
	logOperationSucceeded = "SUCCESS"
	logOperationFailed    = map[string]bool{"FAILURE": true, "REVOKED": true}
)

// ListBaremetalLogKinds shows what a server can send, expanded.
//
// The list route answers with names alone. Each name is then read, because the
// useful part is what a kind actually contains — its display name and the extra
// fields it carries — and a column of one word would send everybody to the API
// documentation.
func ListBaremetalLogKinds(_ *cobra.Command, args []string) {
	server := args[0]

	kinds, err := logKindsOf(server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if len(kinds) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"serviceName": server, "kinds": []any{}},
			"%s has no log kind, so there is nothing to subscribe to yet.", server)
		return
	}

	path := fmt.Sprintf(logKindsPath, url.PathEscape(server))
	ids := make([]any, len(kinds))
	for index, kind := range kinds {
		ids[index] = kind
	}

	details, err := httpLib.FetchObjectsParallel[map[string]any](path+"/%s", ids, flags.IgnoreErrors)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to read the log kinds of %s: %s", server, err)
		return
	}

	display.RenderTable(details,
		[]string{"name", "displayName", "additionalReturnedFields", "kindId"},
		&flags.OutputFormatConfig)
}

// ShowBaremetalLogURL asks for a temporary link to the logs of a server.
//
// It prints a link and says when it dies. The route is named "log/url" and that
// is exactly what it returns: a signed https://get.logs.ovh.com/search address
// into the Log Data Platform search interface, not a socket and not a stream.
// Measured on 20 August 2026, it is valid for thirty minutes — short enough
// that a command which printed it without the expiry would be handing over
// something that stops working while it is still on screen.
func ShowBaremetalLogURL(_ *cobra.Command, args []string) {
	server := args[0]

	kind, err := chooseLogKind(server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf(logURLPath, url.PathEscape(server))
	body := map[string]any{"kind": kind}
	if reportLogDryRun(http.MethodPost, endpoint, body) {
		return
	}

	var link struct {
		URL            string `json:"url"`
		ExpirationDate string `json:"expirationDate"`
	}
	if err := httpLib.Client.Post(endpoint, body, &link); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to get a log link for %s: %s", server, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{
		"serviceName":    server,
		"kind":           kind,
		"url":            link.URL,
		"expirationDate": link.ExpirationDate,
	}, "%s logs of %s, %s:\n\n%s\n\nThis link carries its own authorisation — anyone holding it can read these logs.",
		strings.ToUpper(kind[:1])+kind[1:], server, expiryPhrase(link.ExpirationDate), link.URL)
}

// ListBaremetalLogSubscriptions lists where the logs of a server go.
func ListBaremetalLogSubscriptions(_ *cobra.Command, args []string) {
	server := args[0]

	path := fmt.Sprintf(logSubscriptionsPath, url.PathEscape(server))

	// The kind filter belongs to the collection call and to nothing else. The
	// expansion below builds one URL per subscription as path + "/%s", so a
	// query string carried over would produce ".../log/subscription?kind=install/1234"
	// — a route that does not exist, for a filter that would have looked like
	// it worked.
	query := ""
	if LogKind != "" {
		query = "?kind=" + url.QueryEscape(LogKind)
	}

	ids, err := httpLib.FetchArray(path+query, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to list the log subscriptions of %s: %s", server, err)
		return
	}

	subscriptions, err := httpLib.FetchObjectsParallel[map[string]any](path+"/%s", ids, flags.IgnoreErrors)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to read the log subscriptions of %s: %s", server, err)
		return
	}

	common.RenderFilteredTable(subscriptions,
		[]string{"subscriptionId", "kind", "streamId", "serviceName", "createdAt"})
}

// ShowBaremetalLogSubscription reads one subscription.
func ShowBaremetalLogSubscription(_ *cobra.Command, args []string) {
	server, id := args[0], args[1]
	common.ManageObjectRequest(fmt.Sprintf(logSubscriptionsPath, url.PathEscape(server)), id, "")
}

// SubscribeBaremetalLogs starts sending the logs of a server to a stream.
func SubscribeBaremetalLogs(_ *cobra.Command, args []string) {
	server := args[0]

	kind, err := chooseLogKind(server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	stream, err := resolveStream(LogStream)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Disruptive rather than destructive: nothing is lost, but the logs of a
	// machine start landing somewhere new and the indexing they cause is
	// billed on the receiving service.
	if !common.ConfirmAction(common.Disruptive, server, fmt.Sprintf(
		"This sends the %s logs of %s to stream %s. Indexing them is billed on the receiving Log Data Platform service.",
		kind, server, streamLabel(stream))) {
		display.OutputError(&flags.OutputFormatConfig, "subscription of %s cancelled", server)
		return
	}

	endpoint := fmt.Sprintf(logSubscriptionsPath, url.PathEscape(server))
	body := map[string]any{"kind": kind, "streamId": stream.StreamID}
	if reportLogDryRun(http.MethodPost, endpoint, body) {
		return
	}

	var response struct {
		OperationID string `json:"operationId"`
		ServiceName string `json:"serviceName"`
	}
	if err := httpLib.Client.Post(endpoint, body, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to subscribe %s to stream %s: %s", server, stream.StreamID, err)
		return
	}

	if !LogWait {
		display.OutputInfo(&flags.OutputFormatConfig, response,
			"⚡️ The %s logs of %s are being subscribed to stream %s. Follow it with: ovhcloud baremetal logs subscription list %s",
			kind, server, stream.StreamID, server)
		return
	}

	operation, err := waitForLogOperation(response.ServiceName, response.OperationID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// The operation reporting SUCCESS is not the subscription existing. Every
	// wait in this CLI reads the state afterwards, and this one is handed the
	// identifier to read it with: the operation object carries subscriptionId.
	subscriptionID, _ := operation["subscriptionId"].(string)
	if subscriptionID == "" {
		display.OutputInfo(&flags.OutputFormatConfig, operation,
			"✅ The operation finished, but it did not say which subscription it created. List them with: ovhcloud baremetal logs subscription list %s",
			server)
		return
	}

	var subscription map[string]any
	path := fmt.Sprintf("%s/%s", endpoint, url.PathEscape(subscriptionID))
	if err := httpLib.Client.Get(path, &subscription); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"the operation finished but subscription %s cannot be read back: %s", subscriptionID, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, subscription,
		"✅ The %s logs of %s now go to stream %s (subscription %s).",
		kind, server, stream.StreamID, subscriptionID)
}

// UnsubscribeBaremetalLogs stops sending the logs of a server to a stream.
func UnsubscribeBaremetalLogs(_ *cobra.Command, args []string) {
	server, id := args[0], args[1]

	endpoint := fmt.Sprintf(logSubscriptionsPath, url.PathEscape(server))
	path := fmt.Sprintf("%s/%s", endpoint, url.PathEscape(id))

	// Read it first, so the prompt says what is about to stop rather than
	// quoting back an identifier the operator just pasted.
	var subscription map[string]any
	if err := httpLib.Client.Get(path, &subscription); err != nil {
		if common.IsNotFound(err) {
			display.OutputError(&flags.OutputFormatConfig,
				"%s has no log subscription %s.\n   List them with: ovhcloud baremetal logs subscription list %s",
				server, id, server)
			return
		}

		display.OutputError(&flags.OutputFormatConfig, "failed to read subscription %s: %s", id, err)
		return
	}

	kind, _ := subscription["kind"].(string)
	streamID, _ := subscription["streamId"].(string)
	ldpService, _ := subscription["serviceName"].(string)

	if !common.ConfirmAction(common.Disruptive, server, fmt.Sprintf(
		"This stops the %s logs of %s reaching stream %s on %s. Logs produced from now on are not collected.",
		kind, server, streamID, ldpService)) {
		display.OutputError(&flags.OutputFormatConfig, "removal of subscription %s cancelled", id)
		return
	}

	if reportLogDryRun(http.MethodDelete, path, nil) {
		return
	}

	var response struct {
		OperationID string `json:"operationId"`
		ServiceName string `json:"serviceName"`
	}
	if err := httpLib.Client.Delete(path, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to remove subscription %s: %s", id, err)
		return
	}

	if !LogWait {
		display.OutputInfo(&flags.OutputFormatConfig, response,
			"⚡️ Subscription %s of %s is being removed.", id, server)
		return
	}

	if _, err := waitForLogOperation(response.ServiceName, response.OperationID); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Same rule the other way round: the operation says it finished, the state
	// says whether the subscription is gone.
	if err := httpLib.Client.Get(path, &subscription); err == nil {
		display.OutputError(&flags.OutputFormatConfig,
			"the operation finished but subscription %s is still there; read it with: ovhcloud baremetal logs subscription get %s %s",
			id, server, id)
		return
	} else if !common.IsNotFound(err) {
		display.OutputError(&flags.OutputFormatConfig,
			"the operation finished but subscription %s cannot be checked: %s", id, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"serviceName": server, "subscriptionId": id},
		"✅ Subscription %s of %s is gone.", id, server)
}

// logKindsOf lists the kinds of log a server can send.
func logKindsOf(server string) ([]string, error) {
	var kinds []string

	path := fmt.Sprintf(logKindsPath, url.PathEscape(server))
	if err := httpLib.Client.Get(path, &kinds); err != nil {
		return nil, fmt.Errorf("failed to list the log kinds of %s: %w", server, err)
	}

	sort.Strings(kinds)

	return kinds, nil
}

// chooseLogKind settles which kind a command works on.
//
// One kind means there is nothing to choose, and saying so is better than
// making somebody type it. Several means the command must not pick for them.
// None means this server sends nothing, which is an answer and not a failure of
// the command.
func chooseLogKind(server string) (string, error) {
	if LogKind != "" {
		return LogKind, nil
	}

	kinds, err := logKindsOf(server)
	if err != nil {
		return "", err
	}

	switch len(kinds) {
	case 1:
		return kinds[0], nil

	case 0:
		return "", fmt.Errorf("%s offers no kind of log, so there is nothing to read or subscribe to", server)

	default:
		return "", fmt.Errorf("%s offers %d kinds of log; name one with --kind: %s",
			server, len(kinds), strings.Join(kinds, ", "))
	}
}

// waitForLogOperation follows a subscription change to its end.
//
// The change is made against the server on /v2 and returns an operation
// identifier, and no v2 route can follow it: it is followed on the v1 Log Data
// Platform service. The response says which one, so the caller does not have to
// work it out — the same field, serviceName, that a subscription object carries.
func waitForLogOperation(ldpService, operationID string) (map[string]any, error) {
	if ldpService == "" || operationID == "" {
		return nil, fmt.Errorf("the API accepted the change but did not say how to follow it (service %q, operation %q)",
			ldpService, operationID)
	}

	path := fmt.Sprintf("/v1/dbaas/logs/%s/operation/%s",
		url.PathEscape(ldpService), url.PathEscape(operationID))

	var last string
	for attempt := 0; attempt < logPollAttempts; attempt++ {
		var operation map[string]any
		if err := httpLib.Client.Get(path, &operation); err != nil {
			return nil, fmt.Errorf("failed to follow operation %s on %s: %w", operationID, ldpService, err)
		}

		state, _ := operation["state"].(string)
		last = state

		switch {
		case state == logOperationSucceeded:
			return operation, nil
		case logOperationFailed[state]:
			return nil, fmt.Errorf("operation %s on %s ended in %s", operationID, ldpService, state)
		}

		time.Sleep(logPollInterval)
	}

	return nil, fmt.Errorf("stopped waiting after %s; operation %s on %s still reports %q",
		time.Duration(logPollAttempts)*logPollInterval, operationID, ldpService, last)
}

// streamLabel names a stream the way the operator asked for it.
//
// When a title was resolved, the title is what they recognise and the
// identifier is what the API acted on, so the prompt shows both. When they
// pasted an identifier, repeating it twice says nothing.
func streamLabel(stream ldpStream) string {
	if stream.Title == "" {
		return stream.StreamID
	}

	return fmt.Sprintf("%q (%s on %s)", stream.Title, stream.StreamID, stream.ServiceName)
}

// expiryPhrase says how long a temporary link has left, in the words somebody
// reads rather than as a timestamp they have to subtract from now.
func expiryPhrase(expiration string) string {
	if expiration == "" {
		return "with no stated expiry"
	}

	deadline, err := time.Parse(time.RFC3339, expiration)
	if err != nil {
		return "valid until " + expiration
	}

	remaining := time.Until(deadline).Round(time.Minute)
	if remaining <= 0 {
		return "already expired (" + expiration + ")"
	}

	return fmt.Sprintf("valid for %s, until %s", remaining, expiration)
}

// reportLogDryRun previews a call with the body it would carry.
//
// common.ReportDryRun shows the method and the path, which is the whole story
// for a DELETE and half of it for these two POSTs: what an operator needs to
// check before agreeing to a subscription is which stream and which kind, and
// neither is in the URL. Printing them as a second message would mean two JSON
// documents on one stdout under -o json, so this builds the one document
// itself — the same reason `baremetal ticket` does.
func reportLogDryRun(method, endpoint string, body map[string]any) bool {
	if !flags.DryRun {
		return false
	}

	message := fmt.Sprintf("🔍 Dry run: nothing was sent. This would have been called:\n  %s %s", method, endpoint)
	details := map[string]any{
		"calls": []map[string]any{{"method": method, "endpoint": endpoint}},
	}

	if len(body) > 0 {
		rendered, err := json.MarshalIndent(body, "  ", "  ")
		if err != nil {
			// Nothing here can fail to marshal — the maps are strings — but a
			// preview that swallowed an error would be a preview that lied.
			display.OutputError(&flags.OutputFormatConfig, "failed to render the request body: %s", err)
			return true
		}

		message += fmt.Sprintf("\n\nwith:\n  %s", rendered)
		details["payload"] = body
	}

	display.OutputInfo(&flags.OutputFormatConfig, details, "%s", message)

	return true
}
