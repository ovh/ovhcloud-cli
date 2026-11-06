// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/completion"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/ovh/ovhcloud-cli/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	// Cluster related vars
	ciliumClusterID          uint8
	hubbleEnabled            bool
	hubbleRelayEnabled       bool
	hubbleUIEnabled          bool
	clusterMeshEnabled       bool
	autoUpgradePatchEnabled  bool
	apiServerAuthorizedIPs   []string
	admissionPluginsEnabled  []string
	admissionPluginsDisabled []string

	// Nodepool related vars
	desiredNodes            int
	autoscalerEnabled       bool
	autoscalerMaxNodes      int
	autoscalerMinNodes      int
	antiAffinityEnabled     bool
	attachFloatingIpEnabled bool
	templateUnschedulable   bool
	templateTaints          []string
)

func initManagedKubernetesCommand(cloudCmd *cobra.Command) {
	managedKubernetesCmd := &cobra.Command{
		Use:     "managed-kubernetes",
		Aliases: []string{"mks"},
		Short:   "Manage Kubernetes clusters in the given cloud project",
	}
	managedKubernetesCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	// Command to list Kubernetes clusters
	managedKubernetesListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your Kubernetes clusters",
		Run:     cloud.ListManagedKubernetes,
	}
	managedKubernetesCmd.AddCommand(withFilterFlag(managedKubernetesListCmd))

	// Command to get a Kubernetes cluster
	managedKubernetesCmd.AddCommand(&cobra.Command{
		Use:               "get <cluster_id>",
		Short:             "Get the given Kubernetes cluster",
		ValidArgsFunction: completion.CloudResources(cloud.ManagedKubernetesV2EndpointFormat),
		Run:               cloud.GetManagedKubernetes,
		Args:              cobra.ExactArgs(1),
	})

	// Command to create a Kubernetes cluster
	managedKubernetesCmd.AddCommand(getManagedKubernetesCreateCmd())

	// Command to edit a Kubernetes cluster
	managedKubernetesCmd.AddCommand(getManagedKubernetesEditCmd())

	// Command to delete a Kubernetes cluster
	managedKubernetesCmd.AddCommand(&cobra.Command{
		Use:               "delete <cluster_id>",
		Short:             "Delete the given Kubernetes cluster",
		ValidArgsFunction: completion.CloudResources(cloud.ManagedKubernetesV2EndpointFormat),
		Run:               cloud.DeleteManagedKubernetes,
		Args:              cobra.ExactArgs(1),
	})

	kubeConfigCmd := &cobra.Command{
		Use:   "kubeconfig",
		Short: "Manage the kubeconfig for the given Kubernetes cluster",
	}
	managedKubernetesCmd.AddCommand(kubeConfigCmd)

	kubeConfigCmd.AddCommand(&cobra.Command{
		Use:               "generate <cluster_id>",
		Short:             "Generate the kubeconfig for the given Kubernetes cluster",
		ValidArgsFunction: completion.CloudResources(cloud.ManagedKubernetesV2EndpointFormat),
		Run:               cloud.GenerateKubeConfig,
		Args:              cobra.ExactArgs(1),
	})

	kubeConfigCmd.AddCommand(&cobra.Command{
		Use:               "reset <cluster_id>",
		Short:             "Reset the kubeconfig for the given Kubernetes cluster. Certificates will be regenerated and nodes will be reinstalled",
		ValidArgsFunction: completion.CloudResources(cloud.ManagedKubernetesV2EndpointFormat),
		Run:               cloud.ResetKubeConfig,
		Args:              cobra.ExactArgs(1),
	})

	nodeCmd := &cobra.Command{
		Use:   "node",
		Short: "Inspect Kubernetes nodes",
	}
	managedKubernetesCmd.AddCommand(nodeCmd)

	nodeListCmd := &cobra.Command{
		Use:               "list <cluster_id> <nodepool_id>",
		Aliases:           []string{"ls"},
		Short:             "List nodes in the given Kubernetes node pool",
		ValidArgsFunction: completion.CloudResourceWithChild(cloud.ManagedKubernetesV2EndpointFormat, cloud.ManagedKubernetesNodepoolsEndpointFormat),
		Run:               cloud.ListManagedKubernetesNodes,
		Args:              cobra.ExactArgs(2),
	}
	nodeListCmd.Flags().BoolVar(&cloud.ManagedKubernetesNodeIncludeDeleted, "include-deleted", false, "Include deleted nodes still kept in retention")
	nodeListCmd.Flags().StringVar(&cloud.ManagedKubernetesNodeResourceStatus, "resource-status", "", "List only the nodes in this lifecycle status (CREATING, DELETED, DELETING, ERROR, OUT_OF_SYNC, READY, UNKNOWN, UPDATING). DELETED requires --include-deleted")
	nodeCmd.AddCommand(withFilterFlag(nodeListCmd))

	nodeCmd.AddCommand(&cobra.Command{
		Use:               "get <cluster_id> <nodepool_id> <node_id>",
		Short:             "Get the given Kubernetes node",
		ValidArgsFunction: completion.CloudResourceChain(cloud.ManagedKubernetesV2EndpointFormat, cloud.ManagedKubernetesNodepoolsEndpointFormat, cloud.ManagedKubernetesNodesEndpointFormat),
		Run:               cloud.GetManagedKubernetesNode,
		Args:              cobra.ExactArgs(3),
	})

	nodepoolCmd := &cobra.Command{
		Use:   "nodepool",
		Short: "Manage Kubernetes node pools",
	}
	managedKubernetesCmd.AddCommand(nodepoolCmd)

	nodepoolCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list <cluster_id>",
		Aliases:           []string{"ls"},
		Short:             "List node pools in the given Kubernetes cluster",
		ValidArgsFunction: completion.CloudResources(cloud.ManagedKubernetesV2EndpointFormat),
		Run:               cloud.ListManagedKubernetesNodepools,
		Args:              cobra.ExactArgs(1),
	}))

	nodepoolCmd.AddCommand(&cobra.Command{
		Use:               "get <cluster_id> <nodepool_id>",
		Short:             "Get the given Kubernetes node pool",
		ValidArgsFunction: completion.CloudResourceWithChild(cloud.ManagedKubernetesV2EndpointFormat, cloud.ManagedKubernetesNodepoolsEndpointFormat),
		Run:               cloud.GetManagedKubernetesNodepool,
		Args:              cobra.ExactArgs(2),
	})

	nodepoolCmd.AddCommand(getManagedKubernetesNodepoolEditCmd())

	nodepoolCmd.AddCommand(&cobra.Command{
		Use:               "delete <cluster_id> <nodepool_id>",
		Short:             "Delete the given Kubernetes node pool",
		ValidArgsFunction: completion.CloudResourceWithChild(cloud.ManagedKubernetesV2EndpointFormat, cloud.ManagedKubernetesNodepoolsEndpointFormat),
		Run:               cloud.DeleteManagedKubernetesNodepool,
		Args:              cobra.ExactArgs(2),
	})

	nodepoolCmd.AddCommand(getManagedKubernetesNodepoolCreateCmd())

	nodepoolTaskCmd := &cobra.Command{
		Use:   "task",
		Short: "Inspect the asynchronous tasks of a Kubernetes node pool",
	}
	nodepoolCmd.AddCommand(nodepoolTaskCmd)

	nodepoolTaskCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list <cluster_id> <nodepool_id>",
		Aliases:           []string{"ls"},
		Short:             "List the tasks of the given Kubernetes node pool",
		ValidArgsFunction: completion.CloudResourceWithChild(cloud.ManagedKubernetesV2EndpointFormat, cloud.ManagedKubernetesNodepoolsEndpointFormat),
		Run:               cloud.ListManagedKubernetesNodepoolTasks,
		Args:              cobra.ExactArgs(2),
	}))

	nodepoolTaskCmd.AddCommand(&cobra.Command{
		Use:               "get <cluster_id> <nodepool_id> <task_id>",
		Short:             "Get the given Kubernetes node pool task",
		ValidArgsFunction: completion.CloudResourceChain(cloud.ManagedKubernetesV2EndpointFormat, cloud.ManagedKubernetesNodepoolsEndpointFormat, cloud.ManagedKubernetesNodepoolTasksEndpointFormat),
		Run:               cloud.GetManagedKubernetesNodepoolTask,
		Args:              cobra.ExactArgs(3),
	})

	nodepoolEventCmd := &cobra.Command{
		Use:   "event",
		Short: "Inspect the events of a Kubernetes node pool",
	}
	nodepoolCmd.AddCommand(nodepoolEventCmd)

	nodepoolEventListCmd := &cobra.Command{
		Use:               "list <cluster_id> <nodepool_id>",
		Aliases:           []string{"ls"},
		Short:             "List the events of the given Kubernetes node pool",
		ValidArgsFunction: completion.CloudResourceWithChild(cloud.ManagedKubernetesV2EndpointFormat, cloud.ManagedKubernetesNodepoolsEndpointFormat),
		Run:               cloud.ListManagedKubernetesNodepoolEvents,
		Args:              cobra.ExactArgs(2),
	}
	nodepoolEventListCmd.Flags().StringVar(&cloud.ManagedKubernetesEventTaskID, "task-id", "", "List only the events of the given task")
	nodepoolEventCmd.AddCommand(withFilterFlag(nodepoolEventListCmd))

	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Inspect the asynchronous tasks of a Kubernetes cluster",
	}
	managedKubernetesCmd.AddCommand(taskCmd)

	taskCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list <cluster_id>",
		Aliases:           []string{"ls"},
		Short:             "List the tasks of the given Kubernetes cluster",
		ValidArgsFunction: completion.CloudResources(cloud.ManagedKubernetesV2EndpointFormat),
		Run:               cloud.ListManagedKubernetesTasks,
		Args:              cobra.ExactArgs(1),
	}))

	taskCmd.AddCommand(&cobra.Command{
		Use:               "get <cluster_id> <task_id>",
		Short:             "Get the given Kubernetes cluster task",
		ValidArgsFunction: completion.CloudResourceWithChild(cloud.ManagedKubernetesV2EndpointFormat, cloud.ManagedKubernetesTasksEndpointFormat),
		Run:               cloud.GetManagedKubernetesTask,
		Args:              cobra.ExactArgs(2),
	})

	eventCmd := &cobra.Command{
		Use:   "event",
		Short: "Inspect the events of a Kubernetes cluster",
	}
	managedKubernetesCmd.AddCommand(eventCmd)

	eventListCmd := &cobra.Command{
		Use:               "list <cluster_id>",
		Aliases:           []string{"ls"},
		Short:             "List the events of the given Kubernetes cluster",
		ValidArgsFunction: completion.CloudResources(cloud.ManagedKubernetesV2EndpointFormat),
		Run:               cloud.ListManagedKubernetesEvents,
		Args:              cobra.ExactArgs(1),
	}
	eventListCmd.Flags().StringVar(&cloud.ManagedKubernetesEventTaskID, "task-id", "", "List only the events of the given task")
	eventCmd.AddCommand(withFilterFlag(eventListCmd))

	usageCmd := &cobra.Command{
		Use:   "usage",
		Short: "Inspect the resource usage of a Kubernetes cluster",
	}
	managedKubernetesCmd.AddCommand(usageCmd)

	usageCmd.AddCommand(&cobra.Command{
		Use:               "etcd <cluster_id>",
		Short:             "Get the etcd storage usage of the given Kubernetes cluster",
		ValidArgsFunction: completion.CloudResources(cloud.ManagedKubernetesV2EndpointFormat),
		Run:               cloud.GetManagedKubernetesEtcdUsage,
		Args:              cobra.ExactArgs(1),
	})

	planCmd := &cobra.Command{
		Use:   "plan",
		Short: "Explore the Kubernetes cluster plans",
	}
	managedKubernetesCmd.AddCommand(planCmd)

	planListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List the plans available for Kubernetes cluster creation",
		Run:     cloud.ListManagedKubernetesPlans,
		Args:    cobra.NoArgs,
	}
	planListCmd.Flags().StringVar(&cloud.ManagedKubernetesReferenceRegion, "region", "", "Region to list the plans for (required)")
	planListCmd.MarkFlagRequired("region") //nolint:errcheck
	planCmd.AddCommand(withFilterFlag(planListCmd))

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Explore the Kubernetes versions",
	}
	managedKubernetesCmd.AddCommand(versionCmd)

	versionCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List the versions available for Kubernetes cluster creation",
		Run:     cloud.ListManagedKubernetesVersions,
		Args:    cobra.NoArgs,
	}))

	versionCmd.AddCommand(&cobra.Command{
		Use:               "capabilities <cluster_id>",
		Short:             "Get the patch/minor versions the given Kubernetes cluster can be upgraded to",
		ValidArgsFunction: completion.CloudResources(cloud.ManagedKubernetesV2EndpointFormat),
		Run:               cloud.GetManagedKubernetesVersionCapabilities,
		Args:              cobra.ExactArgs(1),
	})

	regionCmd := &cobra.Command{
		Use:   "region",
		Short: "Explore the Kubernetes regions",
	}
	managedKubernetesCmd.AddCommand(regionCmd)

	regionCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List the regions available for Kubernetes cluster creation",
		Run:     cloud.ListManagedKubernetesRegions,
		Args:    cobra.NoArgs,
	}))

	flavorCmd := &cobra.Command{
		Use:   "flavor",
		Short: "Explore the Kubernetes node flavors",
	}
	managedKubernetesCmd.AddCommand(flavorCmd)

	flavorListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List the node flavors available in the given region",
		Run:     cloud.ListManagedKubernetesFlavors,
		Args:    cobra.NoArgs,
	}
	flavorListCmd.Flags().StringVar(&cloud.ManagedKubernetesReferenceRegion, "region", "", "Region to list the flavors for (required)")
	flavorListCmd.MarkFlagRequired("region") //nolint:errcheck
	flavorCmd.AddCommand(withFilterFlag(flavorListCmd))

	managedKubernetesRestartCmd := &cobra.Command{
		Use:               "restart <cluster_id>",
		Short:             "Restart control plane apiserver to invalidate cache without downtime",
		ValidArgsFunction: completion.CloudResources(cloud.ManagedKubernetesV2EndpointFormat),
		Run:               cloud.RestartManagedKubernetes,
		Args:              cobra.ExactArgs(1),
	}
	managedKubernetesRestartCmd.Flags().BoolVar(&cloud.ManagedKubernetesForceAction, "force", false, "Force restart the Kubernetes cluster (will create a slight downtime)")
	managedKubernetesCmd.AddCommand(managedKubernetesRestartCmd)

	cloudCmd.AddCommand(managedKubernetesCmd)
}

func getManagedKubernetesCreateCmd() *cobra.Command {
	managedKubernetesCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Kubernetes cluster",
		Long: `Use this command to create a managed Kubernetes cluster in the given public cloud project.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud managed-kubernetes create --name MyNewCluster --region SBG5 --plan FREE

2. Using a configuration file:

  First you can generate an example of installation file using the following command:

	ovhcloud cloud managed-kubernetes create --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud managed-kubernetes create --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud managed-kubernetes create

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud managed-kubernetes create --from-file ./params.json --name NameOverriden

3. Using your default text editor:

	ovhcloud cloud managed-kubernetes create --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-kubernetes create --editor --region BHS5
`,
		Run: cloud.CreateManagedKubernetes,
	}

	// Immutable parameters, only settable at creation time
	managedKubernetesCreateCmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Region, "region", "", "Region for the Kubernetes cluster (required)")
	managedKubernetesCreateCmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Plan, "plan", "", "Plan of the Kubernetes cluster: FREE, STANDARD (required)")
	managedKubernetesCreateCmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Type, "cni-type", "", "CNI plugin of the cluster (CANAL for the FREE plan, CILIUM for the STANDARD plan). Derived from --plan when omitted")
	managedKubernetesCreateCmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.IPAllocationPolicy.Pods, "ip-allocation-policy-pods", "", "IPv4 CIDR block allocated for pod networking (e.g. '10.0.0.0/16')")
	managedKubernetesCreateCmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.IPAllocationPolicy.Services, "ip-allocation-policy-services", "", "IPv4 CIDR block allocated for service networking (e.g. '10.96.0.0/12')")

	// Network topology, immutable after creation except for the default egress
	managedKubernetesCreateCmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Network.Type, "network-type", "", "Network topology of the cluster nodes (PUBLIC, VRACK). When omitted, VRACK if a vRack private network is given, PUBLIC otherwise")
	managedKubernetesCreateCmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Network.Vrack.PrivateNetworkId, "network-vrack-private-network-id", "", "Private network ID used by the cluster nodes (VRACK topology)")
	managedKubernetesCreateCmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Network.Vrack.SubnetId, "network-vrack-subnet-id", "", "Subnet ID used by the cluster nodes (VRACK topology)")
	managedKubernetesCreateCmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Network.Vrack.LoadBalancerSubnetId, "network-vrack-load-balancer-subnet-id", "", "Subnet ID used by the load balancers (VRACK topology)")
	managedKubernetesCreateCmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Canal.KubeProxy.Mode, "kube-proxy-mode", "", "Kube-proxy mode (IPTABLES, IPVS), immutable after creation")

	setManagedKubernetesMutableFlags(managedKubernetesCreateCmd)

	// Common flags for other means to define parameters.
	addParameterFileFlags(managedKubernetesCreateCmd, false, assets.CloudV2OpenapiSchema, cloud.ManagedKubernetesV2SchemaPath, "post", "", nil)
	setPreRunE(managedKubernetesCreateCmd, func(cmd *cobra.Command, args []string) error {
		// Required only with flags alone: a file, the editor or a pipe can provide them.
		if flags.ParametersFile == "" && !flags.ParametersViaEditor && !utils.IsInputFromPipe() {
			var missing []string
			for _, name := range []string{"region", "plan"} {
				if !cmd.Flags().Changed(name) {
					missing = append(missing, fmt.Sprintf("%q", name))
				}
			}
			if len(missing) > 0 {
				return fmt.Errorf("required flag(s) %s not set", strings.Join(missing, ", "))
			}
		}
		return managedKubernetesPreRunE(cmd, args)
	})
	addInteractiveEditorFlag(managedKubernetesCreateCmd)
	markFlagsMutuallyExclusive(managedKubernetesCreateCmd, "from-file", "editor")

	return managedKubernetesCreateCmd
}

func getManagedKubernetesEditCmd() *cobra.Command {
	managedKubernetesEditCmd := &cobra.Command{
		Use:               "edit <cluster_id>",
		Short:             "Edit the given Kubernetes cluster",
		Long:              "Edit the given Kubernetes cluster. Only the mutable part of the cluster specification can be changed: to change the region, the plan, the network topology or the pods and services CIDRs, a new cluster must be created.",
		ValidArgsFunction: completion.CloudResources(cloud.ManagedKubernetesV2EndpointFormat),
		Run:               cloud.EditManagedKubernetes,
		Args:              cobra.ExactArgs(1),
	}

	setManagedKubernetesMutableFlags(managedKubernetesEditCmd)

	setPreRunE(managedKubernetesEditCmd, managedKubernetesPreRunE)
	addInteractiveEditorFlag(managedKubernetesEditCmd)

	return managedKubernetesEditCmd
}

// setManagedKubernetesMutableFlags registers the flags mapping onto the mutable part of the cluster
// specification, shared by the create and edit commands.
func setManagedKubernetesMutableFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Name, "name", "", "Name of the Kubernetes cluster (auto-generated at creation if omitted)")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Version, "version", "", "Kubernetes version (a minor version such as '1.34' is only accepted at creation time, which defaults to the latest supported version if omitted)")
	cmd.Flags().BoolVar(&autoUpgradePatchEnabled, "auto-upgrade-patch-enabled", true, "Automatically upgrade the cluster to the latest patch version")

	// API server
	cmd.Flags().StringSliceVar(&apiServerAuthorizedIPs, "api-server-authorized-ips", nil, "CIDR blocks allowed to reach the API server, '0.0.0.0/0' to allow all (replaces the current allowlist when editing)")
	cmd.Flags().StringSliceVar(&admissionPluginsEnabled, "api-server-admission-plugins-enabled", nil, "Admission plugins to enable on API server (AlwaysPullImages, NodeRestriction)")
	cmd.Flags().StringSliceVar(&admissionPluginsDisabled, "api-server-admission-plugins-disabled", nil, "Admission plugins to disable on API server (AlwaysPullImages, NodeRestriction)")

	// Default egress of the cluster nodes, only relevant for a VRACK network topology
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Network.Vrack.DefaultEgress.Type, "network-vrack-default-egress-type", "", "Default egress route of the cluster nodes (MANAGED_GATEWAY, NODE_PUBLIC_INTERFACE, USER_GATEWAY). Defaults to MANAGED_GATEWAY on STANDARD plan, NODE_PUBLIC_INTERFACE on FREE plan")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Network.Vrack.DefaultEgress.UserGateway.IP, "network-vrack-default-egress-user-gateway-ip", "", "IP address of the egress gateway, required when the default egress type is USER_GATEWAY")

	// Kube-proxy, only available with the CANAL CNI
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Canal.KubeProxy.IPTables.MinSyncPeriod, "kube-proxy-iptables-min-sync-period", "", "Minimum period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Canal.KubeProxy.IPTables.SyncPeriod, "kube-proxy-iptables-sync-period", "", "Period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Canal.KubeProxy.IPVS.MinSyncPeriod, "kube-proxy-ipvs-min-sync-period", "", "Minimum period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Canal.KubeProxy.IPVS.Scheduler, "kube-proxy-ipvs-scheduler", "", "Scheduler for kube-proxy ipvs (DESTINATION_HASHING, LEAST_CONNECTION, NEVER_QUEUE, ROUND_ROBIN, SHORTEST_EXPECTED_DELAY, SOURCE_HASHING)")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Canal.KubeProxy.IPVS.SyncPeriod, "kube-proxy-ipvs-sync-period", "", "Period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Canal.KubeProxy.IPVS.TCPFinTimeout, "kube-proxy-ipvs-tcp-fin-timeout", "", "Timeout value used for IPVS TCP sessions after receiving a FIN in RFC3339 duration format (e.g. 'PT60S')")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Canal.KubeProxy.IPVS.TCPTimeout, "kube-proxy-ipvs-tcp-timeout", "", "Timeout value used for idle IPVS TCP sessions in RFC3339 duration format (e.g. 'PT60S')")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Canal.KubeProxy.IPVS.UDPTimeout, "kube-proxy-ipvs-udp-timeout", "", "Timeout value used for IPVS UDP packets in RFC3339 duration format (e.g. 'PT60S')")

	// Cilium, only available with the CILIUM CNI
	cmd.Flags().Uint8Var(&ciliumClusterID, "cilium-cluster-id", 0, "Cilium cluster ID")
	cmd.Flags().BoolVar(&hubbleEnabled, "cilium-hubble-enabled", false, "Enable Hubble observability")
	cmd.Flags().BoolVar(&hubbleRelayEnabled, "cilium-hubble-relay-enabled", false, "Enable Hubble Relay")
	cmd.Flags().BoolVar(&hubbleUIEnabled, "cilium-hubble-ui-enabled", false, "Enable Hubble UI")
	cmd.Flags().BoolVar(&clusterMeshEnabled, "cilium-cluster-mesh-enabled", false, "Enable Cilium ClusterMesh")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Cilium.ClusterMesh.APIServer.ServiceType, "cilium-cluster-mesh-apiserver-service-type", "", "ClusterMesh API server service type (LoadBalancer, NodePort)")
	cmd.Flags().Uint16Var(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Cilium.ClusterMesh.APIServer.NodePort, "cilium-cluster-mesh-apiserver-node-port", 0, "ClusterMesh API server node port")

	// Hubble UI frontend resources
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Cilium.Hubble.UI.FrontendResources.Limits.Cpu, "cilium-hubble-ui-frontend-limits-cpu", "", "Hubble UI frontend CPU limit (e.g. '500m')")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Cilium.Hubble.UI.FrontendResources.Limits.Memory, "cilium-hubble-ui-frontend-limits-memory", "", "Hubble UI frontend memory limit (e.g. '256Mi')")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Cilium.Hubble.UI.FrontendResources.Requests.Cpu, "cilium-hubble-ui-frontend-requests-cpu", "", "Hubble UI frontend CPU request (e.g. '100m')")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Cilium.Hubble.UI.FrontendResources.Requests.Memory, "cilium-hubble-ui-frontend-requests-memory", "", "Hubble UI frontend memory request (e.g. '128Mi')")

	// Hubble UI backend resources
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Cilium.Hubble.UI.BackendResources.Limits.Cpu, "cilium-hubble-ui-backend-limits-cpu", "", "Hubble UI backend CPU limit (e.g. '500m')")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Cilium.Hubble.UI.BackendResources.Limits.Memory, "cilium-hubble-ui-backend-limits-memory", "", "Hubble UI backend memory limit (e.g. '256Mi')")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Cilium.Hubble.UI.BackendResources.Requests.Cpu, "cilium-hubble-ui-backend-requests-cpu", "", "Hubble UI backend CPU request (e.g. '100m')")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Cni.Cilium.Hubble.UI.BackendResources.Requests.Memory, "cilium-hubble-ui-backend-requests-memory", "", "Hubble UI backend memory request (e.g. '128Mi')")

	// OpenID Connect authentication
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Authentication.Oidc.CaContent, "oidc-ca-content", "", "CA certificate content for the OIDC provider")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Authentication.Oidc.ClientId, "oidc-client-id", "", "OIDC client ID")
	cmd.Flags().StringSliceVar(&cloud.ManagedKubernetesSpec.TargetSpec.Authentication.Oidc.GroupsClaim, "oidc-groups-claim", nil, "OIDC groups claim(s)")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Authentication.Oidc.GroupsPrefix, "oidc-groups-prefix", "", "Prefix prepended to group claims")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Authentication.Oidc.IssuerUrl, "oidc-issuer-url", "", "OIDC issuer URL")
	cmd.Flags().StringSliceVar(&cloud.ManagedKubernetesSpec.TargetSpec.Authentication.Oidc.RequiredClaim, "oidc-required-claim", nil, "OIDC required claim(s)")
	cmd.Flags().StringSliceVar(&cloud.ManagedKubernetesSpec.TargetSpec.Authentication.Oidc.SigningAlgs, "oidc-signing-algs", nil, "OIDC signing algorithm(s) (ES256, ES384, ES512, PS256, PS384, PS512, RS256, RS384, RS512)")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Authentication.Oidc.UsernameClaim, "oidc-username-claim", "", "OIDC username claim")
	cmd.Flags().StringVar(&cloud.ManagedKubernetesSpec.TargetSpec.Authentication.Oidc.UsernamePrefix, "oidc-username-prefix", "", "Prefix prepended to username claims")
}

// setPreRunE installs preRunE, keeping the PreRun previously set by addParameterFileFlags:
// cobra ignores PreRun as soon as PreRunE is defined.
func setPreRunE(cmd *cobra.Command, preRunE func(*cobra.Command, []string) error) {
	parameterFilePreRun := cmd.PreRun
	cmd.PreRunE = func(c *cobra.Command, args []string) error {
		if parameterFilePreRun != nil {
			parameterFilePreRun(c, args)
		}
		return preRunE(c, args)
	}
}

// managedKubernetesPreRunE builds the parts of the cluster specification that don't map
// directly onto a flag: optional booleans must stay absent when not given, and the
// IP allowlist and the admission plugins are flat flags feeding nested structures.
func managedKubernetesPreRunE(cmd *cobra.Command, _ []string) error {
	spec := &cloud.ManagedKubernetesSpec.TargetSpec

	spec.AutoUpgrade.Patch.Enabled = flagValueIfSet(cmd, "auto-upgrade-patch-enabled", &autoUpgradePatchEnabled)
	spec.Cni.Cilium.ClusterID = flagValueIfSet(cmd, "cilium-cluster-id", &ciliumClusterID)
	spec.Cni.Cilium.ClusterMesh.Enabled = flagValueIfSet(cmd, "cilium-cluster-mesh-enabled", &clusterMeshEnabled)
	spec.Cni.Cilium.Hubble.Enabled = flagValueIfSet(cmd, "cilium-hubble-enabled", &hubbleEnabled)
	spec.Cni.Cilium.Hubble.Relay.Enabled = flagValueIfSet(cmd, "cilium-hubble-relay-enabled", &hubbleRelayEnabled)
	spec.Cni.Cilium.Hubble.UI.Enabled = flagValueIfSet(cmd, "cilium-hubble-ui-enabled", &hubbleUIEnabled)

	var rules []cloud.ManagedKubernetesAuthorizedIpRule
	for _, cidr := range apiServerAuthorizedIPs {
		rules = append(rules, cloud.ManagedKubernetesAuthorizedIpRule{Cidr: cidr})
	}
	spec.ApiServer.Access.AuthorizedIps.Rules = rules

	// Stays nil when empty: a non-nil map would make the omitzero apiServer object non-zero.
	var plugins map[string]cloud.ManagedKubernetesAdmissionPlugin
	if len(admissionPluginsEnabled)+len(admissionPluginsDisabled) > 0 {
		plugins = make(map[string]cloud.ManagedKubernetesAdmissionPlugin, len(admissionPluginsEnabled)+len(admissionPluginsDisabled))
		for _, plugin := range admissionPluginsEnabled {
			plugins[plugin] = cloud.ManagedKubernetesAdmissionPlugin{Enabled: true}
		}
		for _, plugin := range admissionPluginsDisabled {
			if _, ok := plugins[plugin]; ok {
				return fmt.Errorf("admission plugin %q is both enabled and disabled", plugin)
			}
			plugins[plugin] = cloud.ManagedKubernetesAdmissionPlugin{Enabled: false}
		}
	}
	spec.ApiServer.AdmissionPlugins = plugins

	// The OIDC payload is only accepted along with its authentication variant.
	spec.Authentication.Type = ""
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if strings.HasPrefix(f.Name, "oidc-") {
			spec.Authentication.Type = "OIDC"
		}
	})

	return nil
}

// flagValueIfSet returns value only when the flag was explicitly given, so that an
// optional field stays out of the request body otherwise.
func flagValueIfSet[T any](cmd *cobra.Command, name string, value *T) *T {
	if !cmd.Flags().Changed(name) {
		return nil
	}
	return value
}

func getManagedKubernetesNodepoolEditCmd() *cobra.Command {
	nodepoolEditCmd := &cobra.Command{
		Use:               "edit <cluster_id> <nodepool_id>",
		Short:             "Edit the given Kubernetes node pool",
		Long:              "Edit the given Kubernetes node pool. Only the node count, the autoscaler and the node template can be changed: the name, flavor, availability zone, billing type, anti-affinity and floating IP settings are immutable.",
		ValidArgsFunction: completion.CloudResourceWithChild(cloud.ManagedKubernetesV2EndpointFormat, cloud.ManagedKubernetesNodepoolsEndpointFormat),
		Run:               cloud.EditManagedKubernetesNodepool,
		Args:              cobra.ExactArgs(2),
	}

	setManagedKubernetesNodepoolMutableFlags(nodepoolEditCmd)
	nodepoolEditCmd.PreRunE = managedKubernetesNodepoolPreRunE
	addInteractiveEditorFlag(nodepoolEditCmd)

	return nodepoolEditCmd
}

func getManagedKubernetesNodepoolCreateCmd() *cobra.Command {
	nodepoolCreateCmd := &cobra.Command{
		Use:   "create <cluster_id>",
		Short: "Create a new Kubernetes node pool",
		Long: `Use this command to create a node pool in the given managed Kubernetes cluster.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud managed-kubernetes nodepool create <cluster_id> --flavor b3-16 --desired-nodes 3 --name newnodepool

2. Using a configuration file:

  First you can generate an example of installation file using the following command:

	ovhcloud cloud managed-kubernetes nodepool create <cluster_id> --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud managed-kubernetes nodepool create <cluster_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud managed-kubernetes nodepool create <cluster_id>

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud managed-kubernetes nodepool create <cluster_id> --from-file ./params.json --name NameOverriden

  It is also possible to use the interactive flavor selector to define the flavor parameter, like the following:

	ovhcloud cloud managed-kubernetes nodepool create <cluster_id> --init-file ./params.json --flavor-selector

3. Using your default text editor:

	ovhcloud cloud managed-kubernetes nodepool create <cluster_id> --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-kubernetes nodepool create <cluster_id> --editor --flavor b3-16

  You can also use the interactive flavor selector to define the flavor parameter, like the following:

	ovhcloud cloud managed-kubernetes nodepool create <cluster_id> --editor --flavor-selector
`,
		ValidArgsFunction: completion.CloudResources(cloud.ManagedKubernetesV2EndpointFormat),
		Run:               cloud.CreateManagedKubernetesNodepool,
		Args:              cobra.ExactArgs(1),
	}

	// Immutable parameters, only settable at creation time
	spec := &cloud.ManagedKubernetesNodepoolSpec.TargetSpec
	nodepoolCreateCmd.Flags().StringVar(&spec.Name, "name", "", "Name of the node pool")
	nodepoolCreateCmd.Flags().StringVar(&spec.NodeConfig.Flavor, "flavor", "", "Flavor of the nodes, e.g. b3-8, b3-16 (required)")
	nodepoolCreateCmd.Flags().BoolVar(&attachFloatingIpEnabled, "attach-floating-ip", false, "Attach a floating IP to each node")
	nodepoolCreateCmd.Flags().StringVar(&spec.AvailabilityZone, "availability-zone", "", "Availability zone the nodes are placed in, only available in 3-AZ regions")
	nodepoolCreateCmd.Flags().StringVar(&spec.BillingType, "billing-type", "", "Billing period of the nodes (HOURLY, MONTHLY)")
	nodepoolCreateCmd.Flags().BoolVar(&antiAffinityEnabled, "anti-affinity-enabled", false, "Spread the nodes across distinct hypervisors, capping the node pool size to 5")

	setManagedKubernetesNodepoolMutableFlags(nodepoolCreateCmd)

	// Common flags for other means to define parameters
	addParameterFileFlags(nodepoolCreateCmd, false, assets.CloudV2OpenapiSchema, cloud.ManagedKubernetesV2NodepoolsSchemaPath, "post", "", cloud.GetManagedKubernetesFlavorInteractiveSelector)

	setPreRunE(nodepoolCreateCmd, func(cmd *cobra.Command, args []string) error {
		// Required only with flags alone: a file, the editor, a pipe or the selector can provide it.
		if flags.ParametersFile == "" && !flags.ParametersViaEditor && !utils.IsInputFromPipe() &&
			!cmd.Flags().Changed("flavor") && !cmd.Flags().Changed("flavor-selector") {
			return errors.New(`required flag(s) "flavor" not set`)
		}
		return managedKubernetesNodepoolPreRunE(cmd, args)
	})

	addInteractiveEditorFlag(nodepoolCreateCmd)
	if !(runtime.GOARCH == "wasm" && runtime.GOOS == "js") {
		nodepoolCreateCmd.Flags().BoolVar(&cloud.InstanceFlavorViaInteractiveSelector, "flavor-selector", false, "Use the interactive flavor selector")
	}
	markFlagsMutuallyExclusive(nodepoolCreateCmd, "from-file", "editor")

	return nodepoolCreateCmd
}

// setManagedKubernetesNodepoolMutableFlags registers the flags mapping onto the mutable part of the
// node pool specification, shared by the create and edit commands.
func setManagedKubernetesNodepoolMutableFlags(cmd *cobra.Command) {
	spec := &cloud.ManagedKubernetesNodepoolSpec.TargetSpec
	cmd.Flags().IntVar(&desiredNodes, "desired-nodes", 0, "Desired number of nodes")

	// Autoscaler
	cmd.Flags().BoolVar(&autoscalerEnabled, "autoscaler-enabled", false, "Enable node autoscaling")
	cmd.Flags().IntVar(&autoscalerMinNodes, "autoscaler-min-nodes", 0, "Minimum number of nodes the autoscaler maintains")
	cmd.Flags().IntVar(&autoscalerMaxNodes, "autoscaler-max-nodes", 0, "Maximum number of nodes the autoscaler provisions")
	cmd.Flags().IntVar(&spec.Autoscaler.ScaleDown.UnneededTimeSeconds, "autoscaler-scale-down-unneeded-time-seconds", 0, "Time in seconds a node must stay below the utilization threshold before it is removed")
	cmd.Flags().IntVar(&spec.Autoscaler.ScaleDown.UnreadyTimeSeconds, "autoscaler-scale-down-unready-time-seconds", 0, "Time in seconds an unready node is kept before it is removed")
	cmd.Flags().Float64Var(&spec.Autoscaler.ScaleDown.UtilizationThreshold, "autoscaler-scale-down-utilization-threshold", 0, "Node utilization ratio (between 0 and 1) below which a node becomes a scale-down candidate")

	// Node template
	cmd.Flags().StringToStringVar(&spec.Template.Metadata.Annotations, "template-annotations", nil, "Annotations to apply to each node")
	cmd.Flags().StringSliceVar(&spec.Template.Metadata.Finalizers, "template-finalizers", nil, "Finalizers to apply to each node")
	cmd.Flags().StringToStringVar(&spec.Template.Metadata.Labels, "template-labels", nil, "Labels to apply to each node")
	cmd.Flags().StringSliceVar(&templateTaints, "template-taints", nil, "Taints to apply to each node in key[=value]:effect format (effect: NoExecute, NoSchedule, PreferNoSchedule)")
	cmd.Flags().BoolVar(&templateUnschedulable, "template-unschedulable", false, "Mark the nodes as unschedulable")
}

// managedKubernetesNodepoolPreRunE builds the parts of the node pool specification that don't map
// directly onto a flag: optional values must stay absent when not given, and taints need parsing.
func managedKubernetesNodepoolPreRunE(cmd *cobra.Command, _ []string) error {
	spec := &cloud.ManagedKubernetesNodepoolSpec.TargetSpec

	spec.DesiredNodes = flagValueIfSet(cmd, "desired-nodes", &desiredNodes)
	spec.Autoscaler.Enabled = flagValueIfSet(cmd, "autoscaler-enabled", &autoscalerEnabled)
	spec.Autoscaler.MinNodes = flagValueIfSet(cmd, "autoscaler-min-nodes", &autoscalerMinNodes)
	spec.Autoscaler.MaxNodes = flagValueIfSet(cmd, "autoscaler-max-nodes", &autoscalerMaxNodes)
	spec.AntiAffinity.Enabled = flagValueIfSet(cmd, "anti-affinity-enabled", &antiAffinityEnabled)
	spec.NodeConfig.AttachFloatingIp.Enabled = flagValueIfSet(cmd, "attach-floating-ip", &attachFloatingIpEnabled)
	spec.Template.Spec.Unschedulable = flagValueIfSet(cmd, "template-unschedulable", &templateUnschedulable)

	var taints []cloud.ManagedKubernetesTaint
	for _, taint := range templateTaints {
		// The value is optional, as with kubectl taint.
		keyValue, effect, ok := strings.Cut(taint, ":")
		key, value, _ := strings.Cut(keyValue, "=")
		if !ok || key == "" || effect == "" {
			return fmt.Errorf("invalid taint format: %q, expected format is key[=value]:effect", taint)
		}
		taints = append(taints, cloud.ManagedKubernetesTaint{Key: key, Value: value, Effect: effect})
	}
	spec.Template.Spec.Taints = taints

	return nil
}
