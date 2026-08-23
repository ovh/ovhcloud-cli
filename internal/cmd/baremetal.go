// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/completion"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/baremetal"
	"github.com/spf13/cobra"
)

func init() {
	baremetalCmd := &cobra.Command{
		Use:   "baremetal",
		Short: "Retrieve information and manage your Bare Metal services",
	}

	// Command to list Baremetal services
	baremetalListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your Baremetal services",
		Run:     baremetal.ListBaremetal,
	}
	baremetalCmd.AddCommand(withFilterFlag(baremetalListCmd))

	// Command to get a single Baremetal
	baremetalCmd.AddCommand(&cobra.Command{
		Use:               "get <service_name>",
		Short:             "Retrieve information of a specific baremetal",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.GetBaremetal,
	})

	// Command to edit a single Baremetal
	editBaremetalCmd := &cobra.Command{
		Use:               "edit <service_name>",
		Short:             "Update the given baremetal",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.EditBaremetal,
	}
	editBaremetalCmd.Flags().IntVar(&baremetal.EditBaremetalParams.BootId, "boot-id", 0, "Boot ID")
	editBaremetalCmd.Flags().StringVar(&baremetal.EditBaremetalParams.BootScript, "boot-script", "", "Boot script")
	editBaremetalCmd.Flags().StringVar(&baremetal.EditBaremetalParams.EfiBootloaderPath, "efi-bootloader-path", "", "EFI bootloader path")
	editBaremetalCmd.Flags().BoolVar(&baremetal.EditBaremetalParams.Monitoring, "monitoring", false, "Enable monitoring")
	editBaremetalCmd.Flags().BoolVar(&baremetal.EditBaremetalParams.NoIntervention, "no-intervention", false, "Disable interventions")
	editBaremetalCmd.Flags().StringVar(&baremetal.EditBaremetalParams.RescueMail, "rescue-mail", "", "Rescue mail")
	editBaremetalCmd.Flags().StringVar(&baremetal.EditBaremetalParams.RescueSshKey, "rescue-ssh-key", "", "Rescue SSH key")
	editBaremetalCmd.Flags().StringVar(&baremetal.EditBaremetalParams.RootDevice, "root-device", "", "Root device")
	editBaremetalCmd.Flags().StringVar(&baremetal.EditBaremetalParams.State, "state", "", "State (e.g., error)")
	addInteractiveEditorFlag(editBaremetalCmd)
	baremetalCmd.AddCommand(editBaremetalCmd)

	// Command to list baremetal tasks
	baremetalListTasksCmd := &cobra.Command{
		Use:               "list-tasks <service_name>",
		Short:             "Retrieve tasks of the given baremetal",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ListBaremetalTasks,
	}
	baremetalCmd.AddCommand(withFilterFlag(baremetalListTasksCmd))

	// Catalogue: what can be ordered, where, how fast and at what price.
	baremetalCatalogCmd := &cobra.Command{
		Use:   "catalog",
		Short: "List orderable servers, their availability and their price",
		Long: `List orderable servers, where they can be delivered, how long delivery takes
and what they cost.

Delivery is a delay, not a yes or no: "24h" means a server ordered now is
expected within a day, while "2mo" is a two-month wait. Rows are sorted with the
soonest first.

Scale, High Grade and SAP servers appear with their availability but no price:
they are not sold from the public price list, and show as "on quotation".`,
		Args: cobra.NoArgs,
		Run:  baremetal.GetBaremetalCatalog,
	}
	baremetalCatalogCmd.Flags().StringVar(&baremetal.CatalogPlanCode, "plan-code", "", "Only this plan code (the identifier used to order)")
	baremetalCatalogCmd.Flags().StringVar(&baremetal.CatalogServer, "server", "", "Only this base hardware, for example 24ska01")
	baremetalCatalogCmd.Flags().StringVar(&baremetal.CatalogMemory, "memory", "", "Only this memory reference")
	baremetalCatalogCmd.Flags().StringVar(&baremetal.CatalogStorage, "storage", "", "Only this storage reference")
	baremetalCatalogCmd.Flags().StringVar(&baremetal.CatalogSystemStorage, "system-storage", "", "Only this system storage reference")
	baremetalCatalogCmd.Flags().StringVar(&baremetal.CatalogGPU, "gpu", "", "Only this GPU reference")
	baremetalCatalogCmd.Flags().StringSliceVar(&baremetal.CatalogDatacenters, "datacenter", nil, "Only these datacenters, named short (gra) or long (eu-west-gra-a) (repeatable)")
	baremetalCatalogCmd.Flags().StringVar(&baremetal.CatalogCommitment, "commitment", "default", "Which price to show: default (monthly), 12 or 24 (months paid upfront)")
	baremetalCatalogCmd.Flags().StringVar(&baremetal.CatalogCountry, "country", "", "Subsidiary whose price list to read (default: the one this account belongs to)")
	baremetalCatalogCmd.Flags().BoolVar(&baremetal.CatalogAvailableOnly, "available-only", false, "Hide what cannot be delivered today")
	baremetalCatalogCmd.Flags().BoolVar(&baremetal.CatalogRefresh, "refresh", false, "Download the price list again instead of reusing the one cached today")
	baremetalCatalogCmd.RegisterFlagCompletionFunc("datacenter", baremetal.CompleteCatalogDatacenter)
	baremetalCatalogCmd.RegisterFlagCompletionFunc("country", baremetal.CompleteCatalogCountry)
	baremetalCatalogCmd.RegisterFlagCompletionFunc("commitment", baremetal.CompleteCatalogCommitment)
	baremetalCmd.AddCommand(withFilterFlag(baremetalCatalogCmd))

	// Service information and life cycle
	baremetalServiceInfoCmd := &cobra.Command{
		Use:   "service-info",
		Short: "Manage service information of the given baremetal",
	}
	baremetalCmd.AddCommand(baremetalServiceInfoCmd)

	baremetalServiceInfoCmd.AddCommand(&cobra.Command{
		Use:               "get <service_name>",
		Short:             "Get service information of the given baremetal",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.GetBaremetalServiceInfo,
	})

	baremetalServiceInfoEditCmd := &cobra.Command{
		Use:               "edit <service_name>",
		Short:             "Edit service information of the given baremetal",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.EditBaremetalServiceInfo,
	}
	addServiceInfoRenewFlags(baremetalServiceInfoEditCmd)
	addInteractiveEditorFlag(baremetalServiceInfoEditCmd)
	baremetalServiceInfoCmd.AddCommand(baremetalServiceInfoEditCmd)

	baremetalTerminateCmd := &cobra.Command{
		Use:   "terminate <service_name>",
		Short: "Ask for the termination of the given baremetal",
		Long: `Ask for the termination of the given baremetal.

Nothing stops when this returns: OVHcloud emails a termination token to the
administrative contact of the service, and the server keeps running until that
token is confirmed with "baremetal confirm-termination".`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.TerminateBaremetal,
	}
	addConfirmationFlags(baremetalTerminateCmd, "Print the call that would be made without making it")
	baremetalCmd.AddCommand(baremetalTerminateCmd)

	baremetalConfirmTerminationCmd := &cobra.Command{
		Use:   "confirm-termination <service_name> <token>",
		Short: "Confirm the termination of the given baremetal",
		Long: `Confirm the termination of the given baremetal, using the token emailed to
the administrative contact by "baremetal terminate".

This ends the contract: the server is returned to OVHcloud at expiry.`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ConfirmBaremetalTermination,
	}
	baremetalConfirmTerminationCmd.Flags().StringVar(&baremetal.TerminationReason, "reason", "", "Why the service is being terminated (press <tab> for the accepted values)")
	baremetalConfirmTerminationCmd.Flags().StringVar(&baremetal.TerminationFutureUse, "future-use", "", "What comes next after this termination (press <tab> for the accepted values)")
	baremetalConfirmTerminationCmd.Flags().StringVar(&baremetal.TerminationComment, "commentary", "", "Free-text comment attached to the termination request")
	baremetalConfirmTerminationCmd.RegisterFlagCompletionFunc("reason", baremetal.CompleteTerminationReason)
	baremetalConfirmTerminationCmd.RegisterFlagCompletionFunc("future-use", baremetal.CompleteTerminationFutureUse)
	addConfirmationFlags(baremetalConfirmTerminationCmd, "Print the call that would be made without making it")
	baremetalCmd.AddCommand(baremetalConfirmTerminationCmd)

	// Command to reboot a baremetal
	baremetalRebootCmd := &cobra.Command{
		Use:               "reboot <service_name>",
		Short:             "Reboot the given baremetal",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.RebootBaremetal,
	}
	addConfirmationFlags(baremetalRebootCmd, "Print the call that would be made without making it")
	baremetalCmd.AddCommand(baremetalRebootCmd)

	// Command to reboot a baremetal in rescue mode
	baremetalRebootRescueCmd := &cobra.Command{
		Use:               "reboot-rescue <service_name>",
		Short:             "Reboot the given baremetal in rescue mode",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.RebootRescueBaremetal,
	}
	baremetalRebootRescueCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for reboot to be done before exiting")
	addConfirmationFlags(baremetalRebootRescueCmd, "Print the call that would be made without making it")
	baremetalCmd.AddCommand(baremetalRebootRescueCmd)

	// Command to reinstall a baremetal
	reinstallBaremetalCmd := &cobra.Command{
		Use:   "reinstall <service_name>",
		Short: "Reinstall the given baremetal",
		Long: `Use this command to reinstall the given dedicated server.
There are three ways to define the installation parameters:

1. Using only CLI flags:

	ovhcloud baremetal reinstall ns1234.ip-11.22.33.net --os byolinux_64 --language fr-fr --image-url https://...

2. Using a configuration file:

  First you can generate an example of installation file using the following command:

	ovhcloud baremetal reinstall --init-file ./install.json

  You will be able to choose from several installation examples. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct installation parameters, run:

	ovhcloud baremetal reinstall ns1234.ip-11.22.33.net --from-file ./install.json

  Note that you can also pipe the content of the file to reinstall, like the following:

	cat ./install.json | ovhcloud baremetal reinstall ns1234.ip-11.22.33.net --yes

  Piped input is not a terminal, so there is nobody to answer the confirmation:
  such a run refuses to start unless --yes is given.

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud baremetal reinstall ns1234.ip-11.22.33.net --from-file ./install.json --hostname new-hostname

3. Using your default text editor:

	ovhcloud baremetal reinstall ns1234.ip-11.22.33.net --editor

  You will be able to choose from several installation examples. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the reinstallation will be run.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud baremetal reinstall ns1234.ip-11.22.33.net --editor --os debian12_64

You can visit https://eu.api.ovh.com/console/?section=%2Fdedicated%2Fserver&branch=v1#post-/dedicated/server/-serviceName-/reinstall
to see all the available parameters and real life examples.

Please note that all parameters are not compatible with all OSes.

Reinstalling wipes every disk of the server, so the command asks for a
confirmation: type the server name when prompted. Unattended runs (pipelines,
piped input) must pass --yes, and --dry-run prints the parameters that would
be sent without sending them.
`,
		Args:              cobra.MaximumNArgs(1),
		ArgAliases:        []string{"service_name"},
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ReinstallBaremetal,
	}

	addParameterFileFlags(reinstallBaremetalCmd, false, assets.BaremetalOpenapiSchema, "/dedicated/server/{serviceName}/reinstall", "post", baremetal.BaremetalInstallationExample, nil)
	addInteractiveEditorFlag(reinstallBaremetalCmd)
	reinstallBaremetalCmd.Flags().StringVar(&baremetal.OperatingSystem, "os", "", "Operating system to install")
	reinstallBaremetalCmd.Flags().StringVar(&baremetal.Customizations.ConfigDriveUserData, "config-drive-user-data", "", "Config Drive UserData")
	reinstallBaremetalCmd.Flags().StringVar(&baremetal.Customizations.EfiBootloaderPath, "efi-bootloader-path", "", "Path of the EFI bootloader from the OS installed on the server")
	reinstallBaremetalCmd.Flags().StringVar(&baremetal.Customizations.Hostname, "hostname", "", "Custom hostname")
	reinstallBaremetalCmd.Flags().StringToStringVar(&baremetal.Customizations.HttpHeaders, "http-headers", nil, "Image HTTP headers")
	reinstallBaremetalCmd.Flags().StringVar(&baremetal.Customizations.ImageCheckSum, "image-checksum", "", "Image checksum")
	reinstallBaremetalCmd.Flags().StringVar(&baremetal.Customizations.ImageCheckSumType, "image-checksum-type", "", "Image checksum type")
	reinstallBaremetalCmd.Flags().StringVar(&baremetal.Customizations.ImageType, "image-type", "", "Image type (qcow, raw)")
	reinstallBaremetalCmd.Flags().StringVar(&baremetal.Customizations.ImageURL, "image-url", "", "Image URL")
	reinstallBaremetalCmd.Flags().StringVar(&baremetal.Customizations.Language, "language", "", "Display language")
	reinstallBaremetalCmd.Flags().StringVar(&baremetal.Customizations.PostInstallationScript, "post-installation-script", "", "Post-installation script")
	reinstallBaremetalCmd.Flags().StringVar(&baremetal.Customizations.PostInstallationScriptExtension, "post-installation-script-extension", "", "Post-installation script extension (cmd, ps1)")
	reinstallBaremetalCmd.Flags().StringVar(&baremetal.Customizations.SshKey, "ssh-key", "", "SSH public key")
	reinstallBaremetalCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for reinstall to be done before exiting")
	addConfirmationFlags(reinstallBaremetalCmd, "Print the installation parameters without sending anything")
	markFlagsMutuallyExclusive(reinstallBaremetalCmd, "from-file", "editor")
	baremetalCmd.AddCommand(reinstallBaremetalCmd)

	// List boots and their options
	baremetalBootCmd := &cobra.Command{
		Use:   "boot",
		Short: "Manage boot options for the given baremetal",
	}
	baremetalCmd.AddCommand(baremetalBootCmd)
	baremetalBootCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list <service_name>",
		Aliases:           []string{"ls"},
		Short:             "List boot options for the given baremetal",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ListBaremetalBoots,
	}))
	baremetalBootCmd.AddCommand(&cobra.Command{
		Use:               "set <service_name> <boot_id>",
		Short:             "Configure a boot ID on the given baremetal",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.SetBaremetalBootId,
	})

	baremetalBootSetScriptCmd := &cobra.Command{
		Use:               "set-script <service_name>",
		Short:             "Configure a boot script on the given baremetal",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.SetBaremetalBootScript,
	}
	baremetalBootSetScriptCmd.Flags().StringVar(&baremetal.EditBaremetalParams.BootScript, "script", "", "Boot script to set on the baremetal")
	addInteractiveEditorFlag(baremetalBootSetScriptCmd)
	addParameterFileFlags(baremetalBootSetScriptCmd, true, nil, "", "", "", nil)
	markFlagsOneRequired(baremetalBootSetScriptCmd, "script", "from-file", "editor")
	markFlagsMutuallyExclusive(baremetalBootSetScriptCmd, "script", "from-file", "editor")
	baremetalBootCmd.AddCommand(baremetalBootSetScriptCmd)

	baremetalListInterventionsCmd := &cobra.Command{
		Use:               "list-interventions <service_name>",
		Short:             "List past and planned interventions for the given baremetal",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ListBaremetalInterventions,
	}
	baremetalCmd.AddCommand(withFilterFlag(baremetalListInterventionsCmd))

	baremetalCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list-ips <service_name>",
		Short:             "List all IPs that are routed to the given baremetal",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.GetBaremetalRelatedIPs,
	}))

	baremetalCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list-secrets <service_name>",
		Short:             "Retrieve secrets to connect to the server",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.GetBaremetalAuthenticationSecrets,
	}))

	baremetalCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list-compatible-os <service_name>",
		Short:             "Retrieve OSes that can be installed on this baremetal",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.GetBaremetalCompatibleOses,
	}))

	// Commands to manage virtual network interfaces
	baremetalVNICmd := &cobra.Command{
		Use:   "vni",
		Short: "Manage Virtual Network Interfaces of the given baremetal",
	}
	baremetalCmd.AddCommand(baremetalVNICmd)
	baremetalVNICmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list <service_name>",
		Aliases:           []string{"ls"},
		Short:             "List Virtual Network Interfaces of the given baremetal",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ListBaremetalVNIs,
	}))
	baremetalVNICreateOLAAggregationCmd := &cobra.Command{
		Use:               "ola-create-aggregation <service_name> --name <name> --interface <uuid> --interface <uuid>",
		Short:             "Group interfaces into an aggregation",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.CreateBaremetalOLAAggregation,
	}
	baremetalVNICreateOLAAggregationCmd.Flags().StringArrayVar(&baremetal.BaremetalOLAInterfaces, "interface", nil, "Interfaces to group")
	baremetalVNICreateOLAAggregationCmd.MarkFlagRequired("interface")
	baremetalVNICreateOLAAggregationCmd.Flags().StringVar(&baremetal.BaremetalOLAName, "name", "", "Name of the aggregation")
	baremetalVNICreateOLAAggregationCmd.MarkFlagRequired("name")
	baremetalVNICmd.AddCommand(baremetalVNICreateOLAAggregationCmd)

	baremetalVNIResetOLAAggregationCmd := &cobra.Command{
		Use:               "ola-reset <service_name> --interface <uuid> --interface <uuid>",
		Short:             "Reset interfaces to default configuration",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ResetBaremetalOLAAggregation,
	}
	baremetalVNIResetOLAAggregationCmd.Flags().StringArrayVar(&baremetal.BaremetalOLAInterfaces, "interface", nil, "Interfaces to group")
	baremetalVNIResetOLAAggregationCmd.MarkFlagRequired("interface")
	addConfirmationFlags(baremetalVNIResetOLAAggregationCmd, "Print the call that would be made without making it")
	baremetalVNICmd.AddCommand(baremetalVNIResetOLAAggregationCmd)

	baremetalIPMICmd := &cobra.Command{
		Use:   "ipmi",
		Short: "Manage IPMI on your baremetal",
	}
	baremetalCmd.AddCommand(baremetalIPMICmd)

	baremetalIPMIGetAccessCmd := &cobra.Command{
		Use:               "get-access <service_name> --type serialOverLanURL --ttl 5",
		Short:             "Request an acces on KVM IPMI interface",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.BaremetalGetIPMIAccess,
	}
	baremetalIPMIGetAccessCmd.Flags().IntVar(&baremetal.BaremetalIpmiTTL, "ttl", 1, "Time to live in minutes for cache (1, 3, 5, 10, 15)")
	baremetalIPMIGetAccessCmd.MarkFlagRequired("ttl")
	baremetalIPMIGetAccessCmd.Flags().StringVar(&baremetal.BaremetalIpmiAccessType, "type", "", "Distinct way to acces a KVM IPMI session (kvmipHtml5URL, kvmipJnlp, serialOverLanSshKey, serialOverLanURL)")
	baremetalIPMIGetAccessCmd.MarkFlagRequired("type")
	baremetalIPMIGetAccessCmd.Flags().StringVar(&baremetal.BaremetalIpmiIP, "allowed-ip", "", "IPv4 address that can use the access")
	baremetalIPMIGetAccessCmd.Flags().StringVar(&baremetal.BaremetalIpmiSshKey, "ssh-key", "", "Public SSH key for Serial Over Lan SSH access")
	baremetalIPMICmd.AddCommand(baremetalIPMIGetAccessCmd)

	baremetalIPMICmd.AddCommand(&cobra.Command{
		Use:               "reset-sessions <service_name>",
		Short:             "Reset IPMI sessions on a baremetal server",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.BaremetalResetIPMISessions,
	})

	rootCmd.AddCommand(baremetalCmd)
}
