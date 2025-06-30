package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
)

func initKubeCommand(cloudCmd *cobra.Command) {
	kubeCmd := &cobra.Command{
		Use:   "kube",
		Short: "List Kubernetes clusters in the given cloud project",
	}
	kubeCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	// Command to list Kuberetes clusters
	kubeListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Kubernetes clusters",
		Run:   cloud.ListKubes,
	}
	kubeCmd.AddCommand(withFilterFlag(kubeListCmd))

	kubeCmd.AddCommand(&cobra.Command{
		Use:   "get <cluster_id>",
		Short: "Get the given Kubernetes cluster",
		Run:   cloud.GetKube,
		Args:  cobra.ExactArgs(1),
	})

	kubeCmd.AddCommand(getKubeCreateCmd())

	kubeCmd.AddCommand(&cobra.Command{
		Use:   "edit <cluster_id>",
		Short: "Edit the given Kubernetes cluster",
		Run:   cloud.EditKube,
	})

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

	customizationCmd.AddCommand(&cobra.Command{
		Use:   "edit <cluster_id>",
		Short: "Edit the customization of the given Kubernetes cluster",
		Run:   cloud.EditKubeCustomization,
		Args:  cobra.ExactArgs(1),
	})

	ipRestrictionsCmd := &cobra.Command{
		Use:   "ip-restrictions",
		Short: "Manage IP restrictions for Kubernetes clusters",
	}
	kubeCmd.AddCommand(ipRestrictionsCmd)

	ipRestrictionsCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list <cluster_id>",
		Short: "List IP restrictions for the given Kubernetes cluster",
		Run:   cloud.ListKubeIPRestrictions,
		Args:  cobra.ExactArgs(1),
	}))

	ipRestrictionsCmd.AddCommand(&cobra.Command{
		Use:   "edit <cluster_id>",
		Short: "Edit IP restrictions for the given Kubernetes cluster",
		Run:   cloud.EditKubeIPRestrictions,
		Args:  cobra.ExactArgs(1),
	})

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
		Use:   "list <cluster_id>",
		Short: "List nodes in the given Kubernetes cluster",
		Run:   cloud.ListKubeNodes,
		Args:  cobra.ExactArgs(1),
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
		Use:   "list <cluster_id>",
		Short: "List node pools in the given Kubernetes cluster",
		Run:   cloud.ListKubeNodepools,
		Args:  cobra.ExactArgs(1),
	}))

	nodepoolCmd.AddCommand(&cobra.Command{
		Use:   "get <cluster_id> <nodepool_id>",
		Short: "Get the given Kubernetes node pool",
		Run:   cloud.GetKubeNodepool,
		Args:  cobra.ExactArgs(2),
	})

	nodepoolCmd.AddCommand(&cobra.Command{
		Use:   "edit <cluster_id> <nodepool_id>",
		Short: "Edit the given Kubernetes node pool",
		Run:   cloud.EditKubeNodepool,
		Args:  cobra.ExactArgs(2),
	})

	nodepoolCmd.AddCommand(&cobra.Command{
		Use:   "delete <cluster_id> <nodepool_id>",
		Short: "Delete the given Kubernetes node pool",
		Run:   cloud.DeleteKubeNodepool,
		Args:  cobra.ExactArgs(2),
	})

	nodepoolCmd.AddCommand(getKubeNodePoolCreateCmd())

	cloudCmd.AddCommand(kubeCmd)
}

func getKubeCreateCmd() *cobra.Command {
	kubeCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Kubernetes cluster",
		Run:   cloud.CreateKube,
	}

	// All flags for Kubernetes cluster creation
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.Name, "name", "", "Name of the Kubernetes cluster")
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.Region, "region", "", "Region for the Kubernetes cluster")
	kubeCreateCmd.Flags().StringVar(&cloud.KubeSpec.Version, "version", "", "Kubernetes version")
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
	addInitParameterFileFlag(kubeCreateCmd, cloud.CloudOpenapiSchema, "/cloud/project/{serviceName}/kube", "post", cloud.CloudKubeCreationExample, nil)
	kubeCreateCmd.Flags().StringVar(&flags.ParametersFile, "from-file", "", "File containing creation parameters")
	kubeCreateCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define creation parameters")

	return kubeCreateCmd
}

func getKubeNodePoolCreateCmd() *cobra.Command {
	nodepoolCreateCmd := &cobra.Command{
		Use:   "create <cluster_id>",
		Short: "Create a new Kubernetes node pool",
		Run:   cloud.CreateKubeNodepool,
		Args:  cobra.ExactArgs(1),
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
	addInitParameterFileFlag(nodepoolCreateCmd, cloud.CloudOpenapiSchema, "/cloud/project/{serviceName}/kube/{kubeId}/nodepool", "post", cloud.CloudKubeNodePoolCreationExample, cloud.GetKubeFlavorInteractiveSelector)
	nodepoolCreateCmd.Flags().StringVar(&flags.ParametersFile, "from-file", "", "File containing creation parameters")
	nodepoolCreateCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define creation parameters")
	nodepoolCreateCmd.Flags().BoolVar(&cloud.InstanceFlavorViaInteractiveSelector, "flavor-selector", false, "Use the interactive flavor selector")

	return nodepoolCreateCmd
}
