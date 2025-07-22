package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/xdsl"
)

func init() {
	xdslCmd := &cobra.Command{
		Use:   "xdsl",
		Short: "Retrieve information and manage your XDSL services",
	}

	// Command to list Xdsl services
	xdslListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your XDSL services",
		Run:     xdsl.ListXdsl,
	}
	xdslCmd.AddCommand(withFilterFlag(xdslListCmd))

	// Command to get a single Xdsl
	xdslCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific XDSL",
		Args:  cobra.ExactArgs(1),
		Run:   xdsl.GetXdsl,
	})

	// Command to update a single Xdsl
	xdslEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given XDSL",
		Args:  cobra.ExactArgs(1),
		Run:   xdsl.EditXdsl,
	}
	xdslEditCmd.Flags().StringVar(&xdsl.XdslSpec.Description, "description", "", "Description of the XDSL")
	xdslEditCmd.Flags().IntVar(&xdsl.XdslSpec.LnsRateLimit, "lns-rate-limit", 0, "Rate limit on the LNS in kbps. Must be a multiple of 64 - Min value 64 / Max value 100032")
	xdslEditCmd.Flags().BoolVar(&xdsl.XdslSpec.Monitoring, "monitoring", false, "Enable monitoring of the access")
	addInteractiveEditorFlag(xdslEditCmd)
	xdslCmd.AddCommand(xdslEditCmd)

	rootCmd.AddCommand(xdslCmd)
}
