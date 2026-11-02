// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/completion"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

const swiftContainerCompletionPath = "/v2/publicCloud/project/%s/storage/swift/container"

func initCloudStorageSwiftCommand(cloudCmd *cobra.Command) {
	storageSwiftCmd := &cobra.Command{
		Use:   "swift",
		Short: "Manage Swift storage in the given cloud project",
	}
	storageSwiftCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	containerCmd := &cobra.Command{
		Use:   "container",
		Short: "Manage Swift storage containers in the given cloud project",
	}
	storageSwiftCmd.AddCommand(containerCmd)

	containerCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List Swift containers",
		Run:     cloud.ListCloudStorageSwift,
	}))

	containerCmd.AddCommand(&cobra.Command{
		Use:               "get <container_id>",
		Short:             "Get a specific Swift container",
		ValidArgsFunction: completion.CloudResources(swiftContainerCompletionPath),
		Run:               cloud.GetStorageSwift,
		Args:              cobra.ExactArgs(1),
	})

	containerCmd.AddCommand(getCloudStorageSwiftCreateCmd())

	editCmd := &cobra.Command{
		Use:               "edit <container_id>",
		Short:             "Edit the given Swift container",
		ValidArgsFunction: completion.CloudResources(swiftContainerCompletionPath),
		Run:               cloud.EditStorageSwift,
		Args:              cobra.ExactArgs(1),
	}
	editCmd.Flags().StringVar(&cloud.SwiftContainerEditSpec.TargetSpec.Type, "type", "", "Container type (PRIVATE, PUBLIC, STATIC)")
	editCmd.Flags().StringSliceVar(&cloud.SwiftContainerEditSpec.TargetSpec.Cors, "cors", nil, "CORS origins to add to the container (use --editor to replace or remove origins)")
	editCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for the container to be READY before exiting")
	addInteractiveEditorFlag(editCmd)
	containerCmd.AddCommand(editCmd)

	containerCmd.AddCommand(&cobra.Command{
		Use:               "delete <container_id>",
		Short:             "Delete the given Swift container and all its objects",
		ValidArgsFunction: completion.CloudResources(swiftContainerCompletionPath),
		Run:               cloud.DeleteStorageSwift,
		Args:              cobra.ExactArgs(1),
	})

	presignCmd := &cobra.Command{
		Use:               "presign <container_id> <object_key>",
		Short:             "Generate a presigned URL to download an object of the given Swift container",
		ValidArgsFunction: completion.CloudResources(swiftContainerCompletionPath),
		Run:               cloud.PresignStorageSwift,
		Args:              cobra.ExactArgs(2),
	}
	presignCmd.Flags().IntVar(&cloud.SwiftContainerPresignSpec.ExpiresIn, "expires-in", 0, "Validity of the URL in seconds (default 3600, max 604800)")
	containerCmd.AddCommand(presignCmd)

	objectCmd := &cobra.Command{
		Use:   "object",
		Short: "Manage objects of the given Swift container",
	}
	containerCmd.AddCommand(objectCmd)

	objectListCmd := &cobra.Command{
		Use:               "list <container_id>",
		Aliases:           []string{"ls"},
		Short:             "List objects of the given Swift container",
		ValidArgsFunction: completion.CloudResources(swiftContainerCompletionPath),
		Run:               cloud.ListStorageSwiftObjects,
		Args:              cobra.ExactArgs(1),
	}
	objectListCmd.Flags().StringVar(&cloud.SwiftObjectListPrefix, "prefix", "", "Only list objects whose key begins with this prefix")
	objectCmd.AddCommand(withFilterFlag(objectListCmd))

	objectCmd.AddCommand(&cobra.Command{
		Use:               "get <container_id> <object_key>",
		Short:             "Get a specific object of the given Swift container",
		ValidArgsFunction: completion.CloudResources(swiftContainerCompletionPath),
		Run:               cloud.GetStorageSwiftObject,
		Args:              cobra.ExactArgs(2),
	})

	objectCmd.AddCommand(&cobra.Command{
		Use:               "delete <container_id> <object_key>",
		Short:             "Delete the given object from the given Swift container",
		ValidArgsFunction: completion.CloudResources(swiftContainerCompletionPath),
		Run:               cloud.DeleteStorageSwiftObject,
		Args:              cobra.ExactArgs(2),
	})

	cloudCmd.AddCommand(storageSwiftCmd)
}

func getCloudStorageSwiftCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create <region>",
		Short: "Create a new Swift container",
		Long: `Use this command to create a Swift container in the given cloud project.
The region is always taken from the <region> argument.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud storage swift container create GRA --name my-container --type PUBLIC

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud storage swift container create --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud storage swift container create GRA --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud storage swift container create GRA

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud storage swift container create GRA --from-file ./params.json --name name-overridden

3. Using your default text editor:

	ovhcloud cloud storage swift container create GRA --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud storage swift container create GRA --editor --name name-overridden
`,
		Run:  cloud.CreateStorageSwift,
		Args: cobra.MaximumNArgs(1),
	}

	createCmd.Flags().StringVar(&cloud.SwiftContainerSpec.TargetSpec.Name, "name", "", "Container name")
	createCmd.Flags().BoolVar(&cloud.SwiftContainerSpec.TargetSpec.Archive, "archive", false, "Create an archive container (Public Cloud Archive storage policy)")
	createCmd.Flags().StringVar(&cloud.SwiftContainerSpec.TargetSpec.Type, "type", "", "Container type (PRIVATE, PUBLIC, STATIC)")
	createCmd.Flags().StringSliceVar(&cloud.SwiftContainerSpec.TargetSpec.Cors, "cors", nil, "Origins allowed to access the container from a browser")
	createCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for the container to be READY before exiting")

	addParameterFileFlags(createCmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/storage/swift/container", "post", cloud.SwiftContainerCreateExample, nil)
	addInteractiveEditorFlag(createCmd)
	markFlagsMutuallyExclusive(createCmd, "from-file", "editor")

	return createCmd
}
