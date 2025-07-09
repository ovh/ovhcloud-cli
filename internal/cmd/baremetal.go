package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/baremetal"
)

func init() {
	baremetalCmd := &cobra.Command{
		Use:   "baremetal",
		Short: "Retrieve information and manage your baremetal services",
	}

	// Command to list Baremetal services
	baremetalListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Baremetal services",
		Run:   baremetal.ListBaremetal,
	}
	baremetalCmd.AddCommand(withFilterFlag(baremetalListCmd))

	// Command to get a single Baremetal
	baremetalCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific baremetal",
		Args:  cobra.ExactArgs(1),
		Run:   baremetal.GetBaremetal,
	})

	// Command to edit a single Baremetal
	editBaremetalCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Update the given baremetal",
		Args:  cobra.ExactArgs(1),
		Run:   baremetal.EditBaremetal,
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
	editBaremetalCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define parameters")
	baremetalCmd.AddCommand(editBaremetalCmd)

	// Command to list baremetal tasks
	baremetalListTasksCmd := &cobra.Command{
		Use:   "list-tasks <service_name>",
		Short: "Retrieve tasks of the given baremetal",
		Args:  cobra.ExactArgs(1),
		Run:   baremetal.ListBaremetalTasks,
	}
	baremetalCmd.AddCommand(withFilterFlag(baremetalListTasksCmd))

	// Command to reboot a baremetal
	baremetalRebootCmd := &cobra.Command{
		Use:   "reboot <service_name>",
		Short: "Reboot the given baremetal",
		Args:  cobra.ExactArgs(1),
		Run:   baremetal.RebootBaremetal,
	}
	removeRootFlagsFromCommand(baremetalRebootCmd)
	baremetalCmd.AddCommand(baremetalRebootCmd)

	// Command to reboot a baremetal in rescue mode
	baremetalRebootRescueCmd := &cobra.Command{
		Use:   "reboot-rescue <service_name>",
		Short: "Reboot the given baremetal in rescue mode",
		Args:  cobra.ExactArgs(1),
		Run:   baremetal.RebootRescueBaremetal,
	}
	removeRootFlagsFromCommand(baremetalRebootRescueCmd)
	baremetalRebootRescueCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for reboot to be done before exiting")
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

	cat ./install.json | ovhcloud baremetal reinstall ns1234.ip-11.22.33.net

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
`,
		Args:       cobra.MaximumNArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        baremetal.ReinstallBaremetal,
	}

	addInitParameterFileFlag(reinstallBaremetalCmd, baremetal.BaremetalOpenapiSchema, "/dedicated/server/{serviceName}/reinstall", "post", baremetal.BaremetalInstallationExample, nil)
	reinstallBaremetalCmd.Flags().StringVar(&flags.ParametersFile, "from-file", "", "File containing installation parameters")
	reinstallBaremetalCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define installation parameters")
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
	removeRootFlagsFromCommand(reinstallBaremetalCmd)
	reinstallBaremetalCmd.MarkFlagsMutuallyExclusive("from-file", "editor")
	baremetalCmd.AddCommand(reinstallBaremetalCmd)

	// List boots and their options
	baremetalBootCmd := &cobra.Command{
		Use:   "boot",
		Short: "Manage boot options for the given baremetal",
	}
	baremetalCmd.AddCommand(baremetalBootCmd)
	baremetalBootCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list <service_name>",
		Short: "List boot options for the given baremetal",
		Args:  cobra.ExactArgs(1),
		Run:   baremetal.ListBaremetalBoots,
	}))
	baremetalBootCmd.AddCommand(&cobra.Command{
		Use:   "set <service_name> <boot_id>",
		Short: "Configure a boot ID on the given baremetal",
		Args:  cobra.ExactArgs(2),
		Run:   baremetal.SetBaremetalBootId,
	})
	baremetalBootSetScriptCmd := &cobra.Command{
		Use:   "set-script <service_name>",
		Short: "Configure a boot script on the given baremetal",
		Args:  cobra.ExactArgs(1),
		Run:   baremetal.SetBaremetalBootScript,
	}
	baremetalBootSetScriptCmd.Flags().StringVar(&flags.ParametersFile, "from-file", "", "File containing boot script")
	baremetalBootSetScriptCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define the boot script")
	baremetalBootSetScriptCmd.MarkFlagsOneRequired("from-file", "editor")
	baremetalBootSetScriptCmd.MarkFlagsMutuallyExclusive("from-file", "editor")
	baremetalBootCmd.AddCommand(baremetalBootSetScriptCmd)

	baremetalListInterventionsCmd := &cobra.Command{
		Use:   "list-interventions <service_name>",
		Short: "List past and planned interventions for the given baremetal",
		Args:  cobra.ExactArgs(1),
		Run:   baremetal.ListBaremetalInterventions,
	}
	baremetalCmd.AddCommand(withFilterFlag(baremetalListInterventionsCmd))

	baremetalCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list-ips <service_name>",
		Short: "List all IPs that are routed to the given baremetal",
		Args:  cobra.ExactArgs(1),
		Run:   baremetal.GetBaremetalRelatedIPs,
	}))

	baremetalCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list-secrets <service_name>",
		Short: "Retrieve secrets to connect to the server",
		Args:  cobra.ExactArgs(1),
		Run:   baremetal.GetBaremetalAuthenticationSecrets,
	}))

	baremetalCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list-compatible-os <service_name>",
		Short: "Retrieve OSes that can be installed on this baremetal",
		Args:  cobra.ExactArgs(1),
		Run:   baremetal.GetBaremetalCompatibleOses,
	}))

	// Commands to manage virtual network interfaces
	baremetalVNICmd := &cobra.Command{
		Use:   "vni",
		Short: "Manage Virtual Network Interfaces of the given baremetal",
	}
	baremetalCmd.AddCommand(baremetalVNICmd)
	baremetalVNICmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list <service_name>",
		Short: "List Virtual Network Interfaces of the given baremetal",
		Args:  cobra.ExactArgs(1),
		Run:   baremetal.ListBaremetalVNIs,
	}))
	baremetalVNICreateOLAAggregationCmd := &cobra.Command{
		Use:   "ola-create-aggregation <service_name> --name <name> --interface <uuid> --interface <uuid>",
		Short: "Group interfaces into an aggregation",
		Args:  cobra.ExactArgs(1),
		Run:   baremetal.CreateBaremetalOLAAggregation,
	}
	baremetalVNICreateOLAAggregationCmd.Flags().StringArrayVar(&baremetal.BaremetalOLAInterfaces, "interface", nil, "Interfaces to group")
	baremetalVNICreateOLAAggregationCmd.MarkFlagRequired("interface")
	baremetalVNICreateOLAAggregationCmd.Flags().StringVar(&baremetal.BaremetalOLAName, "name", "", "Name of the aggregation")
	baremetalVNICreateOLAAggregationCmd.MarkFlagRequired("name")
	baremetalVNICmd.AddCommand(baremetalVNICreateOLAAggregationCmd)

	baremetalVNIResetOLAAggregationCmd := &cobra.Command{
		Use:   "ola-reset <service_name> --interface <uuid> --interface <uuid>",
		Short: "Reset interfaces to default configuration",
		Args:  cobra.ExactArgs(1),
		Run:   baremetal.ResetBaremetalOLAAggregation,
	}
	baremetalVNIResetOLAAggregationCmd.Flags().StringArrayVar(&baremetal.BaremetalOLAInterfaces, "interface", nil, "Interfaces to group")
	baremetalVNIResetOLAAggregationCmd.MarkFlagRequired("interface")
	baremetalVNICmd.AddCommand(baremetalVNIResetOLAAggregationCmd)

	baremetalIPMICmd := &cobra.Command{
		Use:   "ipmi",
		Short: "Manage IPMI on your baremetal",
	}
	baremetalCmd.AddCommand(baremetalIPMICmd)

	baremetalIPMIGetAccessCmd := &cobra.Command{
		Use:   "get-access <service_name> --type serialOverLanURL --ttl 5",
		Short: "Request an acces on KVM IPMI interface",
		Args:  cobra.ExactArgs(1),
		Run:   baremetal.BaremetalGetIPMIAccess,
	}
	baremetalIPMIGetAccessCmd.Flags().IntVar(&baremetal.BaremetalIpmiTTL, "ttl", 1, "Time to live in minutes for cache (1, 3, 5, 10, 15)")
	baremetalIPMIGetAccessCmd.MarkFlagRequired("ttl")
	baremetalIPMIGetAccessCmd.Flags().StringVar(&baremetal.BaremetalIpmiAccessType, "type", "", "Distinct way to acces a KVM IPMI session (kvmipHtml5URL, kvmipJnlp, serialOverLanSshKey, serialOverLanURL)")
	baremetalIPMIGetAccessCmd.MarkFlagRequired("type")
	baremetalIPMIGetAccessCmd.Flags().StringVar(&baremetal.BaremetalIpmiIP, "allowed-ip", "", "IPv4 address that can use the access")
	baremetalIPMIGetAccessCmd.Flags().StringVar(&baremetal.BaremetalIpmiSshKey, "ssh-key", "", "Public SSH key for Serial Over Lan SSH access")
	baremetalIPMICmd.AddCommand(baremetalIPMIGetAccessCmd)

	rootCmd.AddCommand(baremetalCmd)
}
