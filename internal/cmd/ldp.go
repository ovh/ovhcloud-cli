package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/ldp"
)

func init() {
	ldpCmd := &cobra.Command{
		Use:   "ldp",
		Short: "Retrieve information and manage your Ldp services",
	}

	// Command to list Ldp services
	ldpListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your Ldp services",
		Run:     ldp.ListLdp,
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
	ldpEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given Ldp",
		Args:  cobra.ExactArgs(1),
		Run:   ldp.EditLdp,
	}
	ldpEditCmd.Flags().StringVar(&ldp.LdpSpec.DisplayName, "display-name", "", "Display name of the LDP")
	ldpEditCmd.Flags().BoolVar(&ldp.LdpSpec.EnableIAM, "enable-iam", false, "Enable IAM for the LDP")
	ldpEditCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define parameters")
	ldpCmd.AddCommand(ldpEditCmd)

	rootCmd.AddCommand(ldpCmd)
}
