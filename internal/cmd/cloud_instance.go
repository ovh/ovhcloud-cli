package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
)

func initInstanceCommand(cloudCmd *cobra.Command) {
	instanceCmd := &cobra.Command{
		Use:   "instance",
		Short: "Manage instances in the given cloud project",
	}
	instanceCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	instanceListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your instances",
		Run:   cloud.ListInstances,
	}
	instanceCmd.AddCommand(withFilterFlag(instanceListCmd))

	instanceCmd.AddCommand(&cobra.Command{
		Use:   "get <instance_id>",
		Short: "Get a specific instance",
		Run:   cloud.GetInstance,
		Args:  cobra.ExactArgs(1),
	})

	instanceCmd.AddCommand(&cobra.Command{
		Use:   "start <instance_id>",
		Short: "Start the given instance",
		Run:   cloud.StartInstance,
		Args:  cobra.ExactArgs(1),
	})

	instanceCmd.AddCommand(&cobra.Command{
		Use:   "stop <instance_id>",
		Short: "Stop the given instance",
		Run:   cloud.StopInstance,
		Args:  cobra.ExactArgs(1),
	})

	instanceCmd.AddCommand(&cobra.Command{
		Use:   "shelve <instance_id>",
		Short: "Shelve the given instance",
		Long: `The resources dedicated to the Public Cloud instance are released.
The data of the local storage will be stored, the duration of the operation depends on the size of the local disk.
The instance can be unshelved at any time. Meanwhile hourly instances will not be billed.
The Snapshot Storage used to store the instance's data will be billed.`,
		Run:  cloud.ShelveInstance,
		Args: cobra.ExactArgs(1),
	})

	instanceCmd.AddCommand(&cobra.Command{
		Use:   "unshelve <instance_id>",
		Short: "Unshelve the given instance",
		Long: `The resources dedicated to the Public Cloud instance are restored.
The duration of the operation depends on the size of the local disk.
Instance billing will get back to normal and the snapshot used to store the instance's data will be deleted.`,
		Run:  cloud.UnshelveInstance,
		Args: cobra.ExactArgs(1),
	})

	instanceCmd.AddCommand(&cobra.Command{
		Use:   "resume <instance_id>",
		Short: "Resume the given suspended instance",
		Run:   cloud.ResumeInstance,
		Args:  cobra.ExactArgs(1),
	})

	rebootCmd := &cobra.Command{
		Use:   "reboot <instance_id>",
		Short: "Reboot the given instance",
		Run:   cloud.RebootInstance,
		Args:  cobra.ExactArgs(1),
	}
	rebootCmd.Flags().StringVarP(&cloud.InstanceRebootType, "type", "t", "soft", "Reboot type: hard or soft (default is soft)")
	instanceCmd.AddCommand(rebootCmd)

	reinstallCmd := &cobra.Command{
		Use:   "reinstall <instance_id>",
		Short: "Reinstall the given instance",
		Long: `Use this command to reinstall the given instance.

There are three ways to define the installation parameters:
(the following examples assume that you have already configured your default cloud project using "ovhcloud config set default_cloud_project <project_id>")

1. Using only CLI flags:

  ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --image <image_id>

2. Using the interactive image selector:

  ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --image-selector

2. Using a configuration file

  First you can generate an example of installation file using the following command:

	ovhcloud cloud instance reinstall --init-file ./install.json

  You will be able to choose from several installation examples. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct installation parameters, run:

	ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --from-file ./install.json

  Note that you can also pipe the content of the file to reinstall, like the following:

	cat ./install.json | ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --from-file ./install.json --image <image_id>

3. Using your default text editor

  ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --editor

  You will be able to choose from several installation examples. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the reinstallation will be run.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --editor --image <image_id>
`,
		Run:  cloud.ReinstallInstance,
		Args: cobra.MaximumNArgs(1),
	}
	addInitParameterFileFlag(reinstallCmd, cloud.CloudOpenapiSchema, "/cloud/project/{serviceName}/instance/{instanceId}/reinstall", "post", "")
	reinstallCmd.Flags().StringVar(&cloud.InstanceInstallationFile, "from-file", "", "File containing installation parameters")
	reinstallCmd.Flags().BoolVar(&cloud.InstanceInstallViaEditor, "editor", false, "Use a text editor to define installation parameters")
	reinstallCmd.Flags().BoolVar(&cloud.InstanceInstallationViaInteractiveSelector, "image-selector", false, "Use the interactive image selector to define installation parameters")
	reinstallCmd.Flags().StringVar(&cloud.InstanceImageID, "image", "", "Image to use for reinstallation")
	reinstallCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for reinstall to be done before exiting")
	removeRootFlagsFromCommand(reinstallCmd)
	reinstallCmd.MarkFlagsMutuallyExclusive("from-file", "editor", "image-selector")
	instanceCmd.AddCommand(reinstallCmd)

	instanceCmd.AddCommand(&cobra.Command{
		Use:   "activate-monthly-billing <instance_id>",
		Short: "Activate monthly billing for the given instance",
		Run:   cloud.ActivateMonthlyBilling,
		Args:  cobra.ExactArgs(1),
	})

	interfacesCommand := &cobra.Command{
		Use:   "interface",
		Short: "Manage interfaces of the given instance",
	}
	instanceCmd.AddCommand(interfacesCommand)

	interfacesCommand.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list <instance_id>",
		Short: "List interfaces of the given instance",
		Run:   cloud.ListInstanceInterfaces,
		Args:  cobra.ExactArgs(1),
	}))

	interfacesCommand.AddCommand(&cobra.Command{
		Use:   "get <instance_id> <interface_id>",
		Short: "Get a specific interface of the given instance",
		Run:   cloud.GetInstanceInterface,
		Args:  cobra.ExactArgs(2),
	})

	interfacesCommand.AddCommand(&cobra.Command{
		Use:   "create <instance_id> <network_id> <ip (optional)>",
		Short: "Create interface on the given instance and attach it to a network",
		Run:   cloud.CreateInstanceInterface,
		Args:  cobra.RangeArgs(2, 3),
	})

	interfacesCommand.AddCommand(&cobra.Command{
		Use:   "delete <instance_id> <interface_id>",
		Short: "Delete a specific interface of the given instance",
		Run:   cloud.DeleteInstanceInterface,
		Args:  cobra.ExactArgs(2),
	})

	enableRescueCmd := &cobra.Command{
		Use:   "reboot-rescue <instance_id>",
		Short: "Reboot the given instance in rescue mode",
		Run:   cloud.EnableInstanceInRescueMode,
		Args:  cobra.ExactArgs(1),
	}
	enableRescueCmd.Flags().StringVar(&cloud.InstanceImageID, "image", "", "Image to boot from")
	enableRescueCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for instance to be in rescue mode before exiting")
	instanceCmd.AddCommand(enableRescueCmd)

	disableRescueCmd := &cobra.Command{
		Use:   "exit-rescue <instance_id>",
		Short: "Exit the given instance from rescue mode",
		Run:   cloud.DisableInstanceRescueMode,
		Args:  cobra.ExactArgs(1),
	}
	disableRescueCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for instance to have exited rescue mode before exiting")
	instanceCmd.AddCommand(disableRescueCmd)

	setFlavorCmd := &cobra.Command{
		Use:   "set-flavor <instance_id> <flavor_id>",
		Short: "Migrate the given instance to the specified flavor",
		Run:   cloud.SetInstanceFlavor,
		Args:  cobra.RangeArgs(1, 2),
	}
	setFlavorCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for instance to run with the desired flavor before exiting")
	setFlavorCmd.Flags().BoolVar(&cloud.InstanceFlavorViaInteractiveSelector, "flavor-selector", false, "Use the interactive flavor selector")
	instanceCmd.AddCommand(setFlavorCmd)

	cloudCmd.AddCommand(instanceCmd)
}
