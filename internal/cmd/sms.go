package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/sms"
	"github.com/spf13/cobra"
)

func init() {
	smsCmd := &cobra.Command{
		Use:   "sms",
		Short: "Retrieve information and manage your SMS services",
	}

	// Command to list Sms services
	smsListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your SMS services",
		Run:     sms.ListSms,
	}
	smsCmd.AddCommand(withFilterFlag(smsListCmd))

	// Command to get a single Sms
	smsCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific SMS account",
		Args:  cobra.ExactArgs(1),
		Run:   sms.GetSms,
	})

	// Command to update a single Sms
	smsEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given SMS account",
		Args:  cobra.ExactArgs(1),
		Run:   sms.EditSms,
	}
	smsEditCmd.Flags().IntVar(&sms.SmsSpec.AutomaticRecreditAmount, "automatic-recredit-amount", 0, "Amount for automatic recredit (100, 200, 250, 500, 1000, 5000, 10000)")
	smsEditCmd.Flags().StringVar(&sms.SmsSpec.CallBack, "callback", "", "URL called when state of a sent SMS changes")
	smsEditCmd.Flags().StringVar(&sms.SmsSpec.Description, "description", "", "Description of the SMS account")
	smsEditCmd.Flags().IntVar(&sms.SmsSpec.CreditThresholdForAutomaticRecredit, "credit-threshold-for-automatic-recredit", 0, "Credit threshold after which an automatic recredit is launched")
	smsEditCmd.Flags().StringVar(&sms.SmsSpec.SmsResponse.CgiUrl, "sms-response-cgi-url", "", "Default url callback used for a given response")
	smsEditCmd.Flags().StringVar(&sms.SmsSpec.SmsResponse.ResponseType, "sms-response-type", "", "Response type (cgi, none, text)")
	smsEditCmd.Flags().StringVar(&sms.SmsSpec.SmsResponse.Text, "sms-response-text", "", "Automatic notification sent by text in case of customer reply")
	smsEditCmd.Flags().StringVar(&sms.SmsSpec.SmsResponse.TrackingDefaultSmsSender, "sms-response-tracking-default-sms-sender", "", "Tracking default SMS sender for SMS response")
	smsEditCmd.Flags().StringVar(&sms.SmsSpec.StopCallBack, "stop-callback", "", "URL called when a STOP is received after a receiver replied stop to a SMS")
	smsEditCmd.Flags().BoolVar(&sms.SmsSpec.Templates.CustomizedEmailMode, "templates-customized-email-mode", false, "Enable customized email mode")
	smsEditCmd.Flags().BoolVar(&sms.SmsSpec.Templates.CustomizedSmsMode, "templates-customized-sms-mode", false, "Enable customized SMS mode")
	smsEditCmd.Flags().StringVar(&sms.SmsSpec.Templates.EmailBody, "templates-email-body", "", "Email body for templates")
	smsEditCmd.Flags().StringVar(&sms.SmsSpec.Templates.EmailFrom, "templates-email-from", "", "Email from for templates")
	smsEditCmd.Flags().StringVar(&sms.SmsSpec.Templates.EmailSubject, "templates-email-subject", "", "Email subject for templates")
	smsEditCmd.Flags().StringVar(&sms.SmsSpec.Templates.SmsBody, "templates-sms-body", "", "SMS body for templates")
	smsEditCmd.Flags().StringVar(&sms.SmsSpec.Templates.Time2chatAutomaticResponse, "templates-time2chat-automatic-response", "", "Time2chat automatic response")
	addInteractiveEditorFlag(smsEditCmd)
	smsCmd.AddCommand(smsEditCmd)

	rootCmd.AddCommand(smsCmd)
}
