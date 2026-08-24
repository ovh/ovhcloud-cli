// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"time"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/completion"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/baremetal"
	"github.com/ovh/ovhcloud-cli/internal/services/vrack"
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
	editBaremetalCmd.Flags().BoolVar(&baremetal.EditBaremetalMonitoring, "monitoring", false, "Enable monitoring")
	editBaremetalCmd.Flags().BoolVar(&baremetal.EditBaremetalNoIntervention, "no-intervention", false, "Disable interventions")
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

	// Powering a server off and back on.
	//
	// There is no power endpoint in this API. `power` is a boot type: setting
	// the server's boot to its "Power-off server" entry and rebooting into it
	// halts the machine, and putting the previous boot back and rebooting
	// starts it again.
	baremetalPowerCmd := &cobra.Command{
		Use:   "power",
		Short: "Power the given baremetal off and on",
		Long: `Power the given dedicated server off and on.

There is no power switch in the API: powering off works by setting the server's
"Power-off server" boot entry and rebooting into it. That has a consequence
worth knowing — a server left on that entry shuts itself down at every reboot,
including a reboot asked for from the manager. "power on" therefore puts the
previous boot back before starting the server, and "power status" says so when
a server is sitting on the power-off entry.`,
	}
	baremetalCmd.AddCommand(baremetalPowerCmd)

	baremetalPowerStatusCmd := &cobra.Command{
		Use:               "status <service_name>",
		Short:             "Show whether the server is on, and what it will boot on",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.GetBaremetalPowerStatus,
	}
	// No --filter here: `power status` answers about one server and renders one
	// object, so there are no rows to filter. Registering the flag would have
	// docgen document it and cobra accept it, on a command where it can only
	// ever be ignored.
	baremetalPowerCmd.AddCommand(baremetalPowerStatusCmd)

	baremetalPowerOffCmd := &cobra.Command{
		Use:               "off <service_name>",
		Short:             "Power off the given baremetal",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.PowerOffBaremetal,
	}
	baremetalPowerOffCmd.Flags().BoolVar(&baremetal.PowerWait, "wait", false, "Wait until the server reports itself off before exiting")
	baremetalPowerOffCmd.Flags().DurationVar(&baremetal.PowerTimeout, "timeout", 10*time.Minute, "How long --wait waits")
	addConfirmationFlags(baremetalPowerOffCmd, "Print the two calls that would be made without making them")
	baremetalPowerCmd.AddCommand(baremetalPowerOffCmd)

	baremetalPowerOnCmd := &cobra.Command{
		Use:               "on <service_name>",
		Short:             "Power on the given baremetal, restoring the boot it was on",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.PowerOnBaremetal,
	}
	baremetalPowerOnCmd.Flags().BoolVar(&baremetal.PowerWait, "wait", false, "Wait until the server reports itself on before exiting")
	baremetalPowerOnCmd.Flags().DurationVar(&baremetal.PowerTimeout, "timeout", 10*time.Minute, "How long --wait waits")
	baremetalPowerOnCmd.Flags().IntVar(&baremetal.PowerBootID, "boot", 0, "Boot to start on, instead of the one the server was on before it was powered off")
	// --dry-run without --yes: powering a server on interrupts nothing, so there
	// is no prompt to skip, and offering a flag that skips nothing would be one
	// more thing to learn that turns out to mean nothing.
	baremetalPowerOnCmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "Print the calls that would be made without making them")
	baremetalPowerCmd.AddCommand(baremetalPowerOnCmd)

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

	// What to put in the storage block of a reinstall. The block itself already
	// travels through --from-file; these are the three questions you have to
	// answer before writing it, and on a reinstall the disks are wiped before
	// the API tells you the answer was wrong.
	listPartitionSchemesCmd := &cobra.Command{
		Use:               "list-partition-schemes <service_name>",
		Short:             "List the partition schemes an OS template allows on this baremetal",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ListBaremetalPartitionSchemes,
	}
	listPartitionSchemesCmd.Flags().StringVar(&baremetal.InstallTemplate, "os", "",
		"OS template the schemes are relative to (see `ovhcloud baremetal list-compatible-os`)")
	listPartitionSchemesCmd.MarkFlagRequired("os")
	baremetalCmd.AddCommand(withFilterFlag(listPartitionSchemesCmd))

	baremetalCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "raid-profile <service_name>",
		Short:             "Show the hardware RAID controllers of this baremetal, if it has any",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.GetBaremetalRaidProfile,
	}))

	// A reinstall runs for tens of minutes and, until now, said nothing while
	// it did. This reads the same progress `reinstall --wait` follows, for
	// somebody who started the install in another terminal — or who answered
	// the confirmation, walked away, and wants to know whether to keep waiting.
	// Traffic graphs, from the only mrtg route that is not deprecated. The
	// controllers are resolved from the server, because a MAC address is
	// otherwise not something this CLI can tell anybody.
	baremetalTrafficCmd := &cobra.Command{
		Use:               "traffic <service_name>",
		Short:             "Show the traffic graphs of this baremetal's network controllers",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ShowBaremetalTraffic,
	}
	baremetalTrafficCmd.Flags().StringVar(&baremetal.TrafficPeriod, "period", "daily",
		"Window the graph covers: hourly, daily, weekly, monthly or yearly")
	// Declared with no default on purpose: PostExecute resets a stringSlice by
	// replacing it with nil rather than with DefValue — DefValue is "[]" for a
	// slice, so it cannot be used — and a default that survives only the first
	// command of a process is worse than no default. The default lives in
	// defaultTrafficTypes, next to the code that reads it, and is stated here.
	baremetalTrafficCmd.Flags().StringSliceVar(&baremetal.TrafficTypes, "type", nil,
		"Series to read: traffic, packets or errors, each :download or :upload (default: traffic:download,traffic:upload)")
	baremetalTrafficCmd.Flags().StringVar(&baremetal.TrafficNIC, "nic", "",
		"Read only this controller, by MAC address (default: every controller of the server)")
	baremetalCmd.AddCommand(baremetalTrafficCmd)

	baremetalCmd.AddCommand(&cobra.Command{
		Use:               "install-status <service_name>",
		Short:             "Show how far the running installation of this baremetal has got",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ShowBaremetalInstallStatus,
	})

	// Commands to manage virtual network interfaces
	// Private network, seen from the machine. The same work lives under
	// `ovhcloud vrack`; this is where somebody holding a server looks for it.
	baremetalVrackCmd := &cobra.Command{
		Use:   "vrack",
		Short: "Attach the given baremetal to a vRack, or detach it",
	}
	baremetalCmd.AddCommand(baremetalVrackCmd)

	baremetalVrackCmd.AddCommand(&cobra.Command{
		Use:               "show <service_name>",
		Short:             "Show the vRack the given baremetal is in",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ShowBaremetalVrack,
	})

	baremetalVrackAttachCmd := &cobra.Command{
		Use:               "attach <service_name> <vrack>",
		Short:             "Attach the given baremetal to a vRack",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.AttachBaremetalToVrack,
	}
	addVrackTaskFlags(baremetalVrackAttachCmd)
	baremetalVrackAttachCmd.Flags().StringVar(&vrack.VrackInterface, "interface", "",
		"Interface to attach, when the server has several")
	addConfirmationFlags(baremetalVrackAttachCmd, "Print the call that would be made without making it")
	baremetalVrackCmd.AddCommand(baremetalVrackAttachCmd)

	// The vRack argument is optional: a server is in at most one, and this CLI
	// can read which. Making the operator look it up first would be work the
	// tool exists to remove.
	baremetalVrackDetachCmd := &cobra.Command{
		Use:               "detach <service_name> [vrack]",
		Short:             "Detach the given baremetal from its vRack",
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.DetachBaremetalFromVrack,
	}
	addVrackTaskFlags(baremetalVrackDetachCmd)
	addConfirmationFlags(baremetalVrackDetachCmd, "Print the call that would be made without making it")
	baremetalVrackCmd.AddCommand(baremetalVrackDetachCmd)

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

	// What quietly breaks a dedicated server is spread across five routes and
	// none of them is where somebody would look. This reads them together.
	baremetalDoctorCmd := &cobra.Command{
		Use:   "doctor [service_name...]",
		Short: "Report what is wrong with a server, or with every server",
		Long: `Check the things that silently break a dedicated server: a machine left on the
rescue system, monitoring switched off, hardware intervention refused, a renewal
that will not happen, work still running, maintenance already planned.

With no argument it checks every server of the account.

The exit code stays 0 when findings are reported, because the command ran and
answered. Use --strict to make findings fail the command instead, which is what
a pipeline gating on it wants.

--strict fails on a warning or a critical. A note never fails it: every server
renewing inside the next 30 days reports one, so a --strict that counted notes
would be red permanently, and a gate that is always red is read like no gate.
Narrowing with --filter narrows the gate too — 'severity=="critical"' fails only
on criticals.`,
		Args:              cobra.ArbitraryArgs,
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.Doctor,
	}
	baremetalDoctorCmd.Flags().IntVar(&baremetal.DoctorExpiryDays, "expiry-days", 30,
		"Report a server expiring within this many days")
	baremetalDoctorCmd.Flags().BoolVar(&baremetal.DoctorStrict, "strict", false,
		"Exit non-zero when a warning or a critical is reported (notes never fail it)")
	baremetalCmd.AddCommand(withFilterFlag(baremetalDoctorCmd))

	// Nine backup routes, none of them reachable: the space included with the
	// server, the access list that guards it, the two passwords and the cloud
	// backup beside it.
	baremetalBackupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage the backup spaces of a dedicated server",
	}
	baremetalCmd.AddCommand(baremetalBackupCmd)

	baremetalBackupCmd.AddCommand(&cobra.Command{
		Use:               "orderable <service_name>",
		Short:             "Show the backup storage capacities this server accepts",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ShowOrderableBackupStorage,
	})

	// Backup FTP
	baremetalBackupFtpCmd := &cobra.Command{
		Use:   "ftp",
		Short: "Manage the Backup FTP space included with the server",
	}
	baremetalBackupCmd.AddCommand(baremetalBackupFtpCmd)

	baremetalBackupFtpCmd.AddCommand(&cobra.Command{
		Use:               "show <service_name>",
		Short:             "Show the Backup FTP space of this server",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ShowBackupFtp,
	})

	baremetalBackupFtpCreateCmd := &cobra.Command{
		Use:               "create <service_name>",
		Short:             "Create the Backup FTP space included with this server",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.CreateBackupFtp,
	}
	baremetalBackupFtpCreateCmd.Flags().BoolVar(&baremetal.BackupWait, "wait", false,
		"Wait until the space actually exists before exiting")
	addConfirmationFlags(baremetalBackupFtpCreateCmd, "Print the call that would be made without making it")
	baremetalBackupFtpCmd.AddCommand(baremetalBackupFtpCreateCmd)

	baremetalBackupFtpDeleteCmd := &cobra.Command{
		Use:               "delete <service_name>",
		Short:             "Terminate the Backup FTP space — all data is permanently deleted",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.DeleteBackupFtp,
	}
	baremetalBackupFtpDeleteCmd.Flags().BoolVar(&baremetal.BackupWait, "wait", false,
		"Wait until the space is actually gone before exiting")
	addConfirmationFlags(baremetalBackupFtpDeleteCmd, "Print the call that would be made without making it")
	baremetalBackupFtpCmd.AddCommand(baremetalBackupFtpDeleteCmd)

	baremetalBackupFtpPasswordCmd := &cobra.Command{
		Use:               "password <service_name>",
		Short:             "Change the Backup FTP password (the new one is emailed)",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ChangeBackupFtpPassword,
	}
	addConfirmationFlags(baremetalBackupFtpPasswordCmd, "Print the call that would be made without making it")
	baremetalBackupFtpCmd.AddCommand(baremetalBackupFtpPasswordCmd)

	baremetalBackupFtpCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "authorizable-blocks <service_name>",
		Short:             "List the IP blocks that may be allowed on this Backup FTP space",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ListAuthorizableBlocks,
	}))

	baremetalBackupAclCmd := &cobra.Command{
		Use:   "acl",
		Short: "Manage who may reach the Backup FTP space, and how",
	}
	baremetalBackupFtpCmd.AddCommand(baremetalBackupAclCmd)

	baremetalBackupAclCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:               "list <service_name>",
		Aliases:           []string{"ls"},
		Short:             "List the access rules of the Backup FTP space",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ListBackupFtpAcl,
	}))

	baremetalBackupAclCmd.AddCommand(&cobra.Command{
		Use:               "get <service_name> <ip_block>",
		Short:             "Show one access rule",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.GetBackupFtpAcl,
	})

	baremetalBackupAclAddCmd := &cobra.Command{
		Use:               "add <service_name> <ip_block>",
		Short:             "Allow an IP block to reach the Backup FTP space",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.AddBackupFtpAcl,
	}
	addBackupAclProtocolFlags(baremetalBackupAclAddCmd)
	addConfirmationFlags(baremetalBackupAclAddCmd, "Print the call that would be made without making it")
	baremetalBackupAclCmd.AddCommand(baremetalBackupAclAddCmd)

	baremetalBackupAclSetCmd := &cobra.Command{
		Use:               "set <service_name> <ip_block>",
		Short:             "Change the protocols an access rule opens",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.SetBackupFtpAcl,
	}
	addBackupAclProtocolFlags(baremetalBackupAclSetCmd)
	addConfirmationFlags(baremetalBackupAclSetCmd, "Print the call that would be made without making it")
	baremetalBackupAclCmd.AddCommand(baremetalBackupAclSetCmd)

	baremetalBackupAclDeleteCmd := &cobra.Command{
		Use:               "delete <service_name> <ip_block>",
		Short:             "Stop an IP block reaching the Backup FTP space",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.DeleteBackupFtpAcl,
	}
	addConfirmationFlags(baremetalBackupAclDeleteCmd, "Print the call that would be made without making it")
	baremetalBackupAclCmd.AddCommand(baremetalBackupAclDeleteCmd)

	// Cloud backup
	baremetalBackupCloudCmd := &cobra.Command{
		Use:   "cloud",
		Short: "Manage the cloud backup containers of the server",
	}
	baremetalBackupCmd.AddCommand(baremetalBackupCloudCmd)

	// --reveal on both, and for the same reason it exists on `password`: the
	// object these print carries the four container passwords, one level down.
	baremetalBackupCloudShowCmd := &cobra.Command{
		Use:               "show <service_name>",
		Short:             "Show the cloud backup containers of this server",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ShowBackupCloud,
	}
	baremetalBackupCloudShowCmd.Flags().BoolVar(&baremetal.RevealBackupPassword, "reveal", false,
		"Print the container passwords instead of their fingerprints")
	baremetalBackupCloudCmd.AddCommand(baremetalBackupCloudShowCmd)

	baremetalBackupCloudCmd.AddCommand(&cobra.Command{
		Use:               "offer <service_name>",
		Short:             "Show what a cloud backup would hold for this server",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ShowBackupCloudOffer,
	})

	baremetalBackupCloudCreateCmd := &cobra.Command{
		Use:               "create <service_name>",
		Short:             "Create the cloud backup containers of this server",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.CreateBackupCloud,
	}
	baremetalBackupCloudCreateCmd.Flags().StringVar(&baremetal.BackupCloudProjectId, "project-id", "",
		"Public cloud project to hold the containers")
	baremetalBackupCloudCreateCmd.Flags().StringVar(&baremetal.BackupCloudProjectDescription, "project-description", "",
		"Description of the project to create, when none is given")
	baremetalBackupCloudCreateCmd.Flags().BoolVar(&baremetal.RevealBackupPassword, "reveal", false,
		"Print the container passwords instead of their fingerprints")
	addConfirmationFlags(baremetalBackupCloudCreateCmd, "Print the call that would be made without making it")
	baremetalBackupCloudCmd.AddCommand(baremetalBackupCloudCreateCmd)

	baremetalBackupCloudDeleteCmd := &cobra.Command{
		Use:               "delete <service_name>",
		Short:             "Deactivate the cloud backup — the container data is kept",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.DeleteBackupCloud,
	}
	addConfirmationFlags(baremetalBackupCloudDeleteCmd, "Print the call that would be made without making it")
	baremetalBackupCloudCmd.AddCommand(baremetalBackupCloudDeleteCmd)

	baremetalBackupCloudPasswordCmd := &cobra.Command{
		Use:               "password <service_name>",
		Short:             "Reset the four cloud backup passwords",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/dedicated/server"),
		Run:               baremetal.ChangeBackupCloudPassword,
	}
	baremetalBackupCloudPasswordCmd.Flags().BoolVar(&baremetal.RevealBackupPassword, "reveal", false,
		"Print the new passwords instead of their fingerprints")
	addConfirmationFlags(baremetalBackupCloudPasswordCmd, "Print the call that would be made without making it")
	baremetalBackupCloudCmd.AddCommand(baremetalBackupCloudPasswordCmd)

	rootCmd.AddCommand(baremetalCmd)
}

// addBackupAclProtocolFlags registers the three protocols an access rule can
// open. They are registered together because the API takes all three and
// accepts all three false — a rule that allows an IP block to reach nothing.
func addBackupAclProtocolFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&baremetal.BackupAclFtp, "ftp", false, "Allow the FTP protocol")
	cmd.Flags().BoolVar(&baremetal.BackupAclNfs, "nfs", false, "Allow the NFS protocol")
	cmd.Flags().BoolVar(&baremetal.BackupAclCifs, "cifs", false, "Allow the CIFS (SMB) protocol")
}
