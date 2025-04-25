package cmd

import (
	_ "embed"
	"fmt"
	"log"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/config"
)

var (
	cloudprojectKubeColumnsToDisplay = []string{"id", "name", "region", "version", "status"}
	cloudProject                     string

	//go:embed templates/cloud_kube.tmpl
	cloudKubeTemplate string
)

func listKubes(_ *cobra.Command, _ []string) {
	if cloudProject == "" {
		projectID, err := config.GetConfigValue(cliConfig, "", "default_cloud_project")
		if err != nil {
			log.Fatalf("failed to fetch default cloud project: %s", err)
		}
		if projectID == "" {
			log.Fatal("no project ID configured, please use --cloud-project <id> or set a default cloud project in your configuration")
		}
		cloudProject = projectID
	}

	manageListRequest(fmt.Sprintf("/cloud/project/%s/kube", cloudProject), cloudprojectKubeColumnsToDisplay, genericFilters)
}

func getKube(_ *cobra.Command, args []string) {
	if cloudProject == "" {
		projectID, err := config.GetConfigValue(cliConfig, "", "default_cloud_project")
		if err != nil {
			log.Fatalf("failed to fetch default cloud project: %s", err)
		}
		if projectID == "" {
			log.Fatal("no project ID configured, please use --cloud-project <id> or set a default cloud project in your configuration")
		}
		cloudProject = projectID
	}

	manageObjectRequest(fmt.Sprintf("/cloud/project/%s/kube", url.PathEscape(cloudProject)), args[0], cloudKubeTemplate)
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
	kubeCmd.AddCommand(kubeListCmd)

	kubeCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Get a specific Kubernetes cluster",
		Run:        getKube,
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"cluster_id"},
	})

	cloudCmd.AddCommand(kubeCmd)
}
