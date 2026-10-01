// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

const (
	authorizationCheckPath = "/v2/iam/authorization/check"
	referenceActionPath    = "/v2/iam/reference/action"
	referenceTypePath      = "/v2/iam/reference/resource/type"
)

var (
	// CheckResources are the URNs the actions are checked against.
	CheckResources []string

	// ActionResourceType narrows the action reference to one product.
	ActionResourceType string

	// ActionCategory narrows it to READ, EDIT, DELETE and the rest.
	ActionCategory string

	// ActionSearch keeps the actions whose name or description contains it.
	ActionSearch string
)

// referenceTooManyToList is the point past which the reference refuses to
// print itself. The account measured returns 9158 actions across 117 product
// families; dedicatedServer alone has 183.
const referenceTooManyToList = 400

type referenceAction struct {
	Action             string   `json:"action"`
	Description        string   `json:"description"`
	ResourceType       string   `json:"resourceType"`
	Categories         []string `json:"categories"`
	HasQueryParameters bool     `json:"hasQueryParameters"`
}

type checkVerdict struct {
	ResourceURN         string   `json:"resourceURN"`
	AuthorizedActions   []string `json:"authorizedActions"`
	UnauthorizedActions []string `json:"unauthorizedActions"`
}

// CheckAuthorization answers "why can this credential not do that", which is
// the question BM-06 leaves an operator with: a restricted account shows no
// server and says nothing about why.
func CheckAuthorization(_ *cobra.Command, args []string) {
	if len(CheckResources) == 0 {
		display.OutputError(&flags.OutputFormatConfig,
			"no resource to check against; give at least one --on <urn>.\n   Find one with: ovhcloud iam resource list")
		return
	}

	var verdicts []checkVerdict
	err := httpLib.Client.Post(authorizationCheckPath, map[string]any{
		"actions":      args,
		"resourceURNs": CheckResources,
	}, &verdicts)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", explainCheckFailure(err, args))
		return
	}

	rows := make([]map[string]any, 0, len(verdicts)*len(args))
	for _, verdict := range verdicts {
		for _, action := range verdict.AuthorizedActions {
			rows = append(rows, checkRow(verdict.ResourceURN, action, true))
		}
		for _, action := range verdict.UnauthorizedActions {
			rows = append(rows, checkRow(verdict.ResourceURN, action, false))
		}
	}

	display.RenderTable(rows, []string{"resource", "action", "allowed"}, &flags.OutputFormatConfig)
}

func checkRow(urn, action string, allowed bool) map[string]any {
	verdict := "no"
	if allowed {
		verdict = "yes"
	}

	return map[string]any{"resource": shortURN(urn), "action": action, "allowed": verdict}
}

// shortURN keeps the part of a URN that identifies the thing. A full URN is
// "urn:v1:eu:resource:dedicatedServer:ns1.example" and repeating the first four
// segments on every row buys nothing.
func shortURN(urn string) string {
	parts := strings.Split(urn, ":")
	if len(parts) >= 6 {
		return strings.Join(parts[4:], ":")
	}

	return urn
}

// explainCheckFailure turns the one refusal that is not about permissions into
// something actionable. The API validates the action names first, and a single
// unknown one fails the whole batch — so a typo in the fourth action reads as
// if none of them could be checked.
func explainCheckFailure(err error, actions []string) string {
	message := err.Error()
	if !strings.Contains(message, "Unknown action") {
		return fmt.Sprintf("failed to check authorization: %s", message)
	}

	return fmt.Sprintf("%s\n   One unknown action fails the whole check, so the other %d were not answered either.\n   List what exists with: ovhcloud iam reference actions --type <resource type>",
		message, len(actions)-1)
}

// ListReferenceActions lists what can be granted on a product.
func ListReferenceActions(_ *cobra.Command, _ []string) {
	query := url.Values{}
	if ActionResourceType != "" {
		query.Set("resourceType", ActionResourceType)
	}

	path := referenceActionPath
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var actions []referenceAction
	if err := httpLib.Client.Get(path, &actions); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to read the action reference: %s", err)
		return
	}

	kept := make([]referenceAction, 0, len(actions))
	for _, action := range actions {
		if ActionCategory != "" && !hasCategory(action, ActionCategory) {
			continue
		}
		if ActionSearch != "" && !matchesSearch(action, ActionSearch) {
			continue
		}
		kept = append(kept, action)
	}

	// Refusing beats printing 9158 rows nobody can read. It says how to narrow
	// rather than how many exist, because the count is not the useful half.
	if len(kept) > referenceTooManyToList {
		display.OutputError(&flags.OutputFormatConfig,
			"%d actions match, which is more than a table can say anything with.\n   Narrow with --type <resource type>, --category, or --search.\n   The types are listed by: ovhcloud iam reference resource-types",
			len(kept))
		return
	}

	sort.Slice(kept, func(i, j int) bool { return kept[i].Action < kept[j].Action })

	rows := make([]map[string]any, 0, len(kept))
	for _, action := range kept {
		rows = append(rows, map[string]any{
			"action":      action.Action,
			"categories":  strings.Join(action.Categories, ","),
			"description": action.Description,
		})
	}

	common.RenderFilteredTable(rows, []string{"action", "categories", "description"})
}

func hasCategory(action referenceAction, wanted string) bool {
	for _, category := range action.Categories {
		if strings.EqualFold(category, wanted) {
			return true
		}
	}

	return false
}

func matchesSearch(action referenceAction, needle string) bool {
	lowered := strings.ToLower(needle)

	return strings.Contains(strings.ToLower(action.Action), lowered) ||
		strings.Contains(strings.ToLower(action.Description), lowered)
}

// ListReferenceResourceTypes lists the product families actions are grouped by.
func ListReferenceResourceTypes(_ *cobra.Command, _ []string) {
	var types []string
	if err := httpLib.Client.Get(referenceTypePath, &types); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to read the resource types: %s", err)
		return
	}

	sort.Strings(types)

	rows := make([]map[string]any, 0, len(types))
	for _, name := range types {
		rows = append(rows, map[string]any{"resourceType": name})
	}

	common.RenderFilteredTable(rows, []string{"resourceType"})
}

// CompleteResourceType offers the product families on <tab>.
func CompleteResourceType(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	var types []string
	if err := httpLib.Client.Get(referenceTypePath, &types); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return types, cobra.ShellCompDirectiveNoFileComp
}
