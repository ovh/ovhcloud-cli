package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
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
	emailmxplanEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given Email MXPlan",
		Args:  cobra.ExactArgs(1),
		Run:   emailmxplan.EditEmailMXPlan,
	}
	emailmxplanEditCmd.Flags().BoolVar(&emailmxplan.EmailMXPlanSpec.ComplexityEnabled, "complexity-enabled", false, "Enable policy for strong and secure passwords")
	emailmxplanEditCmd.Flags().StringVar(&emailmxplan.EmailMXPlanSpec.DisplayName, "display-name", "", "Service displayName")
	emailmxplanEditCmd.Flags().IntVar(&emailmxplan.EmailMXPlanSpec.LockoutDuration, "lockout-duration", 0, "Number of minutes account will remain locked if it occurs")
	emailmxplanEditCmd.Flags().IntVar(&emailmxplan.EmailMXPlanSpec.LockoutObservationWindow, "lockout-observation-window", 0, "Number of minutes that must elapse after a failed logon to reset lockout trigger")
	emailmxplanEditCmd.Flags().IntVar(&emailmxplan.EmailMXPlanSpec.LockoutThreshold, "lockout-threshold", 0, "Number of attempts before account to be locked")
	emailmxplanEditCmd.Flags().IntVar(&emailmxplan.EmailMXPlanSpec.MaxPasswordAge, "max-password-age", 0, "Maximum number of days that account's password is valid before expiration")
	emailmxplanEditCmd.Flags().IntVar(&emailmxplan.EmailMXPlanSpec.MaxReceiveSize, "max-receive-size", 0, "Maximum message size that you can receive in MB")
	emailmxplanEditCmd.Flags().IntVar(&emailmxplan.EmailMXPlanSpec.MaxSendSize, "max-send-size", 0, "Maximum message size that you can send in MB")
	emailmxplanEditCmd.Flags().IntVar(&emailmxplan.EmailMXPlanSpec.MinPasswordAge, "min-password-age", 0, "Minimum number of days before able to change account's password")
	emailmxplanEditCmd.Flags().IntVar(&emailmxplan.EmailMXPlanSpec.MinPasswordLength, "min-password-length", 0, "Minimum number of characters password must contain")
	emailmxplanEditCmd.Flags().BoolVar(&emailmxplan.EmailMXPlanSpec.SpamAndVirusConfiguration.CheckDKIM, "spam-check-dkim", false, "Check DKIM of message")
	emailmxplanEditCmd.Flags().BoolVar(&emailmxplan.EmailMXPlanSpec.SpamAndVirusConfiguration.CheckSPF, "spam-check-spf", false, "Check SPF of message")
	emailmxplanEditCmd.Flags().BoolVar(&emailmxplan.EmailMXPlanSpec.SpamAndVirusConfiguration.DeleteSpam, "spam-delete-spam", false, "If message is a spam delete it")
	emailmxplanEditCmd.Flags().BoolVar(&emailmxplan.EmailMXPlanSpec.SpamAndVirusConfiguration.DeleteVirus, "spam-delete-virus", false, "If message is a virus delete it")
	emailmxplanEditCmd.Flags().BoolVar(&emailmxplan.EmailMXPlanSpec.SpamAndVirusConfiguration.PutInJunk, "spam-put-in-junk", false, "If message is a spam or virus put in junk. Overridden by deleteSpam or deleteVirus")
	emailmxplanEditCmd.Flags().BoolVar(&emailmxplan.EmailMXPlanSpec.SpamAndVirusConfiguration.TagSpam, "spam-tag-spam", false, "If message is a spam change its subject")
	emailmxplanEditCmd.Flags().BoolVar(&emailmxplan.EmailMXPlanSpec.SpamAndVirusConfiguration.TagVirus, "spam-tag-virus", false, "If message is a virus change its subject")
	emailmxplanEditCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define parameters")
	emailmxplanCmd.AddCommand(emailmxplanEditCmd)

	rootCmd.AddCommand(emailmxplanCmd)
}
