// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudReferenceCmd(cloudCmd *cobra.Command) {
	referenceCmd := &cobra.Command{
		Use:   "reference",
		Short: "Fetch reference data in the given cloud project",
	}

	referenceCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	// Flavors
	var region string
	flavorListCmd := withFilterFlag(&cobra.Command{
		Use:   "list-flavors",
		Short: "List available flavors in the given cloud project",
		Run: func(_ *cobra.Command, _ []string) {
			cloud.GetFlavors(region)
		},
		Args: cobra.NoArgs,
	})
	flavorListCmd.Flags().StringVarP(&region, "region", "r", "", "Region to filter flavors (e.g., GRA9, BHS5)")
	referenceCmd.AddCommand(flavorListCmd)

	// Images
	var osType string
	imageListCmd := withFilterFlag(&cobra.Command{
		Use:   "list-images",
		Short: "List available images in the given cloud project",
		Run: func(_ *cobra.Command, _ []string) {
			cloud.GetImages(region, osType)
		},
		Args: cobra.NoArgs,
	})
	imageListCmd.Flags().StringVarP(&region, "region", "r", "", "Region to filter images (e.g., GRA9, BHS5)")
	imageListCmd.Flags().StringVarP(&osType, "os-type", "o", "", "OS type to filter images (baremetal-linux, bsd, linux, windows)")
	referenceCmd.AddCommand(imageListCmd)

	// Container registry reference commands
	containerRegistryReferenceCmd := &cobra.Command{
		Use:   "container-registry",
		Short: "Fetch container registry reference data in the given cloud project",
	}
	referenceCmd.AddCommand(containerRegistryReferenceCmd)

	containerRegistryReferenceCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list-plans",
		Short: "List available container registry plans in the given cloud project",
		Run:   cloud.ListContainerRegistryPlans,
		Args:  cobra.NoArgs,
	}))

	// Rancher reference commands
	rancherReferenceCmd := &cobra.Command{
		Use:   "rancher",
		Short: "Fetch Rancher reference data in the given cloud project",
	}
	referenceCmd.AddCommand(rancherReferenceCmd)

	rancherVersionsListCmd := withFilterFlag(&cobra.Command{
		Use:   "list-versions",
		Short: "List available Rancher versions in the given cloud project",
		Run:   cloud.ListRancherAvailableVersions,
		Args:  cobra.NoArgs,
	})
	rancherVersionsListCmd.Flags().StringP("rancher-id", "r", "", "Rancher service ID to filter available versions")
	rancherReferenceCmd.AddCommand(rancherVersionsListCmd)

	rancherPlansListCmd := withFilterFlag(&cobra.Command{
		Use:   "list-plans",
		Short: "List available Rancher plans in the given cloud project",
		Run:   cloud.ListRancherAvailablePlans,
		Args:  cobra.NoArgs,
	})
	rancherPlansListCmd.Flags().StringP("rancher-id", "r", "", "Rancher service ID to filter available plans")
	rancherReferenceCmd.AddCommand(rancherPlansListCmd)

	// Databases reference commands
	databaseReferenceCmd := &cobra.Command{
		Use:   "database",
		Short: "Fetch database reference data in the given cloud project",
	}
	referenceCmd.AddCommand(databaseReferenceCmd)

	databaseReferenceCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list-plans",
		Short: "List available database plans in the given cloud project",
		Run:   cloud.ListDatabasesPlans,
		Args:  cobra.NoArgs,
	}))

	databaseReferenceCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list-node-flavors",
		Short: "List available database node flavors in the given cloud project",
		Run:   cloud.ListDatabasesNodeFlavors,
		Args:  cobra.NoArgs,
	}))

	databaseReferenceCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list-engines",
		Short: "List available database engines in the given cloud project",
		Run:   cloud.ListDatabaseEngines,
		Args:  cobra.NoArgs,
	}))

	// Loadbalancer reference commands
	loadbalancerReferenceCmd := &cobra.Command{
		Use:   "loadbalancer",
		Short: "Fetch loadbalancer reference data in the given cloud project",
	}
	referenceCmd.AddCommand(loadbalancerReferenceCmd)

	loadbalancerReferenceCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list-flavors <region (GRA9, BHS5, ...)>",
		Short: "List available loadbalancer flavors in the given cloud project",
		Run:   cloud.ListLoadbalancerFlavors,
		Args:  cobra.ExactArgs(1),
	}))

	cloudCmd.AddCommand(referenceCmd)
}
