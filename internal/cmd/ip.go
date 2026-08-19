// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/completion"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
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

	// Where IPs can live, and which registries each site accepts for a
	// bring-your-own-IP announcement. The only route of this domain that
	// answers something about OVHcloud rather than about the account.
	ipCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "campus",
		Short: "List the IP campuses and the registries they accept",
		Args:  cobra.NoArgs,
		Run:   ip.ListCampus,
	}))

	// The billed services. `ip list` and `ip service list` are different
	// resources: a block is what gets routed, a service is what gets renewed,
	// has contacts and can be terminated.
	ipServiceCmd := &cobra.Command{
		Use:   "service",
		Short: "Manage your IP services: contacts, renewal and termination",
	}
	ipCmd.AddCommand(ipServiceCmd)

	ipServiceCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your IP services",
		Args:    cobra.NoArgs,
		Run:     ip.ListIpServices,
	}))

	ipServiceCmd.AddCommand(&cobra.Command{
		Use:               "get <service_name>",
		Short:             "Show one IP service",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip/service"),
		Run:               ip.GetIpService,
	})

	ipServiceEditCmd := &cobra.Command{
		Use:               "edit <service_name>",
		Short:             "Edit an IP service",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip/service"),
		Run:               ip.EditIpService,
	}
	ipServiceEditCmd.Flags().StringVar(&ip.IPServiceSpec.Description, "description", "", "Description of the service")
	addInteractiveEditorFlag(ipServiceEditCmd)
	ipServiceCmd.AddCommand(ipServiceEditCmd)

	ipServiceInfoCmd := &cobra.Command{
		Use:   "service-info",
		Short: "Manage the billing information of an IP service",
	}
	ipServiceCmd.AddCommand(ipServiceInfoCmd)

	ipServiceInfoCmd.AddCommand(&cobra.Command{
		Use:               "get <service_name>",
		Short:             "Show the billing information of an IP service",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip/service"),
		Run:               ip.GetIpServiceInfo,
	})

	ipServiceInfoEditCmd := &cobra.Command{
		Use:               "edit <service_name>",
		Short:             "Edit the billing information of an IP service",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip/service"),
		Run:               ip.EditIpServiceInfo,
	}
	common.AddServiceInfoRenewFlags(ipServiceInfoEditCmd)
	addInteractiveEditorFlag(ipServiceInfoEditCmd)
	ipServiceInfoCmd.AddCommand(ipServiceInfoEditCmd)

	ipServiceContactCmd := &cobra.Command{
		Use:               "change-contact <service_name>",
		Short:             "Start a contact change procedure on an IP service",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip/service"),
		Run:               ip.ChangeIpServiceContact,
	}
	ipServiceContactCmd.Flags().StringVar(&ip.ContactAdmin, "admin", "", "New administrative contact")
	ipServiceContactCmd.Flags().StringVar(&ip.ContactBilling, "billing", "", "New billing contact")
	ipServiceContactCmd.Flags().StringVar(&ip.ContactTech, "tech", "", "New technical contact")
	addConfirmationFlags(ipServiceContactCmd, "Print the call that would be made without making it")
	ipServiceCmd.AddCommand(ipServiceContactCmd)

	ipServiceTerminateCmd := &cobra.Command{
		Use:               "terminate <service_name>",
		Short:             "Ask for the termination of an IP service",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip/service"),
		Run:               ip.TerminateIpService,
	}
	addConfirmationFlags(ipServiceTerminateCmd, "Print the call that would be made without making it")
	ipServiceCmd.AddCommand(ipServiceTerminateCmd)

	ipServiceConfirmCmd := &cobra.Command{
		Use:               "confirm-termination <service_name> <token>",
		Short:             "Confirm the termination of an IP service with the emailed token",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip/service"),
		Run:               ip.ConfirmIpServiceTermination,
	}
	ipServiceConfirmCmd.Flags().StringVar(&ip.TerminationReason, "reason", "", "Why the service is being terminated (press <tab> for the accepted values)")
	ipServiceConfirmCmd.Flags().StringVar(&ip.TerminationFutureUse, "future-use", "", "What comes next after this termination (press <tab> for the accepted values)")
	ipServiceConfirmCmd.Flags().StringVar(&ip.TerminationComment, "commentary", "", "Free-text comment attached to the termination request")
	ipServiceConfirmCmd.RegisterFlagCompletionFunc("reason", ip.CompleteTerminationReason)
	ipServiceConfirmCmd.RegisterFlagCompletionFunc("future-use", ip.CompleteTerminationFutureUse)
	addConfirmationFlags(ipServiceConfirmCmd, "Print the call that would be made without making it")
	ipServiceCmd.AddCommand(ipServiceConfirmCmd)

	// Reverse delegation
	ipDelegationCmd := &cobra.Command{
		Use:   "delegation",
		Short: "Manage the reverse delegation of an IPv6 subnet",
	}
	ipCmd.AddCommand(ipDelegationCmd)

	ipDelegationCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list <ip_block>",
		Aliases:           []string{"ls"},
		Short:             "List the name servers the reverse is delegated to",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.ListDelegation,
	}))

	ipDelegationCmd.AddCommand(&cobra.Command{
		Use:               "get <ip_block> <target>",
		Short:             "Show one reverse delegation target",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.GetDelegation,
	})

	ipDelegationAddCmd := &cobra.Command{
		Use:               "add <ip_block> <target>",
		Short:             "Delegate the reverse of this subnet to a name server",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.AddDelegation,
	}
	addConfirmationFlags(ipDelegationAddCmd, "Print the call that would be made without making it")
	ipDelegationCmd.AddCommand(ipDelegationAddCmd)

	ipDelegationRemoveCmd := &cobra.Command{
		Use:               "remove <ip_block> <target>",
		Short:             "Stop delegating the reverse of this subnet to a name server",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.RemoveDelegation,
	}
	addConfirmationFlags(ipDelegationRemoveCmd, "Print the call that would be made without making it")
	ipDelegationCmd.AddCommand(ipDelegationRemoveCmd)

	// The API has one licence route per product and no index, so asking
	// "does this address carry a licence" is eight requests.
	ipCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "licenses <ip_block>",
		Aliases:           []string{"licences"},
		Short:             "List every licence attached to this IP",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.ListLicenses,
	}))

	// RIPE record
	ipRipeCmd := &cobra.Command{
		Use:   "ripe",
		Short: "Read and change the RIPE record published for an IP block",
	}
	ipCmd.AddCommand(ipRipeCmd)

	ipRipeCmd.AddCommand(&cobra.Command{
		Use:               "get <ip_block>",
		Short:             "Show the RIPE record of this block",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.GetRipe,
	})

	ipRipeSetCmd := &cobra.Command{
		Use:               "set <ip_block>",
		Short:             "Change the RIPE record of this block",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.SetRipe,
	}
	ipRipeSetCmd.Flags().StringVar(&ip.RipeNetname, "netname", "", "Netname published in the registry")
	ipRipeSetCmd.Flags().StringVar(&ip.RipeDescription, "description", "", "Description published in the registry")
	addConfirmationFlags(ipRipeSetCmd, "Print the call that would be made without making it")
	ipRipeCmd.AddCommand(ipRipeSetCmd)

	// Bring your own IP
	ipByoipCmd := &cobra.Command{
		Use:   "byoip",
		Short: "Aggregate or slice a bring-your-own-IP block",
	}
	ipCmd.AddCommand(ipByoipCmd)

	ipByoipCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "aggregations <ip_block>",
		Short:             "List the blocks this one could be merged into",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.ListByoipAggregations,
	}))

	ipByoipCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "slices <ip_block>",
		Short:             "List the ways this block could be split",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.ListByoipSlices,
	}))

	ipByoipAggregateCmd := &cobra.Command{
		Use:               "aggregate <ip_block>",
		Short:             "Merge this block with its neighbours",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.AggregateByoip,
	}
	ipByoipAggregateCmd.Flags().StringVar(&ip.ByoipAggregationIp, "into", "",
		"Parent block to merge into, as listed by `byoip aggregations` (required)")
	addConfirmationFlags(ipByoipAggregateCmd, "Print the call that would be made without making it")
	ipByoipCmd.AddCommand(ipByoipAggregateCmd)

	ipByoipSliceCmd := &cobra.Command{
		Use:               "slice <ip_block>",
		Short:             "Split this block into smaller ones",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.SliceByoip,
	}
	ipByoipSliceCmd.Flags().IntVar(&ip.ByoipSlicingSize, "size", 0,
		"Prefix length of the smaller blocks, as listed by `byoip slices` (required)")
	addConfirmationFlags(ipByoipSliceCmd, "Print the call that would be made without making it")
	ipByoipCmd.AddCommand(ipByoipSliceCmd)

	// Migration token
	ipMigrationCmd := &cobra.Command{
		Use:   "migration-token",
		Short: "Manage the token letting another account claim this IP",
	}
	ipCmd.AddCommand(ipMigrationCmd)

	ipMigrationGetCmd := &cobra.Command{
		Use:               "get <ip_block>",
		Short:             "Show the pending migration token of this IP",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.GetMigrationToken,
	}
	ipMigrationGetCmd.Flags().BoolVar(&ip.RevealMigrationToken, "reveal", false,
		"Print the token itself instead of its fingerprint")
	ipMigrationCmd.AddCommand(ipMigrationGetCmd)

	ipMigrationCreateCmd := &cobra.Command{
		Use:               "create <ip_block>",
		Short:             "Generate a token letting another account claim this IP",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.CreateMigrationToken,
	}
	ipMigrationCreateCmd.Flags().StringVar(&ip.MigrationCustomerId, "customer-id", "",
		"Account that will be able to claim the IP (required)")
	ipMigrationCreateCmd.Flags().BoolVar(&ip.RevealMigrationToken, "reveal", false,
		"Print the token itself instead of its fingerprint")
	addConfirmationFlags(ipMigrationCreateCmd, "Print the call that would be made without making it")
	ipMigrationCmd.AddCommand(ipMigrationCreateCmd)

	ipChangeOrgCmd := &cobra.Command{
		Use:               "change-org <ip_block> <organisation>",
		Short:             "Register this IP to another organisation in the regional registry",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.ChangeIpOrganisation,
	}
	addConfirmationFlags(ipChangeOrgCmd, "Print the call that would be made without making it")
	ipCmd.AddCommand(ipChangeOrgCmd)

	// The incident surface. Three mechanisms can block an address, each with
	// its own list and its own release route, and an operator hit by one of
	// them does not know which. `blocked` reads the three, and what it prints
	// is what `unblock` takes.
	ipCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "blocked <ip_block>",
		Short:             "List the addresses of this block held by anti-hack, ARP or anti-spam",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.ListBlocked,
	}))

	ipUnblockCmd := &cobra.Command{
		Use:               "unblock <ip_block> <ip>",
		Short:             "Release an address from the mechanism blocking it",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.UnblockIp,
	}
	ipUnblockCmd.Flags().StringVar(&ip.UnblockReason, "reason", "",
		"Mechanism to release from: antihack, arp or spam (read from the API when omitted)")
	addConfirmationFlags(ipUnblockCmd, "Print the call that would be made without making it")
	ipCmd.AddCommand(ipUnblockCmd)

	ipSpamStatsCmd := &cobra.Command{
		Use:               "spam-stats <ip_block> <ip>",
		Short:             "Show what an address sent while the anti-spam system held it",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.SpamStats,
	}
	ipSpamStatsCmd.Flags().StringVar(&ip.SpamStatsFrom, "from", "",
		"Start of the window, RFC3339 (defaults to the date the address was blocked)")
	ipSpamStatsCmd.Flags().StringVar(&ip.SpamStatsTo, "to", "",
		"End of the window, RFC3339 (defaults to now)")
	ipCmd.AddCommand(withFilterFlag(ipSpamStatsCmd))

	// Anti-phishing
	ipPhishingCmd := &cobra.Command{
		Use:   "phishing",
		Short: "Read the phishing URLs reported on the given IP block",
	}
	ipCmd.AddCommand(ipPhishingCmd)

	ipPhishingCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list <ip_block>",
		Aliases:           []string{"ls"},
		Short:             "List the phishing entries reported on this block",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.ListPhishing,
	}))

	ipPhishingCmd.AddCommand(&cobra.Command{
		Use:               "get <ip_block> <id>",
		Short:             "Show one phishing entry",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.GetPhishing,
	})

	// DDoS mitigation
	ipMitigationCmd := &cobra.Command{
		Use:   "mitigation",
		Short: "Manage DDoS mitigation on the given IP block",
	}
	ipCmd.AddCommand(ipMitigationCmd)

	ipMitigationListCmd := &cobra.Command{
		Use:               "list <ip_block>",
		Aliases:           []string{"ls"},
		Short:             "List the addresses under mitigation",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.ListMitigation,
	}
	ipMitigationListCmd.Flags().StringVar(&ip.MitigationStateFilter, "state", "",
		"Keep only this state: creationPending, ok or removalPending")
	ipMitigationListCmd.Flags().StringVar(&ip.MitigationAutoFilter, "auto", "",
		"Keep only auto-mitigated (true) or manually mitigated (false) addresses")
	ipMitigationCmd.AddCommand(withFilterFlag(ipMitigationListCmd))

	ipMitigationCmd.AddCommand(&cobra.Command{
		Use:               "get <ip_block> <ip>",
		Short:             "Show the mitigation state of an address",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.GetMitigation,
	})

	ipMitigationAddCmd := &cobra.Command{
		Use:               "add <ip_block> <ip>",
		Short:             "Put an address on permanent mitigation",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.AddMitigation,
	}
	addConfirmationFlags(ipMitigationAddCmd, "Print the call that would be made without making it")
	ipMitigationCmd.AddCommand(ipMitigationAddCmd)

	ipMitigationSetCmd := &cobra.Command{
		Use:               "set <ip_block> <ip>",
		Short:             "Turn permanent mitigation on or off for an address already known to the system",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.SetMitigation,
	}
	ipMitigationSetCmd.Flags().BoolVar(&ip.MitigationPermanent, "permanent", false,
		"Keep the traffic of this address in the scrubbing centre at all times")
	addConfirmationFlags(ipMitigationSetCmd, "Print the call that would be made without making it")
	ipMitigationCmd.AddCommand(ipMitigationSetCmd)

	ipMitigationRemoveCmd := &cobra.Command{
		Use:               "remove <ip_block> <ip>",
		Short:             "Remove an address from mitigation",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.RemoveMitigation,
	}
	addConfirmationFlags(ipMitigationRemoveCmd, "Print the call that would be made without making it")
	ipMitigationCmd.AddCommand(ipMitigationRemoveCmd)

	// Auto-mitigation profiles
	ipMitigationProfileCmd := &cobra.Command{
		Use:   "mitigation-profile",
		Short: "Manage how long auto-mitigation stays on after an attack",
	}
	ipCmd.AddCommand(ipMitigationProfileCmd)

	ipMitigationProfileCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list <ip_block>",
		Aliases:           []string{"ls"},
		Short:             "List the auto-mitigation profiles of this block",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.ListMitigationProfiles,
	}))

	ipMitigationProfileCmd.AddCommand(&cobra.Command{
		Use:               "get <ip_block> <ip>",
		Short:             "Show one auto-mitigation profile",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.GetMitigationProfile,
	})

	ipMitigationProfileSetCmd := &cobra.Command{
		Use:               "set <ip_block> <ip>",
		Short:             "Set the auto-mitigation delay of an address, creating its profile if needed",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.SetMitigationProfile,
	}
	ipMitigationProfileSetCmd.Flags().IntVar(&ip.MitigationTimeout, "timeout", 0,
		"Minutes auto-mitigation stays on after an attack: 0, 15, 60, 360 or 1560")
	addConfirmationFlags(ipMitigationProfileSetCmd, "Print the call that would be made without making it")
	ipMitigationProfileCmd.AddCommand(ipMitigationProfileSetCmd)

	ipMitigationProfileDeleteCmd := &cobra.Command{
		Use:               "delete <ip_block> <ip>",
		Short:             "Delete an auto-mitigation profile",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.DeleteMitigationProfile,
	}
	addConfirmationFlags(ipMitigationProfileDeleteCmd, "Print the call that would be made without making it")
	ipMitigationProfileCmd.AddCommand(ipMitigationProfileDeleteCmd)

	// Game anti-DDoS
	ipGameCmd := &cobra.Command{
		Use:   "game",
		Short: "Manage the game anti-DDoS filter on the given IP block",
	}
	ipCmd.AddCommand(ipGameCmd)

	ipGameCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list <ip_block>",
		Aliases:           []string{"ls"},
		Short:             "List the addresses under game anti-DDoS",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.ListGame,
	}))

	ipGameCmd.AddCommand(&cobra.Command{
		Use:               "get <ip_block> <ip>",
		Short:             "Show the game anti-DDoS configuration of an address",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.GetGame,
	})

	ipGameEditCmd := &cobra.Command{
		Use:               "edit <ip_block> <ip>",
		Short:             "Turn UDP firewall mode on or off",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.EditGame,
	}
	ipGameEditCmd.Flags().BoolVar(&ip.GameFirewallMode, "firewall-mode", false,
		"In UDP, let through only the traffic matching a rule")
	addConfirmationFlags(ipGameEditCmd, "Print the call that would be made without making it")
	ipGameCmd.AddCommand(ipGameEditCmd)

	ipGameRuleCmd := &cobra.Command{
		Use:   "rule",
		Short: "Manage game anti-DDoS rules",
	}
	ipGameCmd.AddCommand(ipGameRuleCmd)

	ipGameRuleCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list <ip_block> <ip>",
		Aliases:           []string{"ls"},
		Short:             "List the game anti-DDoS rules of an address",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.ListGameRules,
	}))

	ipGameRuleCmd.AddCommand(&cobra.Command{
		Use:               "get <ip_block> <ip> <id>",
		Short:             "Show one game anti-DDoS rule",
		Args:              cobra.ExactArgs(3),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.GetGameRule,
	})

	ipGameRuleAddCmd := &cobra.Command{
		Use:               "add <ip_block> <ip>",
		Short:             "Open a port range for one game protocol",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.AddGameRule,
	}
	ipGameRuleAddCmd.Flags().StringVar(&ip.GameProtocol, "protocol", "",
		"Game protocol, among the ones this address supports (required)")
	ipGameRuleAddCmd.Flags().StringVar(&ip.GamePorts, "ports", "",
		"Single port (7777) or range (7777-7778) (required)")
	addConfirmationFlags(ipGameRuleAddCmd, "Print the call that would be made without making it")
	ipGameRuleCmd.AddCommand(ipGameRuleAddCmd)

	ipGameRuleDeleteCmd := &cobra.Command{
		Use:               "delete <ip_block> <ip> <id>",
		Short:             "Delete a game anti-DDoS rule",
		Args:              cobra.ExactArgs(3),
		ValidArgsFunction: completion.ServiceList("/v1/ip"),
		Run:               ip.DeleteGameRule,
	}
	addConfirmationFlags(ipGameRuleDeleteCmd, "Print the call that would be made without making it")
	ipGameRuleCmd.AddCommand(ipGameRuleDeleteCmd)

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
