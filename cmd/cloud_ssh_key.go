package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
)

func initCloudSSHKeyCommand(cloudCmd *cobra.Command) {
	sshKeyCmd := &cobra.Command{
		Use:   "ssh-key",
		Short: "Manage SSH keys in the given cloud project",
	}
	sshKeyCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	sshKeyListCmd := &cobra.Command{
		Use:   "list",
		Short: "List SSH keys",
		Run:   cloud.ListCloudSSHKeys,
	}
	sshKeyCmd.AddCommand(withFilterFlag(sshKeyListCmd))

	sshKeyCmd.AddCommand(&cobra.Command{
		Use:   "get <ssh_key_id>",
		Short: "Get information about a SSH key",
		Run:   cloud.GetCloudSSHKey,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(sshKeyCmd)
}
