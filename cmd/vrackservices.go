package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/vrackservices"
)

func init() {
	vrackservicesCmd := &cobra.Command{
		Use:   "vrackservices",
		Short: "Retrieve information and manage your vRackServices services",
	}

	// Command to list VrackServices services
	vrackservicesListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your vRackServices services",
		Run:   vrackservices.ListVrackServices,
	}
	vrackservicesCmd.AddCommand(withFilterFlag(vrackservicesListCmd))

	// Command to get a single VrackServices
	vrackservicesCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific vRackServices",
		Args:  cobra.ExactArgs(1),
		Run:   vrackservices.GetVrackServices,
	})

	// Command to update a single VrackServices
	vrackservicesCmd.AddCommand(&cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given vRackServices",
		Run:   vrackservices.EditVrackServices,
	})

	rootCmd.AddCommand(vrackservicesCmd)
}
