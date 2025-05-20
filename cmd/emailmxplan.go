package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	emailmxplanColumnsToDisplay = []string{"domain", "displayName", "state", "offer"}

	//go:embed templates/emailmxplan.tmpl
	emailmxplanTemplate string
)

func listEmailMXPlan(_ *cobra.Command, _ []string) {
	manageListRequest("/email/mxplan", "", emailmxplanColumnsToDisplay, genericFilters)
}

func getEmailMXPlan(_ *cobra.Command, args []string) {
	manageObjectRequest("/email/mxplan", args[0], emailmxplanTemplate)
}

func init() {
	emailmxplanCmd := &cobra.Command{
		Use:   "email-mxplan",
		Short: "Retrieve information and manage your Email MXPlan services",
	}

	// Command to list EmailMXPlan services
	emailmxplanListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Email MXPlan services",
		Run:   listEmailMXPlan,
	}
	emailmxplanCmd.AddCommand(withFilterFlag(emailmxplanListCmd))

	// Command to get a single EmailMXPlan
	emailmxplanCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific Email MXPlan",
		Args:  cobra.ExactArgs(1),
		Run:   getEmailMXPlan,
	})

	rootCmd.AddCommand(emailmxplanCmd)
}
