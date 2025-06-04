package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/emailmxplan"
)

func init() {
	emailmxplanCmd := &cobra.Command{
		Use:   "email-mxplan",
		Short: "Retrieve information and manage your Email MXPlan services",
	}

	// Command to list EmailMXPlan services
	emailmxplanListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Email MXPlan services",
		Run:   emailmxplan.ListEmailMXPlan,
	}
	emailmxplanCmd.AddCommand(withFilterFlag(emailmxplanListCmd))

	// Command to get a single EmailMXPlan
	emailmxplanCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific Email MXPlan",
		Args:  cobra.ExactArgs(1),
		Run:   emailmxplan.GetEmailMXPlan,
	})

	// Command to update a single EmailMXPlan
	emailmxplanCmd.AddCommand(&cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given Email MXPlan",
		Run:   emailmxplan.EditEmailMXPlan,
	})

	rootCmd.AddCommand(emailmxplanCmd)
}
