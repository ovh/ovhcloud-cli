// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initLoadbalancerSubCommands(loadbalancerCmd *cobra.Command) {
	// Loadbalancer create
	loadbalancerCmd.AddCommand(getLoadbalancerCreationCmd())

	// Loadbalancer delete
	loadbalancerCmd.AddCommand(&cobra.Command{
		Use:   "delete <loadbalancer_id>",
		Short: "Delete a specific loadbalancer",
		Run:   cloud.DeleteCloudLoadbalancer,
		Args:  cobra.ExactArgs(1),
	})

	// Loadbalancer stats
	loadbalancerCmd.AddCommand(&cobra.Command{
		Use:   "stats <loadbalancer_id>",
		Short: "Get statistics for a loadbalancer",
		Run:   cloud.GetCloudLoadbalancerStats,
		Args:  cobra.ExactArgs(1),
	})

	// Associate floating IP
	loadbalancerCmd.AddCommand(getLoadbalancerAssociateFloatingIpCmd())

	// Create floating IP
	loadbalancerCmd.AddCommand(getLoadbalancerCreateFloatingIpCmd())

	// Listener sub-commands
	initListenerSubCommands(loadbalancerCmd)

	// Pool sub-commands
	initPoolSubCommands(loadbalancerCmd)

	// Health Monitor sub-commands
	initHealthMonitorSubCommands(loadbalancerCmd)

	// L7 Policy sub-commands
	initL7PolicySubCommands(loadbalancerCmd)

	// Log sub-commands
	initLogSubCommands(loadbalancerCmd)
}

// Loadbalancer creation command helpers

func getLoadbalancerCreationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <region>",
		Short: "Create a loadbalancer in the given cloud project",
		Long: `Use this command to create a loadbalancer.
There are three ways to define the parameters:

1. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud network loadbalancer create <region> --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud network loadbalancer create <region> --from-file ./params.json

  Note that you can also pipe the content of the parameters file.

2. Using your default text editor:

	ovhcloud cloud network loadbalancer create <region> --editor

3. Using only CLI flags:

	ovhcloud cloud network loadbalancer create <region> --name my-lb --flavor <flavor_id>
`,
		Run:  cloud.CreateCloudLoadbalancer,
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&cloud.CloudLoadbalancerCreateSpec.Name, "name", "", "Name of the loadbalancer")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerCreateSpec.FlavorId, "flavor", "", "Flavor ID (can be retrieved with 'cloud reference loadbalancer list-flavors <region>')")

	addParameterFileFlags(cmd, false, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/loadbalancing/loadbalancer", "post", cloud.LoadbalancerCreationExample, nil)
	addInteractiveEditorFlag(cmd)
	cmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return cmd
}

func getLoadbalancerAssociateFloatingIpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "associate-floating-ip <loadbalancer_id>",
		Short: "Associate an existing floating IP to a loadbalancer",
		Run:   cloud.AssociateFloatingIpToLoadbalancer,
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&cloud.CloudLoadbalancerAssociateFloatingIpSpec.FloatingIpId, "floating-ip-id", "", "Floating IP ID")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerAssociateFloatingIpSpec.Ip, "ip", "", "Private loadbalancer IP to associate the floating IP with")

	addParameterFileFlags(cmd, false, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/loadbalancing/loadbalancer/{loadBalancerId}/associateFloatingIp", "post", cloud.LoadbalancerAssociateFloatingIpExample, nil)
	addInteractiveEditorFlag(cmd)
	cmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return cmd
}

func getLoadbalancerCreateFloatingIpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-floating-ip <loadbalancer_id>",
		Short: "Create a floating IP and attach it to a loadbalancer",
		Run:   cloud.CreateFloatingIpForLoadbalancer,
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&cloud.CloudLoadbalancerCreateFloatingIpSpec.Ip, "ip", "", "Private loadbalancer IP to associate the floating IP with")

	addParameterFileFlags(cmd, false, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/loadbalancing/loadbalancer/{loadBalancerId}/floatingIp", "post", cloud.LoadbalancerCreateFloatingIpExample, nil)
	addInteractiveEditorFlag(cmd)
	cmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return cmd
}

// Listener sub-commands

func initListenerSubCommands(loadbalancerCmd *cobra.Command) {
	listenerCmd := &cobra.Command{
		Use:   "listener",
		Short: "Manage listeners of loadbalancers",
	}
	loadbalancerCmd.AddCommand(listenerCmd)

	listenerCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all listeners",
		Run:     cloud.ListCloudLoadbalancerListeners,
	}))

	listenerCmd.AddCommand(&cobra.Command{
		Use:   "get <listener_id>",
		Short: "Get a specific listener",
		Run:   cloud.GetCloudLoadbalancerListener,
		Args:  cobra.ExactArgs(1),
	})

	listenerCmd.AddCommand(getListenerCreationCmd())

	listenerEditCmd := &cobra.Command{
		Use:   "edit <listener_id>",
		Short: "Edit a specific listener",
		Run:   cloud.EditCloudLoadbalancerListener,
		Args:  cobra.ExactArgs(1),
	}
	listenerEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerListenerUpdateSpec.Name, "name", "", "Name of the listener")
	listenerEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerListenerUpdateSpec.Description, "description", "", "Description of the listener")
	listenerEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerListenerUpdateSpec.DefaultPoolId, "default-pool-id", "", "Default pool ID")
	listenerEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerListenerUpdateSpec.CertificateId, "certificate-id", "", "Certificate ID")
	addInteractiveEditorFlag(listenerEditCmd)
	listenerCmd.AddCommand(listenerEditCmd)

	listenerCmd.AddCommand(&cobra.Command{
		Use:   "delete <listener_id>",
		Short: "Delete a specific listener",
		Run:   cloud.DeleteCloudLoadbalancerListener,
		Args:  cobra.ExactArgs(1),
	})
}

func getListenerCreationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <region>",
		Short: "Create a listener in the given region",
		Run:   cloud.CreateCloudLoadbalancerListener,
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&cloud.CloudLoadbalancerListenerCreateSpec.LoadbalancerId, "loadbalancer-id", "", "Loadbalancer ID")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerListenerCreateSpec.Name, "name", "", "Name of the listener")
	cmd.Flags().IntVar(&cloud.CloudLoadbalancerListenerCreateSpec.Port, "port", 0, "Port to listen on")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerListenerCreateSpec.Protocol, "protocol", "", "Protocol (http, https, tcp, udp, sctp, prometheus)")

	addParameterFileFlags(cmd, false, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/loadbalancing/listener", "post", cloud.LoadbalancerListenerCreationExample, nil)
	addInteractiveEditorFlag(cmd)
	cmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return cmd
}

// Pool sub-commands

func initPoolSubCommands(loadbalancerCmd *cobra.Command) {
	poolCmd := &cobra.Command{
		Use:   "pool",
		Short: "Manage pools of loadbalancers",
	}
	loadbalancerCmd.AddCommand(poolCmd)

	poolCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all pools",
		Run:     cloud.ListCloudLoadbalancerPools,
	}))

	poolCmd.AddCommand(&cobra.Command{
		Use:   "get <pool_id>",
		Short: "Get a specific pool",
		Run:   cloud.GetCloudLoadbalancerPool,
		Args:  cobra.ExactArgs(1),
	})

	poolCmd.AddCommand(getPoolCreationCmd())

	poolEditCmd := &cobra.Command{
		Use:   "edit <pool_id>",
		Short: "Edit a specific pool",
		Run:   cloud.EditCloudLoadbalancerPool,
		Args:  cobra.ExactArgs(1),
	}
	poolEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerPoolUpdateSpec.Algorithm, "algorithm", "", "Algorithm (roundRobin, leastConnections, sourceIp)")
	poolEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerPoolUpdateSpec.Name, "name", "", "Name of the pool")
	addInteractiveEditorFlag(poolEditCmd)
	poolCmd.AddCommand(poolEditCmd)

	poolCmd.AddCommand(&cobra.Command{
		Use:   "delete <pool_id>",
		Short: "Delete a specific pool",
		Run:   cloud.DeleteCloudLoadbalancerPool,
		Args:  cobra.ExactArgs(1),
	})

	// Pool Member sub-commands
	initPoolMemberSubCommands(poolCmd)
}

func getPoolCreationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <region>",
		Short: "Create a pool in the given region",
		Run:   cloud.CreateCloudLoadbalancerPool,
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&cloud.CloudLoadbalancerPoolCreateSpec.Algorithm, "algorithm", "", "Algorithm (roundRobin, leastConnections, sourceIp)")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerPoolCreateSpec.ListenerId, "listener-id", "", "Listener ID")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerPoolCreateSpec.LoadbalancerId, "loadbalancer-id", "", "Loadbalancer ID")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerPoolCreateSpec.Name, "name", "", "Name of the pool")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerPoolCreateSpec.Protocol, "protocol", "", "Protocol (http, https, tcp, udp, sctp, prometheus)")

	addParameterFileFlags(cmd, false, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/loadbalancing/pool", "post", cloud.LoadbalancerPoolCreationExample, nil)
	addInteractiveEditorFlag(cmd)
	cmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return cmd
}

func initPoolMemberSubCommands(poolCmd *cobra.Command) {
	memberCmd := &cobra.Command{
		Use:   "member",
		Short: "Manage members of a loadbalancer pool",
	}
	poolCmd.AddCommand(memberCmd)

	memberCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <pool_id>",
		Aliases: []string{"ls"},
		Short:   "List members of a specific pool",
		Run:     cloud.ListCloudLoadbalancerPoolMembers,
		Args:    cobra.ExactArgs(1),
	}))

	memberCmd.AddCommand(&cobra.Command{
		Use:   "get <pool_id> <member_id>",
		Short: "Get a specific pool member",
		Run:   cloud.GetCloudLoadbalancerPoolMember,
		Args:  cobra.ExactArgs(2),
	})

	memberCmd.AddCommand(getPoolMemberCreationCmd())

	memberEditCmd := &cobra.Command{
		Use:   "edit <pool_id> <member_id>",
		Short: "Edit a specific pool member",
		Run:   cloud.EditCloudLoadbalancerPoolMember,
		Args:  cobra.ExactArgs(2),
	}
	memberEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerPoolMemberUpdateSpec.Name, "name", "", "Name of the member")
	memberEditCmd.Flags().IntVar(&cloud.CloudLoadbalancerPoolMemberUpdateSpec.Weight, "weight", 0, "Weight of the member (1-256)")
	addInteractiveEditorFlag(memberEditCmd)
	memberCmd.AddCommand(memberEditCmd)

	memberCmd.AddCommand(&cobra.Command{
		Use:   "delete <pool_id> <member_id>",
		Short: "Delete a specific pool member",
		Run:   cloud.DeleteCloudLoadbalancerPoolMember,
		Args:  cobra.ExactArgs(2),
	})
}

func getPoolMemberCreationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <pool_id>",
		Short: "Create member(s) in a specific pool",
		Run:   cloud.CreateCloudLoadbalancerPoolMember,
		Args:  cobra.ExactArgs(1),
	}

	addParameterFileFlags(cmd, false, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/loadbalancing/pool/{poolId}/member", "post", cloud.LoadbalancerPoolMemberCreationExample, nil)
	addInteractiveEditorFlag(cmd)
	cmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return cmd
}

// Health Monitor sub-commands

func initHealthMonitorSubCommands(loadbalancerCmd *cobra.Command) {
	healthMonitorCmd := &cobra.Command{
		Use:   "health-monitor",
		Short: "Manage health monitors of loadbalancers",
	}
	loadbalancerCmd.AddCommand(healthMonitorCmd)

	healthMonitorCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all health monitors",
		Run:     cloud.ListCloudLoadbalancerHealthMonitors,
	}))

	healthMonitorCmd.AddCommand(&cobra.Command{
		Use:   "get <health_monitor_id>",
		Short: "Get a specific health monitor",
		Run:   cloud.GetCloudLoadbalancerHealthMonitor,
		Args:  cobra.ExactArgs(1),
	})

	healthMonitorCmd.AddCommand(getHealthMonitorCreationCmd())

	healthMonitorEditCmd := &cobra.Command{
		Use:   "edit <health_monitor_id>",
		Short: "Edit a specific health monitor",
		Run:   cloud.EditCloudLoadbalancerHealthMonitor,
		Args:  cobra.ExactArgs(1),
	}
	healthMonitorEditCmd.Flags().IntVar(&cloud.CloudLoadbalancerHealthMonitorUpdateSpec.Delay, "delay", 0, "Duration between sending probes to members, in seconds")
	healthMonitorEditCmd.Flags().IntVar(&cloud.CloudLoadbalancerHealthMonitorUpdateSpec.MaxRetries, "max-retries", 0, "Number of successful checks before changing status to ONLINE")
	healthMonitorEditCmd.Flags().IntVar(&cloud.CloudLoadbalancerHealthMonitorUpdateSpec.MaxRetriesDown, "max-retries-down", 0, "Number of allowed check failures before changing status to ERROR")
	healthMonitorEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerHealthMonitorUpdateSpec.Name, "name", "", "Name of the health monitor")
	healthMonitorEditCmd.Flags().IntVar(&cloud.CloudLoadbalancerHealthMonitorUpdateSpec.Timeout, "timeout", 0, "Maximum time in seconds to connect before timeout")
	addInteractiveEditorFlag(healthMonitorEditCmd)
	healthMonitorCmd.AddCommand(healthMonitorEditCmd)

	healthMonitorCmd.AddCommand(&cobra.Command{
		Use:   "delete <health_monitor_id>",
		Short: "Delete a specific health monitor",
		Run:   cloud.DeleteCloudLoadbalancerHealthMonitor,
		Args:  cobra.ExactArgs(1),
	})
}

func getHealthMonitorCreationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <region>",
		Short: "Create a health monitor in the given region",
		Run:   cloud.CreateCloudLoadbalancerHealthMonitor,
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().IntVar(&cloud.CloudLoadbalancerHealthMonitorCreateSpec.Delay, "delay", 0, "Duration between sending probes to members, in seconds")
	cmd.Flags().IntVar(&cloud.CloudLoadbalancerHealthMonitorCreateSpec.MaxRetries, "max-retries", 0, "Number of successful checks before changing status to ONLINE")
	cmd.Flags().IntVar(&cloud.CloudLoadbalancerHealthMonitorCreateSpec.MaxRetriesDown, "max-retries-down", 0, "Number of allowed check failures before changing status to ERROR")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerHealthMonitorCreateSpec.MonitorType, "monitor-type", "", "Type of the monitor (http, https, ping, tcp, tls-hello, udp-connect, sctp)")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerHealthMonitorCreateSpec.Name, "name", "", "Name of the health monitor")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerHealthMonitorCreateSpec.PoolId, "pool-id", "", "Pool ID")
	cmd.Flags().IntVar(&cloud.CloudLoadbalancerHealthMonitorCreateSpec.Timeout, "timeout", 0, "Maximum time in seconds to connect before timeout")

	addParameterFileFlags(cmd, false, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/loadbalancing/healthMonitor", "post", cloud.LoadbalancerHealthMonitorCreationExample, nil)
	addInteractiveEditorFlag(cmd)
	cmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return cmd
}

// L7 Policy sub-commands

func initL7PolicySubCommands(loadbalancerCmd *cobra.Command) {
	l7PolicyCmd := &cobra.Command{
		Use:   "l7policy",
		Short: "Manage L7 policies of loadbalancers",
	}
	loadbalancerCmd.AddCommand(l7PolicyCmd)

	l7PolicyCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all L7 policies",
		Run:     cloud.ListCloudLoadbalancerL7Policies,
	}))

	l7PolicyCmd.AddCommand(&cobra.Command{
		Use:   "get <l7policy_id>",
		Short: "Get a specific L7 policy",
		Run:   cloud.GetCloudLoadbalancerL7Policy,
		Args:  cobra.ExactArgs(1),
	})

	l7PolicyCmd.AddCommand(getL7PolicyCreationCmd())

	l7PolicyEditCmd := &cobra.Command{
		Use:   "edit <l7policy_id>",
		Short: "Edit a specific L7 policy",
		Run:   cloud.EditCloudLoadbalancerL7Policy,
		Args:  cobra.ExactArgs(1),
	}
	l7PolicyEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerL7PolicyUpdateSpec.Action, "action", "", "L7 policy action")
	l7PolicyEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerL7PolicyUpdateSpec.Description, "description", "", "Description")
	l7PolicyEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerL7PolicyUpdateSpec.ListenerId, "listener-id", "", "Listener ID")
	l7PolicyEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerL7PolicyUpdateSpec.Name, "name", "", "Name of the L7 policy")
	l7PolicyEditCmd.Flags().IntVar(&cloud.CloudLoadbalancerL7PolicyUpdateSpec.Position, "position", 0, "Position on the listener")
	l7PolicyEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerL7PolicyUpdateSpec.RedirectPoolId, "redirect-pool-id", "", "Redirect pool ID")
	l7PolicyEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerL7PolicyUpdateSpec.RedirectPrefix, "redirect-prefix", "", "Redirect prefix URL")
	l7PolicyEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerL7PolicyUpdateSpec.RedirectUrl, "redirect-url", "", "Redirect URL")
	addInteractiveEditorFlag(l7PolicyEditCmd)
	l7PolicyCmd.AddCommand(l7PolicyEditCmd)

	l7PolicyCmd.AddCommand(&cobra.Command{
		Use:   "delete <l7policy_id>",
		Short: "Delete a specific L7 policy",
		Run:   cloud.DeleteCloudLoadbalancerL7Policy,
		Args:  cobra.ExactArgs(1),
	})

	// L7 Rule sub-commands
	initL7RuleSubCommands(l7PolicyCmd)
}

func getL7PolicyCreationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <region>",
		Short: "Create an L7 policy in the given region",
		Run:   cloud.CreateCloudLoadbalancerL7Policy,
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&cloud.CloudLoadbalancerL7PolicyCreateSpec.Action, "action", "", "L7 policy action (redirectToPool, redirectToUrl, redirectPrefix, reject)")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerL7PolicyCreateSpec.Description, "description", "", "Description")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerL7PolicyCreateSpec.ListenerId, "listener-id", "", "Listener ID")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerL7PolicyCreateSpec.Name, "name", "", "Name of the L7 policy")
	cmd.Flags().IntVar(&cloud.CloudLoadbalancerL7PolicyCreateSpec.Position, "position", 0, "Position on the listener")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerL7PolicyCreateSpec.RedirectPoolId, "redirect-pool-id", "", "Redirect pool ID")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerL7PolicyCreateSpec.RedirectPrefix, "redirect-prefix", "", "Redirect prefix URL")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerL7PolicyCreateSpec.RedirectUrl, "redirect-url", "", "Redirect URL")

	addParameterFileFlags(cmd, false, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/loadbalancing/l7Policy", "post", cloud.LoadbalancerL7PolicyCreationExample, nil)
	addInteractiveEditorFlag(cmd)
	cmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return cmd
}

func initL7RuleSubCommands(l7PolicyCmd *cobra.Command) {
	l7RuleCmd := &cobra.Command{
		Use:   "l7rule",
		Short: "Manage L7 rules of a loadbalancer L7 policy",
	}
	l7PolicyCmd.AddCommand(l7RuleCmd)

	l7RuleCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <l7policy_id>",
		Aliases: []string{"ls"},
		Short:   "List L7 rules of a specific L7 policy",
		Run:     cloud.ListCloudLoadbalancerL7Rules,
		Args:    cobra.ExactArgs(1),
	}))

	l7RuleCmd.AddCommand(&cobra.Command{
		Use:   "get <l7policy_id> <l7rule_id>",
		Short: "Get a specific L7 rule",
		Run:   cloud.GetCloudLoadbalancerL7Rule,
		Args:  cobra.ExactArgs(2),
	})

	l7RuleCmd.AddCommand(getL7RuleCreationCmd())

	l7RuleEditCmd := &cobra.Command{
		Use:   "edit <l7policy_id> <l7rule_id>",
		Short: "Edit a specific L7 rule",
		Run:   cloud.EditCloudLoadbalancerL7Rule,
		Args:  cobra.ExactArgs(2),
	}
	l7RuleEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerL7RuleUpdateSpec.CompareType, "compare-type", "", "Comparison type")
	l7RuleEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerL7RuleUpdateSpec.Key, "key", "", "Key to use for comparison")
	l7RuleEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerL7RuleUpdateSpec.RuleType, "rule-type", "", "Rule type")
	l7RuleEditCmd.Flags().StringVar(&cloud.CloudLoadbalancerL7RuleUpdateSpec.Value, "value", "", "Value to compare")
	addInteractiveEditorFlag(l7RuleEditCmd)
	l7RuleCmd.AddCommand(l7RuleEditCmd)

	l7RuleCmd.AddCommand(&cobra.Command{
		Use:   "delete <l7policy_id> <l7rule_id>",
		Short: "Delete a specific L7 rule",
		Run:   cloud.DeleteCloudLoadbalancerL7Rule,
		Args:  cobra.ExactArgs(2),
	})
}

func getL7RuleCreationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <l7policy_id>",
		Short: "Create an L7 rule in a specific L7 policy",
		Run:   cloud.CreateCloudLoadbalancerL7Rule,
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&cloud.CloudLoadbalancerL7RuleCreateSpec.CompareType, "compare-type", "", "Comparison type (contains, endsWith, equalTo, regex, startsWith)")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerL7RuleCreateSpec.Key, "key", "", "Key to use for comparison")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerL7RuleCreateSpec.RuleType, "rule-type", "", "Rule type (cookie, fileType, header, hostName, path, sslConnHasCert, sslDNField, sslVerifyResult)")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerL7RuleCreateSpec.Value, "value", "", "Value to compare")

	addParameterFileFlags(cmd, false, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/loadbalancing/l7Policy/{l7PolicyId}/l7Rule", "post", cloud.LoadbalancerL7RuleCreationExample, nil)
	addInteractiveEditorFlag(cmd)
	cmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return cmd
}

// Log sub-commands

func initLogSubCommands(loadbalancerCmd *cobra.Command) {
	logCmd := &cobra.Command{
		Use:   "log",
		Short: "Manage loadbalancer logs",
	}
	loadbalancerCmd.AddCommand(logCmd)

	logCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list-kinds <region>",
		Short: "List available log kinds",
		Run:   cloud.ListCloudLoadbalancerLogKinds,
		Args:  cobra.ExactArgs(1),
	}))

	logCmd.AddCommand(&cobra.Command{
		Use:   "get-kind <region> <kind_name>",
		Short: "Get a specific log kind",
		Run:   cloud.GetCloudLoadbalancerLogKind,
		Args:  cobra.ExactArgs(2),
	})

	logCmd.AddCommand(&cobra.Command{
		Use:   "generate-url <loadbalancer_id>",
		Short: "Generate a temporary URL to retrieve logs",
		Run:   cloud.GenerateCloudLoadbalancerLogURL,
		Args:  cobra.ExactArgs(1),
	})

	// Log Subscription sub-commands
	subscriptionCmd := &cobra.Command{
		Use:   "subscription",
		Short: "Manage log subscriptions for a loadbalancer",
	}
	logCmd.AddCommand(subscriptionCmd)

	subscriptionCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <loadbalancer_id>",
		Aliases: []string{"ls"},
		Short:   "List log subscriptions of a loadbalancer",
		Run:     cloud.ListCloudLoadbalancerLogSubscriptions,
		Args:    cobra.ExactArgs(1),
	}))

	subscriptionCmd.AddCommand(&cobra.Command{
		Use:   "get <loadbalancer_id> <subscription_id>",
		Short: "Get a specific log subscription",
		Run:   cloud.GetCloudLoadbalancerLogSubscription,
		Args:  cobra.ExactArgs(2),
	})

	subscriptionCmd.AddCommand(getLogSubscriptionCreationCmd())

	subscriptionCmd.AddCommand(&cobra.Command{
		Use:   "delete <loadbalancer_id> <subscription_id>",
		Short: "Delete a log subscription",
		Run:   cloud.DeleteCloudLoadbalancerLogSubscription,
		Args:  cobra.ExactArgs(2),
	})
}

func getLogSubscriptionCreationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <loadbalancer_id>",
		Short: "Create a log subscription for a loadbalancer",
		Run:   cloud.CreateCloudLoadbalancerLogSubscription,
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&cloud.CloudLoadbalancerLogSubscriptionCreateSpec.Kind, "kind", "", "Log kind (e.g., haproxy)")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerLogSubscriptionCreateSpec.StreamId, "stream-id", "", "LDP stream ID")

	addParameterFileFlags(cmd, false, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/loadbalancing/loadbalancer/{loadBalancerId}/log/subscription", "post", cloud.LoadbalancerLogSubscriptionCreationExample, nil)
	addInteractiveEditorFlag(cmd)
	cmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return cmd
}
