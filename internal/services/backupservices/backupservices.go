// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package backupservices

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

// The Veeam Backup service — tenants, storage vaults, the Service Provider
// Console that drives the agents — had no command at all. Its seventeen paths
// live in the v2 catalogue, which this repository could not fetch a schema for
// until #262, so the surface was invisible rather than skipped.
//
// Every operation of this API is badged "Beta version" upstream.

// EditSpec carries what the two editable resources accept: a name, and nothing
// else. The API's own TargetSpec for a vault and for a VSPC tenant has exactly
// one property.
var EditSpec struct {
	Name string `json:"name,omitempty"`
}

// ListTenants shows the backup tenants of the account.
func ListTenants(_ *cobra.Command, _ []string) {
	var tenants []resource
	if err := httpLib.Client.Get(TenantsPath, &tenants); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to list the backup tenants: %s", err)
		return
	}

	common.RenderFilteredTable(rowsOf(tenants, func(r resource) map[string]any {
		state := r.CurrentState
		return map[string]any{
			"vaults":      countOf(state["vaults"]),
			"vspcTenants": countOf(state["vspcTenants"]),
		}
	}), []string{"id", "name", "resourceStatus status", "vaults", "vspcTenants", "tasks"})
}

// ShowTenant reads one backup tenant, resolved when the account has only one.
func ShowTenant(_ *cobra.Command, args []string) {
	tenant, err := tenantFromArgs(args)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(TenantsPath, tenant, "")
}

// ListVaults shows the storage vaults of a tenant.
func ListVaults(_ *cobra.Command, _ []string) {
	tenant, err := ResolveTenant()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var vaults []resource
	path := fmt.Sprintf("%s/%s/vault", TenantsPath, url.PathEscape(tenant))
	if err := httpLib.Client.Get(path, &vaults); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to list the vaults of %s: %s", tenant, err)
		return
	}

	// allowedIps is deliberately absent. It looks like a field of a vault and it
	// is not: the addresses allowed to reach one are listed in the VSPC
	// tenant's own view of its vaults, and reading them off the vault resource
	// answers 0 on an account where the real number is 9 — a column that is
	// wrong reads worse than a column that is missing.
	common.RenderFilteredTable(rowsOf(vaults, func(r resource) map[string]any {
		return map[string]any{
			"buckets":     countOf(r.CurrentState["buckets"]),
			"vspcTenants": countOf(r.CurrentState["vspcTenants"]),
			"type":        r.CurrentState["type"],
			"regions":     strings.Join(bucketRegions(r), ", "),
		}
	}), []string{"id", "name", "resourceStatus status", "regions", "type", "buckets", "vspcTenants", "tasks"})
}

// ShowVault reads one vault.
func ShowVault(_ *cobra.Command, args []string) {
	tenant, err := ResolveTenant()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("%s/%s/vault", TenantsPath, url.PathEscape(tenant)), args[0], "")
}

// ListBuckets shows the buckets behind a vault.
//
// There is no command to read one bucket on its own: the collection already
// answers with the whole object — identifier, name, region, performance, role
// and status — so a get would repeat a row of the table it was read from.
func ListBuckets(_ *cobra.Command, args []string) {
	tenant, err := ResolveTenant()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var buckets []map[string]any
	path := VaultPath(tenant, args[0]) + "/bucket"
	if err := httpLib.Client.Get(path, &buckets); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to list the buckets of vault %s: %s", args[0], err)
		return
	}

	common.RenderFilteredTable(buckets, []string{"id", "name", "region", "performance", "role", "status"})
}

// EditVault renames a vault.
func EditVault(cmd *cobra.Command, args []string) {
	tenant, err := ResolveTenant()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	editNamedResource(cmd, VaultPath(tenant, args[0]), "vault "+args[0])
}

// ListVspc shows the Service Provider Console tenants.
func ListVspc(_ *cobra.Command, _ []string) {
	tenant, err := ResolveTenant()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	vspcs, err := listVspc(tenant)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.RenderFilteredTable(rowsOf(vspcs, func(r resource) map[string]any {
		state := r.CurrentState
		addons, _ := state["enabledAddons"].([]any)
		names := make([]string, 0, len(addons))
		for _, addon := range addons {
			names = append(names, fmt.Sprint(addon))
		}

		return map[string]any{
			"vspcType":  state["vspcType"],
			"region":    state["region"],
			"accessUrl": state["accessUrl"],
			"addons":    strings.Join(names, ", "),
			"agents":    countOf(state["backupAgents"]),
		}
	}), []string{"id", "name", "resourceStatus status", "vspcType type", "region", "agents", "addons", "tasks"})
}

// ShowVspc reads one VSPC tenant.
func ShowVspc(_ *cobra.Command, args []string) {
	tenant, err := ResolveTenant()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	vspc := Vspc
	if len(args) > 0 {
		vspc = args[0]
	}
	if vspc == "" {
		if vspc, err = ResolveVspc(tenant); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}
	}

	common.ManageObjectRequest(fmt.Sprintf("%s/%s/vspc", TenantsPath, url.PathEscape(tenant)), vspc, "")
}

// EditVspc renames a VSPC tenant.
func EditVspc(cmd *cobra.Command, args []string) {
	tenant, err := ResolveTenant()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	editNamedResource(cmd, fmt.Sprintf("%s/%s/vspc/%s", TenantsPath, url.PathEscape(tenant), url.PathEscape(args[0])),
		"VSPC tenant "+args[0])
}

// ListPolicies shows the retention policies an agent can be put on.
//
// It is a list of names and nothing else — the API answers with strings — and
// those names are what `baremetal backup-agent edit --policy` takes.
func ListPolicies(_ *cobra.Command, _ []string) {
	tenant, vspc, err := ResolveBoth()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	policies, err := PoliciesOf(tenant, vspc)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if len(policies) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"policies": []any{}},
			"This VSPC tenant defines no retention policy, so an agent has nothing to be put on.")
		return
	}

	rows := make([]map[string]any, 0, len(policies))
	for _, policy := range policies {
		rows = append(rows, map[string]any{"policy": policy})
	}

	common.RenderFilteredTable(rows, []string{"policy"})
}

// PoliciesOf lists the retention policies of a VSPC tenant.
func PoliciesOf(tenant, vspc string) ([]string, error) {
	var policies []string

	path := VspcPath(tenant, vspc) + "/backupPolicies"
	if err := httpLib.Client.Get(path, &policies); err != nil {
		return nil, fmt.Errorf("failed to list the retention policies: %w", err)
	}

	sort.Strings(policies)

	return policies, nil
}

// ShowDeployScript prints what puts the agent on a machine.
//
// This is the command that makes an agent installable from a terminal, and it
// is also the reason the whole lot is worth having: an agent created through
// the API is NOT_INSTALLED until this script has run on the machine.
//
// The links are pre-signed S3 URLs. They carry their own authorisation and, on
// this account, an X-Amz-Expires of seven days — so they are printed, because
// they are the answer, and what they are is said beside them.
func ShowDeployScript(_ *cobra.Command, _ []string) {
	tenant, vspc, err := ResolveBoth()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var agent struct {
		LinuxDeployScript string `json:"linuxDeployScript"`
		LinuxURL          string `json:"linuxUrl"`
		MacURL            string `json:"macUrl"`
		WindowsURL        string `json:"windowsUrl"`
	}

	path := VspcPath(tenant, vspc) + "/managementAgent"
	if err := httpLib.Client.Get(path, &agent); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to read the management agent: %s", err)
		return
	}

	message := fmt.Sprintf("Run this on the machine to install the backup agent:\n\n  %s\n\n"+
		"Windows: %s\nmacOS:   %s\n\n"+
		"These links carry their own authorisation — anyone holding one can install against this tenant.",
		agent.LinuxDeployScript, agent.WindowsURL, agent.MacURL)

	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{
		"linuxDeployScript": agent.LinuxDeployScript,
		"linuxUrl":          agent.LinuxURL,
		"macUrl":            agent.MacURL,
		"windowsUrl":        agent.WindowsURL,
	}, "%s", message)
}

// ListLicenses shows the Veeam licences held by a VSPC tenant.
func ListLicenses(_ *cobra.Command, _ []string) {
	tenant, vspc, err := ResolveBoth()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var licenses []resource
	path := VspcPath(tenant, vspc) + "/backupLicenses"
	if err := httpLib.Client.Get(path, &licenses); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to list the backup licences: %s", err)
		return
	}

	if len(licenses) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"licenses": []any{}},
			"This VSPC tenant holds no Veeam licence, so it drives no backup server.")
		return
	}

	common.RenderFilteredTable(rowsOf(licenses, nil),
		[]string{"id", "name", "resourceStatus status", "tasks"})
}

// ListLicenseServers shows the backup servers driven by one licence.
func ListLicenseServers(_ *cobra.Command, args []string) {
	tenant, vspc, err := ResolveBoth()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var servers []resource
	path := fmt.Sprintf("%s/backupLicenses/%s/backupServer", VspcPath(tenant, vspc), url.PathEscape(args[0]))
	if err := httpLib.Client.Get(path, &servers); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to list the backup servers of licence %s: %s", args[0], err)
		return
	}

	common.RenderFilteredTable(rowsOf(servers, func(r resource) map[string]any {
		return map[string]any{
			"licenseType": r.TargetSpec["licenseType"],
			"displayName": r.TargetSpec["displayName"],
		}
		// "resourceStatus status" and not "status": rowsOf writes the value under
		// resourceStatus, and the second word is only the header. Asking for
		// "status" asked for a key no row has, so the column was empty on every
		// line — and an empty status column reads as "nothing to report" on
		// exactly the resources somebody is checking on.
	}), []string{"id", "displayName", "licenseType", "resourceStatus status", "tasks"})
}

// editNamedResource sends the one field these two resources accept.
func editNamedResource(cmd *cobra.Command, path, label string) {
	if strings.TrimSpace(EditSpec.Name) == "" {
		display.OutputError(&flags.OutputFormatConfig,
			"--name is what this command changes, and a blank one is not a name")
		return
	}

	if !common.ConfirmAction(common.Disruptive, label,
		fmt.Sprintf("This renames %s to %q.", label, EditSpec.Name)) {
		display.OutputError(&flags.OutputFormatConfig, "rename of %s cancelled", label)
		return
	}

	if common.ReportDryRun(common.Call{Method: "PUT", Endpoint: path}) {
		return
	}

	if err := httpLib.Client.Put(path, map[string]any{"name": EditSpec.Name}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to rename %s: %s", label, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"name": EditSpec.Name}, "✅ %s is now named %q.", label, EditSpec.Name)
}

// rowsOf flattens the resource shape into what a table can hold.
//
// Every object of this API answers with the same four parts, and three of them
// are nested: the name lives in the spec, the counts live in the current state,
// and what is happening lives in a list of tasks. A table that showed the raw
// object would show four columns of JSON.
func rowsOf(resources []resource, extra func(resource) map[string]any) []map[string]any {
	rows := make([]map[string]any, 0, len(resources))
	for _, r := range resources {
		row := map[string]any{
			"id":             r.ID,
			"name":           r.name(),
			"resourceStatus": statusOf(r),
			"tasks":          taskSummary(r.CurrentTasks),
			"createdAt":      r.CreatedAt,
			"updatedAt":      r.UpdatedAt,
		}
		if extra != nil {
			for key, value := range extra(r) {
				row[key] = value
			}
		}
		rows = append(rows, row)
	}

	return rows
}

// statusOf reads whichever status field the resource carries.
//
// Most of them answer with resourceStatus, but an agent and a backup server
// answer with status instead — same idea, different field, different
// enumeration. Reading only one of the two would print an empty column on
// exactly the resources somebody is checking on.
func statusOf(r resource) string {
	if r.ResourceStatus != "" {
		return r.ResourceStatus
	}
	if status, ok := r.CurrentState["status"].(string); ok {
		return status
	}

	return ""
}

// taskSummary says what is happening to a resource, in one cell.
//
// This generation of the API has no task route to poll: a resource carries the
// tasks running on it, and that list is the only place a failure is reported.
// An empty list is the normal case and reads as a dash rather than as nothing.
func taskSummary(tasks []currentTask) string {
	if len(tasks) == 0 {
		return "—"
	}

	parts := make([]string, 0, len(tasks))
	for _, task := range tasks {
		// The type is what says which operation is stuck; the status alone
		// answers "ERROR" to the question "error doing what". Measured on this
		// account: three BACKUP_VAULT_CREATE in ERROR and a VSPC_AGENT_UPDATE
		// beside them, none of which the status column could have named.
		part := task.Status
		if task.Type != "" {
			part = task.Type + " " + task.Status
		}
		// The schema declares errors on a task, and the API leaves the list
		// empty even on a failed one, so this is printed when it is there and
		// nothing pretends it will be.
		if len(task.Errors) > 0 {
			part += ": " + task.Errors[0].Message
		}
		parts = append(parts, part)
	}

	return strings.Join(parts, ", ")
}

func countOf(value any) int {
	list, _ := value.([]any)
	return len(list)
}

// bucketRegions says where a vault actually stores, which is the thing that
// distinguishes three vaults whose names are otherwise generated.
func bucketRegions(r resource) []string {
	buckets, _ := r.CurrentState["buckets"].([]any)

	seen := make(map[string]bool)
	var regions []string
	for _, item := range buckets {
		bucket, ok := item.(map[string]any)
		if !ok {
			continue
		}
		region, _ := bucket["region"].(string)
		if region == "" || seen[region] {
			continue
		}
		seen[region] = true
		regions = append(regions, region)
	}
	sort.Strings(regions)

	return regions
}

func tenantFromArgs(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	return ResolveTenant()
}

// ListAgents shows every backup agent of the VSPC tenant, whatever it protects.
//
// `baremetal backup-agent show` answers for one machine, which is the question
// an operator asks about a server. This is the other question — what is the
// backup posture of the estate — and it is the one that made the state of this
// account visible: nine agents provisioned, none deployed, none on a policy.
func ListAgents(_ *cobra.Command, _ []string) {
	tenant, vspc, err := ResolveBoth()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var agents []struct {
		ID           string         `json:"id"`
		Status       string         `json:"status"`
		TargetSpec   map[string]any `json:"targetSpec"`
		CurrentState map[string]any `json:"currentState"`
		CurrentTasks []currentTask  `json:"currentTasks"`
		CreatedAt    string         `json:"createdAt"`
	}

	path := VspcPath(tenant, vspc) + "/backupAgent"
	if err := httpLib.Client.Get(path, &agents); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to list the backup agents: %s", err)
		return
	}

	if len(agents) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"agents": []any{}},
			"This VSPC tenant has no backup agent.\n"+
				"   Create one for a dedicated server with: ovhcloud baremetal backup-agent create <service_name>")
		return
	}

	rows := make([]map[string]any, 0, len(agents))
	for _, agent := range agents {
		policy, _ := agent.CurrentState["policy"].(string)
		if policy == "" {
			policy = "none"
		}

		rows = append(rows, map[string]any{
			"id":       agent.ID,
			"name":     agent.TargetSpec["displayName"],
			"protects": agent.CurrentState["productResourceName"],
			"type":     agent.CurrentState["type"],
			"status":   agent.Status,
			"policy":   policy,
			"ips":      agent.CurrentState["ips"],
			"tasks":    taskSummary(agent.CurrentTasks),
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return fmt.Sprint(rows[i]["protects"]) < fmt.Sprint(rows[j]["protects"])
	})

	common.RenderFilteredTable(rows,
		[]string{"protects", "status", "policy", "type", "ips", "id", "tasks"})
}
