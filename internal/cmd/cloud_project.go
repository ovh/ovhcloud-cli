// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func init() {
	cloudCmd := &cobra.Command{
		Use:   "cloud",
		Short: "Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)",
	}

	cloudprojectCmd := &cobra.Command{
		Use:   "project",
		Short: "Retrieve information and manage your CloudProject services",
	}
	cloudprojectCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	// Command to list CloudProject services
	cloudprojectCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your cloud projects",
		Run:     cloud.ListCloudProject,
	}))

	// Command to get a single CloudProject
	cloudprojectCmd.AddCommand(&cobra.Command{
		Use:   "get <project_id>",
		Short: "Retrieve information of a specific cloud project",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.GetCloudProject,
	})

	editCloudProjectCmd := &cobra.Command{
		Use:   "edit <project_id>",
		Short: "Edit the given cloud project",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.EditCloudProject,
	}
	editCloudProjectCmd.Flags().StringVar(&cloud.CloudProjectSpec.Description, "description", "", "Description of the project")
	editCloudProjectCmd.Flags().BoolVar(&cloud.CloudProjectSpec.ManualQuota, "manual-quota", false, "Prevent automatic quota upgrade")
	addInteractiveEditorFlag(editCloudProjectCmd)
	cloudprojectCmd.AddCommand(editCloudProjectCmd)

	// Project management commands
	cloudprojectCmd.AddCommand(&cobra.Command{
		Use:   "service-info",
		Short: "Get service information for the project",
		Run:   cloud.GetServiceInfo,
	})

	changeContactCmd := &cobra.Command{
		Use:   "change-contact",
		Short: "Change project contacts",
		Run:   cloud.ChangeContact,
	}
	changeContactCmd.Flags().StringVar(&cloud.ChangeContactSpec.ContactAdmin, "contact-admin", "", "Admin contact NIC handle")
	changeContactCmd.Flags().StringVar(&cloud.ChangeContactSpec.ContactBilling, "contact-billing", "", "Billing contact NIC handle")
	changeContactCmd.Flags().StringVar(&cloud.ChangeContactSpec.ContactTech, "contact-tech", "", "Technical contact NIC handle")
	cloudprojectCmd.AddCommand(changeContactCmd)

	// Termination commands
	terminationCmd := &cobra.Command{
		Use:   "termination",
		Short: "Manage project termination lifecycle",
	}

	terminationCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initiate project termination",
		Long:  "Initiate project termination. A termination token will be returned to confirm the operation.",
		Run:   cloud.TerminateProject,
	})

	confirmTerminateCmd := &cobra.Command{
		Use:   "confirm",
		Short: "Confirm project termination with token",
		Run:   cloud.ConfirmTermination,
	}
	confirmTerminateCmd.Flags().String("token", "", "Termination token received from init command")
	confirmTerminateCmd.MarkFlagRequired("token")
	terminationCmd.AddCommand(confirmTerminateCmd)

	terminationCmd.AddCommand(&cobra.Command{
		Use:   "cancel",
		Short: "Cancel a project scheduled for termination",
		Run:   cloud.RetainProject,
	})

	cloudprojectCmd.AddCommand(terminationCmd)

	cloudprojectCmd.AddCommand(&cobra.Command{
		Use:   "unleash",
		Short: "Unleash a project",
		Run:   cloud.UnleashProject,
	})

	initKubeCommand(cloudCmd)
	initContainerRegistryCommand(cloudCmd)
	initCloudDatabaseCommand(cloudCmd)
	initInstanceCommand(cloudCmd)
	initCloudNetworkCommand(cloudCmd)
	initCloudOperationCommand(cloudCmd)
	initCloudQuotaCommand(cloudCmd)
	initCloudRegionCommand(cloudCmd)
	initCloudSSHKeyCommand(cloudCmd)
	initCloudUserCommand(cloudCmd)
	initCloudStorageS3Command(cloudCmd)
	initCloudStorageSwiftCommand(cloudCmd)
	initCloudVolumeCommand(cloudCmd)
	initCloudRancherCommand(cloudCmd)
	initCloudReferenceCmd(cloudCmd)
	initCloudSavingsPlanCommand(cloudCmd)
	initCloudIPFailoverCommand(cloudCmd)
	initCloudAlertingCommand(cloudCmd)

	cloudCmd.AddCommand(cloudprojectCmd)
	rootCmd.AddCommand(cloudCmd)
}
