// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"sync"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

const confirmTerminationPath = "/dedicated/server/{serviceName}/confirmTermination"

// Flags of `confirm-termination`. The reason and the intent are what the
// termination survey asks for on the web side; sending them from the CLI is the
// difference between ending a contract and ending it with a reason attached.
var (
	TerminationReason    string
	TerminationFutureUse string
	TerminationComment   string
)

// The accepted values live in the specification embedded in this binary, so
// they are read rather than transcribed: `reason` carries fourteen values
// today, and a list copied into Go drifts the day the API gains a fifteenth —
// silently, into a 400 nobody can explain.
//
// Reading it costs about 50 ms, because the whole specification is loaded and
// validated. That is why it is behind a sync.OnceValues and never called while
// flags are being registered: every invocation of the CLI registers them, and
// only a completion or an actual termination needs the values.
//
// sync.OnceValues caches a failure as firmly as a success, which is what we
// want: the input is compiled into the binary by //go:embed and cannot change
// while the process runs, so a second attempt would read the same bytes and
// fail the same way.
var (
	terminationReasons = sync.OnceValues(func() ([]string, error) {
		return openapi.GetRequestFieldEnum(assets.BaremetalOpenapiSchema, confirmTerminationPath, "post", "reason")
	})
	terminationFutureUses = sync.OnceValues(func() ([]string, error) {
		return openapi.GetRequestFieldEnum(assets.BaremetalOpenapiSchema, confirmTerminationPath, "post", "futureUse")
	})
)

// CompleteTerminationReason and CompleteTerminationFutureUse offer the accepted
// values on <tab>. They are the discoverable half of the pair: the flag help
// cannot list fourteen values without becoming unreadable, and it must not read
// the specification to find them.
func CompleteTerminationReason(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return completeEnum(terminationReasons)
}

func CompleteTerminationFutureUse(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return completeEnum(terminationFutureUses)
}

func completeEnum(read func() ([]string, error)) ([]string, cobra.ShellCompDirective) {
	values, err := read()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return values, cobra.ShellCompDirectiveNoFileComp
}

// checkEnumFlag rejects a value the API would reject, and names the ones it
// would accept. A 400 from the other side of the network says "invalid value"
// and stops there.
func checkEnumFlag(name, value string, read func() ([]string, error)) error {
	if value == "" {
		return nil
	}

	accepted, err := read()
	if err != nil {
		return fmt.Errorf("failed to read the values accepted for --%s: %w", name, err)
	}

	if slices.Contains(accepted, value) {
		return nil
	}

	return fmt.Errorf("--%s does not accept %q; accepted values are: %s",
		name, value, strings.Join(accepted, ", "))
}

func GetBaremetalServiceInfo(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/serviceInfos", url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(endpoint, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching service information for %s: %s", args[0], err)
		return
	}

	display.OutputObject(object, args[0], common.ServiceInfoTemplate, &flags.OutputFormatConfig)
}

func EditBaremetalServiceInfo(cmd *cobra.Command, args []string) {
	renewPayload := common.ServiceInfoRenewPayload(cmd)

	if err := common.EditResource(
		cmd,
		"/dedicated/server/{serviceName}/serviceInfos",
		fmt.Sprintf("/v1/dedicated/server/%s/serviceInfos", url.PathEscape(args[0])),
		renewPayload,
		assets.BaremetalOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

// TerminateBaremetal asks for the termination of a server. It stops nothing:
// the API sends a token to the administrative contact by email, and the server
// keeps running until that token is confirmed.
//
// The guard is therefore the interrupting kind rather than the destroying kind.
// Making the operator type the server name here would spend the strong
// confirmation on the reversible half of the pair, and teach them to reach for
// --yes before the half that is not reversible.
func TerminateBaremetal(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/terminate", url.PathEscape(args[0]))

	if !common.ConfirmAction(common.Disruptive, args[0], fmt.Sprintf(
		"This asks for the termination of %s. Nothing stops now: a termination token is emailed to the administrative contact, and the server runs until that token is confirmed.",
		args[0])) {
		display.OutputError(&flags.OutputFormatConfig, "termination request for %s cancelled", args[0])
		return
	}

	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	// The response is a free-form string, and the token is not in it — the API
	// mails it. Printing the body as though it were the token is how an
	// operator ends up pasting a sentence into --token.
	var response string
	if err := httpLib.Client.Post(endpoint, nil, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error requesting termination of %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"service": args[0], "response": response},
		"⚡️ Termination of %s requested. A token has been emailed to the administrative contact; confirm with:\n  ovhcloud baremetal confirm-termination %s <token>",
		args[0], args[0])
}

// ConfirmBaremetalTermination is the irreversible half: it ends the contract and
// the server goes back to OVHcloud at expiry. Hence the strongest guard the CLI
// has, the server's own name typed by hand.
//
// The token is not that guard. It proves the request reached the administrative
// contact's mailbox, which says nothing about whether this is the server they
// meant.
func ConfirmBaremetalTermination(_ *cobra.Command, args []string) {
	if err := checkEnumFlag("reason", TerminationReason, terminationReasons); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	if err := checkEnumFlag("future-use", TerminationFutureUse, terminationFutureUses); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// cobra counts the arguments, it does not look at them: an empty second
	// argument satisfies ExactArgs(2) and would travel all the way to a 400.
	token := strings.TrimSpace(args[1])
	if token == "" {
		display.OutputError(&flags.OutputFormatConfig,
			"no termination token given; it is the token emailed to the administrative contact by `baremetal terminate %s`", args[0])
		return
	}

	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/confirmTermination", url.PathEscape(args[0]))

	if !common.ConfirmAction(common.Destructive, args[0], fmt.Sprintf(
		"This confirms the termination of %s. The contract ends and the server is returned to OVHcloud at expiry; there is no undoing it from here.",
		args[0])) {
		display.OutputError(&flags.OutputFormatConfig, "termination of %s not confirmed", args[0])
		return
	}

	body := map[string]any{"token": token}
	if TerminationReason != "" {
		body["reason"] = TerminationReason
	}
	if TerminationFutureUse != "" {
		body["futureUse"] = TerminationFutureUse
	}
	if TerminationComment != "" {
		body["commentary"] = TerminationComment
	}

	// The preview names every field it would send and withholds one value: the
	// token is a single-use credential, and a --dry-run is the one command an
	// operator runs with the output on screen or in a pipeline log.
	if common.ReportDryRun(common.Call{
		Method:   "POST",
		Endpoint: fmt.Sprintf("%s  (%s)", endpoint, describeTerminationBody(body)),
	}) {
		return
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error confirming termination of %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"service": args[0]},
		"✅ Termination of %s confirmed", args[0])
}

// fingerprint identifies a token without reproducing it. Withholding it
// entirely was the first answer, and it costs the one thing an operator
// legitimately checks in a preview: that the shell handed over the token they
// pasted, rather than one it truncated or expanded. Four characters and a
// length settle that question and reconstruct nothing.
func fingerprint(token string) string {
	if len(token) < 8 {
		return fmt.Sprintf("(%d characters, too short to show)", len(token))
	}
	return fmt.Sprintf("%s… (%d characters)", token[:4], len(token))
}

func describeTerminationBody(body map[string]any) string {
	fields := make([]string, 0, len(body))
	for _, name := range []string{"token", "reason", "futureUse", "commentary"} {
		value, set := body[name]
		if !set {
			continue
		}
		if name == "token" {
			fields = append(fields, fmt.Sprintf("token: %s", fingerprint(fmt.Sprint(value))))
			continue
		}
		fields = append(fields, fmt.Sprintf("%s: %v", name, value))
	}
	return strings.Join(fields, ", ")
}
