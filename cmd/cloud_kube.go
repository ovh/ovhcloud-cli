package cmd

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var (
	cloudprojectKubeColumnsToDisplay = []string{"id", "name", "region", "version", "status"}

	//go:embed templates/cloud_kube.tmpl
	cloudKubeTemplate string
)

func listKubes(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageListRequest(fmt.Sprintf("/cloud/project/%s/kube", projectID), "", cloudprojectKubeColumnsToDisplay, genericFilters)
}

func getKube(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageObjectRequest(fmt.Sprintf("/cloud/project/%s/kube", projectID), args[0], cloudKubeTemplate)
}

func initKubeCommand(cloudCmd *cobra.Command) {
	kubeCmd := &cobra.Command{
		Use:   "kube",
		Short: "List Kubernetes clusters in the given cloud project",
	}
	kubeCmd.PersistentFlags().StringVar(&cloudProject, "cloud-project", "", "Cloud project ID")

	// Command to list Kuberetes clusters
	kubeListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Kubernetes clusters",
		Run:   listKubes,
	}
	kubeCmd.AddCommand(withFilterFlag(kubeListCmd))

	kubeCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Get a specific Kubernetes cluster",
		Run:        getKube,
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"cluster_id"},
	})

	cloudCmd.AddCommand(kubeCmd)
}
