package cmd

import (
	_ "embed"
	"runtime"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/ovh/ovhcloud-cli/internal/services/vps"
	"github.com/spf13/cobra"
)

func init() {
	vpsCmd := &cobra.Command{
		Use:   "vps",
		Short: "Retrieve information and manage your VPS services",
	}

	// Command to list VPS services
	vpsListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your VPS services",
		Run:     vps.ListVps,
	}
	vpsCmd.AddCommand(withFilterFlag(vpsListCmd))

	// Command to get a single VPS
	vpsCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.GetVps,
	})

	// Command to update a single VPS
	vpsEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.EditVps,
	}
	vpsEditCmd.Flags().StringVar(&vps.VpsSpec.DisplayName, "display-name", "", "Display name of the VPS")
	vpsEditCmd.Flags().StringVar(&vps.VpsSpec.Keymap, "keymap", "", "Keymap of the VPS (fr, us)")
	vpsEditCmd.Flags().StringVar(&vps.VpsSpec.NetbootMode, "netboot-mode", "", "Netboot mode of the VPS (local, rescue)")
	vpsEditCmd.Flags().BoolVar(&vps.VpsSpec.SlaMonitoring, "sla-monitoring", false, "Enable or disable SLA monitoring for the VPS")
	addInteractiveEditorFlag(vpsEditCmd)
	vpsCmd.AddCommand(vpsEditCmd)

	// Snapshot commands
	vpsSnapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage VPS snapshots",
	}
	vpsCmd.AddCommand(vpsSnapshotCmd)

	vpsSnapshotCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific VPS snapshot",
		Args:  cobra.ExactArgs(1),
		Run:   vps.GetVpsSnapshot,
	})

	vpsSnapshotCreateCmd := &cobra.Command{
		Use:   "create <service_name>",
		Short: "Create a snapshot of the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.CreateVpsSnapshot,
	}
	vpsSnapshotCreateCmd.Flags().StringVar(&vps.VpsSnapshotSpec.Description, "description", "", "Description of the snapshot")
	vpsSnapshotCmd.AddCommand(vpsSnapshotCreateCmd)

	vpsSnapshotEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given VPS snapshot",
		Args:  cobra.ExactArgs(1),
		Run:   vps.EditVpsSnapshot,
	}
	vpsSnapshotEditCmd.Flags().StringVar(&vps.VpsSnapshotSpec.Description, "description", "", "Description of the snapshot")
	addInteractiveEditorFlag(vpsSnapshotEditCmd)
	vpsSnapshotCmd.AddCommand(vpsSnapshotEditCmd)

	vpsSnapshotCmd.AddCommand(&cobra.Command{
		Use:   "delete <service_name>",
		Short: "Delete the given VPS snapshot",
		Args:  cobra.ExactArgs(1),
		Run:   vps.DeleteVpsSnapshot,
	})

	vpsSnapshotCmd.AddCommand(&cobra.Command{
		Use:   "abort <service_name>",
		Short: "Abort the creation of a VPS snapshot",
		Args:  cobra.ExactArgs(1),
		Run:   vps.AbortVpsSnapshot,
	})

	vpsSnapshotCmd.AddCommand(&cobra.Command{
		Use:   "restore <service_name>",
		Short: "Restore the snapshot of the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.RestoreVpsSnapshot,
	})

	vpsSnapshotCmd.AddCommand(&cobra.Command{
		Use:   "download <service_name>",
		Short: "Download the snapshot of the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.DownloadVpsSnapshot,
	})

	// Automated backup commands
	vpsBackupCmd := &cobra.Command{
		Use:   "automated-backup",
		Short: "Manage VPS automated backups",
	}
	vpsCmd.AddCommand(vpsBackupCmd)

	vpsBackupCmd.AddCommand(&cobra.Command{
		Use:   "get-config <service_name>",
		Short: "Retrieve automated backup configuration of the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.GetVpsAutomatedBackup,
	})

	vpsBackupCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_name>",
		Aliases: []string{"ls"},
		Short:   "List all automated backups of the given VPS",
		Args:    cobra.ExactArgs(1),
		Run:     vps.ListVpsAutomatedBackups,
	}))

	vpsBackupCmd.AddCommand(&cobra.Command{
		Use:     "reschedule <service_name> <time>",
		Example: "ovh-cli vps automated-backup reschedule my-vps 15:04:05",
		Short:   "Reschedule the automated backup of the given VPS",
		Args:    cobra.ExactArgs(2),
		Run:     vps.RescheduleVpsAutomatedBackup,
	})

	vpsBackupRestoreCmd := &cobra.Command{
		Use:   "restore <service_name>",
		Short: "Restore the automated backup of the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.RestoreVpsAutomatedBackup,
	}
	vpsBackupRestoreCmd.Flags().StringVar(&vps.VpsSnapshotRestoreSpec.RestorePoint, "restore-point", "", "Restore point to use for the restoration")
	vpsBackupRestoreCmd.Flags().BoolVar(&vps.VpsSnapshotRestoreSpec.ChangePassword, "change-password", false, "Change the password after restoration (only with restore full on VPS Cloud 2014)")
	vpsBackupRestoreCmd.Flags().StringVar(&vps.VpsSnapshotRestoreSpec.Type, "type", "file", "Type of restoration (file, full)")
	vpsBackupCmd.AddCommand(vpsBackupRestoreCmd)

	vpsBackupListRestorePointsCmd := &cobra.Command{
		Use:   "list-restore-points <service_name>",
		Short: "List all restore points of the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.ListVpsAutomatedBackupRestorePoints,
	}
	vpsBackupListRestorePointsCmd.Flags().StringVar(&vps.VpsBackupRestorePointsState, "state", "available", "State of the restore points to list (available, restored, restoring)")
	vpsBackupCmd.AddCommand(withFilterFlag(vpsBackupListRestorePointsCmd))

	// Commands to list available upgrades
	vpsCmd.AddCommand(&cobra.Command{
		Use:   "list-available-upgrades <service_name>",
		Short: "List available upgrades for your VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.ListVpsAvailableUpgrades,
	})

	// Contacts and billing commands
	vpsChangeContactCmd := &cobra.Command{
		Use:   "change-contacts <service_name>",
		Short: "Change contacts for the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.ChangeVpsContacts,
	}
	vpsChangeContactCmd.Flags().StringVar(&vps.VpsContacts.ContactAdmin, "contact-admin", "", "Contact admin for the VPS")
	vpsChangeContactCmd.Flags().StringVar(&vps.VpsContacts.ContactBilling, "contact-billing", "", "Contact billing for the VPS")
	vpsChangeContactCmd.Flags().StringVar(&vps.VpsContacts.ContactTech, "contact-tech", "", "Contact tech for the VPS")
	vpsCmd.AddCommand(vpsChangeContactCmd)

	serviceInfoCmd := &cobra.Command{
		Use:   "service-info",
		Short: "Manage service information for the given VPS",
	}
	vpsCmd.AddCommand(serviceInfoCmd)

	serviceInfoCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Get service information for the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.GetVpsServiceInfo,
	})

	serviceInfoEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit service information for the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.EditVpsServiceInfo,
	}
	serviceInfoEditCmd.Flags().BoolVar(&common.ServiceInfoSpec.Renew.Automatic, "renew-automatic", false, "Enable automatic renewal")
	serviceInfoEditCmd.Flags().BoolVar(&common.ServiceInfoSpec.Renew.DeleteAtExpiration, "renew-delete-at-expiration", false, "Delete service at expiration")
	serviceInfoEditCmd.Flags().BoolVar(&common.ServiceInfoSpec.Renew.Forced, "renew-forced", false, "Force renewal")
	serviceInfoEditCmd.Flags().BoolVar(&common.ServiceInfoSpec.Renew.ManualPayment, "renew-manual-payment", false, "Enable manual payment for renewal")
	serviceInfoEditCmd.Flags().IntVar(&common.ServiceInfoSpec.Renew.Period, "renew-period", 0, "Renewal period (in months)")
	addInteractiveEditorFlag(serviceInfoEditCmd)
	serviceInfoCmd.AddCommand(serviceInfoEditCmd)

	vpsCmd.AddCommand(&cobra.Command{
		Use:   "terminate <service_name>",
		Short: "Ask for termination of the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.TerminateVps,
	})

	vpsCmd.AddCommand(&cobra.Command{
		Use:   "confirm-termination <service_name> <token>",
		Short: "Confirm termination of the given VPS",
		Args:  cobra.ExactArgs(2),
		Run:   vps.ConfirmVpsTermination,
	})

	// Disks commands
	vpsDiskCmd := &cobra.Command{
		Use:   "disk",
		Short: "Manage disks of the given VPS",
	}
	vpsCmd.AddCommand(vpsDiskCmd)

	vpsDiskCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_name>",
		Aliases: []string{"ls"},
		Short:   "List disks of the given VPS",
		Args:    cobra.ExactArgs(1),
		Run:     vps.ListVpsDisks,
	}))

	vpsDiskCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name> <disk_id>",
		Short: "Get information about a specific disk of the given VPS",
		Args:  cobra.ExactArgs(2),
		Run:   vps.GetVpsDisk,
	})

	vpsDiskEditCmd := &cobra.Command{
		Use:   "edit <service_name> <disk_id>",
		Short: "Edit a specific disk of the given VPS",
		Args:  cobra.ExactArgs(2),
		Run:   vps.EditVpsDisk,
	}
	vpsDiskEditCmd.Flags().IntVar(&vps.VpsDiskSpec.LowFreeSpaceThreshold, "low-free-space-threshold", 0, "Low free space threshold for the disk")
	vpsDiskEditCmd.Flags().BoolVar(&vps.VpsDiskSpec.Monitoring, "monitoring", false, "Enable or disable monitoring for the disk")
	addInteractiveEditorFlag(vpsDiskEditCmd)
	vpsDiskCmd.AddCommand(vpsDiskEditCmd)

	// VNC Console command
	vpsCmd.AddCommand(&cobra.Command{
		Use:   "get-console-url <service_name>",
		Short: "Get the console URL for the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.VpsGetConsoleURL,
	})

	// Image commands
	vpsImageCmd := &cobra.Command{
		Use:   "image",
		Short: "Manage images of the given VPS",
	}
	vpsCmd.AddCommand(vpsImageCmd)

	vpsImageCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_name>",
		Aliases: []string{"ls"},
		Short:   "List images of the given VPS",
		Args:    cobra.ExactArgs(1),
		Run:     vps.GetVpsImages,
	}))

	// IPs commands
	vpsIPCmd := &cobra.Command{
		Use:   "ip",
		Short: "Manage IPs of the given VPS",
	}
	vpsCmd.AddCommand(vpsIPCmd)

	vpsIPCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_name>",
		Aliases: []string{"ls"},
		Short:   "List IPs of the given VPS",
		Args:    cobra.ExactArgs(1),
		Run:     vps.ListVpsIPs,
	}))

	vpsIPCmd.AddCommand(&cobra.Command{
		Use:   "set-reverse <service_name> <ip> <reverse>",
		Short: "Set reverse DNS for the given IP of the VPS",
		Args:  cobra.ExactArgs(3),
		Run:   vps.SetVpsIPReverse,
	})

	vpsIPCmd.AddCommand(&cobra.Command{
		Use:   "release <service_name> <ip>",
		Short: "Release the given IP of the VPS",
		Args:  cobra.ExactArgs(2),
		Run:   vps.ReleaseVpsIP,
	})

	// Options command
	vpsCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list-options <service_name>",
		Short: "List options of the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.ListVPSOptions,
	}))

	// Start, stop, reboot commands
	vpsStartCmd := &cobra.Command{
		Use:   "start <service_name>",
		Short: "Start the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.StartVps,
	}
	vpsStartCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for the start task to complete")
	vpsCmd.AddCommand(vpsStartCmd)

	vpsStopCmd := &cobra.Command{
		Use:   "stop <service_name>",
		Short: "Stop the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.StopVps,
	}
	vpsStopCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for the stop task to complete")
	vpsCmd.AddCommand(vpsStopCmd)

	vpsRebootCmd := &cobra.Command{
		Use:   "reboot <service_name>",
		Short: "Reboot the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.RebootVps,
	}
	vpsRebootCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for the reboot task to complete")
	vpsCmd.AddCommand(vpsRebootCmd)

	// Reinstall command
	vpsReinstallCmd := &cobra.Command{
		Use:   "reinstall <service_name>",
		Short: "Reinstall the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.ReinstallVps,
	}
	vpsReinstallCmd.Flags().BoolVar(&vps.VpsReinstallSpec.DoNotSendPassword, "do-not-send-password", false, "Do not send the new password after reinstallation (only if sshKey defined)")
	vpsReinstallCmd.Flags().StringVar(&vps.VpsReinstallSpec.ImageId, "image-id", "", "ID of the image to use for reinstallation")
	vpsReinstallCmd.Flags().BoolVar(&vps.VpsReinstallSpec.InstallRTM, "install-rtm", false, "Install RTM during reinstallation")
	vpsReinstallCmd.Flags().StringVar(&vps.VpsReinstallSpec.PublicSshKey, "public-ssh-key", "", "Public SSH key to pre-install on your VPS")
	vpsReinstallCmd.Flags().StringVar(&vps.VpsReinstallSpec.SshKey, "ssh-key", "", "SSH key name to pre-install on your VPS (name can be found running `ovh-cli account ssh-key list`)")
	addInitParameterFileFlag(vpsReinstallCmd, assets.VpsOpenapiSchema, "/vps/{serviceName}/rebuild", "post", vps.VpsReinstallExample, nil)
	addInteractiveEditorFlag(vpsReinstallCmd)
	addFromFileFlag(vpsReinstallCmd)
	if !(runtime.GOARCH == "wasm" && runtime.GOOS == "js") {
		vpsReinstallCmd.Flags().BoolVar(&vps.VpsImageViaInteractiveSelector, "image-selector", false, "Use the interactive image selector")
		vpsReinstallCmd.Flags().BoolVar(&vps.VpsSSHKeyViaInteractiveSelector, "ssh-key-selector", false, "Use the interactive SSH key selector")
		vpsReinstallCmd.MarkFlagsMutuallyExclusive("from-file", "editor")
	}
	vpsReinstallCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for reinstall to be done before exiting")
	removeRootFlagsFromCommand(vpsReinstallCmd)
	vpsCmd.AddCommand(vpsReinstallCmd)

	// Secondary DNS Domains commands
	vpsSecondaryDNSCmd := &cobra.Command{
		Use:   "secondary-dns-domain",
		Short: "Manage secondary DNS domains of the given VPS",
	}
	vpsCmd.AddCommand(vpsSecondaryDNSCmd)

	vpsSecondaryDNSCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_name>",
		Aliases: []string{"ls"},
		Short:   "List secondary DNS domains of the given VPS",
		Args:    cobra.ExactArgs(1),
		Run:     vps.ListVpsSecondaryDNSDomains,
	}))

	vpsSecondaryDnsAddCmd := &cobra.Command{
		Use:   "add <service_name>",
		Short: "Add a secondary DNS domain to the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.AddVpsSecondaryDNSDomain,
	}
	vpsSecondaryDnsAddCmd.Flags().StringVar(&vps.VpsSecondaryDNSDomainSpec.Domain, "domain", "", "Domain name for the secondary DNS")
	vpsSecondaryDnsAddCmd.Flags().StringVar(&vps.VpsSecondaryDNSDomainSpec.IP, "ip", "", "IP address for the secondary DNS")
	addInteractiveEditorFlag(vpsSecondaryDnsAddCmd)
	vpsSecondaryDNSCmd.AddCommand(vpsSecondaryDnsAddCmd)

	vpsSecondaryDnsEditCmd := &cobra.Command{
		Use:   "edit <service_name> <domain>",
		Short: "Edit a secondary DNS domain of the given VPS",
		Args:  cobra.ExactArgs(2),
		Run:   vps.EditVpsSecondaryDNSDomain,
	}
	vpsSecondaryDnsEditCmd.Flags().StringVar(&vps.VpsSecondaryDNSDomainSpec.IPMaster, "ip-master", "", "IP address of the master server for the secondary DNS")
	addInteractiveEditorFlag(vpsSecondaryDnsEditCmd)
	vpsSecondaryDNSCmd.AddCommand(vpsSecondaryDnsEditCmd)

	vpsSecondaryDNSCmd.AddCommand(&cobra.Command{
		Use:   "delete <service_name> <domain>",
		Short: "Remove a secondary DNS domain from the given VPS",
		Args:  cobra.ExactArgs(2),
		Run:   vps.DeleteVpsSecondaryDNSDomain,
	})

	// Set password command
	vpsSetPasswordCmd := &cobra.Command{
		Use:   "set-password <service_name>",
		Short: "Start the process in order to set the root password of the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.ChangeVpsPassword,
	}
	vpsSetPasswordCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for the task to complete before exiting")
	vpsCmd.AddCommand(vpsSetPasswordCmd)

	// Tasks command
	vpsCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list-tasks <service_name>",
		Short: "List tasks of the given VPS",
		Args:  cobra.ExactArgs(1),
		Run:   vps.ListVpsTasks,
	}))

	rootCmd.AddCommand(vpsCmd)
}
