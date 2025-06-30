package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
)

func getInstanceCreationCmd() *cobra.Command {
	instanceCreateCmd := &cobra.Command{
		Use:   "create <region (e.g. GRA, BHS, SBG)>",
		Short: "Create a new instance",
		Long: `Use this command to create an instance in the given public cloud project.
There are three ways to define the creation parameters:

1. Using only CLI flags:

  ovhcloud cloud instance create GRA9 --name MyNewInstance --boot-from.image <image_id> --flavor <flavor_id> ...

2. Using a configuration file:

  First you can generate an example of installation file using the following command:

	ovhcloud cloud instance create BHS5 --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud instance create BHS5 --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud instance create BHS5

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud instance create GRA11 --from-file ./params.json --name NameOverriden

  It is also possible to use the interactive image and flavor selector to define the image and flavor parameters, like the following:

  	ovhcloud cloud instance create BHS5 --init-file ./params.json --image-selector --flavor-selector

3. Using your default text editor:

  ovhcloud cloud instance create GRA11 --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud instance create RBX8 --editor --flavor <flavor_id>

  You can also use the interactive image and flavor selector to define the image and flavor parameters, like the following:

  	ovhcloud cloud instance create RBX8 --editor --image-selector --flavor-selector
`,
		Run:  cloud.CreateInstance,
		Args: cobra.MaximumNArgs(1),
	}

	// Add flags for instance creation parameters
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.AvailabilityZone, "availability-zone", "", "Availability zone")
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.BillingPeriod, "billing-period", "hourly", "Billing period (hourly, monthly), default is hourly")
	instanceCreateCmd.Flags().IntVar(&cloud.InstanceCreationParameters.Bulk, "bulk", 0, "Number of instances to create")
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.Flavor.ID, "flavor", "", "Flavor ID")
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.Group.ID, "group", "", "Group ID")
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.Name, "name", "", "Instance name")

	// Boot options
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.BootFrom.ImageID, "boot-from.image", "", "Image ID to boot from")
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.BootFrom.VolumeID, "boot-from.volume", "", "Volume ID to boot from")
	instanceCreateCmd.MarkFlagsMutuallyExclusive("boot-from.image", "boot-from.volume")

	// Private Network - Floating IP
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.Network.Private.FloatingIp.ID, "network.private.floating-ip.id", "", "ID of an existing floating IP")
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.Network.Private.FloatingIpCreate.Description, "network.private.floating-ip.create.description", "", "Description for the floating IP to create")
	instanceCreateCmd.MarkFlagsMutuallyExclusive("network.private.floating-ip.id", "network.private.floating-ip.create.description")

	// Private Network - Gateway
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.Network.Private.Gateway.ID, "network.private.gateway.id", "", "ID of the existing gateway to attach to the private network")
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.Network.Private.GatewayCreate.Model, "network.private.gateway.create.model", "", "Model for the gateway to create (s, m, l)")
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.Network.Private.GatewayCreate.Name, "network.private.gateway.create.name", "", "Name for the gateway to create")
	instanceCreateCmd.MarkFlagsMutuallyExclusive("network.private.gateway.id", "network.private.gateway.create.model")
	instanceCreateCmd.MarkFlagsMutuallyExclusive("network.private.gateway.id", "network.private.gateway.create.name")

	// Private network - IP
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.Network.Private.IP, "network.private.ip", "", "Instance IP in the private network")

	// Private Network - Existing network information
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.Network.Private.Network.ID, "network.private.id", "", "ID of the existing private network")
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.Network.Private.Network.SubnetID, "network.private.subnet-id", "", "Existing subnet ID")

	// Private Network - Create network
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.Network.Private.NetworkCreate.Name, "network.private.create.name", "", "Name for the private network to create")
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.Network.Private.NetworkCreate.Subnet.CIDR, "network.private.create.subnet-cidr", "", "CIDR for the subnet to create")
	instanceCreateCmd.Flags().BoolVar(&cloud.InstanceCreationParameters.Network.Private.NetworkCreate.Subnet.EnableDhcp, "network.private.create.subnet-enable-dhcp", false, "Enable DHCP for the subnet to create")
	instanceCreateCmd.Flags().IntVar(&cloud.InstanceCreationParameters.Network.Private.NetworkCreate.Subnet.IPVersion, "network.private.create.subnet-ip-version", 0, "IP version for the subnet to create")
	instanceCreateCmd.Flags().IntVar(&cloud.InstanceCreationParameters.Network.Private.NetworkCreate.VlanID, "network.private.create.vlan-id", 0, "VLAN ID for the private network to create")
	instanceCreateCmd.MarkFlagsMutuallyExclusive("network.private.id", "network.private.create.name")

	// Network - Public
	instanceCreateCmd.Flags().BoolVar(&cloud.InstanceCreationParameters.Network.Public, "network.public", false, "Set the new instance as public")
	instanceCreateCmd.MarkFlagsMutuallyExclusive("network.private.id", "network.public")
	instanceCreateCmd.MarkFlagsMutuallyExclusive("network.private.create.name", "network.public")

	// Autobackup
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.Autobackup.Cron, "backup-cron", "", "Autobackup Unix Cron pattern (eg: '0 0 * * *')")
	instanceCreateCmd.Flags().IntVar(&cloud.InstanceCreationParameters.Autobackup.Rotation, "backup-rotation", 0, "Number of backups to keep")

	// SSH Key
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.SshKey.Name, "ssh-key.name", "", "Existing SSH key name")

	// SSH Key Creation
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.SshKeyCreate.Name, "ssh-key.create.name", "", "Name for the SSH key to create")
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.SshKeyCreate.PublicKey, "ssh-key.create.public-key", "", "Public key for the SSH key to create")
	instanceCreateCmd.MarkFlagsMutuallyExclusive("ssh-key.name", "ssh-key.create.name")
	instanceCreateCmd.MarkFlagsMutuallyExclusive("ssh-key.name", "ssh-key.create.public-key")

	// User Data
	instanceCreateCmd.Flags().StringVar(&cloud.InstanceCreationParameters.UserData, "user-data", "", "Configuration information or scripts to use upon launch")

	// Common flags for other mean to define parameters
	addInitParameterFileFlag(instanceCreateCmd, cloud.CloudOpenapiSchema, "/cloud/project/{serviceName}/instance", "post", cloud.CloudInstanceCreationExample, cloud.GetInstanceFlavorAndImageInteractiveSelector)
	instanceCreateCmd.Flags().StringVar(&flags.ParametersFile, "from-file", "", "File containing creation parameters")
	instanceCreateCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define creation parameters")
	instanceCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for instance creation to be done before exiting")
	instanceCreateCmd.Flags().BoolVar(&cloud.InstanceImageViaInteractiveSelector, "image-selector", false, "Use the interactive image selector")
	instanceCreateCmd.Flags().BoolVar(&cloud.InstanceFlavorViaInteractiveSelector, "flavor-selector", false, "Use the interactive flavor selector")

	removeRootFlagsFromCommand(instanceCreateCmd)
	instanceCreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return instanceCreateCmd
}

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

	instanceCmd.AddCommand(getInstanceCreationCmd())

	instanceCmd.AddCommand(&cobra.Command{
		Use:   "delete <instance_id>",
		Short: "Delete the given instance",
		Run:   cloud.DeleteInstance,
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

3. Using a configuration file:

  First you can generate an example of installation file using the following command:

	ovhcloud cloud instance reinstall --init-file ./install.json

  You will be able to choose from several installation examples. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct installation parameters, run:

	ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --from-file ./install.json

  Note that you can also pipe the content of the file to reinstall, like the following:

	cat ./install.json | ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --from-file ./install.json --image <image_id>

4. Using your default text editor:

  ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --editor

  You will be able to choose from several installation examples. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the reinstallation will be run.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --editor --image <image_id>
`,
		Run:  cloud.ReinstallInstance,
		Args: cobra.MaximumNArgs(1),
	}
	addInitParameterFileFlag(reinstallCmd, cloud.CloudOpenapiSchema, "/cloud/project/{serviceName}/instance/{instanceId}/reinstall", "post", "", nil)
	reinstallCmd.Flags().StringVar(&flags.ParametersFile, "from-file", "", "File containing installation parameters")
	reinstallCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define installation parameters")
	reinstallCmd.Flags().BoolVar(&cloud.InstanceImageViaInteractiveSelector, "image-selector", false, "Use the interactive image selector to define installation parameters")
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

	snapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage snapshots of the given instance",
	}
	instanceCmd.AddCommand(snapshotCmd)

	snapshotCmd.AddCommand(&cobra.Command{
		Use:   "create <instance_id> <snapshot_name>",
		Short: "Create a snapshot of the given instance",
		Run:   cloud.CreateInstanceSnapshot,
		Args:  cobra.ExactArgs(2),
	})

	snapshotCmd.AddCommand(&cobra.Command{
		Use:   "abort <instance_id>",
		Short: "Abort the snapshot creation of the given instance",
		Run:   cloud.AbortInstanceSnapshot,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(instanceCmd)
}
