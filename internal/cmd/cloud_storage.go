// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

func initCloudStorageCommand(cloudCmd *cobra.Command) {
	storageCmd := &cobra.Command{
		Use:   "storage",
		Short: "Manage storage services in the given cloud project",
	}

	initCloudVolumeCommand(storageCmd)
	initCloudStorageFileCommand(storageCmd)
	initCloudStorageS3Command(storageCmd)
	initCloudStorageSwiftCommand(storageCmd)

	cloudCmd.AddCommand(storageCmd)
}
