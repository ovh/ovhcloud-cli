// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initManagedDatabaseCommand(cloudCmd *cobra.Command) {
	managedDatabaseCmd := &cobra.Command{
		Use:   "managed-database",
		Short: "Manage managed database services in the given cloud project",
	}
	managedDatabaseCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	// Managed database commands
	managedDatabaseCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your managed database services",
		Run:     cloud.ListManagedDatabases,
	}))

	managedDatabaseCmd.AddCommand(&cobra.Command{
		Use:   "get <service_id>",
		Short: "Get a specific managed database service",
		Run:   cloud.GetManagedDatabase,
		Args:  cobra.ExactArgs(1),
	})

	managedDatabaseCmd.AddCommand(managedDatabaseCreationCmd())
	managedDatabaseCmd.AddCommand(managedDatabaseEditCmd())

	managedDatabaseCmd.AddCommand(&cobra.Command{
		Use:   "delete <service_id>",
		Short: "Delete a specific managed database service",
		Run:   cloud.DeleteManagedDatabase,
		Args:  cobra.ExactArgs(1),
	})

	initManagedDatabaseDatabaseCommand(managedDatabaseCmd)
	initManagedDatabaseUserCommand(managedDatabaseCmd)
	initManagedDatabaseRoleCommand(managedDatabaseCmd)
	initManagedDatabaseCertificateCommand(managedDatabaseCmd)
	initManagedDatabaseBackupCommand(managedDatabaseCmd)

	cloudCmd.AddCommand(managedDatabaseCmd)
}

func initManagedDatabaseDatabaseCommand(managedDatabaseCmd *cobra.Command) {
	databaseCmd := &cobra.Command{
		Use:   "database",
		Short: "Manage databases in a specific managed database service",
	}

	databaseCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_id>",
		Aliases: []string{"ls"},
		Short:   "List all databases in the given managed database service",
		Run:     cloud.ListManagedDatabaseDatabases,
		Args:    cobra.ExactArgs(1),
	}))

	databaseCmd.AddCommand(&cobra.Command{
		Use:   "get <service_id> <database_id>",
		Short: "Get a specific database in the given managed database service",
		Run:   cloud.GetManagedDatabaseDatabase,
		Args:  cobra.ExactArgs(2),
	})

	databaseCmd.AddCommand(managedDatabaseDatabaseCreateCmd())

	databaseCmd.AddCommand(&cobra.Command{
		Use:   "delete <service_id> <database_id>",
		Short: "Delete a specific database in the given managed database service",
		Run:   cloud.DeleteManagedDatabaseDatabase,
		Args:  cobra.ExactArgs(2),
	})

	managedDatabaseCmd.AddCommand(databaseCmd)
}

func initManagedDatabaseUserCommand(managedDatabaseCmd *cobra.Command) {
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users in a specific managed database service",
	}

	userCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_id>",
		Aliases: []string{"ls"},
		Short:   "List all users in the given managed database service",
		Run:     cloud.ListManagedDatabaseUsers,
		Args:    cobra.ExactArgs(1),
	}))

	userCmd.AddCommand(&cobra.Command{
		Use:   "get <service_id> <user_id>",
		Short: "Get a specific user in the given managed database service",
		Run:   cloud.GetManagedDatabaseUser,
		Args:  cobra.ExactArgs(2),
	})

	userCmd.AddCommand(managedDatabaseUserCreateCmd())
	userCmd.AddCommand(managedDatabaseUserEditCmd())

	userCmd.AddCommand(&cobra.Command{
		Use:   "delete <service_id> <user_id>",
		Short: "Delete a specific user in the given managed database service",
		Run:   cloud.DeleteManagedDatabaseUser,
		Args:  cobra.ExactArgs(2),
	})

	userCmd.AddCommand(&cobra.Command{
		Use:   "credentials-reset <service_id> <user_id>",
		Short: "Reset the credentials of a specific user in the given managed database service",
		Run:   cloud.ResetManagedDatabaseUserCredentials,
		Args:  cobra.ExactArgs(2),
	})

	managedDatabaseCmd.AddCommand(userCmd)
}

func initManagedDatabaseRoleCommand(managedDatabaseCmd *cobra.Command) {
	roleCmd := &cobra.Command{
		Use:   "role",
		Short: "Manage roles in a specific managed database service",
	}

	roleCmd.AddCommand(withFilterFlag(managedDatabaseRoleListCmd()))

	managedDatabaseCmd.AddCommand(roleCmd)
}

func initManagedDatabaseCertificateCommand(managedDatabaseCmd *cobra.Command) {
	certificateCmd := &cobra.Command{
		Use:   "certificate",
		Short: "Manage certificates in a specific managed database service",
	}

	certificateCmd.AddCommand(&cobra.Command{
		Use:   "get <service_id>",
		Short: "Get certificates in the given managed database service",
		Run:   cloud.GetManagedDatabaseCertificates,
		Args:  cobra.ExactArgs(1),
	})

	managedDatabaseCmd.AddCommand(certificateCmd)
}

func initManagedDatabaseBackupCommand(managedDatabaseCmd *cobra.Command) {
	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage backups in a specific managed database service",
	}

	backupCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_id>",
		Aliases: []string{"ls"},
		Short:   "List all backups in the given managed database service",
		Run:     cloud.ListManagedDatabaseBackups,
		Args:    cobra.ExactArgs(1),
	}))

	backupCmd.AddCommand(&cobra.Command{
		Use:   "get <service_id> <backup_id>",
		Short: "Get a specific backup in the given managed database service",
		Run:   cloud.GetManagedDatabaseBackup,
		Args:  cobra.ExactArgs(2),
	})

	backupCmd.AddCommand(&cobra.Command{
		Use:   "restore <service_id> <backup_id>",
		Short: "Restore a specific backup in the given managed database service",
		Run:   cloud.RestoreManagedDatabaseBackup,
		Args:  cobra.ExactArgs(2),
	})

	managedDatabaseCmd.AddCommand(backupCmd)
}

func managedDatabaseCreationCmd() *cobra.Command {
	managedDatabaseCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new managed database service",
		Long: `Use this command to create a managed database service in the given public cloud project.
There are two ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud managed-database create --engine mysql --version 8 --plan essential --nodes-pattern.flavor db1-4 --nodes-pattern.region DE

2. Using your default text editor:

	ovhcloud cloud managed-database create --engine mysql --editor
  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-database create --engine mysql --editor --version 8
`,
		PreRunE: cloud.CreateManagedDatabasePreRun,
		Run:     cloud.CreateManagedDatabase,
		Args:    cobra.NoArgs,
	}

	// Database details
	managedDatabaseCreateCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.Engine, "engine", "", "Database engine (you can get the list of available engines using 'ovhcloud cloud reference managed-database list-engines')")
	managedDatabaseCreateCmd.MarkFlagRequired("engine")
	managedDatabaseCreateCmd.RegisterFlagCompletionFunc("engine", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return cloud.ManagedDatabaseValidEngines, cobra.ShellCompDirectiveNoFileComp
	})
	managedDatabaseCreateCmd.Flags().StringSliceVar(&cloud.ManagedDatabaseSpec.Backups.Regions, "backups-regions", nil, "Regions on which the backups are stored")
	managedDatabaseCreateCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.Backups.Time, "backups-time", "", "Time on which backups start every day")
	managedDatabaseCreateCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.Description, "description", "", "Database description")
	managedDatabaseCreateCmd.Flags().IntVar(&cloud.ManagedDatabaseSpec.Disk.Size, "disk-size", 0, "Disk size (GB)")
	managedDatabaseCreateCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.ForkFrom.BackupID, "fork-from.backup-id", "", "Backup ID (not compatible with fork-from.point-in-time)")
	managedDatabaseCreateCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.ForkFrom.PointInTime, "fork-from.point-in-time", "", "Point in time to restore from (not compatible with fork-from.backup-id)")
	managedDatabaseCreateCmd.MarkFlagsMutuallyExclusive("fork-from.backup-id", "fork-from.point-in-time")
	managedDatabaseCreateCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.ForkFrom.ServiceID, "fork-from.service-id", "", "Service ID that owns the backups")
	managedDatabaseCreateCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.MaintenanceTime, "maintenance-time", "", "Time on which maintenances can start every day")
	managedDatabaseCreateCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.Plan, "plan", "", "Database plan (you can get the list of available plans using 'ovhcloud cloud reference managed-database list-plans')")
	managedDatabaseCreateCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.Version, "version", "", "Database version (you can get the list of available versions using 'ovhcloud cloud reference managed-database list-engines')")

	// Network configuration
	managedDatabaseCreateCmd.Flags().StringSliceVar(&cloud.ManagedDatabaseSpec.CLIIPRestrictions, "ip-restrictions", nil, "IP blocks authorized to access the cluster (CIDR format)")
	managedDatabaseCreateCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.NetworkID, "network-id", "", "Private network ID in which the cluster is deployed")
	managedDatabaseCreateCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.SubnetID, "subnet-id", "", "Private subnet ID in which the cluster is deployed")

	// Nodes pattern definition
	managedDatabaseCreateCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.NodesPattern.Flavor, "nodes-pattern.flavor", "", "Flavor of all nodes")
	managedDatabaseCreateCmd.Flags().IntVar(&cloud.ManagedDatabaseSpec.NodesPattern.Number, "nodes-pattern.number", 0, "Number of nodes")
	managedDatabaseCreateCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.NodesPattern.Region, "nodes-pattern.region", "", "Region of all nodes")

	// Nodes list definition
	managedDatabaseCreateCmd.Flags().StringSliceVar(&cloud.ManagedDatabaseSpec.CLINodesList, "nodes-list", nil, "List of nodes (format: flavor1:region1,flavor2:region2...)")
	managedDatabaseCreateCmd.MarkFlagsMutuallyExclusive("nodes-pattern.flavor", "nodes-list")
	managedDatabaseCreateCmd.MarkFlagsMutuallyExclusive("nodes-pattern.number", "nodes-list")
	managedDatabaseCreateCmd.MarkFlagsMutuallyExclusive("nodes-pattern.region", "nodes-list")

	// Common flags for other mean to define parameters
	addInteractiveEditorFlag(managedDatabaseCreateCmd)
	return managedDatabaseCreateCmd
}

func managedDatabaseEditCmd() *cobra.Command {
	managedDatabaseEditCmd := &cobra.Command{
		Use:   "edit <service_id>",
		Short: "Edit a specific managed database service",
		Long: `Use this command to edit a managed database service in the given public cloud project.
There are two ways to define the edition parameters:

1. Using only CLI flags:

	ovhcloud cloud managed-database edit <service_id> --description "My database"

2. Using your default text editor:

	ovhcloud cloud managed-database edit <service_id> --editor
  The CLI will open your default text editor to update the parameters. When saving the file, the edition will be applied.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-database edit <service_id> --editor --description "My database cluster"
`,
		Run:  cloud.EditManagedDatabase,
		Args: cobra.ExactArgs(1),
	}

	// Database details
	managedDatabaseEditCmd.Flags().StringSliceVar(&cloud.ManagedDatabaseSpec.Backups.Regions, "backups-regions", nil, "Regions on which the backups are stored")
	managedDatabaseEditCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.Backups.Time, "backups-time", "", "Time on which backups start every day")
	managedDatabaseEditCmd.Flags().BoolVar(&cloud.ManagedDatabaseSpec.DeletionProtection, "deletion-protection", false, "Enable deletion protection")
	managedDatabaseEditCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.Description, "description", "", "Description of the cluster")
	managedDatabaseEditCmd.Flags().BoolVar(&cloud.ManagedDatabaseSpec.EnablePrometheus, "enable-prometheus", false, "Enable Prometheus")
	managedDatabaseEditCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.NodesPattern.Flavor, "flavor", "", "The VM flavor used for this cluster")
	managedDatabaseEditCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.MaintenanceTime, "maintenance-time", "", "Time on which maintenances can start every day")
	managedDatabaseEditCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.Plan, "plan", "", "Plan of the cluster")
	managedDatabaseEditCmd.Flags().StringVar(&cloud.ManagedDatabaseSpec.Version, "version", "", "Version of the engine deployed on the cluster")

	// Network configuration
	managedDatabaseEditCmd.Flags().StringSliceVar(&cloud.ManagedDatabaseSpec.CLIIPRestrictions, "ip-restrictions", nil, "IP blocks authorized to access the cluster (CIDR format)")

	// Common flags for other mean to define parameters
	addInteractiveEditorFlag(managedDatabaseEditCmd)

	return managedDatabaseEditCmd
}

func managedDatabaseDatabaseCreateCmd() *cobra.Command {
	databaseCreateCmd := &cobra.Command{
		Use:   "create <service_id>",
		Short: "Create a new database in the given managed database service",
		Long: `Use this command to create a database in the given managed database service.

	ovhcloud cloud managed-database database create <service_id> --name mydb
`,
		Run:  cloud.CreateManagedDatabaseDatabase,
		Args: cobra.ExactArgs(1),
	}

	databaseCreateCmd.Flags().StringVar(&cloud.ManagedDatabaseDatabaseSpec.Name, "name", "", "Name of the database to create")
	databaseCreateCmd.MarkFlagRequired("name")

	return databaseCreateCmd
}

func managedDatabaseUserCreateCmd() *cobra.Command {
	userCreateCmd := &cobra.Command{
		Use:   "create <service_id>",
		Short: "Create a new user in the given managed database service",
		Long: `Use this command to create a user in the given managed database service.
There are two ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud managed-database user create <service_id> --name myuser

  For PostgreSQL and MongoDB engines, you can also specify roles:

	ovhcloud cloud managed-database user create <service_id> --name myuser --roles role1,role2

  For Valkey engine, you can specify permissions:

	ovhcloud cloud managed-database user create <service_id> --name myuser --categories "+@read" --commands "+get"

2. Using your default text editor:

	ovhcloud cloud managed-database user create <service_id> --name myuser --editor
  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-database user create <service_id> --name myuser --editor --roles role1,role2
`,
		Run:  cloud.CreateManagedDatabaseUser,
		Args: cobra.ExactArgs(1),
	}

	userCreateCmd.Flags().StringVar(&cloud.ManagedDatabaseUserSpec.Name, "name", "", "Name of the user to create")
	userCreateCmd.MarkFlagRequired("name")

	// PostgreSQL and MongoDB specific flags
	userCreateCmd.Flags().StringSliceVar(&cloud.ManagedDatabaseUserSpec.Roles, "roles", nil, "Roles granted to the user (postgresql and mongodb only)")

	// Valkey specific flags
	userCreateCmd.Flags().StringSliceVar(&cloud.ManagedDatabaseUserSpec.Categories, "categories", nil, "Command categories the user can execute (valkey only)")
	userCreateCmd.Flags().StringSliceVar(&cloud.ManagedDatabaseUserSpec.Channels, "channels", nil, "Channels the user can subscribe to (valkey only)")
	userCreateCmd.Flags().StringSliceVar(&cloud.ManagedDatabaseUserSpec.Commands, "commands", nil, "Commands the user can execute (valkey only)")
	userCreateCmd.Flags().StringSliceVar(&cloud.ManagedDatabaseUserSpec.Keys, "keys", nil, "Keys the user can access (valkey only)")

	addInteractiveEditorFlag(userCreateCmd)

	return userCreateCmd
}

func managedDatabaseUserEditCmd() *cobra.Command {
	userEditCmd := &cobra.Command{
		Use:   "edit <service_id> <user_id>",
		Short: "Edit a user in the given managed database service",
		Long: `Use this command to edit a user in the given managed database service.
There are two ways to define the edition parameters:

1. Using only CLI flags:

  For PostgreSQL and MongoDB engines:

	ovhcloud cloud managed-database user edit <service_id> <user_id> --roles role1,role2

  For Valkey engine:

	ovhcloud cloud managed-database user edit <service_id> <user_id> --categories "+@read" --commands "+get"

2. Using your default text editor:

	ovhcloud cloud managed-database user edit <service_id> <user_id> --editor
  The CLI will open your default text editor to update the parameters. When saving the file, the edition will be applied.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-database user edit <service_id> <user_id> --editor --roles role1,role2
`,
		Run:  cloud.EditManagedDatabaseUser,
		Args: cobra.ExactArgs(2),
	}

	// PostgreSQL and MongoDB specific flags
	userEditCmd.Flags().StringSliceVar(&cloud.ManagedDatabaseUserSpec.Roles, "roles", nil, "Roles granted to the user (postgresql and mongodb only)")

	// Valkey specific flags
	userEditCmd.Flags().StringSliceVar(&cloud.ManagedDatabaseUserSpec.Categories, "categories", nil, "Command categories the user can execute (valkey only)")
	userEditCmd.Flags().StringSliceVar(&cloud.ManagedDatabaseUserSpec.Channels, "channels", nil, "Channels the user can subscribe to (valkey only)")
	userEditCmd.Flags().StringSliceVar(&cloud.ManagedDatabaseUserSpec.Commands, "commands", nil, "Commands the user can execute (valkey only)")
	userEditCmd.Flags().StringSliceVar(&cloud.ManagedDatabaseUserSpec.Keys, "keys", nil, "Keys the user can access (valkey only)")

	addInteractiveEditorFlag(userEditCmd)

	return userEditCmd
}

func managedDatabaseRoleListCmd() *cobra.Command {
	roleListCmd := &cobra.Command{
		Use:     "list <service_id>",
		Aliases: []string{"ls"},
		Short:   "List roles in the given managed database service",
		Run:     cloud.ListManagedDatabaseRoles,
		Args:    cobra.ExactArgs(1),
	}

	// MongoDB specific flags
	roleListCmd.Flags().BoolVar(&cloud.ManagedDatabaseRoleSpec.Advanced, "advanced", false, "Adds the advanced roles to the list of the roles (mongodb only)")

	return roleListCmd
}
