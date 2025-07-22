package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/packxdsl"
)

func init() {
	packxdslCmd := &cobra.Command{
		Use:   "pack-xdsl",
		Short: "Retrieve information and manage your PackXDSL services",
	}

	// Command to list PackXDSL services
	packxdslListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your PackXDSL services",
		Run:     packxdsl.ListPackXDSL,
	}
	packxdslCmd.AddCommand(withFilterFlag(packxdslListCmd))

	// Command to get a single PackXDSL
	packxdslCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific PackXDSL",
		Args:  cobra.ExactArgs(1),
		Run:   packxdsl.GetPackXDSL,
	})

	// Command to update a single PackXDSL
	packxdslEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given PackXDSL",
		Args:  cobra.ExactArgs(1),
		Run:   packxdsl.EditPackXDSL,
	}
	packxdslEditCmd.Flags().StringVar(&packxdsl.PackXDSLSpec.Description, "description", "", "Description of the PackXDSL")
	addInteractiveEditorFlag(packxdslEditCmd)
	packxdslCmd.AddCommand(packxdslEditCmd)

	rootCmd.AddCommand(packxdslCmd)
}
