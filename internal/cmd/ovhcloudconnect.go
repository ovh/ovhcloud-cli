package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/ovhcloudconnect"
	"github.com/spf13/cobra"
)

func init() {
	ovhcloudconnectCmd := &cobra.Command{
		Use:   "ovhcloudconnect",
		Short: "Retrieve information and manage your OvhCloudConnect services",
	}

	// Command to list OvhCloudConnect services
	ovhcloudconnectListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your OvhCloudConnect services",
		Run:     ovhcloudconnect.ListOvhCloudConnect,
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
	ovhcloudconnectEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given OvhCloudConnect",
		Args:  cobra.ExactArgs(1),
		Run:   ovhcloudconnect.EditOvhCloudConnect,
	}
	ovhcloudconnectEditCmd.Flags().StringVar(&ovhcloudconnect.OvhCloudConnectSpec.Description, "description", "", "Description")
	addInteractiveEditorFlag(ovhcloudconnectEditCmd)
	ovhcloudconnectCmd.AddCommand(ovhcloudconnectEditCmd)

	rootCmd.AddCommand(ovhcloudconnectCmd)
}
