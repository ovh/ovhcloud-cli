package cmd

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var (
	cloudprojectStorageSwiftColumnsToDisplay = []string{"id", "name", "region"}

	//go:embed templates/cloud_storage_swift.tmpl
	cloudStorageSwiftTemplate string
)

func listCloudStorageSwift(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageListRequestNoExpand(fmt.Sprintf("/cloud/project/%s/storage", projectID), cloudprojectStorageSwiftColumnsToDisplay, genericFilters)
}

func getStorageSwift(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	manageObjectRequest(fmt.Sprintf("/cloud/project/%s/storage", projectID), args[0], cloudStorageSwiftTemplate)
}

func initCloudStorageSwiftCommand(cloudCmd *cobra.Command) {
	storageSwiftCmd := &cobra.Command{
		Use:   "storage-swift",
		Short: "Manage SWIFT storage containers in the given cloud project",
	}
	storageSwiftCmd.PersistentFlags().StringVar(&cloudProject, "cloud-project", "", "Cloud project ID")

	storageSwiftListCmd := &cobra.Command{
		Use:   "list",
		Short: "List SWIFT storage containers",
		Run:   listCloudStorageSwift,
	}
	storageSwiftCmd.AddCommand(withFilterFlag(storageSwiftListCmd))

	storageSwiftCmd.AddCommand(&cobra.Command{
		Use:   "get <container_id>",
		Short: "Get a specific SWIFT storage container",
		Run:   getStorageSwift,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(storageSwiftCmd)
}
