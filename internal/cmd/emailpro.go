package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/emailpro"
	"github.com/spf13/cobra"
)

func init() {
	emailproCmd := &cobra.Command{
		Use:   "email-pro",
		Short: "Retrieve information and manage your EmailPro services",
	}

	// Command to list EmailPro services
	emailproListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your EmailPro services",
		Run:     emailpro.ListEmailPro,
	}
	emailproCmd.AddCommand(withFilterFlag(emailproListCmd))

	// Command to get a single EmailPro
	emailproCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific EmailPro",
		Args:  cobra.ExactArgs(1),
		Run:   emailpro.GetEmailPro,
	})

	// Command to update a single EmailPro
	editEmailProCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given EmailPro",
		Args:  cobra.ExactArgs(1),
		Run:   emailpro.EditEmailPro,
	}
	emailproCmd.Flags().BoolVar(&emailpro.EmailProSpec.ComplexityEnabled, "complexity-enabled", false, "Enable policy for strong and secure passwords")
	emailproCmd.Flags().StringVar(&emailpro.EmailProSpec.DisplayName, "display-name", "", "Service displayName")
	emailproCmd.Flags().IntVar(&emailpro.EmailProSpec.LockoutDuration, "lockout-duration", 0, "Number of minutes account will remain locked if it occurs")
	emailproCmd.Flags().IntVar(&emailpro.EmailProSpec.LockoutObservationWindow, "lockout-observation-window", 0, "Number of minutes that must elapse after a failed logon to reset lockout trigger")
	emailproCmd.Flags().IntVar(&emailpro.EmailProSpec.LockoutThreshold, "lockout-threshold", 0, "Number of attempts before account to be locked")
	emailproCmd.Flags().IntVar(&emailpro.EmailProSpec.MaxPasswordAge, "max-password-age", 0, "Maximum number of days that account's password is valid before expiration")
	emailproCmd.Flags().IntVar(&emailpro.EmailProSpec.MaxReceiveSize, "max-receive-size", 0, "Maximum message size that you can receive in MB")
	emailproCmd.Flags().IntVar(&emailpro.EmailProSpec.MaxSendSize, "max-send-size", 0, "Maximum message size that you can send in MB")
	emailproCmd.Flags().IntVar(&emailpro.EmailProSpec.MinPasswordAge, "min-password-age", 0, "Minimum number of days before able to change account's password")
	emailproCmd.Flags().IntVar(&emailpro.EmailProSpec.MinPasswordLength, "min-password-length", 0, "Minimum number of characters password must contain")
	emailproCmd.Flags().BoolVar(&emailpro.EmailProSpec.SpamAndVirusConfiguration.CheckDKIM, "spam-check-dkim", false, "Check DKIM of message")
	emailproCmd.Flags().BoolVar(&emailpro.EmailProSpec.SpamAndVirusConfiguration.CheckSPF, "spam-check-spf", false, "Check SPF of message")
	emailproCmd.Flags().BoolVar(&emailpro.EmailProSpec.SpamAndVirusConfiguration.DeleteSpam, "spam-delete-spam", false, "If message is a spam delete it")
	emailproCmd.Flags().BoolVar(&emailpro.EmailProSpec.SpamAndVirusConfiguration.DeleteVirus, "spam-delete-virus", false, "If message is a virus delete it")
	emailproCmd.Flags().BoolVar(&emailpro.EmailProSpec.SpamAndVirusConfiguration.PutInJunk, "spam-put-in-junk", false, "If message is a spam or virus put in junk. Overridden by deleteSpam or deleteVirus")
	emailproCmd.Flags().BoolVar(&emailpro.EmailProSpec.SpamAndVirusConfiguration.TagSpam, "spam-tag-spam", false, "If message is a spam change its subject")
	emailproCmd.Flags().BoolVar(&emailpro.EmailProSpec.SpamAndVirusConfiguration.TagVirus, "spam-tag-virus", false, "If message is a virus change its subject")
	addInteractiveEditorFlag(editEmailProCmd)
	emailproCmd.AddCommand(editEmailProCmd)

	rootCmd.AddCommand(emailproCmd)
}
