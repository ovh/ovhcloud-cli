// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudSSHKeyCommand(cloudCmd *cobra.Command) {
	sshKeyCmd := &cobra.Command{
		Use:   "ssh-key",
		Short: "Manage SSH keys in the given cloud project",
	}
	sshKeyCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	sshKeyListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List SSH keys",
		Run:     cloud.ListCloudSSHKeys,
	}
	sshKeyCmd.AddCommand(withFilterFlag(sshKeyListCmd))

	sshKeyCmd.AddCommand(&cobra.Command{
		Use:   "get <ssh_key_id>",
		Short: "Get information about a SSH key",
		Run:   cloud.GetCloudSSHKey,
		Args:  cobra.ExactArgs(1),
	})

	sshKeyCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new SSH key",
		Run:   cloud.CreateCloudSSHKey,
	}
	sshKeyCreateCmd.Flags().StringVar(&cloud.SSHKeyCreationParameters.Name, "name", "", "Name for the SSH key to create")
	sshKeyCreateCmd.Flags().StringVar(&cloud.SSHKeyCreationParameters.PublicKey, "public-key", "", "Public key for the SSH key to create")
	sshKeyCreateCmd.Flags().StringVar(&cloud.SSHKeyCreationParameters.Region, "region", "", "Region for the SSH key to create (optional)")
	addInitParameterFileFlag(sshKeyCreateCmd, assets.CloudOpenapiSchema, "/v1/cloud/project/{serviceName}/sshkey", "post", cloud.SSHKeyCreationExample, nil)
	addInteractiveEditorFlag(sshKeyCreateCmd)
	addFromFileFlag(sshKeyCreateCmd)
	sshKeyCreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")
	sshKeyCmd.AddCommand(sshKeyCreateCmd)

	cloudCmd.AddCommand(sshKeyCmd)
}
