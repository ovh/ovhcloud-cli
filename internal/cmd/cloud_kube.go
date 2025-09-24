// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"runtime"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initKubeCommand(cloudCmd *cobra.Command) {
	kubeCmd := &cobra.Command{
		Use:   "kube",
		Short: "Manage Kubernetes clusters in the given cloud project",
	}
	kubeCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	// Command to list Kuberetes clusters
	kubeListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your Kubernetes clusters",
		Run:     cloud.ListKubes,
	}
	kubeCmd.AddCommand(withFilterFlag(kubeListCmd))

	kubeCmd.AddCommand(&cobra.Command{
		Use:   "get <cluster_id>",
		Short: "Get the given Kubernetes cluster",
		Run:   cloud.GetKube,
		Args:  cobra.ExactArgs(1),
	})

	kubeCmd.AddCommand(getKubeCreateCmd())

	kubeEditCmd := &cobra.Command{
		Use:   "edit <cluster_id>",
		Short: "Edit the given Kubernetes cluster",
		Run:   cloud.EditKube,
		Args:  cobra.ExactArgs(1),
	}
	kubeEditCmd.Flags().StringVar(&cloud.KubeSpec.Name, "name", "", "Name of the Kubernetes cluster")
	kubeEditCmd.Flags().StringVar(&cloud.KubeSpec.UpdatePolicy, "update-policy", "", "Update policy for the cluster (ALWAYS_UPDATE, MINIMAL_DOWNTIME, NEVER_UPDATE)")
	kubeEditCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define edit parameters")
	kubeCmd.AddCommand(kubeEditCmd)

	kubeCmd.AddCommand(&cobra.Command{
		Use:   "delete <cluster_id>",
		Short: "Delete the given Kubernetes cluster",
		Run:   cloud.DeleteKube,
		Args:  cobra.ExactArgs(1),
	})

	customizationCmd := &cobra.Command{
		Use:   "customization",
		Short: "Manage Kubernetes cluster customizations",
	}
	kubeCmd.AddCommand(customizationCmd)

	customizationCmd.AddCommand(&cobra.Command{
		Use:   "get <cluster_id>",
		Short: "Get the customization of the given Kubernetes cluster",
		Run:   cloud.GetKubeCustomization,
		Args:  cobra.ExactArgs(1),
	})

	customizationEditCmd := &cobra.Command{
		Use:   "edit <cluster_id>",
		Short: "Edit the customization of the given Kubernetes cluster",
		Run:   cloud.EditKubeCustomization,
		Args:  cobra.ExactArgs(1),
	}
	customizationEditCmd.Flags().StringSliceVar(&cloud.KubeSpec.Customization.APIServer.AdmissionPlugins.Enabled, "api-server.admission-plugins.enabled", nil, "Admission plugins to enable on API server (AlwaysPullImages, NodeRestriction)")
	customizationEditCmd.Flags().StringSliceVar(&cloud.KubeSpec.Customization.APIServer.AdmissionPlugins.Disabled, "api-server.admission-plugins.disabled", nil, "Admission plugins to disable on API server (AlwaysPullImages, NodeRestriction)")
	customizationEditCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPTables.MinSyncPeriod, "kube-proxy.iptables.min-sync-period", "", "Minimum period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')")
	customizationEditCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPTables.SyncPeriod, "kube-proxy.iptables.sync-period", "", "Period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')")
	customizationEditCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.MinSyncPeriod, "kube-proxy.ipvs.min-sync-period", "", "Minimum period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')")
	customizationEditCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.Scheduler, "kube-proxy.ipvs.scheduler", "", "Scheduler for kube-proxy ipvs (dh, lc, nq, rr, sed, sh)")
	customizationEditCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.SyncPeriod, "kube-proxy.ipvs.sync-period", "", "Period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')")
	customizationEditCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.TCPFinTimeout, "kube-proxy.ipvs.tcp-fin-timeout", "", "Timeout value used for IPVS TCP sessions after receiving a FIN in RFC3339 duration format (e.g. 'PT60S')")
	customizationEditCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.TCPTimeout, "kube-proxy.ipvs.tcp-timeout", "", "Timeout value used for idle IPVS TCP sessions in RFC3339 duration format (e.g. 'PT60S')")
	customizationEditCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.UDPTimeout, "kube-proxy.ipvs.udp-timeout", "", "Timeout value used for IPVS UDP packets in RFC3339 duration format (e.g. 'PT60S')")
	addInteractiveEditorFlag(customizationEditCmd)
	customizationCmd.AddCommand(customizationEditCmd)

	ipRestrictionsCmd := &cobra.Command{
		Use:   "ip-restrictions",
		Short: "Manage IP restrictions for Kubernetes clusters",
	}
	kubeCmd.AddCommand(ipRestrictionsCmd)

	ipRestrictionsCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <cluster_id>",
		Aliases: []string{"ls"},
		Short:   "List IP restrictions for the given Kubernetes cluster",
		Run:     cloud.ListKubeIPRestrictions,
		Args:    cobra.ExactArgs(1),
	}))

	ipRestrictionsEditCmd := &cobra.Command{
		Use:   "edit <cluster_id>",
		Short: "Edit IP restrictions for the given Kubernetes cluster",
		Run:   cloud.EditKubeIPRestrictions,
		Args:  cobra.ExactArgs(1),
	}
	ipRestrictionsEditCmd.Flags().StringSliceVar(&cloud.KubeIPRestrictions, "ips", nil, "List of IPs to restrict access to the Kubernetes cluster")
	addInteractiveEditorFlag(ipRestrictionsEditCmd)
	ipRestrictionsEditCmd.MarkFlagsMutuallyExclusive("ips", "editor")
	ipRestrictionsCmd.AddCommand(ipRestrictionsEditCmd)

	kubeConfigCmd := &cobra.Command{
		Use:   "kubeconfig",
		Short: "Manage the kubeconfig for the given Kubernetes cluster",
	}
	kubeCmd.AddCommand(kubeConfigCmd)

	kubeConfigCmd.AddCommand(&cobra.Command{
		Use:   "generate <cluster_id>",
		Short: "Generate the kubeconfig for the given Kubernetes cluster",
		Run:   cloud.GenerateKubeConfig,
		Args:  cobra.ExactArgs(1),
	})

	kubeConfigCmd.AddCommand(&cobra.Command{
		Use:   "reset <cluster_id>",
		Short: "Reset the kubeconfig for the given Kubernetes cluster. Certificates will be regenerated and nodes will be reinstalled",
		Run:   cloud.ResetKubeConfig,
		Args:  cobra.ExactArgs(1),
	})

	nodeCmd := &cobra.Command{
		Use:   "node",
		Short: "Manage Kubernetes nodes",
	}
	kubeCmd.AddCommand(nodeCmd)

	nodeCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <cluster_id>",
		Aliases: []string{"ls"},
		Short:   "List nodes in the given Kubernetes cluster",
		Run:     cloud.ListKubeNodes,
		Args:    cobra.ExactArgs(1),
	}))

	nodeCmd.AddCommand(&cobra.Command{
		Use:   "get <cluster_id> <node_id>",
		Short: "Get the given Kubernetes node",
		Run:   cloud.GetKubeNode,
		Args:  cobra.ExactArgs(2),
	})

	nodeCmd.AddCommand(&cobra.Command{
		Use:   "delete <cluster_id> <node_id>",
		Short: "Delete the given Kubernetes node",
		Run:   cloud.DeleteKubeNode,
		Args:  cobra.ExactArgs(2),
	})

	nodepoolCmd := &cobra.Command{
		Use:   "nodepool",
		Short: "Manage Kubernetes node pools",
	}
	kubeCmd.AddCommand(nodepoolCmd)

	nodepoolCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <cluster_id>",
		Aliases: []string{"ls"},
		Short:   "List node pools in the given Kubernetes cluster",
		Run:     cloud.ListKubeNodepools,
		Args:    cobra.ExactArgs(1),
	}))

	nodepoolCmd.AddCommand(&cobra.Command{
		Use:   "get <cluster_id> <nodepool_id>",
		Short: "Get the given Kubernetes node pool",
		Run:   cloud.GetKubeNodepool,
		Args:  cobra.ExactArgs(2),
	})

	nodepoolEditCmd := &cobra.Command{
		Use:   "edit <cluster_id> <nodepool_id>",
		Short: "Edit the given Kubernetes node pool",
		Run:   cloud.EditKubeNodepool,
		Args:  cobra.ExactArgs(2),
	}
	nodepoolEditCmd.Flags().BoolVar(&cloud.KubeNodepoolSpec.Autoscale, "autoscale", false, "Enable autoscaling for the node pool")
	nodepoolEditCmd.Flags().IntVar(&cloud.KubeNodepoolSpec.Autoscaling.ScaleDownUnneededTimeSeconds, "scale-down-unneeded-time-seconds", 0, "How long a node should be unneeded before it is eligible for scale down (seconds)")
	nodepoolEditCmd.Flags().IntVar(&cloud.KubeNodepoolSpec.Autoscaling.ScaleDownUnreadyTimeSeconds, "scale-down-unready-time-seconds", 0, "How long an unready node should be unneeded before it is eligible for scale down (seconds)")
	nodepoolEditCmd.Flags().Float64Var(&cloud.KubeNodepoolSpec.Autoscaling.ScaleDownUtilizationThreshold, "scale-down-utilization-threshold", 0, "Sum of CPU or memory of all pods running on the node divided by node's corresponding allocatable resource, below which a node can be considered for scale down")
	nodepoolEditCmd.Flags().IntVar(&cloud.KubeNodepoolSpec.DesiredNodes, "desired-nodes", 0, "Desired number of nodes")
	nodepoolEditCmd.Flags().IntVar(&cloud.KubeNodepoolSpec.MaxNodes, "max-nodes", 0, "Higher limit you accept for the desiredNodes value (100 by default)")
	nodepoolEditCmd.Flags().IntVar(&cloud.KubeNodepoolSpec.MinNodes, "min-nodes", 0, "Lower limit you accept for the desiredNodes value (0 by default)")
	nodepoolEditCmd.Flags().StringSliceVar(&cloud.KubeNodepoolSpec.NodesToRemove, "nodes-to-remove", nil, "List of node IDs to remove from the node pool")
	nodepoolEditCmd.Flags().StringToStringVar(&cloud.KubeNodepoolSpec.Template.Metadata.Annotations, "template-annotations", nil, "Annotations to apply to each node")
	nodepoolEditCmd.Flags().StringSliceVar(&cloud.KubeNodepoolSpec.Template.Metadata.Finalizers, "template-finalizers", nil, "Finalizers to apply to each node")
	nodepoolEditCmd.Flags().StringToStringVar(&cloud.KubeNodepoolSpec.Template.Metadata.Labels, "template-labels", nil, "Labels to apply to each node")
	nodepoolEditCmd.Flags().StringSliceVar(&cloud.KubeNodepoolSpec.Template.Spec.CommandLineTaints, "template-taints", nil, "Taints to apply to each node in key=value:effect format")
	nodepoolEditCmd.Flags().BoolVar(&cloud.KubeNodepoolSpec.Template.Spec.Unschedulable, "template-unschedulable", false, "Set the nodes as unschedulable")
	addInteractiveEditorFlag(nodepoolEditCmd)
	nodepoolCmd.AddCommand(nodepoolEditCmd)

	nodepoolCmd.AddCommand(&cobra.Command{
		Use:   "delete <cluster_id> <nodepool_id>",
		Short: "Delete the given Kubernetes node pool",
		Run:   cloud.DeleteKubeNodepool,
		Args:  cobra.ExactArgs(2),
	})

	nodepoolCmd.AddCommand(getKubeNodePoolCreateCmd())

	oidcCmd := &cobra.Command{
		Use:   "oidc",
		Short: "Manage OpenID Connect (OIDC) integration for Kubernetes clusters",
	}
	kubeCmd.AddCommand(oidcCmd)

	oidcCmd.AddCommand(&cobra.Command{
		Use:   "get <cluster_id>",
		Short: "Get the OIDC configuration for the given Kubernetes cluster",
		Run:   cloud.GetKubeOIDCIntegration,
		Args:  cobra.ExactArgs(1),
	})

	editCmd := &cobra.Command{
		Use:   "edit <cluster_id>",
		Short: "Edit the OIDC configuration for the given Kubernetes cluster",
		Run:   cloud.EditKubeOIDCIntegration,
		Args:  cobra.ExactArgs(1),
	}
	editCmd.Flags().StringVar(&cloud.KubeOIDCConfig.CaContent, "ca-content", "", "CA certificate content for the OIDC provider")
	editCmd.Flags().StringVar(&cloud.KubeOIDCConfig.ClientId, "client-id", "", "OIDC client ID")
	editCmd.Flags().StringSliceVar(&cloud.KubeOIDCConfig.GroupsClaim, "groups-claim", nil, "OIDC groups claim(s)")
	editCmd.Flags().StringVar(&cloud.KubeOIDCConfig.GroupsPrefix, "groups-prefix", "", "Prefix prepended to group claims")
	editCmd.Flags().StringVar(&cloud.KubeOIDCConfig.IssuerUrl, "issuer-url", "", "OIDC issuer URL")
	editCmd.Flags().StringSliceVar(&cloud.KubeOIDCConfig.RequiredClaim, "required-claim", nil, "OIDC required claim(s)")
	editCmd.Flags().StringSliceVar(&cloud.KubeOIDCConfig.SigningAlgorithms, "signing-algorithms", nil, "OIDC signing algorithm(s) (ES256, ES384, ES512, PS256, PS384, PS512, RS256, RS384, RS512)")
	editCmd.Flags().StringVar(&cloud.KubeOIDCConfig.UsernameClaim, "username-claim", "", "OIDC username claim")
	editCmd.Flags().StringVar(&cloud.KubeOIDCConfig.UsernamePrefix, "username-prefix", "", "Prefix prepended to username claims")
	oidcCmd.AddCommand(editCmd)

	oidcCmd.AddCommand(getKubeOIDCCreateCmd())

	oidcCmd.AddCommand(&cobra.Command{
		Use:   "delete <cluster_id>",
		Short: "Delete the OIDC integration for the given Kubernetes cluster",
		Run:   cloud.DeleteKubeOIDCIntegration,
		Args:  cobra.ExactArgs(1),
	})

	privateNetworkConfigCmd := &cobra.Command{
		Use:   "private-network-configuration",
		Short: "Manage private network configuration for Kubernetes clusters",
	}
	kubeCmd.AddCommand(privateNetworkConfigCmd)

	privateNetworkConfigCmd.AddCommand(&cobra.Command{
		Use:   "get <cluster_id>",
		Short: "Get the private network configuration for the given Kubernetes cluster",
		Run:   cloud.GetKubePrivateNetworkConfiguration,
		Args:  cobra.ExactArgs(1),
	})

	privateNetworkConfigEditCmd := &cobra.Command{
		Use:   "edit <cluster_id>",
		Short: "Edit the private network configuration for the given Kubernetes cluster",
		Run:   cloud.EditKubePrivateNetworkConfiguration,
		Args:  cobra.ExactArgs(1),
	}
	privateNetworkConfigEditCmd.Flags().StringVar(&cloud.KubeSpec.PrivateNetworkConfiguration.DefaultVrackGateway, "default-vrack-gateway", "", "If defined, all egress traffic will be routed towards this IP address, which should belong to the private network")
	privateNetworkConfigEditCmd.Flags().BoolVar(&cloud.KubeSpec.PrivateNetworkConfiguration.PrivateNetworkRoutingAsDefault, "routing-as-default", false, "Set private network routing as default")
	addInteractiveEditorFlag(privateNetworkConfigEditCmd)
	privateNetworkConfigCmd.AddCommand(privateNetworkConfigEditCmd)

	kubeCmd.AddCommand(getKubeResetCmd())

	kubeRestartCmd := &cobra.Command{
		Use:   "restart <cluster_id>",
		Short: "Restart control plane apiserver to invalidate cache without downtime",
		Run:   cloud.RestartKubeCluster,
		Args:  cobra.ExactArgs(1),
	}
	kubeRestartCmd.Flags().BoolVar(&cloud.KubeForceAction, "force", false, "Force restart the Kubernetes cluster (will create a slight downtime)")
	kubeCmd.AddCommand(kubeRestartCmd)

	kubeUpdateCmd := &cobra.Command{
		Use:   "update <cluster_id>",
		Short: "Update the given Kubernetes cluster",
		Run:   cloud.UpdateKubeCluster,
		Args:  cobra.ExactArgs(1),
	}
	kubeUpdateCmd.Flags().StringVar(&cloud.KubeUpdateStrategy, "strategy", "", "Update strategy to apply on your service (LATEST_PATCH, NEXT_MINOR)")
	kubeUpdateCmd.Flags().BoolVar(&cloud.KubeForceAction, "force", false, "Force redeploying the control plane / reinstalling the nodes regardless of their current version")
	kubeCmd.AddCommand(kubeUpdateCmd)

	kubeCmd.AddCommand(&cobra.Command{
		Use:   "set-load-balancers-subnet <cluster_id> <subnet_id>",
		Short: "Update the load balancers subnet ID for the given Kubernetes cluster",
		Run:   cloud.UpdateKubeLoadBalancersSubnet,
		Args:  cobra.ExactArgs(2),
	})

	cloudCmd.AddCommand(kubeCmd)
}

func getKubeCreateCmd() *cobra.Command {
	kubeCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Kubernetes cluster",
		Long: `Use this command to create a managed Kubernetes cluster in the given public cloud project.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud kube create --name MyNewCluster --region SBG5 --version 1.32

2. Using a configuration file:

  First you can generate an example of installation file using the following command:

	ovhcloud cloud kube create --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud kube create --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud kube create

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud kube create --from-file ./params.json --name NameOverriden

3. Using your default text editor:

	ovhcloud cloud kube create --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud kube create --editor --region BHS5
`,
		Run: cloud.CreateKube,
	}

	// All flags for Kubernetes cluster creation
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.Name, "name", "", "Name of the Kubernetes cluster")
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.Region, "region", "", "Region for the Kubernetes cluster")
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.Version, "version", "", "Kubernetes version")
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.Plan, "plan", "", "Kubernetes cluster plan (free or standard, default: free)")
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.KubeProxyMode, "kube-proxy-mode", "", "Kube-proxy mode (iptables or ipvs)")
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.LoadBalancersSubnetId, "load-balancers-subnet-id", "", "OpenStack subnet ID that the load balancers will use")
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.NodesSubnetId, "nodes-subnet-id", "", "OpenStack subnet ID that the cluster nodes will use")
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.PrivateNetworkId, "private-network-id", "", "OpenStack private network ID that the cluster will use")
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.UpdatePolicy, "update-policy", "", "Update policy for the cluster (ALWAYS_UPDATE, MINIMAL_DOWNTIME, NEVER_UPDATE)")

	// Private network configuration
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.PrivateNetworkConfiguration.DefaultVrackGateway, "private-network.default-vrack-gateway", "", "If defined, all egress traffic will be routed towards this IP address, which should belong to the private network")
	kubeCreateCmd.Flags().BoolVar(&cloud.KubeSpec.PrivateNetworkConfiguration.PrivateNetworkRoutingAsDefault, "private-network.routing-as-default", false, "Set private network routing as default")

	// Customization: API Server Admission Plugins
	kubeCreateCmd.Flags().StringSliceVar(&cloud.KubeSpec.Customization.APIServer.AdmissionPlugins.Enabled, "customization.api-server.admission-plugins.enabled", nil, "Admission plugins to enable on API server (AlwaysPullImages, NodeRestriction)")
	kubeCreateCmd.Flags().StringSliceVar(&cloud.KubeSpec.Customization.APIServer.AdmissionPlugins.Disabled, "customization.api-server.admission-plugins.disabled", nil, "Admission plugins to disable on API server (AlwaysPullImages, NodeRestriction)")

	// Customization: KubeProxy IPTables
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPTables.MinSyncPeriod, "customization.kube-proxy.iptables.min-sync-period", "", "Minimum period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')")
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPTables.SyncPeriod, "customization.kube-proxy.iptables.sync-period", "", "Period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')")

	// Customization: KubeProxy IPVS
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.MinSyncPeriod, "customization.kube-proxy.ipvs.min-sync-period", "", "Minimum period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')")
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.Scheduler, "customization.kube-proxy.ipvs.scheduler", "", "Scheduler for kube-proxy ipvs (dh, lc, nq, rr, sed, sh)")
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.SyncPeriod, "customization.kube-proxy.ipvs.sync-period", "", "Period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')")
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.TCPFinTimeout, "customization.kube-proxy.ipvs.tcp-fin-timeout", "", "Timeout value used for IPVS TCP sessions after receiving a FIN in RFC3339 duration format (e.g. 'PT60S')")
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.TCPTimeout, "customization.kube-proxy.ipvs.tcp-timeout", "", "Timeout value used for idle IPVS TCP sessions in RFC3339 duration format (e.g. 'PT60S')")
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.UDPTimeout, "customization.kube-proxy.ipvs.udp-timeout", "", "Timeout value used for IPVS UDP packets in RFC3339 duration format (e.g. 'PT60S')")

	// Common flags for other means to define parameters
	addInitParameterFileFlag(kubeCreateCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/kube", "post", cloud.CloudKubeCreationExample, nil)
	addInteractiveEditorFlag(kubeCreateCmd)
	addFromFileFlag(kubeCreateCmd)
	kubeCreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return kubeCreateCmd
}

func getKubeResetCmd() *cobra.Command {
	kubeResetCmd := &cobra.Command{
		Use:   "reset <cluster_id>",
		Short: "Reset the given Kubernetes cluster",
		Long: `Reset the given Kubernetes cluster.
All Kubernetes data will be erased (pods, services, configuration, etc), nodes will be either deleted or reinstalled.

There are three ways to define the reset parameters:

1. Using only CLI flags:

	ovhcloud cloud kube reset <cluster_id> --name MyNewCluster --version 1.32

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud kube reset <cluster_id> --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct parameters, run:

	ovhcloud cloud kube reset <cluster_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud kube reset <cluster_id>

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud kube reset <cluster_id> --from-file ./params.json --name NameOverriden

3. Using your default text editor:

	ovhcloud cloud kube reset <cluster_id> --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the reset will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud kube reset <cluster_id> --editor --version 1.31
`,
		Run:  cloud.ResetKubeCluster,
		Args: cobra.ExactArgs(1),
	}

	// All flags for Kubernetes cluster reset
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.Name, "name", "", "Name of the Kubernetes cluster")
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.Version, "version", "", "Kubernetes version")
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.KubeProxyMode, "kube-proxy-mode", "", "Kube-proxy mode (iptables or ipvs)")
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.LoadBalancersSubnetId, "load-balancers-subnet-id", "", "OpenStack subnet ID that the load balancers will use")
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.NodesSubnetId, "nodes-subnet-id", "", "OpenStack subnet ID that the cluster nodes will use")
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.PrivateNetworkId, "private-network-id", "", "OpenStack private network ID that the cluster will use")
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.UpdatePolicy, "update-policy", "", "Update policy for the cluster (ALWAYS_UPDATE, MINIMAL_DOWNTIME, NEVER_UPDATE)")
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.WorkerNodesPolicy, "worker-nodes-policy", "", "Worker nodes reset policy (delete, reinstall)")

	// Private network configuration
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.PrivateNetworkConfiguration.DefaultVrackGateway, "private-network.default-vrack-gateway", "", "If defined, all egress traffic will be routed towards this IP address, which should belong to the private network")
	kubeResetCmd.Flags().BoolVar(&cloud.KubeSpec.PrivateNetworkConfiguration.PrivateNetworkRoutingAsDefault, "private-network.routing-as-default", false, "Set private network routing as default")

	// Customization: API Server Admission Plugins
	kubeResetCmd.Flags().StringSliceVar(&cloud.KubeSpec.Customization.APIServer.AdmissionPlugins.Enabled, "customization.api-server.admission-plugins.enabled", nil, "Admission plugins to enable on API server (AlwaysPullImages, NodeRestriction)")
	kubeResetCmd.Flags().StringSliceVar(&cloud.KubeSpec.Customization.APIServer.AdmissionPlugins.Disabled, "customization.api-server.admission-plugins.disabled", nil, "Admission plugins to disable on API server (AlwaysPullImages, NodeRestriction)")

	// Customization: KubeProxy IPTables
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPTables.MinSyncPeriod, "customization.kube-proxy.iptables.min-sync-period", "", "Minimum period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')")
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPTables.SyncPeriod, "customization.kube-proxy.iptables.sync-period", "", "Period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')")

	// Customization: KubeProxy IPVS
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.MinSyncPeriod, "customization.kube-proxy.ipvs.min-sync-period", "", "Minimum period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')")
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.Scheduler, "customization.kube-proxy.ipvs.scheduler", "", "Scheduler for kube-proxy ipvs (dh, lc, nq, rr, sed, sh)")
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.SyncPeriod, "customization.kube-proxy.ipvs.sync-period", "", "Period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')")
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.TCPFinTimeout, "customization.kube-proxy.ipvs.tcp-fin-timeout", "", "Timeout value used for IPVS TCP sessions after receiving a FIN in RFC3339 duration format (e.g. 'PT60S')")
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.TCPTimeout, "customization.kube-proxy.ipvs.tcp-timeout", "", "Timeout value used for idle IPVS TCP sessions in RFC3339 duration format (e.g. 'PT60S')")
	kubeResetCmd.Flags().StringVar(&cloud.KubeSpec.Customization.KubeProxy.IPVS.UDPTimeout, "customization.kube-proxy.ipvs.udp-timeout", "", "Timeout value used for IPVS UDP packets in RFC3339 duration format (e.g. 'PT60S')")

	// Common flags for other means to define parameters
	addInitParameterFileFlag(kubeResetCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/kube/reset", "post", cloud.CloudKubeResetExample, nil)
	addInteractiveEditorFlag(kubeResetCmd)
	addFromFileFlag(kubeResetCmd)
	kubeResetCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return kubeResetCmd
}

func getKubeNodePoolCreateCmd() *cobra.Command {
	nodepoolCreateCmd := &cobra.Command{
		Use:   "create <cluster_id>",
		Short: "Create a new Kubernetes node pool",
		Long: `Use this command to create a node pool in the given managed Kubernetes cluster.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud kube nodepool create <cluster_id> --flavor-name b3-16 --desired-nodes 3 --name newnodepool

2. Using a configuration file:

  First you can generate an example of installation file using the following command:

	ovhcloud cloud kube nodepool create <cluster_id> --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud kube nodepool create <cluster_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud kube nodepool create <cluster_id>

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud kube nodepool create <cluster_id> --from-file ./params.json --name NameOverriden

  It is also possible to use the interactive flavor selector to define the flavor-name parameter, like the following:

	ovhcloud cloud kube nodepool create <cluster_id> --init-file ./params.json --flavor-selector

3. Using your default text editor:

	ovhcloud cloud kube nodepool create <cluster_id> --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud kube nodepool create <cluster_id> --editor --flavor-name b3-16

  You can also use the interactive flavor selector to define the flavor-name parameter, like the following:

	ovhcloud cloud kube nodepool create <cluster_id> --editor --flavor-selector
`,
		Run:  cloud.CreateKubeNodepool,
		Args: cobra.ExactArgs(1),
	}

	// All flags for node pool creation
	nodepoolCreateCmd.Flags().StringVar(&cloud.KubeNodepoolSpec.Name, "name", "", "Name of the node pool")
	nodepoolCreateCmd.Flags().BoolVar(&cloud.KubeNodepoolSpec.AntiAffinity, "anti-affinity", false, "Enable anti-affinity for the node pool")
	nodepoolCreateCmd.Flags().BoolVar(&cloud.KubeNodepoolSpec.Autoscale, "autoscale", false, "Enable autoscaling for the node pool")
	nodepoolCreateCmd.Flags().IntVar(&cloud.KubeNodepoolSpec.Autoscaling.ScaleDownUnneededTimeSeconds, "scale-down-unneeded-time-seconds", 0, "How long a node should be unneeded before it is eligible for scale down (seconds)")
	nodepoolCreateCmd.Flags().IntVar(&cloud.KubeNodepoolSpec.Autoscaling.ScaleDownUnreadyTimeSeconds, "scale-down-unready-time-seconds", 0, "How long an unready node should be unneeded before it is eligible for scale down (seconds)")
	nodepoolCreateCmd.Flags().Float64Var(&cloud.KubeNodepoolSpec.Autoscaling.ScaleDownUtilizationThreshold, "scale-down-utilization-threshold", 0, "Sum of CPU or memory of all pods running on the node divided by node's corresponding allocatable resource, below which a node can be considered for scale down")
	nodepoolCreateCmd.Flags().StringSliceVar(&cloud.KubeNodepoolSpec.AvailabilityZones, "availability-zones", nil, "Availability zones for the node pool")
	nodepoolCreateCmd.Flags().IntVar(&cloud.KubeNodepoolSpec.DesiredNodes, "desired-nodes", 0, "Desired number of nodes")
	nodepoolCreateCmd.Flags().StringVar(&cloud.KubeNodepoolSpec.FlavorName, "flavor-name", "", "Flavor name for the nodes (b2-7, b2-15, etc.)")
	nodepoolCreateCmd.Flags().IntVar(&cloud.KubeNodepoolSpec.MaxNodes, "max-nodes", 0, "Higher limit you accept for the desiredNodes value (100 by default)")
	nodepoolCreateCmd.Flags().IntVar(&cloud.KubeNodepoolSpec.MinNodes, "min-nodes", 0, "Lower limit you accept for the desiredNodes value (0 by default)")
	nodepoolCreateCmd.Flags().BoolVar(&cloud.KubeNodepoolSpec.MonthlyBilled, "monthly-billed", false, "Enable monthly billing for the node pool")

	// Template.Metadata
	nodepoolCreateCmd.Flags().StringToStringVar(&cloud.KubeNodepoolSpec.Template.Metadata.Annotations, "template-annotations", nil, "Annotations to apply to each node")
	nodepoolCreateCmd.Flags().StringSliceVar(&cloud.KubeNodepoolSpec.Template.Metadata.Finalizers, "template-finalizers", nil, "Finalizers to apply to each node")
	nodepoolCreateCmd.Flags().StringToStringVar(&cloud.KubeNodepoolSpec.Template.Metadata.Labels, "template-labels", nil, "Labels to apply to each node")

	// Template.Spec
	nodepoolCreateCmd.Flags().StringSliceVar(&cloud.KubeNodepoolSpec.Template.Spec.CommandLineTaints, "template-taints", nil, "Taints to apply to each node in key=value:effect format")
	nodepoolCreateCmd.Flags().BoolVar(&cloud.KubeNodepoolSpec.Template.Spec.Unschedulable, "template-unschedulable", false, "Set the nodes as unschedulable")

	// Common flags for other means to define parameters
	addInitParameterFileFlag(nodepoolCreateCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/kube/{kubeId}/nodepool", "post", cloud.CloudKubeNodePoolCreationExample, cloud.GetKubeFlavorInteractiveSelector)
	addInteractiveEditorFlag(nodepoolCreateCmd)
	addFromFileFlag(nodepoolCreateCmd)
	if !(runtime.GOARCH == "wasm" && runtime.GOOS == "js") {
		nodepoolCreateCmd.Flags().BoolVar(&cloud.InstanceFlavorViaInteractiveSelector, "flavor-selector", false, "Use the interactive flavor selector")
		nodepoolCreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")
	}

	return nodepoolCreateCmd
}

func getKubeOIDCCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create <cluster_id>",
		Short: "Create a new OIDC integration for the given Kubernetes cluster",
		Long: `Use this command to create a new OIDC integration for the given Kubernetes cluster.
There are three ways to define the parameters:

1. Using only CLI flags:

	ovhcloud cloud kube oidc create <cluster_id> --issuer-url <url>

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud kube oidc create <cluster_id> --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud kube oidc create <cluster_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud kube oidc create <cluster_id>

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud kube oidc create <cluster_id> --from-file ./params.json --client-id <client_id>

3. Using your default text editor:

	ovhcloud cloud kube oidc create <cluster_id> --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud kube oidc create <cluster_id> --editor --client-id <client_id>
`,
		Run:  cloud.CreateKubeOIDCIntegration,
		Args: cobra.ExactArgs(1),
	}

	// All flags for OIDC integration creation
	createCmd.Flags().StringVar(&cloud.KubeOIDCConfig.CaContent, "ca-content", "", "CA certificate content for the OIDC provider")
	createCmd.Flags().StringVar(&cloud.KubeOIDCConfig.ClientId, "client-id", "", "OIDC client ID")
	createCmd.Flags().StringSliceVar(&cloud.KubeOIDCConfig.GroupsClaim, "groups-claim", nil, "OIDC groups claim(s)")
	createCmd.Flags().StringVar(&cloud.KubeOIDCConfig.GroupsPrefix, "groups-prefix", "", "Prefix prepended to group claims")
	createCmd.Flags().StringVar(&cloud.KubeOIDCConfig.IssuerUrl, "issuer-url", "", "OIDC issuer URL")
	createCmd.Flags().StringSliceVar(&cloud.KubeOIDCConfig.RequiredClaim, "required-claim", nil, "OIDC required claim(s)")
	createCmd.Flags().StringSliceVar(&cloud.KubeOIDCConfig.SigningAlgorithms, "signing-algorithms", nil, "OIDC signing algorithm(s) (ES256, ES384, ES512, PS256, PS384, PS512, RS256, RS384, RS512)")
	createCmd.Flags().StringVar(&cloud.KubeOIDCConfig.UsernameClaim, "username-claim", "", "OIDC username claim")
	createCmd.Flags().StringVar(&cloud.KubeOIDCConfig.UsernamePrefix, "username-prefix", "", "Prefix prepended to username claims")

	// Common flags for other means to define parameters
	addInitParameterFileFlag(createCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/kube/{kubeId}/openIdConnect", "post", cloud.CloudKubeOIDCCreationExample, nil)
	addInteractiveEditorFlag(createCmd)
	addFromFileFlag(createCmd)
	createCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return createCmd
}
