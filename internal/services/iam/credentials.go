// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"fmt"
	"net/url"
	"sort"
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

// The credentials live in the /me catalogue while the rest of IAM is v2. They
// are placed under `iam` all the same: Scaleway, AWS and gcloud all keep API
// keys under their identity command, and an operator looking for "what can this
// key do" looks where the policies are, not under an account tree.
const (
	credentialsPath  = "/v1/me/api/credential"
	applicationsPath = "/v1/me/api/application"
)

var (
	// CredentialStatus filters on api.CredentialStateEnum.
	CredentialStatus string

	// CredentialApplication keeps the keys of one application.
	CredentialApplication int64

	// CredentialUnusedOnly keeps the keys that have never been used. 17 of the
	// 66 on the account measured never were.
	CredentialUnusedOnly bool
)

var credentialColumns = []string{"credentialId", "status", "applicationId", "scope", "lastUse", "expiration", "restrictedTo"}

var applicationColumns = []string{"applicationId", "name", "status", "description"}

// credentialStates reads api.CredentialStateEnum from the embedded schema.
var credentialStates = sync.OnceValues(func() ([]string, error) {
	return openapi.GetComponentEnum(assets.MeOpenapiSchema, "auth.CredentialStateEnum")
})

// CompleteCredentialStatus offers the states on <tab>.
func CompleteCredentialStatus(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return common.CompleteEnum(credentialStates)
}

type credentialRule struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type credential struct {
	CredentialID  int64            `json:"credentialId"`
	ApplicationID int64            `json:"applicationId"`
	Status        string           `json:"status"`
	Creation      string           `json:"creation"`
	LastUse       *string          `json:"lastUse"`
	Expiration    *string          `json:"expiration"`
	OvhSupport    bool             `json:"ovhSupport"`
	AllowedIPs    []string         `json:"allowedIPs"`
	Rules         []credentialRule `json:"rules"`
}

// scope says in one cell how far a key reaches. A key holding a rule on "/*"
// or "*" can call the whole API with that verb, and 139 of the rules on the
// account measured are exactly that — so this is the column that matters, not
// a rule count.
func (c credential) scope() string {
	if len(c.Rules) == 0 {
		return "nothing"
	}

	whole := make([]string, 0, len(c.Rules))
	for _, rule := range c.Rules {
		if rule.Path == "/*" || rule.Path == "*" {
			whole = append(whole, rule.Method)
		}
	}

	if len(whole) == 0 {
		return fmt.Sprintf("%d path(s)", len(c.Rules))
	}

	sort.Strings(whole)
	if len(whole) == len(c.Rules) {
		return "whole API: " + strings.Join(whole, ",")
	}

	return fmt.Sprintf("whole API: %s, and %d more", strings.Join(whole, ","), len(c.Rules)-len(whole))
}

// restriction says whether anything narrows where the key may be used from.
// Two of the 66 keys measured were restricted by IP; the other 64 work from
// anywhere.
func (c credential) restriction() string {
	if len(c.AllowedIPs) == 0 {
		return "anywhere"
	}

	return strings.Join(c.AllowedIPs, ", ")
}

// never is what a nil timestamp means on these two fields, and it is not the
// same as "unknown": the API sends null for a key that has never been used and
// for one that never expires.
func never(value *string, absent string) string {
	if value == nil || *value == "" {
		return absent
	}

	return *value
}

// ListCredentials lists the API keys of the account.
func ListCredentials(_ *cobra.Command, _ []string) {
	query := url.Values{}
	if CredentialStatus != "" {
		if err := common.CheckEnumFlag("status", CredentialStatus, credentialStates); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}
		query.Set("status", CredentialStatus)
	}
	if CredentialApplication != 0 {
		query.Set("applicationId", fmt.Sprint(CredentialApplication))
	}

	path := credentialsPath
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	ids, err := httpLib.FetchArray(path, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to list API credentials: %s", err)
		return
	}

	// The filter must ride on the collection call only: FetchObjectsParallel
	// builds each object URL as path + "/%s", so a query carried over would
	// ask for ".../credential?status=expired/1234".
	credentials, err := httpLib.FetchObjectsParallel[credential](credentialsPath+"/%s", ids, flags.IgnoreErrors)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to read API credentials: %s", err)
		return
	}

	rows := make([]map[string]any, 0, len(credentials))
	for _, item := range credentials {
		if CredentialUnusedOnly && item.LastUse != nil && *item.LastUse != "" {
			continue
		}

		rows = append(rows, map[string]any{
			"credentialId":  item.CredentialID,
			"applicationId": item.ApplicationID,
			"status":        item.Status,
			"scope":         item.scope(),
			"lastUse":       never(item.LastUse, "never used"),
			"expiration":    never(item.Expiration, "never expires"),
			"restrictedTo":  item.restriction(),
		})
	}

	display.RenderTable(rows, credentialColumns, &flags.OutputFormatConfig)
}

// GetCredential shows one API key, with its rules spelled out.
func GetCredential(_ *cobra.Command, args []string) {
	items, err := httpLib.FetchObjectsParallel[map[string]any](
		credentialsPath+"/%s", []any{args[0]}, false)
	if err != nil || len(items) == 0 {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to read API credential %s: %s\n   List them with: ovhcloud iam credential list", args[0], err)
		return
	}

	display.OutputObject(items[0], args[0], credentialTemplate, &flags.OutputFormatConfig)
}

// DeleteCredential revokes an API key.
func DeleteCredential(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("%s/%s", credentialsPath, url.PathEscape(args[0]))

	if common.ReportDryRun(common.Call{Method: "DELETE", Endpoint: endpoint}) {
		return
	}

	// Revoking a key breaks whatever holds it, immediately and without a way
	// back: a new key has to be created and re-authorised by the customer.
	if !common.ConfirmAction(common.Destructive, fmt.Sprintf("API credential %s", args[0]),
		"anything still using this key stops working at once, and the key cannot be restored") {
		display.OutputError(&flags.OutputFormatConfig, "aborted")
		return
	}

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to revoke API credential %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ API credential %s is revoked.", args[0])
}

// ListApplications lists the applications keys are issued against.
func ListApplications(_ *cobra.Command, _ []string) {
	common.ManageListRequest(applicationsPath, "", applicationColumns, flags.GenericFilters)
}

// GetApplication shows one application.
func GetApplication(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(applicationsPath, args[0], applicationTemplate)
}

// DeleteApplication removes an application and, with it, every key issued
// against it.
func DeleteApplication(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("%s/%s", applicationsPath, url.PathEscape(args[0]))

	if common.ReportDryRun(common.Call{Method: "DELETE", Endpoint: endpoint}) {
		return
	}

	keys, err := keysOfApplication(args[0])
	warning := "every API key issued against this application stops working"
	if err == nil {
		warning = fmt.Sprintf("%d API key(s) issued against this application stop working", keys)
	}

	if !common.ConfirmAction(common.Destructive, fmt.Sprintf("application %s", args[0]), warning) {
		display.OutputError(&flags.OutputFormatConfig, "aborted")
		return
	}

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete application %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Application %s is gone.", args[0])
}

// keysOfApplication counts what a deletion would take down with it. The count
// is read rather than guessed, and a failure to read it is not a failure to
// delete: the prompt falls back to saying it without a number.
func keysOfApplication(application string) (int, error) {
	ids, err := httpLib.FetchArray(
		fmt.Sprintf("%s?applicationId=%s", credentialsPath, url.QueryEscape(application)), "")
	if err != nil {
		return 0, err
	}

	return len(ids), nil
}

// SetUserState enables or disables an IAM user. Two routes, one command shape,
// because "enable" and "disable" answer the same question in two directions.
func SetUserState(enable bool) func(*cobra.Command, []string) {
	verb := "disable"
	if enable {
		verb = "enable"
	}

	return func(_ *cobra.Command, args []string) {
		endpoint := fmt.Sprintf("/v1/me/identity/user/%s/%s", url.PathEscape(args[0]), verb)

		if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
			return
		}

		// Disabling cuts a person's access to the account. It interrupts, it
		// does not destroy: the user and their grants survive.
		if !enable && !common.ConfirmAction(common.Disruptive, fmt.Sprintf("user %s", args[0]),
			"they lose access to the account until the user is enabled again") {
			display.OutputError(&flags.OutputFormatConfig, "aborted")
			return
		}

		if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to %s user %s: %s", verb, args[0], err)
			return
		}

		display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ User %s is now %sd.", args[0], verb)
	}
}
