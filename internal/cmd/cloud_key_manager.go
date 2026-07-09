// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudKeyManagerCommand(cloudCmd *cobra.Command) {
	keyManagerCmd := &cobra.Command{
		Use:     "key-manager",
		Aliases: []string{"kms"},
		Short:   "Manage Key Management Service (KMS) resources in the given cloud project",
	}
	keyManagerCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	keyManagerCmd.AddCommand(getKeyManagerSecretCmd())
	keyManagerCmd.AddCommand(getKeyManagerContainerCmd())

	cloudCmd.AddCommand(keyManagerCmd)
}

//
// Secret command tree
//

func getKeyManagerSecretCmd() *cobra.Command {
	secretCmd := &cobra.Command{
		Use:   "secret",
		Short: "Manage Key Manager secrets",
	}

	secretListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List Key Manager secrets",
		Run:     cloud.ListKeyManagerSecrets,
	}
	secretCmd.AddCommand(withFilterFlag(secretListCmd))

	secretCmd.AddCommand(&cobra.Command{
		Use:   "get <secret_id>",
		Short: "Get a specific Key Manager secret",
		Run:   cloud.GetKeyManagerSecret,
		Args:  cobra.ExactArgs(1),
	})

	secretCmd.AddCommand(getKeyManagerSecretCreateCmd())

	secretEditCmd := &cobra.Command{
		Use:   "edit <secret_id>",
		Short: "Edit the given Key Manager secret (only metadata is mutable)",
		Run:   cloud.EditKeyManagerSecret,
		Args:  cobra.ExactArgs(1),
	}
	secretEditCmd.Flags().StringToStringVar(&cloud.KeyManagerSecretEditSpec.TargetSpec.Metadata, "metadata", nil, "Metadata key-value pairs (replaces all existing metadata)")
	addInteractiveEditorFlag(secretEditCmd)
	secretCmd.AddCommand(secretEditCmd)

	secretCmd.AddCommand(&cobra.Command{
		Use:   "delete <secret_id>",
		Short: "Delete the given Key Manager secret",
		Run:   cloud.DeleteKeyManagerSecret,
		Args:  cobra.ExactArgs(1),
	})

	secretCmd.AddCommand(&cobra.Command{
		Use:   "payload <secret_id>",
		Short: "Fetch the payload (sensitive material) of the given Key Manager secret",
		Run:   cloud.GetKeyManagerSecretPayload,
		Args:  cobra.ExactArgs(1),
	})

	// Consumer subcommands
	secretCmd.AddCommand(getKeyManagerSecretConsumerCmd())

	return secretCmd
}

func getKeyManagerSecretCreateCmd() *cobra.Command {
	secretCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Key Manager secret",
		Run:   cloud.CreateKeyManagerSecret,
	}

	spec := &cloud.KeyManagerSecretSpec.TargetSpec
	secretCreateCmd.Flags().StringVar(&spec.Name, "name", "", "Human-readable name of the secret")
	secretCreateCmd.Flags().StringVar(&spec.SecretType, "secret-type", "", "Type of the secret (CERTIFICATE, OPAQUE, PASSPHRASE, PRIVATE, PUBLIC, SYMMETRIC)")
	secretCreateCmd.Flags().StringVar(&spec.Algorithm, "algorithm", "", "Algorithm associated with the secret (AES, DH, DSA, EC, RSA)")
	secretCreateCmd.Flags().IntVar(&spec.BitLength, "bit-length", 0, "Bit length of the secret (128, 256, 512, 1024, 2048, 4096)")
	secretCreateCmd.Flags().StringVar(&spec.Mode, "mode", "", "Mode of the algorithm (CBC, CTR)")
	secretCreateCmd.Flags().StringVar(&spec.Expiration, "expiration", "", "Expiration date and time of the secret (RFC3339)")
	secretCreateCmd.Flags().StringVar(&spec.Payload, "payload", "", "Secret payload data (base64-encoded, write-only). Requires --payload-content-type")
	secretCreateCmd.Flags().StringVar(&spec.PayloadContentType, "payload-content-type", "", "Content type of the payload (APPLICATION_OCTET_STREAM, APPLICATION_PKCS8, APPLICATION_PKIX_CERT, TEXT_PLAIN)")
	secretCreateCmd.Flags().StringToStringVar(&spec.Metadata, "metadata", nil, "Metadata key-value pairs for the secret")
	secretCreateCmd.Flags().StringVar(&cloud.KeyManagerSecretCreateRegion, "region", "", "Region code where the secret is located")
	secretCreateCmd.Flags().StringVar(&cloud.KeyManagerSecretCreateAvailabilityZone, "availability-zone", "", "Availability zone within the region")
	secretCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for the secret to be ready before exiting")

	addParameterFileFlags(secretCreateCmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/keyManager/secret", "post", cloud.KeyManagerSecretCreateExample, nil)
	addInteractiveEditorFlag(secretCreateCmd)
	markFlagsMutuallyExclusive(secretCreateCmd, "from-file", "editor")

	return secretCreateCmd
}

func getKeyManagerSecretConsumerCmd() *cobra.Command {
	consumerCmd := &cobra.Command{
		Use:   "consumer",
		Short: "Manage consumers of a Key Manager secret",
	}

	consumerListCmd := &cobra.Command{
		Use:     "list <secret_id>",
		Aliases: []string{"ls"},
		Short:   "List consumers registered for the given secret",
		Run:     cloud.ListKeyManagerSecretConsumers,
		Args:    cobra.ExactArgs(1),
	}
	consumerCmd.AddCommand(withFilterFlag(consumerListCmd))

	consumerCmd.AddCommand(&cobra.Command{
		Use:   "get <secret_id> <consumer_id>",
		Short: "Get a specific consumer of the given secret",
		Run:   cloud.GetKeyManagerSecretConsumer,
		Args:  cobra.ExactArgs(2),
	})

	consumerRegisterCmd := &cobra.Command{
		Use:     "register <secret_id>",
		Aliases: []string{"create"},
		Short:   "Register a consumer for the given secret",
		Run:     cloud.RegisterKeyManagerSecretConsumer,
		Args:    cobra.ExactArgs(1),
	}
	addKeyManagerConsumerFlags(consumerRegisterCmd)
	consumerCmd.AddCommand(consumerRegisterCmd)

	consumerCmd.AddCommand(&cobra.Command{
		Use:   "delete <secret_id> <consumer_id>",
		Short: "Delete a consumer from the given secret",
		Run:   cloud.DeleteKeyManagerSecretConsumer,
		Args:  cobra.ExactArgs(2),
	})

	return consumerCmd
}

//
// Container command tree
//

func getKeyManagerContainerCmd() *cobra.Command {
	containerCmd := &cobra.Command{
		Use:   "container",
		Short: "Manage Key Manager containers",
	}

	containerListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List Key Manager containers",
		Run:     cloud.ListKeyManagerContainers,
	}
	containerCmd.AddCommand(withFilterFlag(containerListCmd))

	containerCmd.AddCommand(&cobra.Command{
		Use:   "get <container_id>",
		Short: "Get a specific Key Manager container",
		Run:   cloud.GetKeyManagerContainer,
		Args:  cobra.ExactArgs(1),
	})

	containerCmd.AddCommand(getKeyManagerContainerCreateCmd())

	containerEditCmd := &cobra.Command{
		Use:   "edit <container_id>",
		Short: "Edit the given Key Manager container (only secret references are mutable)",
		Run:   cloud.EditKeyManagerContainer,
		Args:  cobra.ExactArgs(1),
	}
	containerEditCmd.Flags().StringArrayVar(&cloud.KeyManagerContainerSecretRefs, "secret-ref", nil, "Secret reference as '<name>=<secretId>' (repeatable, replaces all existing references)")
	addInteractiveEditorFlag(containerEditCmd)
	containerCmd.AddCommand(containerEditCmd)

	containerCmd.AddCommand(&cobra.Command{
		Use:   "delete <container_id>",
		Short: "Delete the given Key Manager container",
		Run:   cloud.DeleteKeyManagerContainer,
		Args:  cobra.ExactArgs(1),
	})

	// Consumer subcommands
	containerCmd.AddCommand(getKeyManagerContainerConsumerCmd())

	return containerCmd
}

func getKeyManagerContainerCreateCmd() *cobra.Command {
	containerCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Key Manager container",
		Run:   cloud.CreateKeyManagerContainer,
	}

	spec := &cloud.KeyManagerContainerSpec.TargetSpec
	containerCreateCmd.Flags().StringVar(&spec.Name, "name", "", "Desired container name")
	containerCreateCmd.Flags().StringVar(&spec.Type, "type", "", "Type of the container (CERTIFICATE, GENERIC, RSA)")
	containerCreateCmd.Flags().StringArrayVar(&cloud.KeyManagerContainerSecretRefs, "secret-ref", nil, "Secret reference as '<name>=<secretId>' (repeatable)")
	containerCreateCmd.Flags().StringVar(&cloud.KeyManagerContainerCreateRegion, "region", "", "Region code where the container is located")
	containerCreateCmd.Flags().StringVar(&cloud.KeyManagerContainerCreateAvailabilityZone, "availability-zone", "", "Availability zone within the region")
	containerCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for the container to be ready before exiting")

	addParameterFileFlags(containerCreateCmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/keyManager/container", "post", cloud.KeyManagerContainerCreateExample, nil)
	addInteractiveEditorFlag(containerCreateCmd)
	markFlagsMutuallyExclusive(containerCreateCmd, "from-file", "editor")

	return containerCreateCmd
}

func getKeyManagerContainerConsumerCmd() *cobra.Command {
	consumerCmd := &cobra.Command{
		Use:   "consumer",
		Short: "Manage consumers of a Key Manager container",
	}

	consumerListCmd := &cobra.Command{
		Use:     "list <container_id>",
		Aliases: []string{"ls"},
		Short:   "List consumers registered for the given container",
		Run:     cloud.ListKeyManagerContainerConsumers,
		Args:    cobra.ExactArgs(1),
	}
	consumerCmd.AddCommand(withFilterFlag(consumerListCmd))

	consumerCmd.AddCommand(&cobra.Command{
		Use:   "get <container_id> <consumer_id>",
		Short: "Get a specific consumer of the given container",
		Run:   cloud.GetKeyManagerContainerConsumer,
		Args:  cobra.ExactArgs(2),
	})

	consumerRegisterCmd := &cobra.Command{
		Use:     "register <container_id>",
		Aliases: []string{"create"},
		Short:   "Register a consumer for the given container",
		Run:     cloud.RegisterKeyManagerContainerConsumer,
		Args:    cobra.ExactArgs(1),
	}
	addKeyManagerConsumerFlags(consumerRegisterCmd)
	consumerCmd.AddCommand(consumerRegisterCmd)

	consumerCmd.AddCommand(&cobra.Command{
		Use:   "delete <container_id> <consumer_id>",
		Short: "Delete a consumer from the given container",
		Run:   cloud.DeleteKeyManagerContainerConsumer,
		Args:  cobra.ExactArgs(2),
	})

	return consumerCmd
}

func addKeyManagerConsumerFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&cloud.KeyManagerConsumerSpec.ResourceId, "resource-id", "", "UUID of the resource consuming the secret/container")
	cmd.Flags().StringVar(&cloud.KeyManagerConsumerSpec.ResourceType, "resource-type", "", "Type of the consuming resource (IMAGE, INSTANCE, LOADBALANCER)")
	cmd.Flags().StringVar(&cloud.KeyManagerConsumerSpec.Service, "service", "", "OpenStack service type of the consumer (COMPUTE, IMAGE, LOADBALANCER, NETWORK)")
}
