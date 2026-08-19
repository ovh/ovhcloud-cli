// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/completion"
	"github.com/ovh/ovhcloud-cli/internal/services/ip"
	"github.com/spf13/cobra"
)

func init() {
	ipCmd := &cobra.Command{
		Use:   "ip",
		Short: "Retrieve information and manage your IP services",
	}

	// Command to list Ip services
	ipListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your Ip services",
		Run:     ip.ListIp,
	}
	ipCmd.AddCommand(withFilterFlag(ipListCmd))

	// Command to get a single Ip
	ipCmd.AddCommand(&cobra.Command{
		Use:               "get <service_name>",
		Short:             "Retrieve information of a specific Ip",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.GetIp,
	})

	// Command to update a single Ip
	ipEditCmd := &cobra.Command{
		Use:               "edit <service_name>",
		Short:             "Edit the given IP",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.EditIp,
	}
	ipEditCmd.Flags().StringVar(&ip.IPSpec.Description, "description", "", "Description of the IP")
	addInteractiveEditorFlag(ipEditCmd)
	ipCmd.AddCommand(ipEditCmd)

	// An additional IP is bought to be moved between services; the CLI could
	// show where each one pointed and move none of them.
	ipCmd.AddCommand(&cobra.Command{
		Use:               "destinations <ip_block>",
		Short:             "List the services this IP can be moved to",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.ListIpDestinations,
	})

	ipMoveCmd := &cobra.Command{
		Use:               "move <ip_block> <service>",
		Short:             "Route the given IP to another service",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.MoveIp,
	}
	ipMoveCmd.Flags().StringVar(&ip.MoveNexthop, "nexthop", "",
		"Next hop to use, when the destination offers several")
	ipMoveCmd.Flags().BoolVar(&ip.IPWait, "wait", false,
		"Wait until the IP is actually routed to the destination before exiting")
	addConfirmationFlags(ipMoveCmd, "Print the call that would be made without making it")
	ipCmd.AddCommand(ipMoveCmd)

	ipParkCmd := &cobra.Command{
		Use:               "park <ip_block>",
		Short:             "Detach the given IP from the service it currently serves",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.ParkIp,
	}
	ipParkCmd.Flags().BoolVar(&ip.IPWait, "wait", false,
		"Wait until the IP is actually parked before exiting")
	addConfirmationFlags(ipParkCmd, "Print the call that would be made without making it")
	ipCmd.AddCommand(ipParkCmd)

	ipCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "tasks <ip_block>",
		Short:             "List the tasks of the given IP",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.ListIpTasks,
	}))

	ipReverseCmd := &cobra.Command{
		Use:   "reverse",
		Short: "Manage reverses on the given IP",
	}
	ipCmd.AddCommand(ipReverseCmd)

	ipReverseSetCmd := &cobra.Command{
		Use:               "set <service_name> <ip> <reverse>",
		Short:             "Set reverse on the given IP",
		Args:              cobra.ExactArgs(3),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.IpSetReverse,
	}
	ipReverseCmd.AddCommand(ipReverseSetCmd)

	ipReverseGetCmd := &cobra.Command{
		Use:               "get <service_name>",
		Short:             "List reverse on the given IP range",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.IpGetReverse,
	}
	ipReverseCmd.AddCommand(ipReverseGetCmd)

	ipReverseDeleteCmd := &cobra.Command{
		Use:               "delete <service_name> <ip>",
		Short:             "Delete reverse on the given IP",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.IpDeleteReverse,
	}
	ipReverseCmd.AddCommand(ipReverseDeleteCmd)

	// Firewall commands
	ipFirewallCmd := &cobra.Command{
		Use:   "firewall",
		Short: "Manage firewall (Edge Firewall) on the given IP",
	}
	ipCmd.AddCommand(ipFirewallCmd)

	ipFirewallCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list <ip_block>",
		Aliases:           []string{"ls"},
		Short:             "List IPs registered on the firewall",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.ListFirewall,
	}))

	ipFirewallCmd.AddCommand(&cobra.Command{
		Use:               "add <ip_block> <ip>",
		Short:             "Add an IP to the firewall",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.AddFirewall,
	})

	ipFirewallCmd.AddCommand(&cobra.Command{
		Use:               "get <ip_block> <ip>",
		Short:             "Get firewall status for a specific IP",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.GetFirewall,
	})

	ipFirewallCmd.AddCommand(&cobra.Command{
		Use:               "enable <ip_block> <ip>",
		Short:             "Enable the firewall on the given IP",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.EnableFirewall,
	})

	ipFirewallCmd.AddCommand(&cobra.Command{
		Use:               "disable <ip_block> <ip>",
		Short:             "Disable the firewall on the given IP",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.DisableFirewall,
	})

	ipFirewallCmd.AddCommand(&cobra.Command{
		Use:               "delete <ip_block> <ip>",
		Short:             "Remove IP and all rules from firewall",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.DeleteFirewall,
	})

	// Firewall rule sub-commands
	ipFirewallRuleCmd := &cobra.Command{
		Use:   "rule",
		Short: "Manage firewall rules",
	}
	ipFirewallCmd.AddCommand(ipFirewallRuleCmd)

	ipFirewallRuleCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list <ip_block> <ip>",
		Aliases:           []string{"ls"},
		Short:             "List firewall rules for the given IP",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.ListFirewallRules,
	}))

	ipFirewallRuleCmd.AddCommand(&cobra.Command{
		Use:               "get <ip_block> <ip> <sequence>",
		Short:             "Get a specific firewall rule",
		Args:              cobra.ExactArgs(3),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.GetFirewallRule,
	})

	ipFirewallRuleCreateCmd := &cobra.Command{
		Use:   "create <ip_block> <ip>",
		Short: "Create a new firewall rule",
		Long: `Use this command to create a new firewall rule.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud ip firewall rule create <ip_block> <ip> --action permit --protocol tcp --sequence 0 --destination-port 443

2. Using a configuration file:

  First generate an example parameters file:

	ovhcloud ip firewall rule create <ip_block> <ip> --init-file ./rule.json

  After editing the file, run:

	ovhcloud ip firewall rule create <ip_block> <ip> --from-file ./rule.json

  You can also pipe the content:

	cat ./rule.json | ovhcloud ip firewall rule create <ip_block> <ip>

3. Using your default text editor:

	ovhcloud ip firewall rule create <ip_block> <ip> --editor
`,
		Args: cobra.ExactArgs(2),
		RunE: ip.CreateFirewallRule,
	}
	ipFirewallRuleCreateCmd.Flags().StringVar(&ip.FirewallRuleSpec.Action, "action", "", "Action: deny or permit (required)")
	ipFirewallRuleCreateCmd.Flags().StringVar(&ip.FirewallRuleSpec.Protocol, "protocol", "", "Protocol: ah, esp, gre, icmp, ipv4, tcp, udp (required)")
	ipFirewallRuleCreateCmd.Flags().IntVar(&ip.FirewallRuleSpec.Sequence, "sequence", -1, "Rule priority 0-19 (required)")
	ipFirewallRuleCreateCmd.Flags().StringVar(&ip.FirewallRuleSpec.Source, "source", "", "Source IP/CIDR (defaults to any)")
	ipFirewallRuleCreateCmd.Flags().IntVar(&ip.FirewallRuleSpec.DestinationPort, "destination-port", 0, "Destination port (TCP/UDP only)")
	ipFirewallRuleCreateCmd.Flags().IntVar(&ip.FirewallRuleSpec.SourcePort, "source-port", 0, "Source port (TCP/UDP only)")
	ipFirewallRuleCreateCmd.Flags().BoolVar(&ip.FirewallRuleSpec.TCPFragments, "tcp-fragments", false, "TCP fragments option")
	ipFirewallRuleCreateCmd.Flags().StringVar(&ip.FirewallRuleSpec.TCPOption, "tcp-option", "", "TCP option: established or syn (TCP only)")
	addParameterFileFlags(ipFirewallRuleCreateCmd, false, assets.IpOpenapiSchema, "/ip/{ip}/firewall/{ipOnFirewall}/rule", "post", ip.FirewallRuleCreateExample, nil)
	addInteractiveEditorFlag(ipFirewallRuleCreateCmd)
	markFlagsMutuallyExclusive(ipFirewallRuleCreateCmd, "from-file", "editor")
	ipFirewallRuleCmd.AddCommand(ipFirewallRuleCreateCmd)

	ipFirewallRuleCmd.AddCommand(&cobra.Command{
		Use:               "delete <ip_block> <ip> <sequence>",
		Short:             "Delete a firewall rule",
		Args:              cobra.ExactArgs(3),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.DeleteFirewallRule,
	})

	rootCmd.AddCommand(ipCmd)
}
