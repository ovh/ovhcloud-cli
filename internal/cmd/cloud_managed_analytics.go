// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initManagedAnalyticsCommand(cloudCmd *cobra.Command) {
	managedAnalyticsCmd := &cobra.Command{
		Use:   "managed-analytics",
		Short: "Manage managed analytics services in the given cloud project",
	}
	managedAnalyticsCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")
	// Managed analytics commands
	managedAnalyticsCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your managed analytics services",
		Run:     cloud.ListManagedAnalytics,
	}))

	managedAnalyticsCmd.AddCommand(&cobra.Command{
		Use:   "get <service_id>",
		Short: "Get a specific managed analytics service",
		Run:   cloud.GetManagedAnalytics,
		Args:  cobra.ExactArgs(1),
	})

	managedAnalyticsCmd.AddCommand(managedAnalyticsCreationCmd())
	managedAnalyticsCmd.AddCommand(managedAnalyticsEditCmd())

	managedAnalyticsCmd.AddCommand(&cobra.Command{
		Use:   "delete <service_id>",
		Short: "Delete a specific managed analytics service",
		Run:   cloud.DeleteManagedAnalytics,
		Args:  cobra.ExactArgs(1),
	})

	initManagedAnalyticsDatabaseCommand(managedAnalyticsCmd)
	initManagedAnalyticsUserCommand(managedAnalyticsCmd)
	initManagedAnalyticsRoleCommand(managedAnalyticsCmd)
	initManagedAnalyticsPermissionCommand(managedAnalyticsCmd)
	initManagedAnalyticsPatternCommand(managedAnalyticsCmd)
	initManagedAnalyticsCertificateCommand(managedAnalyticsCmd)
	initManagedAnalyticsBackupCommand(managedAnalyticsCmd)
	initManagedAnalyticsTopicCommand(managedAnalyticsCmd)
	initManagedAnalyticsTopicACLCommand(managedAnalyticsCmd)

	cloudCmd.AddCommand(managedAnalyticsCmd)
}

func initManagedAnalyticsDatabaseCommand(managedAnalyticsCmd *cobra.Command) {
	databaseCmd := &cobra.Command{
		Use:   "database",
		Short: "Manage databases in a specific managed analytics service",
	}

	databaseCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_id>",
		Aliases: []string{"ls"},
		Short:   "List all databases in the given managed analytics service",
		Run:     cloud.ListManagedAnalyticsDatabases,
		Args:    cobra.ExactArgs(1),
	}))

	databaseCmd.AddCommand(&cobra.Command{
		Use:   "get <service_id> <database_id>",
		Short: "Get a specific database in the given managed analytics service",
		Run:   cloud.GetManagedAnalyticsDatabase,
		Args:  cobra.ExactArgs(2),
	})

	databaseCmd.AddCommand(managedAnalyticsDatabaseCreateCmd())
	databaseCmd.AddCommand(&cobra.Command{
		Use:   "delete <service_id> <database_id>",
		Short: "Delete a specific database in the given managed analytics service",
		Run:   cloud.DeleteManagedAnalyticsDatabase,
		Args:  cobra.ExactArgs(2),
	})

	managedAnalyticsCmd.AddCommand(databaseCmd)
}

func initManagedAnalyticsUserCommand(managedAnalyticsCmd *cobra.Command) {
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users in a specific managed analytics service",
	}

	userCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_id>",
		Aliases: []string{"ls"},
		Short:   "List all users in the given managed analytics service",
		Run:     cloud.ListManagedAnalyticsUsers,
		Args:    cobra.ExactArgs(1),
	}))

	userCmd.AddCommand(&cobra.Command{
		Use:   "get <service_id> <user_id>",
		Short: "Get a specific user in the given managed analytics service",
		Run:   cloud.GetManagedAnalyticsUser,
		Args:  cobra.ExactArgs(2),
	})

	userCmd.AddCommand(managedAnalyticsUserCreateCmd())
	userCmd.AddCommand(managedAnalyticsUserEditCmd())

	userCmd.AddCommand(&cobra.Command{
		Use:   "delete <service_id> <user_id>",
		Short: "Delete a specific user in the given managed analytics service",
		Run:   cloud.DeleteManagedAnalyticsUser,
		Args:  cobra.ExactArgs(2),
	})

	userCmd.AddCommand(&cobra.Command{
		Use:   "credentials-reset <service_id> <user_id>",
		Short: "Reset the credentials of a specific user in the given managed analytics service",
		Run:   cloud.ResetManagedAnalyticsUserCredentials,
		Args:  cobra.ExactArgs(2),
	})

	managedAnalyticsCmd.AddCommand(userCmd)
}

func initManagedAnalyticsRoleCommand(managedAnalyticsCmd *cobra.Command) {
	roleCmd := &cobra.Command{
		Use:   "role",
		Short: "Manage roles in a specific managed analytics service",
	}

	roleCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_id>",
		Aliases: []string{"ls"},
		Short:   "List roles in the given managed analytics service",
		Run:     cloud.ListManagedAnalyticsRoles,
		Args:    cobra.ExactArgs(1),
	}))

	managedAnalyticsCmd.AddCommand(roleCmd)
}

func initManagedAnalyticsPermissionCommand(managedAnalyticsCmd *cobra.Command) {
	permissionCmd := &cobra.Command{
		Use:   "permission",
		Short: "Manage permissions in a specific managed analytics service",
	}

	permissionCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_id>",
		Aliases: []string{"ls"},
		Short:   "List permissions in the given managed analytics service",
		Run:     cloud.ListManagedAnalyticsPermissions,
		Args:    cobra.ExactArgs(1),
	}))

	managedAnalyticsCmd.AddCommand(permissionCmd)
}

func initManagedAnalyticsPatternCommand(managedAnalyticsCmd *cobra.Command) {
	patternCmd := &cobra.Command{
		Use:   "pattern",
		Short: "Manage patterns in a specific managed analytics service",
	}

	patternCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_id>",
		Aliases: []string{"ls"},
		Short:   "List patterns in the given managed analytics service",
		Run:     cloud.ListManagedAnalyticsPatterns,
		Args:    cobra.ExactArgs(1),
	}))

	patternCmd.AddCommand(&cobra.Command{
		Use:   "get <service_id> <pattern_id>",
		Short: "Get a specific pattern in the given managed analytics service",
		Run:   cloud.GetManagedAnalyticsPattern,
		Args:  cobra.ExactArgs(2),
	})

	patternCmd.AddCommand(managedAnalyticsPatternCreateCmd())

	patternCmd.AddCommand(&cobra.Command{
		Use:   "delete <service_id> <pattern_id>",
		Short: "Delete a specific pattern in the given managed analytics service",
		Run:   cloud.DeleteManagedAnalyticsPattern,
		Args:  cobra.ExactArgs(2),
	})

	managedAnalyticsCmd.AddCommand(patternCmd)
}

func initManagedAnalyticsCertificateCommand(managedAnalyticsCmd *cobra.Command) {
	certificateCmd := &cobra.Command{
		Use:   "certificate",
		Short: "Manage certificates in a specific managed analytics service",
	}

	certificateCmd.AddCommand(&cobra.Command{
		Use:   "get <service_id>",
		Short: "Get certificates in the given managed analytics service",
		Run:   cloud.GetManagedAnalyticsCertificates,
		Args:  cobra.ExactArgs(1),
	})

	managedAnalyticsCmd.AddCommand(certificateCmd)
}

func initManagedAnalyticsBackupCommand(managedAnalyticsCmd *cobra.Command) {
	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage backups in a specific managed analytics service",
	}

	backupCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_id>",
		Aliases: []string{"ls"},
		Short:   "List all backups in the given managed analytics service",
		Run:     cloud.ListManagedAnalyticsBackups,
		Args:    cobra.ExactArgs(1),
	}))

	backupCmd.AddCommand(&cobra.Command{
		Use:   "get <service_id> <backup_id>",
		Short: "Get a specific backup in the given managed analytics service",
		Run:   cloud.GetManagedAnalyticsBackup,
		Args:  cobra.ExactArgs(2),
	})

	managedAnalyticsCmd.AddCommand(backupCmd)
}

func initManagedAnalyticsTopicCommand(managedAnalyticsCmd *cobra.Command) {
	topicCmd := &cobra.Command{
		Use:   "topic",
		Short: "Manage topics in a specific managed analytics service",
	}

	topicCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_id>",
		Aliases: []string{"ls"},
		Short:   "List topics in the given managed analytics service",
		Run:     cloud.ListManagedAnalyticsTopics,
		Args:    cobra.ExactArgs(1),
	}))

	topicCmd.AddCommand(&cobra.Command{
		Use:   "get <service_id> <topic_id>",
		Short: "Get a specific topic in the given managed analytics service",
		Run:   cloud.GetManagedAnalyticsTopic,
		Args:  cobra.ExactArgs(2),
	})

	topicCmd.AddCommand(managedAnalyticsTopicCreateCmd())
	topicCmd.AddCommand(managedAnalyticsTopicEditCmd())

	topicCmd.AddCommand(&cobra.Command{
		Use:   "delete <service_id> <topic_id>",
		Short: "Delete a specific topic in the given managed analytics service",
		Run:   cloud.DeleteManagedAnalyticsTopic,
		Args:  cobra.ExactArgs(2),
	})

	managedAnalyticsCmd.AddCommand(topicCmd)
}

func initManagedAnalyticsTopicACLCommand(managedAnalyticsCmd *cobra.Command) {
	topicACLCmd := &cobra.Command{
		Use:   "topic-acl",
		Short: "Manage topic ACLs in a specific managed analytics service",
	}

	topicACLCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <service_id>",
		Aliases: []string{"ls"},
		Short:   "List topic ACLs in the given managed analytics service",
		Run:     cloud.ListManagedAnalyticsTopicACLs,
		Args:    cobra.ExactArgs(1),
	}))

	topicACLCmd.AddCommand(&cobra.Command{
		Use:   "get <service_id> <topic-acl_id>",
		Short: "Get a specific topic ACL in the given managed analytics service",
		Run:   cloud.GetManagedAnalyticsTopicACL,
		Args:  cobra.ExactArgs(2),
	})

	topicACLCmd.AddCommand(managedAnalyticsTopicACLCreateCmd())

	topicACLCmd.AddCommand(&cobra.Command{
		Use:   "delete <service_id> <topic-acl_id>",
		Short: "Delete a specific topic ACL in the given managed analytics service",
		Run:   cloud.DeleteManagedAnalyticsTopicACL,
		Args:  cobra.ExactArgs(2),
	})

	managedAnalyticsCmd.AddCommand(topicACLCmd)
}

func managedAnalyticsCreationCmd() *cobra.Command {
	managedAnalyticsCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new managed analytics service",
		Long: `Use this command to create a managed analytics service in the given public cloud project.
There are two ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud managed-analytics create --engine kafka --version 4.0 --plan production --nodes-pattern.flavor db1-4 --nodes-pattern.region GRA

2. Using your default text editor:

	ovhcloud cloud managed-analytics create --engine kafka --editor
  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-analytics create --engine kafka --editor --version 3
`,
		PreRunE: cloud.CreateManagedAnalyticsPreRun,
		Run:     cloud.CreateManagedAnalytics,
		Args:    cobra.NoArgs,
	}

	// Analytics details
	managedAnalyticsCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.Engine, "engine", "", "Analytics engine (you can get the list of available engines using 'ovhcloud cloud reference managed-analytics list-engines')")
	managedAnalyticsCreateCmd.MarkFlagRequired("engine")
	managedAnalyticsCreateCmd.RegisterFlagCompletionFunc("engine", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return cloud.ManagedAnalyticsValidEngines, cobra.ShellCompDirectiveNoFileComp
	})
	managedAnalyticsCreateCmd.Flags().StringSliceVar(&cloud.ManagedAnalyticsSpec.Backups.Regions, "backups-regions", nil, "Regions on which the backups are stored")
	managedAnalyticsCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.Backups.Time, "backups-time", "", "Time on which backups start every day")
	managedAnalyticsCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.Description, "description", "", "Analytics description")
	managedAnalyticsCreateCmd.Flags().IntVar(&cloud.ManagedAnalyticsSpec.Disk.Size, "disk-size", 0, "Disk size (GB)")
	managedAnalyticsCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.ForkFrom.BackupID, "fork-from.backup-id", "", "Backup ID (not compatible with fork-from.point-in-time)")
	managedAnalyticsCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.ForkFrom.PointInTime, "fork-from.point-in-time", "", "Point in time to restore from (not compatible with fork-from.backup-id)")
	managedAnalyticsCreateCmd.MarkFlagsMutuallyExclusive("fork-from.backup-id", "fork-from.point-in-time")
	markFlagsMutuallyExclusive(managedAnalyticsCreateCmd, "fork-from.backup-id", "fork-from.point-in-time")
	managedAnalyticsCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.ForkFrom.ServiceID, "fork-from.service-id", "", "Service ID that owns the backups")
	managedAnalyticsCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.MaintenanceTime, "maintenance-time", "", "Time on which maintenances can start every day")
	managedAnalyticsCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.Plan, "plan", "", "Analytics plan (you can get the list of available plans using 'ovhcloud cloud reference managed-analytics list-plans')")
	managedAnalyticsCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.Version, "version", "", "Analytics version (you can get the list of available versions using 'ovhcloud cloud reference managed-analytics list-engines')")

	// Network configuration
	managedAnalyticsCreateCmd.Flags().StringSliceVar(&cloud.ManagedAnalyticsSpec.CLIIPRestrictions, "ip-restrictions", nil, "IP blocks authorized to access the cluster (CIDR format)")
	managedAnalyticsCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.NetworkID, "network-id", "", "Private network ID in which the cluster is deployed")
	managedAnalyticsCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.SubnetID, "subnet-id", "", "Private subnet ID in which the cluster is deployed")

	// Nodes pattern definition
	managedAnalyticsCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.NodesPattern.Flavor, "nodes-pattern.flavor", "", "Flavor of all nodes")
	managedAnalyticsCreateCmd.Flags().IntVar(&cloud.ManagedAnalyticsSpec.NodesPattern.Number, "nodes-pattern.number", 0, "Number of nodes")
	managedAnalyticsCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.NodesPattern.Region, "nodes-pattern.region", "", "Region of all nodes")

	// Nodes list definition
	managedAnalyticsCreateCmd.Flags().StringSliceVar(&cloud.ManagedAnalyticsSpec.CLINodesList, "nodes-list", nil, "List of nodes (format: flavor1:region1,flavor2:region2...)")
	markFlagsMutuallyExclusive(managedAnalyticsCreateCmd, "nodes-pattern.flavor", "nodes-list")
	markFlagsMutuallyExclusive(managedAnalyticsCreateCmd, "nodes-pattern.number", "nodes-list")
	markFlagsMutuallyExclusive(managedAnalyticsCreateCmd, "nodes-pattern.region", "nodes-list")

	// Common flags for other mean to define parameters
	addInteractiveEditorFlag(managedAnalyticsCreateCmd)
	return managedAnalyticsCreateCmd
}

func managedAnalyticsEditCmd() *cobra.Command {
	managedAnalyticsEditCmd := &cobra.Command{
		Use:   "edit <service_id>",
		Short: "Edit a specific managed analytics service",
		Long: `Use this command to edit a managed analytics service in the given public cloud project.
There are two ways to define the edition parameters:

1. Using only CLI flags:

	ovhcloud cloud managed-analytics edit <service_id> --description "My analytics service"

2. Using your default text editor:

	ovhcloud cloud managed-analytics edit <service_id> --editor
  The CLI will open your default text editor to update the parameters. When saving the file, the edition will be applied.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-analytics edit <service_id> --editor --description "My analytics service"
`,
		Run:  cloud.EditManagedAnalytics,
		Args: cobra.ExactArgs(1),
	}

	// Analytics details
	managedAnalyticsEditCmd.Flags().StringSliceVar(&cloud.ManagedAnalyticsSpec.Backups.Regions, "backups-regions", nil, "Regions on which the backups are stored")
	managedAnalyticsEditCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.Backups.Time, "backups-time", "", "Time on which backups start every day")
	managedAnalyticsEditCmd.Flags().BoolVar(&cloud.ManagedAnalyticsSpec.DeletionProtection, "deletion-protection", false, "Enable deletion protection")
	managedAnalyticsEditCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.Description, "description", "", "Description of the cluster")
	managedAnalyticsEditCmd.Flags().BoolVar(&cloud.ManagedAnalyticsSpec.EnablePrometheus, "enable-prometheus", false, "Enable Prometheus")
	managedAnalyticsEditCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.NodesPattern.Flavor, "flavor", "", "The VM flavor used for this cluster")
	managedAnalyticsEditCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.MaintenanceTime, "maintenance-time", "", "Time on which maintenances can start every day")
	managedAnalyticsEditCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.Plan, "plan", "", "Plan of the cluster")
	managedAnalyticsEditCmd.Flags().StringVar(&cloud.ManagedAnalyticsSpec.Version, "version", "", "Version of the engine deployed on the cluster")

	// Network configuration
	managedAnalyticsEditCmd.Flags().StringSliceVar(&cloud.ManagedAnalyticsSpec.CLIIPRestrictions, "ip-restrictions", nil, "IP blocks authorized to access the cluster (CIDR format)")

	// Common flags for other mean to define parameters
	addInteractiveEditorFlag(managedAnalyticsEditCmd)

	return managedAnalyticsEditCmd
}

func managedAnalyticsDatabaseCreateCmd() *cobra.Command {
	databaseCreateCmd := &cobra.Command{
		Use:   "create <service_id>",
		Short: "Create a new database in the given managed analytics service",
		Long: `Use this command to create a database in the given managed analytics service.

	ovhcloud cloud managed-analytics database create <service_id> --name mydb
`,
		Run:  cloud.CreateManagedAnalyticsDatabase,
		Args: cobra.ExactArgs(1),
	}

	databaseCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsDatabaseSpec.Name, "name", "", "Name of the database to create")
	databaseCreateCmd.MarkFlagRequired("name")

	return databaseCreateCmd
}

func managedAnalyticsUserCreateCmd() *cobra.Command {
	userCreateCmd := &cobra.Command{
		Use:   "create <service_id>",
		Short: "Create a new user in the given managed analytics service",
		Long: `Use this command to create a user in the given managed analytics service.
There are two ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud managed-analytics user create <service_id> --name myuser

  For Clickhouse engine, you can also specify roles:

	ovhcloud cloud managed-analytics user create <service_id> --name myuser --roles role1,role2

  For OpenSearch engine, you can specify ACLs:

	ovhcloud cloud managed-analytics user create <service_id> --name myuser --acls "logs-.*:write" --acls "metrics-.*:read"

2. Using your default text editor:

	ovhcloud cloud managed-analytics user create <service_id> --name myuser --editor
  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-analytics user create <service_id> --name myuser --editor --roles role1,role2
`,
		Run:  cloud.CreateManagedAnalyticsUser,
		Args: cobra.ExactArgs(1),
	}

	userCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsUserSpec.Name, "name", "", "Name of the user to create")
	userCreateCmd.MarkFlagRequired("name")

	// Clickhouse specific flags
	userCreateCmd.Flags().StringSliceVar(&cloud.ManagedAnalyticsUserSpec.Roles, "roles", nil, "Roles granted to the user (clickhouse only)")

	// OpenSearch specific flags
	userCreateCmd.Flags().StringArrayVar(&cloud.ManagedAnalyticsUserSpec.CLIAcls, "acls", nil, "ACL granted to the user (opensearch only, format: pattern:permission)")

	addInteractiveEditorFlag(userCreateCmd)

	return userCreateCmd
}

func managedAnalyticsUserEditCmd() *cobra.Command {
	userEditCmd := &cobra.Command{
		Use:   "edit <service_id> <user_id>",
		Short: "Edit a user in the given managed analytics service",
		Long: `Use this command to edit a user in the given managed analytics service.
There are two ways to define the edition parameters:

1. Using only CLI flags:

  For Clickhouse engine:

	ovhcloud cloud managed-analytics user edit <service_id> <user_id> --roles role1,role2

  For OpenSearch engine:

	ovhcloud cloud managed-analytics user edit <service_id> <user_id> --acls "logs-.*:write" --acls "metrics-.*:read"

2. Using your default text editor:

	ovhcloud cloud managed-analytics user edit <service_id> <user_id> --editor
  The CLI will open your default text editor to update the parameters. When saving the file, the edition will be applied.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-analytics user edit <service_id> <user_id> --editor --roles role1,role2
`,
		Run:  cloud.EditManagedAnalyticsUser,
		Args: cobra.ExactArgs(2),
	}

	// Clickhouse specific flags
	userEditCmd.Flags().StringSliceVar(&cloud.ManagedAnalyticsUserSpec.Roles, "roles", nil, "Roles granted to the user (clickhouse only)")

	// OpenSearch specific flags
	userEditCmd.Flags().StringArrayVar(&cloud.ManagedAnalyticsUserSpec.CLIAcls, "acls", nil, "ACL granted to the user (opensearch only, format: pattern:permission)")

	addInteractiveEditorFlag(userEditCmd)

	return userEditCmd
}

func managedAnalyticsPatternCreateCmd() *cobra.Command {
	patternCreateCmd := &cobra.Command{
		Use:   "create <service_id>",
		Short: "Create a new pattern in the given managed analytics service",
		Run:   cloud.CreateManagedAnalyticsPattern,
		Args:  cobra.ExactArgs(1),
	}

	patternCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsPatternSpec.Pattern, "pattern", "", "Pattern format")
	patternCreateCmd.MarkFlagRequired("pattern")

	patternCreateCmd.Flags().IntVar(&cloud.ManagedAnalyticsPatternSpec.MaxIndexCount, "max-index-count", 0, "Maximum number of index for this pattern (clickhouse only)")

	addInteractiveEditorFlag(patternCreateCmd)

	return patternCreateCmd
}

func managedAnalyticsTopicCreateCmd() *cobra.Command {
	topicCreateCmd := &cobra.Command{
		Use:   "create <service_id>",
		Short: "Create a new topic in the given managed analytics service",
		Run:   cloud.CreateManagedAnalyticsTopic,
		Args:  cobra.ExactArgs(1),
	}

	topicCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsTopicSpec.Name, "name", "", "Topic name")
	topicCreateCmd.MarkFlagRequired("name")
	topicCreateCmd.Flags().IntVar(&cloud.ManagedAnalyticsTopicSpec.MinInsyncReplicas, "min-insync-replicas", 0, "Minimum in-sync replicas")
	topicCreateCmd.Flags().IntVar(&cloud.ManagedAnalyticsTopicSpec.Partitions, "partitions", 0, "Number of partitions")
	topicCreateCmd.Flags().IntVar(&cloud.ManagedAnalyticsTopicSpec.Replication, "replication", 0, "Number of replications")
	topicCreateCmd.Flags().IntVar(&cloud.ManagedAnalyticsTopicSpec.RetentionBytes, "retention-bytes", 0, "Retention size in bytes (-1 for unlimited)")
	topicCreateCmd.Flags().IntVar(&cloud.ManagedAnalyticsTopicSpec.RetentionHours, "retention-hours", 0, "Retention duration in hours (-1 for unlimited)")

	addInteractiveEditorFlag(topicCreateCmd)

	return topicCreateCmd
}

func managedAnalyticsTopicEditCmd() *cobra.Command {
	topicEditCmd := &cobra.Command{
		Use:   "edit <service_id> <topic_id>",
		Short: "Edit a topic in the given managed analytics service",
		Run:   cloud.EditManagedAnalyticsTopic,
		Args:  cobra.ExactArgs(2),
	}

	topicEditCmd.Flags().IntVar(&cloud.ManagedAnalyticsTopicSpec.MinInsyncReplicas, "min-insync-replicas", 0, "Minimum in-sync replicas")
	topicEditCmd.Flags().IntVar(&cloud.ManagedAnalyticsTopicSpec.Partitions, "partitions", 0, "Number of partitions")
	topicEditCmd.Flags().IntVar(&cloud.ManagedAnalyticsTopicSpec.Replication, "replication", 0, "Number of replications")
	topicEditCmd.Flags().IntVar(&cloud.ManagedAnalyticsTopicSpec.RetentionBytes, "retention-bytes", 0, "Retention size in bytes (-1 for unlimited)")
	topicEditCmd.Flags().IntVar(&cloud.ManagedAnalyticsTopicSpec.RetentionHours, "retention-hours", 0, "Retention duration in hours (-1 for unlimited)")

	addInteractiveEditorFlag(topicEditCmd)

	return topicEditCmd
}

func managedAnalyticsTopicACLCreateCmd() *cobra.Command {
	topicACLCreateCmd := &cobra.Command{
		Use:   "create <service_id>",
		Short: "Create a new topic ACL in the given managed analytics service",
		Long: `Use this command to create a topic ACL in the given managed analytics service (kafka only).

	ovhcloud cloud managed-analytics topic-acl create <service_id> --permission read --topic my-topic --username myuser
`,
		Run:  cloud.CreateManagedAnalyticsTopicACL,
		Args: cobra.ExactArgs(1),
	}

	topicACLCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsTopicACLSpec.Permission, "permission", "", "ACL permission (e.g. read, write, readwrite)")
	topicACLCreateCmd.MarkFlagRequired("permission")
	topicACLCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsTopicACLSpec.Topic, "topic", "", "Topic name the ACL applies to")
	topicACLCreateCmd.MarkFlagRequired("topic")
	topicACLCreateCmd.Flags().StringVar(&cloud.ManagedAnalyticsTopicACLSpec.Username, "username", "", "Username the ACL applies to")
	topicACLCreateCmd.MarkFlagRequired("username")

	return topicACLCreateCmd
}
