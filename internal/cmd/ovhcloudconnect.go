package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/ovhcloudconnect"
)

func init() {
	ovhcloudconnectCmd := &cobra.Command{
		Use:   "ovhcloudconnect",
		Short: "Retrieve information and manage your OvhCloudConnect services",
	}

	// Command to list OvhCloudConnect services
	ovhcloudconnectListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your OvhCloudConnect services",
		Run:   ovhcloudconnect.ListOvhCloudConnect,
	}
	ovhcloudconnectCmd.AddCommand(withFilterFlag(ovhcloudconnectListCmd))

	// Command to get a single OvhCloudConnect
	ovhcloudconnectCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific OvhCloudConnect",
		Args:  cobra.ExactArgs(1),
		Run:   ovhcloudconnect.GetOvhCloudConnect,
	})

	// Command to update a single OvhCloudConnect
	ovhcloudconnectCmd.AddCommand(&cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given OvhCloudConnect",
		Run:   ovhcloudconnect.EditOvhCloudConnect,
	})

	rootCmd.AddCommand(ovhcloudconnectCmd)
}
