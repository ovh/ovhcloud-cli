// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudLoadbalancerCommand(cloudCmd *cobra.Command) {
	loadbalancerCmd := &cobra.Command{
		Use:   "loadbalancer",
		Short: "Manage loadbalancers in the given cloud project",
	}
	loadbalancerCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	loadbalancerListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your loadbalancers",
		Run:     cloud.ListCloudLoadbalancers,
	}
	loadbalancerCmd.AddCommand(withFilterFlag(loadbalancerListCmd))

	loadbalancerCmd.AddCommand(&cobra.Command{
		Use:   "get <loadbalancer_id>",
		Short: "Get a specific loadbalancer",
		Run:   cloud.GetCloudLoadbalancer,
		Args:  cobra.ExactArgs(1),
	})

	editLoadbalancerCmd := &cobra.Command{
		Use:   "edit <loadbalancer_id>",
		Short: "Edit the given loadbalancer",
		Run:   cloud.EditCloudLoadbalancer,
		Args:  cobra.ExactArgs(1),
	}
	editLoadbalancerCmd.Flags().StringVar(&cloud.CloudLoadbalancerUpdateSpec.TargetSpec.Name, "name", "", "Name of the loadbalancer")
	editLoadbalancerCmd.Flags().StringVar(&cloud.CloudLoadbalancerUpdateSpec.TargetSpec.Description, "description", "", "Description of the loadbalancer")
	editLoadbalancerCmd.Flags().StringVar(&cloud.CloudLoadbalancerUpdateSpec.Size, "size", "", "Size of the loadbalancer (SMALL, MEDIUM, LARGE, XL). A different size resizes the loadbalancer in place and can briefly interrupt traffic")
	addInteractiveEditorFlag(editLoadbalancerCmd)
	loadbalancerCmd.AddCommand(editLoadbalancerCmd)

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

	// L7 Policy sub-commands
	initL7PolicySubCommands(loadbalancerCmd)

	// Log sub-commands
	initLogSubCommands(loadbalancerCmd)

	// Flavor reference data commands
	flavorCmd := &cobra.Command{
		Use:   "flavor",
		Short: "List available loadbalancer flavors in the given cloud project",
	}
	loadbalancerCmd.AddCommand(flavorCmd)

	flavorCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list [region (GRA9, BHS5, ...)]",
		Short: "List available loadbalancer flavors in the given cloud project, optionally restricted to one region",
		Run:   cloud.ListLoadbalancerFlavors,
		Args:  cobra.MaximumNArgs(1),
	}))

	flavorCmd.AddCommand(&cobra.Command{
		Use:   "get [region] <flavor_name>",
		Short: "Get details of a specific loadbalancer flavor, optionally restricted to one region",
		Long: `Get details of a specific loadbalancer flavor, using its name (SMALL, MEDIUM, ...).
Without a region, a flavor sold in several regions is listed once per region.`,
		Run:  cloud.GetLoadbalancerFlavor,
		Args: cobra.RangeArgs(1, 2),
	})

	cloudCmd.AddCommand(loadbalancerCmd)
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

	ovhcloud cloud loadbalancer create <region> --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud loadbalancer create <region> --from-file ./params.json

  Note that you can also pipe the content of the parameters file.

2. Using your default text editor:

	ovhcloud cloud loadbalancer create <region> --editor

3. Using only CLI flags:

	ovhcloud cloud loadbalancer create <region> --name my-lb --size SMALL --subnet-id <subnet_id>
`,
		Run:  cloud.CreateCloudLoadbalancer,
		Args: cobra.ExactArgs(1),
	}

	spec := &cloud.CloudLoadbalancerCreateSpec.TargetSpec
	cmd.Flags().StringVar(&spec.Name, "name", "", "Name of the loadbalancer")
	cmd.Flags().StringVar(&spec.Description, "description", "", "Description of the loadbalancer")
	cmd.Flags().StringVar(&cloud.CloudLoadbalancerCreateSpec.Size, "size", "", "Size of the loadbalancer (SMALL, MEDIUM, LARGE, XL)")
	cmd.Flags().StringVar(&spec.Network.Id, "network-id", "", "Network ID (optional, defaults to the network of the subnet)")
	cmd.Flags().StringVar(&spec.Network.SubnetId, "subnet-id", "", "Subnet ID")
	cmd.Flags().StringVar(&spec.Network.Ip, "network-ip", "", "IP of the loadbalancer VIP: an address inside the subnet CIDR pins the fixed address, an address outside it designates an existing floating IP to associate (requires the loadbalancer/usePublicIP IAM action)")

	addParameterFileFlags(cmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/loadbalancer", "post", cloud.LoadbalancerCreationExample, nil)
	addInteractiveEditorFlag(cmd)
	markFlagsMutuallyExclusive(cmd, "from-file", "editor")

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
	markFlagsMutuallyExclusive(cmd, "from-file", "editor")

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
	markFlagsMutuallyExclusive(cmd, "from-file", "editor")

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
		Use:     "list <loadbalancer_id>",
		Aliases: []string{"ls"},
		Short:   "List the listeners of a loadbalancer",
		Run:     cloud.ListCloudLoadbalancerListeners,
		Args:    cobra.ExactArgs(1),
	}))

	listenerCmd.AddCommand(&cobra.Command{
		Use:   "get <loadbalancer_id> <listener_id>",
		Short: "Get a specific listener",
		Run:   cloud.GetCloudLoadbalancerListener,
		Args:  cobra.ExactArgs(2),
	})

	listenerCmd.AddCommand(getListenerCreationCmd())

	listenerEditCmd := &cobra.Command{
		Use:   "edit <loadbalancer_id> <listener_id>",
		Short: "Edit a specific listener",
		Run:   cloud.EditCloudLoadbalancerListener,
		Args:  cobra.ExactArgs(2),
	}
	editSpec := &cloud.CloudLoadbalancerListenerUpdateSpec.TargetSpec
	listenerEditCmd.Flags().StringVar(&editSpec.Name, "name", "", "Name of the listener")
	listenerEditCmd.Flags().StringVar(&editSpec.Description, "description", "", "Description of the listener")
	listenerEditCmd.Flags().StringVar(&editSpec.DefaultPool.Id, "default-pool-id", "", "Default pool ID")
	listenerEditCmd.Flags().StringVar(&editSpec.DefaultTlsContainerRef, "default-tls-container-ref", "", "Reference of the default TLS certificate container")
	addInteractiveEditorFlag(listenerEditCmd)
	listenerCmd.AddCommand(listenerEditCmd)

	listenerCmd.AddCommand(&cobra.Command{
		Use:   "delete <loadbalancer_id> <listener_id>",
		Short: "Delete a specific listener",
		Run:   cloud.DeleteCloudLoadbalancerListener,
		Args:  cobra.ExactArgs(2),
	})
}

func getListenerCreationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <loadbalancer_id>",
		Short: "Create a listener on a loadbalancer",
		Long: `Use this command to create a listener, optionally with its default pool.
There are three ways to define the parameters:

1. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud loadbalancer listener create <loadbalancer_id> --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud loadbalancer listener create <loadbalancer_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file.

2. Using your default text editor:

	ovhcloud cloud loadbalancer listener create <loadbalancer_id> --editor

3. Using only CLI flags:

	ovhcloud cloud loadbalancer listener create <loadbalancer_id> --name my-listener --protocol HTTP --port 80 --default-pool-id <pool_id>
`,
		Run:  cloud.CreateCloudLoadbalancerListener,
		Args: cobra.ExactArgs(1),
	}

	spec := &cloud.CloudLoadbalancerListenerCreateSpec.TargetSpec
	cmd.Flags().StringVar(&spec.Name, "name", "", "Name of the listener")
	cmd.Flags().StringVar(&spec.Description, "description", "", "Description of the listener")
	cmd.Flags().IntVar(&spec.ProtocolPort, "port", 0, "Port to listen on")
	cmd.Flags().StringVar(&spec.Protocol, "protocol", "", "Protocol (HTTP, HTTPS, TERMINATED_HTTPS, TCP, UDP, SCTP)")
	cmd.Flags().StringVar(&spec.DefaultPool.Id, "default-pool-id", "", "Default pool ID")
	cmd.Flags().StringVar(&spec.DefaultTlsContainerRef, "default-tls-container-ref", "", "Reference of the default TLS certificate container")

	addParameterFileFlags(cmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/loadbalancer/{loadbalancerId}/listener", "post", cloud.LoadbalancerListenerCreationExample, nil)
	addInteractiveEditorFlag(cmd)
	markFlagsMutuallyExclusive(cmd, "from-file", "editor")

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
		Use:     "list <loadbalancer_id>",
		Aliases: []string{"ls"},
		Short:   "List the pools of a loadbalancer",
		Run:     cloud.ListCloudLoadbalancerPools,
		Args:    cobra.ExactArgs(1),
	}))

	poolCmd.AddCommand(&cobra.Command{
		Use:   "get <loadbalancer_id> <pool_id>",
		Short: "Get a specific pool",
		Run:   cloud.GetCloudLoadbalancerPool,
		Args:  cobra.ExactArgs(2),
	})

	poolCmd.AddCommand(getPoolCreationCmd())

	poolEditCmd := &cobra.Command{
		Use:   "edit <loadbalancer_id> <pool_id>",
		Short: "Edit a specific pool",
		Long: `Edit a specific pool, including the health monitor nested in it.
To remove the health monitor, use --editor and delete the "healthMonitor" block.`,
		Run:  cloud.EditCloudLoadbalancerPool,
		Args: cobra.ExactArgs(2),
	}
	editSpec := &cloud.CloudLoadbalancerPoolUpdateSpec.TargetSpec
	poolEditCmd.Flags().StringVar(&editSpec.Name, "name", "", "Name of the pool")
	poolEditCmd.Flags().StringVar(&editSpec.Description, "description", "", "Description of the pool")
	poolEditCmd.Flags().StringVar(&editSpec.Algorithm, "algorithm", "", "Algorithm (ROUND_ROBIN, LEAST_CONNECTIONS, SOURCE_IP)")
	addPoolHealthMonitorFlags(poolEditCmd, &editSpec.HealthMonitor)
	addInteractiveEditorFlag(poolEditCmd)
	poolCmd.AddCommand(poolEditCmd)

	poolCmd.AddCommand(&cobra.Command{
		Use:   "delete <loadbalancer_id> <pool_id>",
		Short: "Delete a specific pool",
		Run:   cloud.DeleteCloudLoadbalancerPool,
		Args:  cobra.ExactArgs(2),
	})

	// Pool Member sub-commands
	initPoolMemberSubCommands(poolCmd)
}

func getPoolCreationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <loadbalancer_id>",
		Short: "Create a pool on a loadbalancer",
		Long: `Use this command to create a pool, optionally with its health monitor.
There are three ways to define the parameters:

1. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud loadbalancer pool create <loadbalancer_id> --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud loadbalancer pool create <loadbalancer_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file.

2. Using your default text editor:

	ovhcloud cloud loadbalancer pool create <loadbalancer_id> --editor

3. Using only CLI flags, for example a pool created with a full health monitor in one command:

	ovhcloud cloud loadbalancer pool create <loadbalancer_id> --name my-pool --algorithm ROUND_ROBIN --protocol HTTP \
		--health-monitor-type HTTP --health-monitor-delay 10 --health-monitor-timeout 5 --health-monitor-max-retries 3 \
		--health-monitor-url-path /healthz --health-monitor-expected-codes 200

  When any --health-monitor-* flag is used with flags only, type, delay, timeout and max-retries are all required.
`,
		Run:  cloud.CreateCloudLoadbalancerPool,
		Args: cobra.ExactArgs(1),
	}

	spec := &cloud.CloudLoadbalancerPoolCreateSpec.TargetSpec
	cmd.Flags().StringVar(&spec.Name, "name", "", "Name of the pool")
	cmd.Flags().StringVar(&spec.Description, "description", "", "Description of the pool")
	cmd.Flags().StringVar(&spec.Algorithm, "algorithm", "", "Algorithm (ROUND_ROBIN, LEAST_CONNECTIONS, SOURCE_IP)")
	cmd.Flags().StringVar(&spec.Protocol, "protocol", "", "Protocol (HTTP, HTTPS, PROXY, PROXYV2, TCP, UDP, SCTP)")
	addPoolHealthMonitorFlags(cmd, &spec.HealthMonitor)

	addParameterFileFlags(cmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/loadbalancer/{loadbalancerId}/pool", "post", cloud.LoadbalancerPoolCreationExample, nil)
	addInteractiveEditorFlag(cmd)
	markFlagsMutuallyExclusive(cmd, "from-file", "editor")

	return cmd
}

// addPoolHealthMonitorFlags registers the flags describing the health monitor nested in a pool.
func addPoolHealthMonitorFlags(cmd *cobra.Command, hm *cloud.PoolHealthMonitorParams) {
	cmd.Flags().StringVar(&hm.Type, "health-monitor-type", "", "Type of the health monitor (HTTP, HTTPS, PING, TCP, TLS_HELLO, UDP_CONNECT, SCTP)")
	cmd.Flags().IntVar(&hm.Delay, "health-monitor-delay", 0, "Time in seconds between sending health check probes to members")
	cmd.Flags().IntVar(&hm.Timeout, "health-monitor-timeout", 0, "Maximum time in seconds to wait for a health check reply (must be less than or equal to the delay)")
	cmd.Flags().IntVar(&hm.MaxRetries, "health-monitor-max-retries", 0, "Number of consecutive health check failures before marking a member as INACTIVE (1-10)")
	cmd.Flags().IntVar(&hm.MaxRetriesDown, "health-monitor-max-retries-down", 0, "Number of consecutive health check failures before marking a member as ERROR (1-10)")
	cmd.Flags().StringVar(&hm.HttpMethod, "health-monitor-http-method", "", "HTTP method used by HTTP and HTTPS health monitors (GET, HEAD, POST, ...)")
	cmd.Flags().StringVar(&hm.UrlPath, "health-monitor-url-path", "", "URL path requested by HTTP and HTTPS health monitors")
	cmd.Flags().StringVar(&hm.ExpectedCodes, "health-monitor-expected-codes", "", "Expected HTTP status codes of HTTP and HTTPS health monitors (e.g. 200, 200-204)")
}

func initPoolMemberSubCommands(poolCmd *cobra.Command) {
	memberCmd := &cobra.Command{
		Use:   "member",
		Short: "Manage members of a loadbalancer pool",
	}
	poolCmd.AddCommand(memberCmd)

	memberCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <loadbalancer_id> <pool_id>",
		Aliases: []string{"ls"},
		Short:   "List members of a specific pool",
		Run:     cloud.ListCloudLoadbalancerPoolMembers,
		Args:    cobra.ExactArgs(2),
	}))

	memberCmd.AddCommand(&cobra.Command{
		Use:   "get <loadbalancer_id> <pool_id> <member_id>",
		Short: "Get a specific pool member",
		Run:   cloud.GetCloudLoadbalancerPoolMember,
		Args:  cobra.ExactArgs(3),
	})

	memberCmd.AddCommand(getPoolMemberCreationCmd())

	memberEditCmd := &cobra.Command{
		Use:   "edit <loadbalancer_id> <pool_id> <member_id>",
		Short: "Edit a specific pool member",
		Run:   cloud.EditCloudLoadbalancerPoolMember,
		Args:  cobra.ExactArgs(3),
	}
	editSpec := &cloud.CloudLoadbalancerPoolMemberUpdateSpec.TargetSpec
	memberEditCmd.Flags().StringVar(&editSpec.Name, "name", "", "Name of the member")
	memberEditCmd.Flags().IntVar(&editSpec.Weight, "weight", 0, "Weight of the member (1-256)")
	memberEditCmd.Flags().BoolVar(&editSpec.Backup, "backup", false, "Whether this is a backup member, only used when all the other members are down")
	addInteractiveEditorFlag(memberEditCmd)
	memberCmd.AddCommand(memberEditCmd)

	memberCmd.AddCommand(&cobra.Command{
		Use:   "delete <loadbalancer_id> <pool_id> <member_id>",
		Short: "Delete a specific pool member",
		Run:   cloud.DeleteCloudLoadbalancerPoolMember,
		Args:  cobra.ExactArgs(3),
	})
}

func getPoolMemberCreationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <loadbalancer_id> <pool_id>",
		Short: "Add a member to a pool",
		Long: `Use this command to add a member to a pool.
There are three ways to define the parameters:

1. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud loadbalancer pool member create <loadbalancer_id> <pool_id> --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud loadbalancer pool member create <loadbalancer_id> <pool_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file.

2. Using your default text editor:

	ovhcloud cloud loadbalancer pool member create <loadbalancer_id> <pool_id> --editor

3. Using only CLI flags:

	ovhcloud cloud loadbalancer pool member create <loadbalancer_id> <pool_id> --address 10.0.0.10 --protocol-port 80 --name my-member --subnet-id <subnet_id>
`,
		Run:  cloud.CreateCloudLoadbalancerPoolMember,
		Args: cobra.ExactArgs(2),
	}

	spec := &cloud.CloudLoadbalancerPoolMemberCreateSpec.TargetSpec
	cmd.Flags().StringVar(&spec.Address, "address", "", "IP address of the member")
	cmd.Flags().IntVar(&spec.ProtocolPort, "protocol-port", 0, "Protocol port number of the member")
	cmd.Flags().StringVar(&spec.Name, "name", "", "Name of the member")
	cmd.Flags().IntVar(&spec.Weight, "weight", 0, "Weight of the member (1-256)")
	cmd.Flags().BoolVar(&spec.Backup, "backup", false, "Whether this is a backup member, only used when all the other members are down")
	cmd.Flags().StringVar(&spec.Subnet.Id, "subnet-id", "", "ID of the subnet of the member")

	addParameterFileFlags(cmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/loadbalancer/{loadbalancerId}/pool/{poolId}/member", "post", cloud.LoadbalancerPoolMemberCreationExample, nil)
	addInteractiveEditorFlag(cmd)
	markFlagsMutuallyExclusive(cmd, "from-file", "editor")

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
		Use:     "list <loadbalancer_id> <listener_id>",
		Aliases: []string{"ls"},
		Short:   "List the L7 policies of a listener",
		Run:     cloud.ListCloudLoadbalancerL7Policies,
		Args:    cobra.ExactArgs(2),
	}))

	l7PolicyCmd.AddCommand(&cobra.Command{
		Use:   "get <loadbalancer_id> <listener_id> <l7policy_id>",
		Short: "Get a specific L7 policy",
		Run:   cloud.GetCloudLoadbalancerL7Policy,
		Args:  cobra.ExactArgs(3),
	})

	l7PolicyCmd.AddCommand(getL7PolicyCreationCmd())

	l7PolicyEditCmd := &cobra.Command{
		Use:   "edit <loadbalancer_id> <listener_id> <l7policy_id>",
		Short: "Edit a specific L7 policy",
		Long: `Edit a specific L7 policy, including the rules nested in it.
The flags leave the rules of the policy untouched. To add, change or remove rules, use --editor.`,
		Run:  cloud.EditCloudLoadbalancerL7Policy,
		Args: cobra.ExactArgs(3),
	}
	editSpec := &cloud.CloudLoadbalancerL7PolicyUpdateSpec.TargetSpec
	l7PolicyEditCmd.Flags().StringVar(&editSpec.Action, "action", "", "L7 policy action (REDIRECT_TO_POOL, REDIRECT_TO_URL, REDIRECT_PREFIX, REJECT)")
	l7PolicyEditCmd.Flags().StringVar(&editSpec.Name, "name", "", "Name of the L7 policy")
	l7PolicyEditCmd.Flags().StringVar(&editSpec.Description, "description", "", "Description")
	l7PolicyEditCmd.Flags().IntVar(&editSpec.Position, "position", 0, "Position on the listener")
	l7PolicyEditCmd.Flags().StringVar(&editSpec.RedirectPool.Id, "redirect-pool-id", "", "Pool ID to redirect to")
	l7PolicyEditCmd.Flags().StringVar(&editSpec.RedirectUrl, "redirect-url", "", "URL to redirect to")
	l7PolicyEditCmd.Flags().StringVar(&editSpec.RedirectPrefix, "redirect-prefix", "", "Prefix URL to redirect to")
	l7PolicyEditCmd.Flags().IntVar(&editSpec.RedirectHttpCode, "redirect-http-code", 0, "HTTP status code of the redirection (301, 302, 303, 307, 308)")
	addInteractiveEditorFlag(l7PolicyEditCmd)
	l7PolicyCmd.AddCommand(l7PolicyEditCmd)

	l7PolicyCmd.AddCommand(&cobra.Command{
		Use:   "delete <loadbalancer_id> <listener_id> <l7policy_id>",
		Short: "Delete a specific L7 policy",
		Run:   cloud.DeleteCloudLoadbalancerL7Policy,
		Args:  cobra.ExactArgs(3),
	})
}

func getL7PolicyCreationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <loadbalancer_id> <listener_id>",
		Short: "Create an L7 policy on a listener",
		Long: `Use this command to create an L7 policy on a listener, optionally with a rule.
There are three ways to define the parameters:

1. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud loadbalancer l7policy create <loadbalancer_id> <listener_id> --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud loadbalancer l7policy create <loadbalancer_id> <listener_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file.

2. Using your default text editor:

	ovhcloud cloud loadbalancer l7policy create <loadbalancer_id> <listener_id> --editor

3. Using only CLI flags:

	ovhcloud cloud loadbalancer l7policy create <loadbalancer_id> <listener_id> --action REDIRECT_TO_POOL --redirect-pool-id <pool_id> --name my-policy \
		--rule-type HOST_NAME --rule-compare-type EQUAL_TO --rule-value example.com

  At most one rule can be given with the --rule-* flags. To create a policy with several rules, use --from-file or --editor.
`,
		Run:  cloud.CreateCloudLoadbalancerL7Policy,
		Args: cobra.ExactArgs(2),
	}

	spec := &cloud.CloudLoadbalancerL7PolicyCreateSpec.TargetSpec
	cmd.Flags().StringVar(&spec.Action, "action", "", "L7 policy action (REDIRECT_TO_POOL, REDIRECT_TO_URL, REDIRECT_PREFIX, REJECT)")
	cmd.Flags().StringVar(&spec.Name, "name", "", "Name of the L7 policy")
	cmd.Flags().StringVar(&spec.Description, "description", "", "Description")
	cmd.Flags().IntVar(&spec.Position, "position", 0, "Position on the listener")
	cmd.Flags().StringVar(&spec.RedirectPool.Id, "redirect-pool-id", "", "Pool ID to redirect to")
	cmd.Flags().StringVar(&spec.RedirectUrl, "redirect-url", "", "URL to redirect to")
	cmd.Flags().StringVar(&spec.RedirectPrefix, "redirect-prefix", "", "Prefix URL to redirect to")
	cmd.Flags().IntVar(&spec.RedirectHttpCode, "redirect-http-code", 0, "HTTP status code of the redirection (301, 302, 303, 307, 308)")
	cmd.Flags().StringVar(&spec.Rule.Type, "rule-type", "", "Type of the rule (COOKIE, FILE_TYPE, HEADER, HOST_NAME, PATH)")
	cmd.Flags().StringVar(&spec.Rule.CompareType, "rule-compare-type", "", "Comparison type of the rule (CONTAINS, ENDS_WITH, EQUAL_TO, REGEX, STARTS_WITH)")
	cmd.Flags().StringVar(&spec.Rule.Value, "rule-value", "", "Value to compare the rule with")
	cmd.Flags().StringVar(&spec.Rule.Key, "rule-key", "", "Key of the rule, for COOKIE and HEADER rules")
	cmd.Flags().BoolVar(&spec.Rule.Invert, "rule-invert", false, "Invert the result of the rule")

	addParameterFileFlags(cmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/loadbalancer/{loadbalancerId}/listener/{listenerId}/l7policy", "post", cloud.LoadbalancerL7PolicyCreationExample, nil)
	addInteractiveEditorFlag(cmd)
	markFlagsMutuallyExclusive(cmd, "from-file", "editor")

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

	logURLCmd := &cobra.Command{
		Use:   "generate-url <loadbalancer_id>",
		Short: "Generate a temporary URL to retrieve logs",
		Run:   cloud.GenerateCloudLoadbalancerLogURL,
		Args:  cobra.ExactArgs(1),
	}
	logURLCmd.Flags().StringVar(&cloud.CloudLoadbalancerLogURLSpec.Kind, "kind", "", "Log kind (e.g., haproxy)")
	logURLCmd.MarkFlagRequired("kind")
	logCmd.AddCommand(logURLCmd)

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
	markFlagsMutuallyExclusive(cmd, "from-file", "editor")

	return cmd
}
