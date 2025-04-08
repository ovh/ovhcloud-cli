
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	emailmxplanColumnsToDisplay = []string{ "domain","displayName","state","offer" }
)

func listEmailMXPlan(_ *cobra.Command, _ []string) {
	manageListRequest("/email/mxplan", emailmxplanColumnsToDisplay)
}

func getEmailMXPlan(_ *cobra.Command, args []string) {
	manageObjectRequest("/email/mxplan", args[0], emailmxplanColumnsToDisplay[0])
}

func init() {
	emailmxplanCmd := &cobra.Command{
		Use:   "emailmxplan",
		Short: "Retrieve information and manage your EmailMXPlan services",
	}

	// Command to list EmailMXPlan services
	emailmxplanCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your EmailMXPlan services",
		Run:   listEmailMXPlan,
	})

	// Command to get a single EmailMXPlan
	emailmxplanCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific EmailMXPlan",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getEmailMXPlan,
	})

	rootCmd.AddCommand(emailmxplanCmd)
}
