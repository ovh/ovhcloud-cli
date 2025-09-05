// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudUserCommand(cloudCmd *cobra.Command) {
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users in the given cloud project",
	}
	userCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	userListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List users",
		Run:     cloud.ListCloudUsers,
	}
	userCmd.AddCommand(withFilterFlag(userListCmd))

	userCmd.AddCommand(&cobra.Command{
		Use:   "get <user_id>",
		Short: "Get information about a user",
		Run:   cloud.GetCloudUser,
		Args:  cobra.ExactArgs(1),
	})

	userCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		Run:   cloud.CreateCloudUser,
	}
	userCreateCmd.Flags().StringVar(&cloud.UserSpec.Description, "description", "", "Description of the user")
	userCreateCmd.Flags().StringArrayVar(&cloud.UserSpec.Roles, "roles", nil, "Roles assigned to the user")
	addInitParameterFileFlag(userCreateCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/user", "post", cloud.UserCreateExample, nil)
	addInteractiveEditorFlag(userCreateCmd)
	addFromFileFlag(userCreateCmd)
	userCreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")
	userCmd.AddCommand(userCreateCmd)

	userCmd.AddCommand(&cobra.Command{
		Use:   "delete <user_id>",
		Short: "Delete the given user",
		Run:   cloud.DeleteCloudUser,
		Args:  cobra.ExactArgs(1),
	})

	// S3 policy commands
	s3PolicyCmd := &cobra.Command{
		Use:   "s3-policy",
		Short: "Manage policies for users on S3™* compatible storage containers (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)",
	}
	userCmd.AddCommand(s3PolicyCmd)

	s3PolicyCmd.AddCommand(&cobra.Command{
		Use:   "get <user_id>",
		Short: "Get the policy for the given user ID",
		Run:   cloud.GetUserS3Policy,
		Args:  cobra.ExactArgs(1),
	})

	s3PolicyCreateCmd := &cobra.Command{
		Use:   "create <user_id>",
		Short: "Create a policy for the given user ID",
		Run:   cloud.CreateUserS3Policy,
		Args:  cobra.ExactArgs(1),
	}
	s3PolicyCreateCmd.Flags().StringVar(&cloud.StorageS3ContainerPolicySpec.Policy, "policy", "", "Policy in JSON format")
	addInitParameterFileFlag(s3PolicyCreateCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/user/{userId}/policy", "post", cloud.CloudStorageS3ContainerPolicyExample, nil)
	addInteractiveEditorFlag(s3PolicyCreateCmd)
	addFromFileFlag(s3PolicyCreateCmd)
	s3PolicyCreateCmd.MarkFlagsMutuallyExclusive("policy", "from-file", "editor")
	s3PolicyCmd.AddCommand(s3PolicyCreateCmd)

	cloudCmd.AddCommand(userCmd)
}
