// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudStorageS3Command(cloudCmd *cobra.Command) {
	// Storage commands
	storageS3Cmd := &cobra.Command{
		Use:   "storage-s3",
		Short: "Manage S3™* compatible storage containers in the given cloud project (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)",
	}
	storageS3Cmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	storageS3ListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List S3™* compatible storage containers (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)",
		Run:     cloud.ListCloudStorageS3,
	}
	storageS3Cmd.AddCommand(withFilterFlag(storageS3ListCmd))

	// Container commands
	storageS3Cmd.AddCommand(&cobra.Command{
		Use:   "get <container_name>",
		Short: "Get a specific S3™* compatible storage container (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)",
		Run:   cloud.GetStorageS3,
		Args:  cobra.ExactArgs(1),
	})

	editStorageS3Cmd := &cobra.Command{
		Use:   "edit <container_name>",
		Short: "Edit the given S3™* compatible storage container (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)",
		Run:   cloud.EditStorageS3,
		Args:  cobra.ExactArgs(1),
	}
	editStorageS3Cmd.Flags().StringVar(&cloud.StorageS3Spec.Encryption.SSEAlgorithm, "encryption-sse-algorithm", "", "Encryption SSE Algorithm (AES256, plaintext)")
	editStorageS3Cmd.Flags().StringVar(&cloud.StorageS3Spec.ObjectLock.Rule.Mode, "object-lock-rule-mode", "", "Object lock mode (compliance, governance)")
	editStorageS3Cmd.Flags().StringVar(&cloud.StorageS3Spec.ObjectLock.Rule.Period, "object-lock-rule-period", "", "Object lock period (e.g., P3Y6M4DT12H30M5S)")
	editStorageS3Cmd.Flags().StringVar(&cloud.StorageS3Spec.ObjectLock.Status, "object-lock-status", "", "Object lock status (disabled, enabled)")
	editStorageS3Cmd.Flags().StringToStringVar(&cloud.StorageS3Spec.Tags, "tag", nil, "Container tags as key=value pairs")
	editStorageS3Cmd.Flags().StringVar(&cloud.StorageS3Spec.Versioning.Status, "versioning-status", "", "Versioning status (disabled, enabled, suspended)")
	addInteractiveEditorFlag(editStorageS3Cmd)
	storageS3Cmd.AddCommand(editStorageS3Cmd)

	storageS3Cmd.AddCommand(getCloudStorageS3CreateCmd())

	storageS3Cmd.AddCommand(&cobra.Command{
		Use:   "delete <container_name>",
		Short: "Delete the given S3™* compatible storage container (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)",
		Run:   cloud.DeleteStorageS3,
		Args:  cobra.ExactArgs(1),
	})

	// Bulk-delete command
	bulkDeleteCmd := &cobra.Command{
		Use:   "bulk-delete <container_name>",
		Short: "Bulk delete objects in the given storage container",
		Run:   cloud.StorageS3BulkDeleteObjects,
		Args:  cobra.ExactArgs(1),
	}
	bulkDeleteCmd.Flags().StringSliceVar(&cloud.StorageS3ObjectsToDelete, "objects", nil, "List of objects to delete (format is '<object_name>' or '<object_name>:<version_id>'")
	bulkDeleteCmd.Flags().BoolVar(&cloud.StorageS3BulkDeleteAll, "all", false, "Delete all objects in the container")
	bulkDeleteCmd.Flags().StringVar(&cloud.StorageS3BulkDeletePrefix, "prefix", "", "Prefix to filter objects to delete")
	bulkDeleteCmd.MarkFlagsOneRequired("objects", "all", "prefix")
	bulkDeleteCmd.MarkFlagsMutuallyExclusive("objects", "all", "prefix")

	storageS3Cmd.AddCommand(bulkDeleteCmd)

	// Object commands
	objectCmd := &cobra.Command{
		Use:   "object",
		Short: "Manage objects in the given storage container",
	}
	storageS3Cmd.AddCommand(objectCmd)

	objectListCmd := &cobra.Command{
		Use:     "list <container_name>",
		Aliases: []string{"ls"},
		Short:   "List objects in the given storage container",
		Run:     cloud.ListStorageS3Objects,
		Args:    cobra.ExactArgs(1),
	}
	objectListCmd.Flags().StringVar(&cloud.StorageS3ListParams.KeyMarker, "key-marker", "", "Key marker for pagination")
	objectListCmd.Flags().IntVar(&cloud.StorageS3ListParams.Limit, "limit", 1000, "Maximum number of objects to return")
	objectListCmd.Flags().StringVar(&cloud.StorageS3ListParams.Prefix, "prefix", "", "Prefix to filter objects by name")
	objectListCmd.Flags().StringVar(&cloud.StorageS3ListParams.VersionIdMarker, "version-id-marker", "", "Version ID marker for pagination")
	objectListCmd.Flags().BoolVar(&cloud.StorageS3ListParams.WithVersions, "with-versions", false, "Include object versions in the listing")
	objectCmd.AddCommand(objectListCmd)

	objectCmd.AddCommand(&cobra.Command{
		Use:   "get <container_name> <object_name>",
		Short: "Get a specific object from the given storage container",
		Run:   cloud.GetStorageS3Object,
		Args:  cobra.ExactArgs(2),
	})

	objectEditCmd := &cobra.Command{
		Use:   "edit <container_name> <object_name>",
		Short: "Edit the given object in the storage container",
		Run:   cloud.EditStorageS3Object,
		Args:  cobra.ExactArgs(2),
	}
	objectEditCmd.Flags().StringVar(&cloud.StorageS3ObjectSpec.LegalHold, "legal-hold", "", "Legal hold status (on, off)")
	objectEditCmd.Flags().StringVar(&cloud.StorageS3ObjectSpec.Lock.Mode, "lock-mode", "", "Lock mode (compliance, governance)")
	objectEditCmd.Flags().StringVar(&cloud.StorageS3ObjectSpec.Lock.RetainUntil, "lock-retain-until", "", "Lock retain until date (e.g., 2024-12-31T23:59:59Z)")
	addInteractiveEditorFlag(objectEditCmd)
	objectCmd.AddCommand(objectEditCmd)

	objectCmd.AddCommand(&cobra.Command{
		Use:   "delete <container_name> <object_name>",
		Short: "Delete the given object from the storage container",
		Run:   cloud.DeleteStorageS3Object,
		Args:  cobra.ExactArgs(2),
	})

	// Object version commands
	objectVersionCmd := &cobra.Command{
		Use:   "version",
		Short: "Manage versions of objects in the given storage container",
	}
	objectCmd.AddCommand(objectVersionCmd)

	objectVersionListCmd := &cobra.Command{
		Use:     "list <container_name> <object_name>",
		Aliases: []string{"ls"},
		Short:   "List versions of a specific object in the given storage container",
		Run:     cloud.ListStorageS3ObjectVersions,
		Args:    cobra.ExactArgs(2),
	}
	objectVersionListCmd.Flags().StringVar(&cloud.StorageS3ListParams.VersionIdMarker, "version-id-marker", "", "Version ID marker for pagination")
	objectVersionListCmd.Flags().IntVar(&cloud.StorageS3ListParams.Limit, "limit", 1000, "Maximum number of versions to return")
	objectVersionCmd.AddCommand(objectVersionListCmd)

	objectVersionCmd.AddCommand(&cobra.Command{
		Use:   "get <container_name> <object_name> <version_id>",
		Short: "Get a specific version of an object from the given storage container",
		Run:   cloud.GetStorageS3ObjectVersion,
		Args:  cobra.ExactArgs(3),
	})

	objectVersionEditCmd := &cobra.Command{
		Use:   "edit <container_name> <object_name> <version_id>",
		Short: "Edit the given version of an object in the storage container",
		Run:   cloud.EditStorageS3ObjectVersion,
		Args:  cobra.ExactArgs(3),
	}
	objectVersionEditCmd.Flags().StringVar(&cloud.StorageS3ObjectSpec.LegalHold, "legal-hold", "", "Legal hold status (on, off)")
	objectVersionEditCmd.Flags().StringVar(&cloud.StorageS3ObjectSpec.Lock.Mode, "lock-mode", "", "Lock mode (compliance, governance)")
	objectVersionEditCmd.Flags().StringVar(&cloud.StorageS3ObjectSpec.Lock.RetainUntil, "lock-retain-until", "", "Lock retain until date (e.g., 2024-12-31T23:59:59Z)")
	addInteractiveEditorFlag(objectVersionEditCmd)
	objectVersionCmd.AddCommand(objectVersionEditCmd)

	objectVersionCmd.AddCommand(&cobra.Command{
		Use:   "delete <container_name> <object_name> <version_id>",
		Short: "Delete a specific version of an object from the storage container",
		Run:   cloud.DeleteStorageS3ObjectVersion,
		Args:  cobra.ExactArgs(3),
	})

	// Presigned URL command
	presignedURLCmd := &cobra.Command{
		Use:   "generate-presigned-url <container_name>",
		Short: "Generate a presigned URL to upload or download an object in the given storage container",
		Run:   cloud.StorageS3GeneratePresignedURL,
		Args:  cobra.ExactArgs(1),
	}
	presignedURLCmd.Flags().StringVar(&cloud.StorageS3PresignedURLParams.Method, "method", "GET", "HTTP method for the presigned URL (GET, PUT, DELETE)")
	presignedURLCmd.Flags().StringVar(&cloud.StorageS3PresignedURLParams.Object, "object", "", "Name of the object to upload or download")
	presignedURLCmd.Flags().IntVar(&cloud.StorageS3PresignedURLParams.Expire, "expire", 60, "Expiration time in seconds for the presigned URL")
	presignedURLCmd.Flags().StringVar(&cloud.StorageS3PresignedURLParams.VersionId, "version-id", "", "Version ID of the object (if applicable)")
	presignedURLCmd.Flags().StringVar(&cloud.StorageS3PresignedURLParams.StorageClass, "storage-class", "", "Storage class for the object (HIGH_PERF, STANDARD, STANDARD_IA)")
	addInitParameterFileFlag(presignedURLCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/storage/{name}/presign", "post", cloud.CloudStorageS3PresignedURLExample, nil)
	addInteractiveEditorFlag(presignedURLCmd)
	addFromFileFlag(presignedURLCmd)
	presignedURLCmd.MarkFlagsMutuallyExclusive("from-file", "editor")
	storageS3Cmd.AddCommand(presignedURLCmd)

	// Add user to bucket command
	storageS3Cmd.AddCommand(&cobra.Command{
		Use:   "add-user <container_name> <user_id> <role (admin, deny, readOnly, readWrite)>",
		Short: "Add a user to the given storage container with the specified role (admin, deny, readOnly, readWrite)",
		Run:   cloud.StorageS3AddUser,
		Args:  cobra.ExactArgs(3),
	})

	// Credentials command
	credentialsCmd := &cobra.Command{
		Use:   "credentials",
		Short: "Manage storage containers credentials",
	}
	storageS3Cmd.AddCommand(credentialsCmd)

	credentialsCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <user_id>",
		Aliases: []string{"ls"},
		Short:   "List credentials for the given user ID",
		Run:     cloud.ListStorageS3Credentials,
		Args:    cobra.ExactArgs(1),
	}))

	credentialsCmd.AddCommand(&cobra.Command{
		Use:   "create <user_id>",
		Short: "Create credentials for the given user ID",
		Run:   cloud.CreateStorageS3Credentials,
		Args:  cobra.ExactArgs(1),
	})

	credentialsCmd.AddCommand(&cobra.Command{
		Use:   "delete <user_id> <access_id>",
		Short: "Delete credentials for the given user ID and access ID",
		Run:   cloud.DeleteStorageS3Credentials,
		Args:  cobra.ExactArgs(2),
	})

	credentialsCmd.AddCommand(&cobra.Command{
		Use:   "get <user_id> <access_id>",
		Short: "Get credentials for the given user ID and access ID",
		Run:   cloud.GetStorageS3Credentials,
		Args:  cobra.ExactArgs(2),
	})

	cloudCmd.AddCommand(storageS3Cmd)
}

func getCloudStorageS3CreateCmd() *cobra.Command {
	s3CreateCmd := &cobra.Command{
		Use:   "create <region>",
		Short: "Create a new S3™* compatible storage container (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)",
		Long: `Use this command to create a S3™* compatible storage container in the given cloud project.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud storage-s3 create BHS --name mynewContainer

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud storage-s3 create --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud storage-s3 create GRA --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud storage-s3 create GRA

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud storage-s3 create GRA --from-file ./params.json --name nameoverriden

3. Using your default text editor:

	ovhcloud cloud storage-s3 create GRA --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud storage-s3 create GRA --editor --name nameoverriden

*S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.
`,
		Run:  cloud.CreateStorageS3,
		Args: cobra.MaximumNArgs(1),
	}

	s3CreateCmd.Flags().StringVar(&cloud.StorageS3Spec.Name, "name", "", "Name of the storage container")
	s3CreateCmd.Flags().IntVar(&cloud.StorageS3Spec.OwnerId, "owner-id", 0, "Owner ID of the storage container")
	s3CreateCmd.Flags().StringVar(&cloud.StorageS3Spec.Encryption.SSEAlgorithm, "encryption-sse-algorithm", "", "Encryption SSE Algorithm (AES256, plaintext)")
	s3CreateCmd.Flags().StringVar(&cloud.StorageS3Spec.ObjectLock.Rule.Mode, "object-lock-rule-mode", "", "Object lock mode (compliance, governance)")
	s3CreateCmd.Flags().StringVar(&cloud.StorageS3Spec.ObjectLock.Rule.Period, "object-lock-rule-period", "", "Object lock period (e.g., P3Y6M4DT12H30M5S)")
	s3CreateCmd.Flags().StringVar(&cloud.StorageS3Spec.ObjectLock.Status, "object-lock-status", "", "Object lock status (disabled, enabled)")
	s3CreateCmd.Flags().StringToStringVar(&cloud.StorageS3Spec.Tags, "tag", nil, "Container tags as key=value pairs")
	s3CreateCmd.Flags().StringVar(&cloud.StorageS3Spec.Versioning.Status, "versioning-status", "", "Versioning status (disabled, enabled, suspended)")

	// Common flags for other means to define parameters
	addInitParameterFileFlag(s3CreateCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/storage", "post", cloud.CloudStorageS3CreationExample, nil)
	addInteractiveEditorFlag(s3CreateCmd)
	addFromFileFlag(s3CreateCmd)
	s3CreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return s3CreateCmd
}
