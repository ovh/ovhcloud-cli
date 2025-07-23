package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/telephony"
	"github.com/spf13/cobra"
)

func init() {
	telephonyCmd := &cobra.Command{
		Use:   "telephony",
		Short: "Retrieve information and manage your Telephony services",
	}

	// Command to list Telephony services
	telephonyListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your Telephony services",
		Run:     telephony.ListTelephony,
	}
	telephonyCmd.AddCommand(withFilterFlag(telephonyListCmd))

	// Command to get a single Telephony
	telephonyCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific Telephony service",
		Args:  cobra.ExactArgs(1),
		Run:   telephony.GetTelephony,
	})

	// Command to update a single Telephony
	telephonyEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given Telephony service",
		Args:  cobra.ExactArgs(1),
		Run:   telephony.EditTelephony,
	}
	telephonyEditCmd.Flags().StringVar(&telephony.TelephonySpec.Description, "description", "", "Description of service")
	telephonyEditCmd.Flags().BoolVar(&telephony.TelephonySpec.HiddenExternalNumber, "hidden-external-number", false, "Hide called numbers in end-of-month call details CSV")
	telephonyEditCmd.Flags().BoolVar(&telephony.TelephonySpec.OverrideDisplayedNumber, "override-displayed-number", false, "Override number displayed for calls between services of your billing account")
	telephonyEditCmd.Flags().StringVar(&telephony.TelephonySpec.CreditThreshold.CurrencyCode, "credit-threshold-currency", "", "Currency code (AUD, CAD, CZK, EUR, GBP, INR, LTL, MAD, N/A, PLN, SGD, TND, USD, XOF, points)")
	telephonyEditCmd.Flags().StringVar(&telephony.TelephonySpec.CreditThreshold.Text, "credit-threshold-text", "", "Text for credit threshold")
	telephonyEditCmd.Flags().IntVar(&telephony.TelephonySpec.CreditThreshold.Value, "credit-threshold-value", 0, "Value for credit threshold")
	addInteractiveEditorFlag(telephonyEditCmd)
	telephonyCmd.AddCommand(telephonyEditCmd)

	rootCmd.AddCommand(telephonyCmd)
}
