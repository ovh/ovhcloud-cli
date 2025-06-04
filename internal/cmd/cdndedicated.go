package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/cdndedicated"
)

func init() {
	cdndedicatedCmd := &cobra.Command{
		Use:   "cdn-dedicated",
		Short: "Retrieve information and manage your dedicated CDN services",
	}

	// Command to list CdnDedicated services
	cdndedicatedListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your dedicated CDN services",
		Run:   cdndedicated.ListCdnDedicated,
	}
	cdndedicatedCmd.AddCommand(withFilterFlag(cdndedicatedListCmd))

	// Command to get a single CdnDedicated
	cdndedicatedCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific CDN",
		Args:  cobra.ExactArgs(1),
		Run:   cdndedicated.GetCdnDedicated,
	})

	rootCmd.AddCommand(cdndedicatedCmd)
}
