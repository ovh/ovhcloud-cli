// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudAlertingCommand(cloudCmd *cobra.Command) {
	alertingCmd := &cobra.Command{
		Use:   "alerting",
		Short: "Manage billing alert configurations in the given cloud project",
	}
	alertingCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	// List alerting configurations
	alertingListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List billing alert configurations",
		Run:     cloud.ListCloudAlertingConfigs,
	}
	alertingCmd.AddCommand(withFilterFlag(alertingListCmd))

	// Get specific alerting configuration
	alertingCmd.AddCommand(&cobra.Command{
		Use:   "get <alert_id>",
		Short: "Get a specific billing alert configuration",
		Run:   cloud.GetCloudAlertingConfig,
		Args:  cobra.ExactArgs(1),
	})

	// Create alerting configuration
	alertingCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new billing alert configuration",
		Run:   cloud.CreateCloudAlertingConfig,
	}
	alertingCreateCmd.Flags().Int64Var(&cloud.AlertingConfigSpec.Delay, "delay", 3600, "Delay between alerts in seconds (minimum 3600)")
	alertingCreateCmd.Flags().StringSliceVar(&cloud.AlertingConfigSpec.Emails, "emails", nil, "Email addresses to receive alerts (comma-separated)")
	alertingCreateCmd.Flags().Int64Var(&cloud.AlertingConfigSpec.MonthlyThreshold, "monthly-threshold", 0, "Monthly threshold value")
	alertingCreateCmd.Flags().StringVar(&cloud.AlertingConfigSpec.Name, "name", "", "Alert name")
	alertingCreateCmd.Flags().StringVar(&cloud.AlertingConfigSpec.Service, "service", "", "Service of the alert. Allowed: ai_endpoint, all, block_storage, data_platform, instances, instances_gpu, instances_without_gpu, objet_storage, rancher, snapshot")
	addInitParameterFileFlag(alertingCreateCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/alerting", "post", cloud.AlertingConfigCreateExample, nil)
	addInteractiveEditorFlag(alertingCreateCmd)
	addFromFileFlag(alertingCreateCmd)
	alertingCreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")
	alertingCmd.AddCommand(alertingCreateCmd)

	// Edit alerting configuration
	alertingEditCmd := &cobra.Command{
		Use:   "edit <alert_id>",
		Short: "Edit a billing alert configuration",
		Run:   cloud.EditCloudAlertingConfig,
		Args:  cobra.ExactArgs(1),
	}
	alertingEditCmd.Flags().Int64Var(&cloud.AlertingConfigEditSpec.Delay, "delay", 0, "Delay between alerts in seconds (minimum 3600)")
	alertingEditCmd.Flags().StringSliceVar(&cloud.AlertingConfigEditSpec.Emails, "emails", nil, "Email addresses to receive alerts (comma-separated)")
	alertingEditCmd.Flags().Int64Var(&cloud.AlertingConfigEditSpec.MonthlyThreshold, "monthly-threshold", 0, "Monthly threshold value")
	alertingEditCmd.Flags().StringVar(&cloud.AlertingConfigEditSpec.Name, "name", "", "Alert name")
	alertingEditCmd.Flags().StringVar(&cloud.AlertingConfigEditSpec.Service, "service", "", "Service of the alert. Allowed: ai_endpoint, all, block_storage, data_platform, instances, instances_gpu, instances_without_gpu, objet_storage, rancher, snapshot")
	alertingEditCmd.Flags().StringVar(&cloud.AlertingConfigEditSpec.Status, "status", "", "Status of the alert. Allowed: deleted, disabled, ok")
	addInteractiveEditorFlag(alertingEditCmd)
	alertingCmd.AddCommand(alertingEditCmd)

	// Delete alerting configuration
	alertingCmd.AddCommand(&cobra.Command{
		Use:   "delete <alert_id>",
		Short: "Delete a billing alert configuration",
		Run:   cloud.DeleteCloudAlertingConfig,
		Args:  cobra.ExactArgs(1),
	})

	// Subcommand for triggered alerts
	triggeredAlertCmd := &cobra.Command{
		Use:   "alert",
		Short: "Manage triggered alerts for a billing alert configuration",
	}
	alertingCmd.AddCommand(triggeredAlertCmd)

	// List triggered alerts
	triggeredAlertListCmd := &cobra.Command{
		Use:     "list <alert_id>",
		Aliases: []string{"ls"},
		Short:   "List triggered alerts for a specific alert configuration",
		Run:     cloud.ListCloudAlertingTriggeredAlerts,
		Args:    cobra.ExactArgs(1),
	}
	triggeredAlertCmd.AddCommand(withFilterFlag(triggeredAlertListCmd))

	// Get specific triggered alert
	triggeredAlertCmd.AddCommand(&cobra.Command{
		Use:   "get <alert_id> <triggered_alert_id>",
		Short: "Get a specific triggered alert",
		Run:   cloud.GetCloudAlertingTriggeredAlert,
		Args:  cobra.ExactArgs(2),
	})

	cloudCmd.AddCommand(alertingCmd)
}
