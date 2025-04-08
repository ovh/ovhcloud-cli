
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	emailproColumnsToDisplay = []string{ "domain","displayName","state","offer" }
)

func listEmailPro(_ *cobra.Command, _ []string) {
	manageListRequest("/email/pro", emailproColumnsToDisplay)
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
	emailproCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your EmailPro services",
		Run:   listEmailPro,
	})

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
