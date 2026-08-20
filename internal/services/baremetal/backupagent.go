// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
	"github.com/ovh/ovhcloud-cli/internal/services/backupservices"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

// A backup agent belongs to a VSPC tenant, which belongs to a backup tenant,
// and it is reached by two UUIDs nobody knows. But it backs up one machine, and
// it says so: backup.AgentProductTypeEnum lists OVHCLOUD_BAREMETAL as a
// first-class product type, and the agent carries the server it protects in
// productResourceName.
//
// So these commands start where an operator starts, from a server name, and the
// hierarchy above is resolved. Measured on 20 August 2026, all nine agents of
// this account are OVHCLOUD_BAREMETAL, named agent-<server>, with ips equal to
// the server's own address in a /32 and region equal to the server's region —
// nine times out of nine. That correspondence is what lets `create` need
// nothing but the server name.

var (
	// BackupAgentPolicy is the retention policy to put the agent on.
	BackupAgentPolicy string

	// BackupAgentName overrides the generated display name.
	BackupAgentName string

	// BackupAgentIPs override the addresses derived from the server.
	BackupAgentIPs []string

	// BackupAgentRegion overrides the region derived from the server.
	BackupAgentRegion string

	// BackupAgentWait keeps the command running until the agent has settled.
	BackupAgentWait bool
)

const (
	backupAgentPollInterval = 5 * time.Second
	backupAgentPollAttempts = 60
)

// backupAgentSettled: backup.AgentStatusEnum is CREATING, DISABLED, ENABLED,
// NOT_CONFIGURED, NOT_INSTALLED, UPDATING. Only two of those are transitions;
// the other four are places an agent stops, and NOT_INSTALLED is where a freshly
// created one stops — the object exists, the software is not on the machine yet.
var backupAgentTransient = map[string]bool{"CREATING": true, "UPDATING": true}

// backupRegions are the regions the creation accepts. The API has two region
// enumerations and they are not interchangeable: common.RegionEnum spells
// eu-west-rbx and backup.RegionCodeEnum spells rbx. The creation takes the
// first, which is also what a dedicated server reports as its own region.
var backupRegions = sync.OnceValues(func() ([]string, error) {
	return openapi.GetComponentEnum(assets.BackupservicesV2OpenapiSchema, "common.RegionEnum")
})

// backupAgent is one agent, flattened out of the resource shape.
type backupAgent struct {
	ID           string         `json:"id"`
	Status       string         `json:"status"`
	TargetSpec   map[string]any `json:"targetSpec"`
	CurrentState map[string]any `json:"currentState"`
	CreatedAt    string         `json:"createdAt"`
	UpdatedAt    string         `json:"updatedAt"`
}

func (a backupAgent) protects() string {
	name, _ := a.CurrentState["productResourceName"].(string)
	return name
}

func (a backupAgent) policy() string {
	if policy, ok := a.CurrentState["policy"].(string); ok && policy != "" {
		return policy
	}
	policy, _ := a.TargetSpec["policy"].(string)
	return policy
}

// ShowBackupAgent shows the backup agent protecting a server.
func ShowBackupAgent(_ *cobra.Command, args []string) {
	server := args[0]

	tenant, vspc, err := backupservices.ResolveBoth()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	agents, err := agentsProtecting(tenant, vspc, server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if len(agents) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"serviceName": server, "agents": []any{}},
			"%s has no backup agent.\n   Create one with: ovhcloud baremetal backup-agent create %s", server, server)
		return
	}

	rows := make([]map[string]any, 0, len(agents))
	for _, agent := range agents {
		rows = append(rows, map[string]any{
			"id":          agent.ID,
			"displayName": agent.TargetSpec["displayName"],
			"status":      agent.Status,
			"policy":      policyOrNone(agent.policy()),
			"ips":         agent.CurrentState["ips"],
			"type":        agent.CurrentState["type"],
			"createdAt":   agent.CreatedAt,
		})
	}

	common.RenderFilteredTable(rows, []string{"id", "displayName", "status", "policy", "ips", "type"})
}

// CreateBackupAgent provisions a backup agent for a server.
func CreateBackupAgent(_ *cobra.Command, args []string) {
	server := args[0]

	tenant, vspc, err := backupservices.ResolveBoth()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	existing, err := agentsProtecting(tenant, vspc, server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	if len(existing) > 0 {
		display.OutputError(&flags.OutputFormatConfig,
			"%s already has a backup agent (%s, %s).\n   Change it with: ovhcloud baremetal backup-agent edit %s",
			server, existing[0].ID, existing[0].Status, server)
		return
	}

	spec, err := agentSpecFor(server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Creating an agent reaches past the server: on this account the allowed-IP
	// list of every vault is exactly the set of agent addresses, so a new agent
	// adds its address to them. Somebody agreeing to this should know that.
	if !common.ConfirmAction(common.Disruptive, server, fmt.Sprintf(
		"This provisions a Veeam backup agent for %s in %s, reachable at %s, and lists that address among the ones allowed on the backup vaults.",
		server, spec["region"], joinAny(spec["ips"]))) {
		display.OutputError(&flags.OutputFormatConfig, "creation of a backup agent for %s cancelled", server)
		return
	}

	endpoint := backupservices.VspcPath(tenant, vspc) + "/backupAgent"
	if reportBackupAgentDryRun("POST", endpoint, spec) {
		return
	}

	if err := httpLib.Client.Post(endpoint, spec, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to create a backup agent for %s: %s", server, err)
		return
	}

	if !BackupAgentWait {
		display.OutputInfo(&flags.OutputFormatConfig, spec,
			"⚡️ A backup agent is being created for %s. Follow it with: ovhcloud baremetal backup-agent show %s",
			server, server)
		return
	}

	agent, err := waitForBackupAgent(tenant, vspc, server, true)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// NOT_INSTALLED is where a created agent stops, and it is not a failure: the
	// object exists and the software is not on the machine. Saying "created" and
	// stopping there is how nine agents on this account ended up provisioned and
	// none of them deployed.
	display.OutputInfo(&flags.OutputFormatConfig, agent,
		"✅ Backup agent %s created for %s, status %s.\n"+
			"   It protects nothing until the agent software runs on the machine:\n"+
			"     ovhcloud backup-services deploy-script\n"+
			"   And nothing is retained until it is put on a policy:\n"+
			"     ovhcloud baremetal backup-agent edit %s --policy <name>",
		agent.ID, server, agent.Status, server)
}

// EditBackupAgent changes the agent protecting a server.
func EditBackupAgent(cmd *cobra.Command, args []string) {
	server := args[0]

	tenant, vspc, err := backupservices.ResolveBoth()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	agent, err := oneAgentProtecting(tenant, vspc, server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if cmd.Flags().NFlag() == 0 {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
		return
	}

	// The PUT replaces the target spec, so what is not being changed has to be
	// carried over. Sending only --policy would blank the display name and the
	// addresses, which is the trap `iam resource edit --tag` still has.
	spec := map[string]any{
		"displayName": agent.TargetSpec["displayName"],
		"ips":         agent.TargetSpec["ips"],
		"policy":      agent.TargetSpec["policy"],
	}

	if BackupAgentName != "" {
		spec["displayName"] = BackupAgentName
	}
	if len(BackupAgentIPs) > 0 {
		spec["ips"] = BackupAgentIPs
	}
	if cmd.Flags().Changed("policy") {
		if err := checkBackupPolicy(tenant, vspc, BackupAgentPolicy); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}
		spec["policy"] = BackupAgentPolicy
	}

	if !common.ConfirmAction(common.Disruptive, server, fmt.Sprintf(
		"This changes the backup agent of %s (%s).", server, agent.ID)) {
		display.OutputError(&flags.OutputFormatConfig, "change of the backup agent of %s cancelled", server)
		return
	}

	endpoint := fmt.Sprintf("%s/backupAgent/%s", backupservices.VspcPath(tenant, vspc), url.PathEscape(agent.ID))
	if reportBackupAgentDryRun("PUT", endpoint, spec) {
		return
	}

	if err := httpLib.Client.Put(endpoint, spec, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to change the backup agent of %s: %s", server, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, spec,
		"✅ The backup agent of %s has been changed.", server)
}

// DeleteBackupAgent removes the backup agent of a server.
func DeleteBackupAgent(_ *cobra.Command, args []string) {
	server := args[0]

	tenant, vspc, err := backupservices.ResolveBoth()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	agent, err := oneAgentProtecting(tenant, vspc, server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Destructive: the restore points held under this agent go with it, and no
	// other command brings them back. Same guard as terminating a Backup FTP
	// space — the server's name, typed.
	if !common.ConfirmAction(common.Destructive, server, fmt.Sprintf(
		"This removes the backup agent of %s (%s, policy %s). Its restore points go with it.",
		server, agent.ID, policyOrNone(agent.policy()))) {
		display.OutputError(&flags.OutputFormatConfig, "removal of the backup agent of %s cancelled", server)
		return
	}

	endpoint := fmt.Sprintf("%s/backupAgent/%s", backupservices.VspcPath(tenant, vspc), url.PathEscape(agent.ID))
	if common.ReportDryRun(common.Call{Method: "DELETE", Endpoint: endpoint}) {
		return
	}

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to remove the backup agent of %s: %s", server, err)
		return
	}

	if !BackupAgentWait {
		display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"id": agent.ID, "serviceName": server},
			"⚡️ The backup agent of %s is being removed.", server)
		return
	}

	if _, err := waitForBackupAgent(tenant, vspc, server, false); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"id": agent.ID, "serviceName": server},
		"✅ The backup agent of %s is gone.", server)
}

// agentsProtecting lists the agents that back up a given server.
func agentsProtecting(tenant, vspc, server string) ([]backupAgent, error) {
	var agents []backupAgent

	path := backupservices.VspcPath(tenant, vspc) + "/backupAgent"
	if err := httpLib.Client.Get(path, &agents); err != nil {
		return nil, fmt.Errorf("failed to list the backup agents: %w", err)
	}

	var matching []backupAgent
	for _, agent := range agents {
		if agent.protects() == server {
			matching = append(matching, agent)
		}
	}

	return matching, nil
}

// oneAgentProtecting is agentsProtecting for the commands that change one.
func oneAgentProtecting(tenant, vspc, server string) (backupAgent, error) {
	agents, err := agentsProtecting(tenant, vspc, server)
	if err != nil {
		return backupAgent{}, err
	}

	switch len(agents) {
	case 1:
		return agents[0], nil

	case 0:
		return backupAgent{}, fmt.Errorf(
			"%s has no backup agent.\n   Create one with: ovhcloud baremetal backup-agent create %s", server, server)

	default:
		var lines []string
		for _, agent := range agents {
			lines = append(lines, fmt.Sprintf("     %s  (%s)", agent.ID, agent.Status))
		}

		return backupAgent{}, fmt.Errorf(
			"%d backup agents protect %s, so this command cannot tell which one you mean:\n%s",
			len(agents), server, strings.Join(lines, "\n"))
	}
}

// agentSpecFor builds the creation body from the server itself.
//
// Nothing here has to be typed: nine agents out of nine on this account are
// named agent-<server>, carry the server's own address in a /32, and sit in the
// server's region. The flags override each part for the cases that are not
// those nine.
func agentSpecFor(server string) (map[string]any, error) {
	var machine struct {
		IP     string `json:"ip"`
		Region string `json:"region"`
	}

	path := fmt.Sprintf("/v1/dedicated/server/%s", url.PathEscape(server))
	if err := httpLib.Client.Get(path, &machine); err != nil {
		return nil, fmt.Errorf("failed to read %s, so its region and address cannot be derived: %w", server, err)
	}

	ips := BackupAgentIPs
	if len(ips) == 0 {
		if machine.IP == "" {
			return nil, fmt.Errorf("%s reports no address, so the agent has none to be reached at; give --ip", server)
		}
		ips = []string{machine.IP + "/32"}
	}

	region := BackupAgentRegion
	if region == "" {
		region = machine.Region
	}
	if region == "" {
		return nil, fmt.Errorf("%s reports no region, so the agent has nowhere to sit; give --region", server)
	}

	regions, err := backupRegions()
	if err != nil {
		return nil, fmt.Errorf("failed to read the regions from the embedded schema: %w", err)
	}
	if !slicesContain(regions, region) {
		return nil, fmt.Errorf("region %q is not one the backup service accepts; use one of %s",
			region, strings.Join(regions, ", "))
	}

	name := BackupAgentName
	if name == "" {
		name = "agent-" + server
	}

	return map[string]any{
		"displayName":         name,
		"ips":                 ips,
		"productResourceName": server,
		"region":              region,
	}, nil
}

// checkBackupPolicy refuses a policy the tenant does not define.
func checkBackupPolicy(tenant, vspc, policy string) error {
	policies, err := backupservices.PoliciesOf(tenant, vspc)
	if err != nil {
		return err
	}

	// An empty policy is how an agent is taken off retention, and the API
	// accepts it: it is the state all nine agents of this account are in.
	if policy == "" || slicesContain(policies, policy) {
		return nil
	}

	if len(policies) == 0 {
		return fmt.Errorf("this VSPC tenant defines no retention policy, so %q cannot be one of them", policy)
	}

	return fmt.Errorf("unknown retention policy %q; this tenant defines %s",
		policy, strings.Join(policies, ", "))
}

// waitForBackupAgent follows a creation or a removal by reading the agent.
//
// This generation of the API answers a write with 200 and no body, so there is
// no identifier and no operation to follow — the state of the resource is the
// only thing there is to read. CREATING and UPDATING are transitions; the other
// four statuses are places an agent stops.
func waitForBackupAgent(tenant, vspc, server string, want bool) (backupAgent, error) {
	var last string

	for attempt := 0; attempt < backupAgentPollAttempts; attempt++ {
		agents, err := agentsProtecting(tenant, vspc, server)
		if err != nil {
			return backupAgent{}, err
		}

		switch {
		case !want && len(agents) == 0:
			return backupAgent{}, nil

		case want && len(agents) > 0:
			last = agents[0].Status
			if !backupAgentTransient[last] {
				return agents[0], nil
			}
		}

		time.Sleep(backupAgentPollInterval)
	}

	state := "created"
	if !want {
		state = "removed"
	}

	return backupAgent{}, fmt.Errorf(
		"stopped waiting after %s; the backup agent of %s is not %s yet (last status %q), read it with: ovhcloud baremetal backup-agent show %s",
		time.Duration(backupAgentPollAttempts)*backupAgentPollInterval, server, state, last, server)
}

// reportBackupAgentDryRun previews a call with the body it would carry.
//
// The path of an agent is three UUIDs and says nothing; the body is where the
// region, the addresses and the policy are, and those are what somebody checks
// before agreeing. One document rather than two messages, so that -o json stays
// one JSON document.
func reportBackupAgentDryRun(method, endpoint string, body map[string]any) bool {
	if !flags.DryRun {
		return false
	}

	rendered, err := marshalIndent(body)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to render the request body: %s", err)
		return true
	}

	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{
		"calls":   []map[string]any{{"method": method, "endpoint": endpoint}},
		"payload": body,
	}, "🔍 Dry run: nothing was sent. This would have been called:\n  %s %s\n\nwith:\n  %s",
		method, endpoint, rendered)

	return true
}

// policyOrNone says out loud that an agent retains nothing.
//
// An empty policy field printed as an empty cell reads as "not shown". All nine
// agents of this account are in exactly that state, which is worth reading as
// a fact rather than as a gap in the table.
func policyOrNone(policy string) string {
	if policy == "" {
		return "none"
	}

	return policy
}

// marshalIndent is json.MarshalIndent with the indentation these previews use.
func marshalIndent(value any) ([]byte, error) {
	return json.MarshalIndent(value, "  ", "  ")
}

// joinAny renders the addresses of a spec for a prompt.
//
// It reads the spec rather than the flag, because the flag is empty in the case
// that matters: the addresses were derived from the server, and a prompt that
// showed the flag would show nothing exactly when it has something to say.
func joinAny(value any) string {
	list, ok := value.([]string)
	if !ok {
		return fmt.Sprint(value)
	}

	return strings.Join(list, ", ")
}
