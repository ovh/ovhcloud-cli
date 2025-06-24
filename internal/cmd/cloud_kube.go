package cmd

import (
	"github.com/spf13/cobra"
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
		Short: "Get a specific Kubernetes cluster",
		Run:   cloud.GetKube,
		Args:  cobra.ExactArgs(1),
	})

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

	cloudCmd.AddCommand(kubeCmd)
}
