// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package ip

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/utils"
	"github.com/spf13/cobra"
)

const FirewallRuleCreateExample = `{
  "action": "permit",
  "protocol": "tcp",
  "sequence": 0,
  "source": "10.0.0.0/8",
  "destinationPort": 443
}`

var (
	firewallColumnsToDisplay = []string{"ipOnFirewall", "enabled", "state"}

	firewallRuleColumnsToDisplay = []string{
		"sequence", "action", "protocol", "source",
		"destinationPort", "rule", "state",
	}

	//go:embed templates/firewall.tmpl
	firewallTemplate string

	//go:embed templates/firewall_rule.tmpl
	firewallRuleTemplate string

	FirewallRuleSpec struct {
		Action              string `json:"action"`
		Protocol            string `json:"protocol"`
		Sequence            int    `json:"sequence"`
		Source              string `json:"source,omitempty"`
		DestinationPort     int    `json:"destinationPort,omitempty"`
		DestinationPortFrom int    `json:"-"`
		DestinationPortTo   int    `json:"-"`
		SourcePort          int    `json:"sourcePort,omitempty"`
		SourcePortFrom      int    `json:"-"`
		SourcePortTo        int    `json:"-"`
		TCPFragments        bool   `json:"-"`
		TCPOption           string `json:"-"`
	}
)

// ListFirewall lists IPs registered on the firewall for the given IP block.
// API: GET /v1/ip/{ip}/firewall
func ListFirewall(_ *cobra.Command, args []string) {
	baseURL := fmt.Sprintf("/v1/ip/%s/firewall", url.PathEscape(args[0]))

	ids, err := httpLib.FetchArray(baseURL, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch firewall list: %s", err)
		return
	}

	// FetchObjectsParallel uses fmt.Sprintf(fmtPath, id), so we must escape any
	// literal '%' characters in the base URL (e.g. %2F from url.PathEscape).
	fmtPath := strings.ReplaceAll(baseURL, "%", "%%") + "/%s"
	objects, err := httpLib.FetchObjectsParallel[map[string]any](fmtPath, ids, flags.IgnoreErrors)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch firewall details: %s", err)
		return
	}

	var body []map[string]any
	for _, obj := range objects {
		if obj != nil {
			body = append(body, obj)
		}
	}

	body, err = filtersLib.FilterLines(body, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(body, firewallColumnsToDisplay, &flags.OutputFormatConfig)
}

// AddFirewall adds an IP to the firewall.
// API: POST /v1/ip/{ip}/firewall
func AddFirewall(_ *cobra.Command, args []string) {
	apiURL := fmt.Sprintf("/v1/ip/%s/firewall", url.PathEscape(args[0]))
	if err := httpLib.Client.Post(apiURL, map[string]string{"ipOnFirewall": args[1]}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ IP %s successfully added to firewall", args[1])
}

// GetFirewall shows firewall status for a specific IP.
// API: GET /v1/ip/{ip}/firewall/{ipOnFirewall}
func GetFirewall(_ *cobra.Command, args []string) {
	apiURL := fmt.Sprintf("/v1/ip/%s/firewall/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))

	var object map[string]any
	if err := httpLib.Client.Get(apiURL, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching firewall for %s: %s", args[1], err)
		return
	}

	display.OutputObject(object, args[1], firewallTemplate, &flags.OutputFormatConfig)
}

// EnableFirewall enables the firewall on an IP.
// API: PUT /v1/ip/{ip}/firewall/{ipOnFirewall}
func EnableFirewall(_ *cobra.Command, args []string) {
	apiURL := fmt.Sprintf("/v1/ip/%s/firewall/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Put(apiURL, map[string]bool{"enabled": true}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Firewall successfully enabled for %s", args[1])
}

// DisableFirewall disables the firewall on an IP.
// API: PUT /v1/ip/{ip}/firewall/{ipOnFirewall}
func DisableFirewall(_ *cobra.Command, args []string) {
	apiURL := fmt.Sprintf("/v1/ip/%s/firewall/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Put(apiURL, map[string]bool{"enabled": false}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Firewall successfully disabled for %s", args[1])
}

// DeleteFirewall removes an IP and all its rules from the firewall.
// API: DELETE /v1/ip/{ip}/firewall/{ipOnFirewall}
func DeleteFirewall(_ *cobra.Command, args []string) {
	apiURL := fmt.Sprintf("/v1/ip/%s/firewall/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(apiURL, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Firewall and all rules successfully removed for %s", args[1])
}

// ListFirewallRules lists all firewall rules for the given IP.
// API: GET /v1/ip/{ip}/firewall/{ipOnFirewall}/rule
func ListFirewallRules(_ *cobra.Command, args []string) {
	baseURL := fmt.Sprintf("/v1/ip/%s/firewall/%s/rule",
		url.PathEscape(args[0]), url.PathEscape(args[1]))

	ids, err := httpLib.FetchArray(baseURL, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch firewall rules: %s", err)
		return
	}

	// FetchObjectsParallel uses fmt.Sprintf(fmtPath, id), so we must escape any
	// literal '%' characters in the base URL (e.g. %2F from url.PathEscape).
	fmtPath := strings.ReplaceAll(baseURL, "%", "%%") + "/%s"
	objects, err := httpLib.FetchObjectsParallel[map[string]any](fmtPath, ids, flags.IgnoreErrors)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch rule details: %s", err)
		return
	}

	var body []map[string]any
	for _, obj := range objects {
		if obj != nil {
			body = append(body, obj)
		}
	}

	body, err = filtersLib.FilterLines(body, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(body, firewallRuleColumnsToDisplay, &flags.OutputFormatConfig)
}

// GetFirewallRule gets details of a specific firewall rule.
// API: GET /v1/ip/{ip}/firewall/{ipOnFirewall}/rule/{sequence}
func GetFirewallRule(_ *cobra.Command, args []string) {
	apiURL := fmt.Sprintf("/v1/ip/%s/firewall/%s/rule/%s",
		url.PathEscape(args[0]), url.PathEscape(args[1]), args[2])

	var object map[string]any
	if err := httpLib.Client.Get(apiURL, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching rule #%s: %s", args[2], err)
		return
	}

	display.OutputObject(object, args[2], firewallRuleTemplate, &flags.OutputFormatConfig)
}

// CreateFirewallRule creates a new firewall rule.
// API: POST /v1/ip/{ip}/firewall/{ipOnFirewall}/rule
// Uses RunE signature so client-side validation errors are returned to cobra
// (which prints them and exits cleanly without calling os.Exit).
func CreateFirewallRule(cmd *cobra.Command, args []string) error {
	apiURL := fmt.Sprintf("/v1/ip/%s/firewall/%s/rule", url.PathEscape(args[0]), url.PathEscape(args[1]))

	// If --from-file or pipe: load base params, merge explicit CLI flags on top, POST directly
	usingFile := flags.ParametersFile != "" || utils.IsInputFromPipe()
	if usingFile {
		var fileData []byte
		var err error
		if utils.IsInputFromPipe() {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				fileData = append(fileData, scanner.Bytes()...)
			}
			if err = scanner.Err(); err != nil {
				return fmt.Errorf("failed to read from pipe: %w", err)
			}
		} else {
			fileData, err = os.ReadFile(flags.ParametersFile)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", flags.ParametersFile, err)
			}
		}

		var body map[string]any
		if err := json.Unmarshal(fileData, &body); err != nil {
			return fmt.Errorf("failed to parse parameters: %w", err)
		}

		// Override with any explicitly set CLI flags
		if cmd.Flags().Changed("action") {
			body["action"] = FirewallRuleSpec.Action
		}
		if cmd.Flags().Changed("protocol") {
			body["protocol"] = FirewallRuleSpec.Protocol
		}
		if cmd.Flags().Changed("sequence") {
			body["sequence"] = FirewallRuleSpec.Sequence
		}
		if cmd.Flags().Changed("source") {
			body["source"] = FirewallRuleSpec.Source
		}
		if cmd.Flags().Changed("destination-port") {
			body["destinationPort"] = FirewallRuleSpec.DestinationPort
		}
		if cmd.Flags().Changed("destination-port-from") || cmd.Flags().Changed("destination-port-to") {
			body["destinationPortRange"] = map[string]int{
				"from": FirewallRuleSpec.DestinationPortFrom,
				"to":   FirewallRuleSpec.DestinationPortTo,
			}
		}
		if cmd.Flags().Changed("source-port") {
			body["sourcePort"] = FirewallRuleSpec.SourcePort
		}
		if cmd.Flags().Changed("source-port-from") || cmd.Flags().Changed("source-port-to") {
			body["sourcePortRange"] = map[string]int{
				"from": FirewallRuleSpec.SourcePortFrom,
				"to":   FirewallRuleSpec.SourcePortTo,
			}
		}
		if cmd.Flags().Changed("tcp-fragments") || cmd.Flags().Changed("tcp-option") {
			tcpOpt := map[string]any{}
			if cmd.Flags().Changed("tcp-fragments") {
				tcpOpt["fragments"] = FirewallRuleSpec.TCPFragments
			}
			if cmd.Flags().Changed("tcp-option") {
				tcpOpt["option"] = FirewallRuleSpec.TCPOption
			}
			body["tcpOption"] = tcpOpt
		}

		// Validate required fields are present (from file or CLI)
		for _, field := range []string{"action", "protocol", "sequence"} {
			if _, ok := body[field]; !ok {
				return fmt.Errorf("required field %q is missing (must be provided via file or --%s flag)", field, field)
			}
		}

		var createdRule map[string]any
		if err := httpLib.Client.Post(apiURL, body, &createdRule); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "error creating rule: %s", err)
			return nil
		}
		display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Rule #%v successfully created", createdRule["sequence"])
		return nil
	}

	// --- CLI-flags-only path: validate and build body manually ---

	// Validate required flags were provided
	for _, flagName := range []string{"action", "protocol", "sequence"} {
		if !cmd.Flags().Changed(flagName) {
			return fmt.Errorf("required flag %q not set", flagName)
		}
	}

	// Client-side validation
	validActions := map[string]bool{"deny": true, "permit": true}
	if !validActions[FirewallRuleSpec.Action] {
		return fmt.Errorf("invalid action %q: must be 'deny' or 'permit'", FirewallRuleSpec.Action)
	}

	validProtocols := map[string]bool{
		"ah": true, "esp": true, "gre": true, "icmp": true, "ipv4": true, "tcp": true, "udp": true,
	}
	if !validProtocols[FirewallRuleSpec.Protocol] {
		return fmt.Errorf("invalid protocol %q: must be one of ah, esp, gre, icmp, ipv4, tcp, udp", FirewallRuleSpec.Protocol)
	}

	if FirewallRuleSpec.Sequence < 0 || FirewallRuleSpec.Sequence > 19 {
		return fmt.Errorf("invalid sequence %d: must be between 0 and 19", FirewallRuleSpec.Sequence)
	}

	tcpUDP := FirewallRuleSpec.Protocol == "tcp" || FirewallRuleSpec.Protocol == "udp"

	if !tcpUDP {
		for _, portFlag := range []string{"destination-port", "destination-port-from", "destination-port-to", "source-port", "source-port-from", "source-port-to"} {
			if cmd.Flags().Changed(portFlag) {
				return fmt.Errorf("port options are only valid for TCP and UDP protocols")
			}
		}
	}

	if FirewallRuleSpec.Protocol != "tcp" {
		for _, tcpFlag := range []string{"tcp-option", "tcp-fragments"} {
			if cmd.Flags().Changed(tcpFlag) {
				return fmt.Errorf("--%s is only valid for TCP protocol", tcpFlag)
			}
		}
	}

	// Build POST body
	body := map[string]any{
		"action":   FirewallRuleSpec.Action,
		"protocol": FirewallRuleSpec.Protocol,
		"sequence": FirewallRuleSpec.Sequence,
	}

	if cmd.Flags().Changed("source") {
		body["source"] = FirewallRuleSpec.Source
	}

	if cmd.Flags().Changed("destination-port") {
		body["destinationPort"] = FirewallRuleSpec.DestinationPort
	} else if cmd.Flags().Changed("destination-port-from") || cmd.Flags().Changed("destination-port-to") {
		body["destinationPortRange"] = map[string]int{
			"from": FirewallRuleSpec.DestinationPortFrom,
			"to":   FirewallRuleSpec.DestinationPortTo,
		}
	}

	if cmd.Flags().Changed("source-port") {
		body["sourcePort"] = FirewallRuleSpec.SourcePort
	} else if cmd.Flags().Changed("source-port-from") || cmd.Flags().Changed("source-port-to") {
		body["sourcePortRange"] = map[string]int{
			"from": FirewallRuleSpec.SourcePortFrom,
			"to":   FirewallRuleSpec.SourcePortTo,
		}
	}

	if FirewallRuleSpec.Protocol == "tcp" {
		tcpFragmentsChanged := cmd.Flags().Changed("tcp-fragments")
		tcpOptionChanged := cmd.Flags().Changed("tcp-option")

		if tcpFragmentsChanged || tcpOptionChanged {
			tcpOption := map[string]any{}
			if tcpFragmentsChanged {
				tcpOption["fragments"] = FirewallRuleSpec.TCPFragments
			}
			if tcpOptionChanged {
				tcpOption["option"] = FirewallRuleSpec.TCPOption
			}
			body["tcpOption"] = tcpOption
		}
	}

	var createdRule map[string]any
	if err := httpLib.Client.Post(apiURL, body, &createdRule); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error creating rule: %s", err)
		return nil
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Rule #%v successfully created", createdRule["sequence"])
	return nil
}

// DeleteFirewallRule deletes a firewall rule.
// API: DELETE /v1/ip/{ip}/firewall/{ipOnFirewall}/rule/{sequence}
func DeleteFirewallRule(_ *cobra.Command, args []string) {
	apiURL := fmt.Sprintf("/v1/ip/%s/firewall/%s/rule/%s",
		url.PathEscape(args[0]), url.PathEscape(args[1]), args[2])
	if err := httpLib.Client.Delete(apiURL, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Rule #%s successfully deleted", args[2])
}
