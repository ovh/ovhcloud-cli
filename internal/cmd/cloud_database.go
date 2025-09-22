// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudDatabaseCommand(cloudCmd *cobra.Command) {
	databaseCmd := &cobra.Command{
		Use:   "database",
		Short: "Manage databases in the given cloud project",
	}
	databaseCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	databaseListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your databases",
		Run:     cloud.ListCloudDatabases,
	}
	databaseCmd.AddCommand(withFilterFlag(databaseListCmd))

	databaseCmd.AddCommand(&cobra.Command{
		Use:   "get <database_id>",
		Short: "Get a specific database",
		Run:   cloud.GetCloudDatabase,
		Args:  cobra.ExactArgs(1),
	})

	databaseCmd.AddCommand(getDatabaseCreationCmd())

	databaseCmd.AddCommand(&cobra.Command{
		Use:   "delete <database_id>",
		Short: "Delete a specific database",
		Run:   cloud.DeleteDatabase,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(databaseCmd)
}

func getDatabaseCreationCmd() *cobra.Command {
	databaseCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new database",
		Long: `Use this command to create a database in the given public cloud project.
		There are two ways to define the creation parameters:

		1. Using only CLI flags:

		  ovhcloud cloud database create --engine mysql --version 8 --plan essential  --nodes-list "db1-4:DE"

		2. Using your default text editor:

		  ovhcloud cloud database create --engine kafka --editor

		  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
		  default text editor to update the parameters. When saving the file, the creation will start.

		  Note that it is also possible to override values in the presented examples using command line flags like the following:

			ovhcloud cloud database create --engine mysql --editor --version 8
		`,
		Run:  cloud.CreateDatabase,
		Args: cobra.NoArgs,
	}

	// Database details
	databaseCreateCmd.Flags().StringVar(&cloud.DatabaseSpec.Engine, "engine", "", "Database engine (you can get the list of available engines using 'ovhcloud cloud reference database list-engines')")
	databaseCreateCmd.MarkFlagRequired("engine")
	databaseCreateCmd.Flags().StringSliceVar(&cloud.DatabaseSpec.Backups.Regions, "backups-regions", nil, "Regions on which the backups are stored")
	databaseCreateCmd.Flags().StringVar(&cloud.DatabaseSpec.Backups.Time, "backups-time", "", "Time on which backups start every day")
	databaseCreateCmd.Flags().StringVar(&cloud.DatabaseSpec.Description, "description", "", "Database description")
	databaseCreateCmd.Flags().IntVar(&cloud.DatabaseSpec.Disk.Size, "disk-size", 0, "Disk size (GB)")
	databaseCreateCmd.Flags().StringVar(&cloud.DatabaseSpec.ForkFrom.BackupID, "fork-from.backup-id", "", "Backup ID (not compatible with fork-from.point-in-time)")
	databaseCreateCmd.Flags().StringVar(&cloud.DatabaseSpec.ForkFrom.PointInTime, "fork-from.point-in-time", "", "Point in time to restore from (not compatible with fork-from.backup-id)")
	databaseCreateCmd.MarkFlagsMutuallyExclusive("fork-from.backup-id", "fork-from.point-in-time")
	databaseCreateCmd.Flags().StringVar(&cloud.DatabaseSpec.ForkFrom.ServiceID, "fork-from.service-id", "", "Service ID that owns the backups")
	databaseCreateCmd.Flags().StringVar(&cloud.DatabaseSpec.MaintenanceTime, "maintenance-time", "", "Time on which maintenances can start every day")
	databaseCreateCmd.Flags().StringVar(&cloud.DatabaseSpec.Plan, "plan", "", "Database plan (you can get the list of available plans using 'ovhcloud cloud reference database list-plans')")
	databaseCreateCmd.Flags().StringVar(&cloud.DatabaseSpec.Version, "version", "", "Database version (you can get the list of available versions using 'ovhcloud cloud reference database list-engines')")

	// Network configuration
	databaseCreateCmd.Flags().StringSliceVar(&cloud.DatabaseSpec.CLIIPRestrictions, "ip-restrictions", nil, "IP blocks authorized to access the cluster (CIDR format)")
	databaseCreateCmd.Flags().StringVar(&cloud.DatabaseSpec.NetworkID, "network-id", "", "Private network ID in which the cluster is deployed")
	databaseCreateCmd.Flags().StringVar(&cloud.DatabaseSpec.SubnetID, "subnet-id", "", "Private subnet ID in which the cluster is deployed")

	// Nodes pattern definition
	databaseCreateCmd.Flags().StringVar(&cloud.DatabaseSpec.NodesPattern.Flavor, "nodes-pattern.flavor", "", "Flavor of all nodes")
	databaseCreateCmd.Flags().IntVar(&cloud.DatabaseSpec.NodesPattern.Number, "nodes-pattern.number", 0, "Number of nodes")
	databaseCreateCmd.Flags().StringVar(&cloud.DatabaseSpec.NodesPattern.Region, "nodes-pattern.region", "", "Region of all nodes")

	// Nodes list definition
	databaseCreateCmd.Flags().StringSliceVar(&cloud.DatabaseSpec.CLINodesList, "nodes-list", nil, "List of nodes (format: flavor1:region1,flavor2:region2...)")
	databaseCreateCmd.MarkFlagsMutuallyExclusive("nodes-pattern.flavor", "nodes-list")
	databaseCreateCmd.MarkFlagsMutuallyExclusive("nodes-pattern.number", "nodes-list")
	databaseCreateCmd.MarkFlagsMutuallyExclusive("nodes-pattern.region", "nodes-list")

	// Common flags for other mean to define parameters
	addInteractiveEditorFlag(databaseCreateCmd)

	return databaseCreateCmd
}
