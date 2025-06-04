package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/ldp"
)

func init() {
	ldpCmd := &cobra.Command{
		Use:   "ldp",
		Short: "Retrieve information and manage your Ldp services",
	}

	// Command to list Ldp services
	ldpListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Ldp services",
		Run:   ldp.ListLdp,
	}
	ldpCmd.AddCommand(withFilterFlag(ldpListCmd))

	// Command to get a single Ldp
	ldpCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific Ldp",
		Args:  cobra.ExactArgs(1),
		Run:   ldp.GetLdp,
	})

	// Command to update a single Ldp
	ldpCmd.AddCommand(&cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given Ldp",
		Run:   ldp.EditLdp,
	})

	rootCmd.AddCommand(ldpCmd)
}
