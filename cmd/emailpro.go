package cmd

import (
	"github.com/spf13/cobra"
)

var (
	emailproColumnsToDisplay = []string{ "domain","displayName","state","offer" }
)

func listEmailPro(_ *cobra.Command, _ []string) {
	manageListRequest("/email/pro", emailproColumnsToDisplay, genericFilters)
}

func getEmailPro(_ *cobra.Command, args []string) {
	manageObjectRequest("/email/pro", args[0], emailproColumnsToDisplay[0])
}

func init() {
	emailproCmd := &cobra.Command{
		Use:   "emailpro",
		Short: "Retrieve information and manage your EmailPro services",
	}

	// Command to list EmailPro services
	emailproListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your EmailPro services",
		Run:   listEmailPro,
	}
	emailproListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	emailproCmd.AddCommand(emailproListCmd)

	// Command to get a single EmailPro
	emailproCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific EmailPro",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getEmailPro,
	})

	rootCmd.AddCommand(emailproCmd)
}
