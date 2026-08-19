// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package ip

import (
	_ "embed"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

//go:embed templates/game.tmpl
var gameTemplate string

var (
	// MitigationStateFilter and MitigationAutoFilter narrow the mitigation
	// list. They exist as query parameters on the API and nowhere in the CLI.
	MitigationStateFilter string
	MitigationAutoFilter  string

	// MitigationPermanent is the flag `mitigation set` writes.
	MitigationPermanent bool

	// MitigationTimeout is how long auto-mitigation stays on after an attack.
	MitigationTimeout int

	// GameFirewallMode is the flag `game edit` writes.
	GameFirewallMode bool

	// GameProtocol and GamePorts describe a rule to add.
	GameProtocol string
	GamePorts    string
)

// mitigationTimeouts is the set the API accepts. It is spelled out here so the
// refusal can list it: an operator who types --timeout 30 should be told what
// the five accepted values are, not handed a 400.
var mitigationTimeouts = []int{0, 15, 60, 360, 1560}

// ListMitigation lists the addresses of a block under DDoS mitigation.
func ListMitigation(_ *cobra.Command, args []string) {
	ipBlock := args[0]

	query := url.Values{}
	if MitigationStateFilter != "" {
		query.Set("state", MitigationStateFilter)
	}
	if MitigationAutoFilter != "" {
		query.Set("auto", MitigationAutoFilter)
	}

	path := fmt.Sprintf("/v1/ip/%s/mitigation", url.PathEscape(ipBlock))
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	rows, err := expandAddresses(path,
		fmt.Sprintf("/v1/ip/%s/mitigation", url.PathEscape(ipBlock)))
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if len(rows) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"ip": ipBlock, "mitigation": []any{}},
			"No address of %s is under mitigation.", ipBlock)
		return
	}

	renderFiltered(rows, []string{"ipOnMitigation", "state", "permanent", "auto"})
}

// GetMitigation shows the mitigation state of one address.
func GetMitigation(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(
		fmt.Sprintf("/v1/ip/%s/mitigation", url.PathEscape(args[0])), args[1], "")
}

// AddMitigation puts an address on permanent mitigation.
func AddMitigation(_ *cobra.Command, args []string) {
	ipBlock, target := args[0], args[1]

	// Permanent mitigation routes the address through the scrubbing centre at
	// all times, so it is not a free safety net: the traffic that was going
	// straight now takes a detour. That is why it is confirmed rather than
	// applied silently.
	if !common.ConfirmAction(common.Disruptive, target, fmt.Sprintf(
		"Putting %s on permanent mitigation sends all of its traffic through the scrubbing centre, at all times.", target)) {
		display.OutputError(&flags.OutputFormatConfig, "mitigation of %s cancelled", target)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/mitigation", url.PathEscape(ipBlock))
	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	var created map[string]any
	if err := httpLib.Client.Post(endpoint, map[string]string{"ipOnMitigation": target}, &created); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to put %s on permanent mitigation: %s", target, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, created,
		"⚡️ %s is being put on permanent mitigation (state: %s).",
		target, stringField(created, "state"))
}

// SetMitigation turns permanent mitigation on or off for an address already
// known to the mitigation system.
//
// It is not the same call as add and remove, and it is not a convenience over
// them: POST enrols an address, DELETE drops it from the list entirely, and
// this PUT changes the `permanent` flag of a row that stays. Collapsing three
// API verbs into two commands would have made one of the three unreachable.
func SetMitigation(_ *cobra.Command, args []string) {
	ipBlock, target := args[0], args[1]

	warning := fmt.Sprintf("Turning permanent mitigation off for %s sends its traffic straight again, until an attack triggers auto-mitigation.", target)
	if MitigationPermanent {
		warning = fmt.Sprintf("Turning permanent mitigation on for %s sends all of its traffic through the scrubbing centre, at all times.", target)
	}

	if !common.ConfirmAction(common.Disruptive, target, warning) {
		display.OutputError(&flags.OutputFormatConfig, "change on %s cancelled", target)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/mitigation/%s",
		url.PathEscape(ipBlock), url.PathEscape(target))
	if common.ReportDryRun(common.Call{Method: "PUT", Endpoint: endpoint}) {
		return
	}

	if err := httpLib.Client.Put(endpoint, map[string]bool{"permanent": MitigationPermanent}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to change the mitigation of %s: %s", target, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"ip": target, "permanent": MitigationPermanent},
		"✅ Permanent mitigation is now %s for %s.", onOff(MitigationPermanent), target)
}

// RemoveMitigation drops an address from the mitigation system.
func RemoveMitigation(_ *cobra.Command, args []string) {
	ipBlock, target := args[0], args[1]

	if !common.ConfirmAction(common.Disruptive, target, fmt.Sprintf(
		"Removing %s from mitigation stops filtering its traffic.", target)) {
		display.OutputError(&flags.OutputFormatConfig, "removal of %s cancelled", target)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/mitigation/%s",
		url.PathEscape(ipBlock), url.PathEscape(target))
	if common.ReportDryRun(common.Call{Method: "DELETE", Endpoint: endpoint}) {
		return
	}

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to remove %s from mitigation: %s", target, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"ip": target},
		"⚡️ %s is being removed from mitigation.", target)
}

// ListMitigationProfiles lists the auto-mitigation profiles of a block.
func ListMitigationProfiles(_ *cobra.Command, args []string) {
	ipBlock := args[0]
	basePath := fmt.Sprintf("/v1/ip/%s/mitigationProfiles", url.PathEscape(ipBlock))

	rows, err := expandAddresses(basePath, basePath)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if len(rows) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"ip": ipBlock, "profiles": []any{}},
			"No address of %s has an auto-mitigation profile; the account default applies.", ipBlock)
		return
	}

	for _, row := range rows {
		row["timeout"] = mitigationTimeoutLabel(intField(row, "autoMitigationTimeOut"))
	}

	renderFiltered(rows, []string{"ipMitigationProfile", "state", "timeout"})
}

// GetMitigationProfile shows one auto-mitigation profile.
func GetMitigationProfile(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(
		fmt.Sprintf("/v1/ip/%s/mitigationProfiles", url.PathEscape(args[0])), args[1], "")
}

// SetMitigationProfile sets the auto-mitigation delay of an address, creating
// the profile if it has none.
//
// The API has two routes for this — POST to create, PUT to change — and an
// operator setting a delay does not know, and should not have to find out,
// whether a row already exists. So the existence is read here and the right
// verb chosen; the dry run names the one that would be used.
func SetMitigationProfile(_ *cobra.Command, args []string) {
	ipBlock, target := args[0], args[1]

	if !slicesContainInt(mitigationTimeouts, MitigationTimeout) {
		display.OutputError(&flags.OutputFormatConfig,
			"--timeout %d is not accepted; the API takes %s (in minutes, 0 meaning no delay).",
			MitigationTimeout, joinInts(mitigationTimeouts))
		return
	}

	collection := fmt.Sprintf("/v1/ip/%s/mitigationProfiles", url.PathEscape(ipBlock))
	profile := fmt.Sprintf("%s/%s", collection, url.PathEscape(target))

	method, endpoint, payload := "POST", collection, map[string]any{
		"ipMitigationProfile":   target,
		"autoMitigationTimeOut": MitigationTimeout,
	}

	// A 404 is the only answer that means "no profile yet". Treating every
	// failure as one sent a create where an update was due, and made --dry-run
	// name the wrong verb — the one thing that preview exists to get right.
	var existing map[string]any
	switch err := httpLib.Client.Get(profile, &existing); {
	case err == nil:
		method, endpoint, payload = "PUT", profile, map[string]any{
			"autoMitigationTimeOut": MitigationTimeout,
		}
	case !common.IsNotFound(err):
		display.OutputError(&flags.OutputFormatConfig,
			"failed to read the mitigation profile of %s: %s", target, err)
		return
	}

	if !common.ConfirmAction(common.Disruptive, target, mitigationDelayWarning(target, MitigationTimeout)) {
		display.OutputError(&flags.OutputFormatConfig, "change on %s cancelled", target)
		return
	}

	if common.ReportDryRun(common.Call{Method: method, Endpoint: endpoint}) {
		return
	}

	if method == "POST" {
		err := httpLib.Client.Post(endpoint, payload, nil)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig,
				"failed to create the mitigation profile of %s: %s", target, err)
			return
		}
	} else if err := httpLib.Client.Put(endpoint, payload, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to change the mitigation profile of %s: %s", target, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"ip": target, "autoMitigationTimeOut": MitigationTimeout},
		"✅ %s", mitigationDelayWarning(target, MitigationTimeout))
}

// DeleteMitigationProfile drops an auto-mitigation profile.
func DeleteMitigationProfile(_ *cobra.Command, args []string) {
	ipBlock, target := args[0], args[1]

	if !common.ConfirmAction(common.Disruptive, target, fmt.Sprintf(
		"Deleting the profile of %s returns it to the account's default auto-mitigation delay.", target)) {
		display.OutputError(&flags.OutputFormatConfig, "deletion of the profile of %s cancelled", target)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/mitigationProfiles/%s",
		url.PathEscape(ipBlock), url.PathEscape(target))
	if common.ReportDryRun(common.Call{Method: "DELETE", Endpoint: endpoint}) {
		return
	}

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to delete the mitigation profile of %s: %s", target, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"ip": target},
		"✅ The mitigation profile of %s was deleted.", target)
}

// ListGame lists the addresses of a block under game anti-DDoS.
func ListGame(_ *cobra.Command, args []string) {
	ipBlock := args[0]
	basePath := fmt.Sprintf("/v1/ip/%s/game", url.PathEscape(ipBlock))

	rows, err := expandAddresses(basePath, basePath)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if len(rows) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"ip": ipBlock, "game": []any{}},
			"No address of %s is under game anti-DDoS.", ipBlock)
		return
	}

	for _, row := range rows {
		row["protocols"] = len(stringSlice(row["supportedProtocols"]))
	}

	renderFiltered(rows, []string{"ipOnGame", "state", "firewallModeEnabled", "maxRules", "protocols"})
}

// GetGame shows the game anti-DDoS configuration of one address.
func GetGame(_ *cobra.Command, args []string) {
	ipBlock, target := args[0], args[1]

	config, err := gameConfig(ipBlock, target)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	rules, err := gameRuleIDs(ipBlock, target)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	protocols := stringSlice(config["supportedProtocols"])
	sort.Strings(protocols)

	maxRules := intField(config, "maxRules")
	view := map[string]any{}
	for key, value := range config {
		view[key] = value
	}
	view["summary"] = fmt.Sprintf("%d of %d rules · %d protocols supported",
		len(rules), maxRules, len(protocols))
	view["firewallMode"] = onOff(boolField(config, "firewallModeEnabled"))
	view["rules"] = fmt.Sprintf("%d of %d", len(rules), maxRules)
	view["protocols"] = strings.Join(protocols, ", ")
	view["ruleIds"] = rules

	display.OutputObject(view, target, gameTemplate, &flags.OutputFormatConfig)
}

// EditGame turns UDP firewall mode on or off for an address.
func EditGame(_ *cobra.Command, args []string) {
	ipBlock, target := args[0], args[1]

	rules, err := gameRuleIDs(ipBlock, target)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Firewall mode only lets through UDP traffic matching a rule. Turning it
	// on with no rule at all therefore drops every UDP packet to the address —
	// and that is not a hypothetical: an address measured on a real account
	// had firewall mode off and zero rules, one flag away from going dark.
	if firewallModeWouldBlackhole(GameFirewallMode, len(rules)) {
		display.OutputError(&flags.OutputFormatConfig,
			"%s has no game rule, so turning firewall mode on would drop every UDP packet sent to it.\n   Add the rules first with: ovhcloud ip game rule add %s %s --protocol <protocol> --ports <ports>",
			target, ipBlock, target)
		return
	}

	if !common.ConfirmAction(common.Disruptive, target, gameFirewallWarning(target, GameFirewallMode, len(rules))) {
		display.OutputError(&flags.OutputFormatConfig, "change on %s cancelled", target)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/game/%s",
		url.PathEscape(ipBlock), url.PathEscape(target))
	if common.ReportDryRun(common.Call{Method: "PUT", Endpoint: endpoint}) {
		return
	}

	if err := httpLib.Client.Put(endpoint, map[string]bool{"firewallModeEnabled": GameFirewallMode}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to change the game protection of %s: %s", target, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"ip": target, "firewallModeEnabled": GameFirewallMode},
		"✅ UDP firewall mode is now %s for %s.", onOff(GameFirewallMode), target)
}

// ListGameRules lists the game anti-DDoS rules of an address.
func ListGameRules(_ *cobra.Command, args []string) {
	ipBlock, target := args[0], args[1]

	ids, err := gameRuleIDs(ipBlock, target)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if len(ids) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"ip": target, "rules": []any{}},
			"%s has no game rule.", target)
		return
	}

	basePath := fmt.Sprintf("/v1/ip/%s/game/%s/rule",
		url.PathEscape(ipBlock), url.PathEscape(target))

	rows := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		var rule map[string]any
		if err := httpLib.Client.Get(fmt.Sprintf("%s/%d", basePath, id), &rule); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to read the rule %d: %s", id, err)
			return
		}
		rule["portsReadable"] = portsLabel(rule["ports"])
		rows = append(rows, rule)
	}

	// The API answers the rule identifiers in whatever order it holds them —
	// measured on a real address: 9592667, 9592668, 9667301, 9592666. Sorting
	// makes two readings of the same address comparable.
	sort.Slice(rows, func(i, j int) bool {
		return intField(rows[i], "id") < intField(rows[j], "id")
	})

	renderFiltered(rows, []string{"id", "protocol", "portsReadable ports", "state"})
}

// GetGameRule shows one game anti-DDoS rule.
func GetGameRule(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(
		fmt.Sprintf("/v1/ip/%s/game/%s/rule",
			url.PathEscape(args[0]), url.PathEscape(args[1])), args[2], "")
}

// AddGameRule opens a port range for one game protocol.
func AddGameRule(_ *cobra.Command, args []string) {
	ipBlock, target := args[0], args[1]

	from, to, err := parsePorts(GamePorts)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	config, err := gameConfig(ipBlock, target)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	protocol, err := pickProtocol(target, stringSlice(config["supportedProtocols"]), GameProtocol)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	rules, err := gameRuleIDs(ipBlock, target)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if maxRules := intField(config, "maxRules"); maxRules > 0 && int64(len(rules)) >= maxRules {
		display.OutputError(&flags.OutputFormatConfig,
			"%s already carries its %d rules, which is the most it accepts.\n   Free one with: ovhcloud ip game rule delete %s %s <id>",
			target, maxRules, ipBlock, target)
		return
	}

	if !common.ConfirmAction(common.Disruptive, target, fmt.Sprintf(
		"Adding a %s rule on ports %s of %s changes what its anti-DDoS filter lets through.",
		protocol, portsRange(from, to), target)) {
		display.OutputError(&flags.OutputFormatConfig, "rule on %s cancelled", target)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/game/%s/rule",
		url.PathEscape(ipBlock), url.PathEscape(target))
	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	payload := map[string]any{
		"protocol": protocol,
		"ports":    map[string]int{"from": from, "to": to},
	}

	var created map[string]any
	if err := httpLib.Client.Post(endpoint, payload, &created); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to add the rule: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, created,
		"⚡️ Rule %d is being added on %s: %s ports %s.",
		intField(created, "id"), target, protocol, portsRange(from, to))
}

// DeleteGameRule closes a rule.
func DeleteGameRule(_ *cobra.Command, args []string) {
	ipBlock, target, id := args[0], args[1], args[2]

	// Removing the last rule while firewall mode is on drops all UDP traffic,
	// exactly as turning firewall mode on with no rule does. The guard on the
	// other path would have been decorative without this one: the same outcome
	// is one rule deletion away, and this is the command an operator runs
	// while tidying up rather than while changing protection.
	if blackhole, err := deletingLastRuleWouldBlackhole(ipBlock, target); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	} else if blackhole {
		display.OutputError(&flags.OutputFormatConfig,
			"rule %s is the last one on %s and its firewall mode is on, so deleting it would drop every UDP packet sent to the address.\n   Turn firewall mode off first with: ovhcloud ip game edit %s %s --firewall-mode=false",
			id, target, ipBlock, target)
		return
	}

	if !common.ConfirmAction(common.Disruptive, target, fmt.Sprintf(
		"Deleting rule %s changes what the anti-DDoS filter of %s lets through.", id, target)) {
		display.OutputError(&flags.OutputFormatConfig, "deletion of rule %s cancelled", id)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/game/%s/rule/%s",
		url.PathEscape(ipBlock), url.PathEscape(target), url.PathEscape(id))
	if common.ReportDryRun(common.Call{Method: "DELETE", Endpoint: endpoint}) {
		return
	}

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete the rule %s: %s", id, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"ip": target, "rule": id},
		"⚡️ Rule %s is being deleted from %s.", id, target)
}

// deletingLastRuleWouldBlackhole answers whether this deletion leaves an
// address with firewall mode on and nothing to match.
func deletingLastRuleWouldBlackhole(ipBlock, target string) (bool, error) {
	config, err := gameConfig(ipBlock, target)
	if err != nil {
		return false, err
	}

	if !boolField(config, "firewallModeEnabled") {
		return false, nil
	}

	rules, err := gameRuleIDs(ipBlock, target)
	if err != nil {
		return false, err
	}

	return len(rules) <= 1, nil
}

// gameConfig reads the game anti-DDoS configuration of an address.
func gameConfig(ipBlock, target string) (map[string]any, error) {
	var config map[string]any

	path := fmt.Sprintf("/v1/ip/%s/game/%s", url.PathEscape(ipBlock), url.PathEscape(target))
	if err := httpLib.Client.Get(path, &config); err != nil {
		return nil, fmt.Errorf("%s is not under game anti-DDoS: %w\n   List the protected addresses with: ovhcloud ip game list %s",
			target, err, ipBlock)
	}

	return config, nil
}

// gameRuleIDs reads the rule identifiers of an address.
func gameRuleIDs(ipBlock, target string) ([]int64, error) {
	var ids []int64

	path := fmt.Sprintf("/v1/ip/%s/game/%s/rule",
		url.PathEscape(ipBlock), url.PathEscape(target))
	if err := httpLib.Client.Get(path, &ids); err != nil {
		return nil, fmt.Errorf("failed to read the game rules of %s: %w", target, err)
	}

	return ids, nil
}

// pickProtocol checks a protocol against the ones this address supports.
//
// It is checked against the address rather than against the enum shipped in
// the CLI's embedded schema, because the two disagree: an account measured
// while writing this carries a rule using `arkSurvivalAscended`, which the API
// reports in supportedProtocols and the embedded enum does not list at all.
// Validating against the schema would have refused a protocol already in use.
// The list also differs between addresses — 14 protocols on one, 20 on another
// — so there is no single right list to check against anyway.
func pickProtocol(target string, supported []string, wanted string) (string, error) {
	if wanted == "" {
		return "", fmt.Errorf("--protocol is required; %s supports %s",
			target, strings.Join(sortedCopy(supported), ", "))
	}

	for _, candidate := range supported {
		if strings.EqualFold(candidate, wanted) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("%s does not support the protocol %q.\n   It supports: %s",
		target, wanted, strings.Join(sortedCopy(supported), ", "))
}

// parsePorts reads "443" or "27015-27020".
//
// Both forms exist in the wild: rules measured on a real account use a single
// port and a two-port range, and the API takes a range object either way.
func parsePorts(spec string) (int, int, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return 0, 0, fmt.Errorf("--ports is required, as a single port (7777) or a range (7777-7778)")
	}

	from, to, hasRange := strings.Cut(spec, "-")
	start, err := strconv.Atoi(strings.TrimSpace(from))
	if err != nil {
		return 0, 0, fmt.Errorf("%q is not a port number", from)
	}

	end := start
	if hasRange {
		if end, err = strconv.Atoi(strings.TrimSpace(to)); err != nil {
			return 0, 0, fmt.Errorf("%q is not a port number", to)
		}
	}

	if start < 1 || end > 65535 {
		return 0, 0, fmt.Errorf("ports run from 1 to 65535, %s does not", spec)
	}

	if end < start {
		return 0, 0, fmt.Errorf("the range %s ends before it starts", spec)
	}

	return start, end, nil
}

// expandAddresses reads a list of identifiers and then each one's object.
//
// Every list route of this domain answers identifiers, never objects: a table
// built straight from the list would have exactly one column, and it would be
// the argument the operator already typed.
func expandAddresses(listPath, detailBase string) ([]map[string]any, error) {
	var ids []string
	if err := httpLib.Client.Get(listPath, &ids); err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", listPath, err)
	}

	rows := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		var object map[string]any
		path := fmt.Sprintf("%s/%s", detailBase, url.PathEscape(id))
		if err := httpLib.Client.Get(path, &object); err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", path, err)
		}
		rows = append(rows, object)
	}

	return rows, nil
}

// renderFiltered applies --filter and prints the table.
//
// It is one call rather than two in every list command because the two must
// happen in this order and on the enriched rows: filtering before the readable
// columns exist makes --filter silently match nothing on the very columns the
// table shows, which reads as an answer rather than as a missing feature.
func renderFiltered(rows []map[string]any, columns []string) {
	filtered, err := filtersLib.FilterLines(rows, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(filtered, columns, &flags.OutputFormatConfig)
}

// gameFirewallWarning is the sentence somebody reads before UDP traffic starts
// being dropped. It is built here so a test can read the wording.
func gameFirewallWarning(target string, enabling bool, rules int) string {
	if enabling {
		return fmt.Sprintf("Turning firewall mode on for %s drops every UDP packet that does not match one of its %d rules.",
			target, rules)
	}

	return fmt.Sprintf("Turning firewall mode off for %s lets UDP traffic through on the ports its %d rules do not cover.",
		target, rules)
}

// firewallModeWouldBlackhole answers whether enabling firewall mode on an
// address carrying this many rules drops all of its UDP traffic.
func firewallModeWouldBlackhole(enabling bool, rules int) bool {
	return enabling && rules == 0
}

// mitigationDelayWarning says what the delay does, including for the value the
// API spells 0. "stays on for no delay" named the setting and described no
// behaviour, which is the one thing a confirmation prompt has to do.
func mitigationDelayWarning(target string, minutes int) string {
	if minutes <= 0 {
		return fmt.Sprintf("Auto-mitigation on %s will stop as soon as an attack ends, with no delay.", target)
	}

	return fmt.Sprintf("Auto-mitigation on %s will stay on for %s after an attack ends.",
		target, mitigationTimeoutLabel(int64(minutes)))
}

func mitigationTimeoutLabel(minutes int64) string {
	if minutes <= 0 {
		return "no delay"
	}
	if minutes%60 == 0 {
		return fmt.Sprintf("%dh", minutes/60)
	}
	return fmt.Sprintf("%dmin", minutes)
}

func portsLabel(value any) string {
	ports, ok := value.(map[string]any)
	if !ok {
		return ""
	}

	from, to := intField(ports, "from"), intField(ports, "to")
	if from == 0 && to == 0 {
		return ""
	}

	return portsRange(int(from), int(to))
}

func portsRange(from, to int) string {
	if from == to {
		return strconv.Itoa(from)
	}
	return fmt.Sprintf("%d-%d", from, to)
}

func onOff(enabled bool) string {
	if enabled {
		return "on"
	}
	return "off"
}

func boolField(object map[string]any, key string) bool {
	value, _ := object[key].(bool)
	return value
}

func stringSlice(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}

	result := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok {
			result = append(result, text)
		}
	}
	return result
}

func sortedCopy(values []string) []string {
	copied := append([]string(nil), values...)
	sort.Strings(copied)
	return copied
}

func slicesContainInt(haystack []int, needle int) bool {
	for _, candidate := range haystack {
		if candidate == needle {
			return true
		}
	}
	return false
}

func joinInts(values []int) string {
	texts := make([]string, 0, len(values))
	for _, value := range values {
		texts = append(texts, strconv.Itoa(value))
	}
	return strings.Join(texts, ", ")
}
