// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0
// Crafted with Codex

package cmd

import (
	"fmt"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/ovh/ovhcloud-cli/internal/services/webhosting"
	"github.com/spf13/cobra"
)

func init() {
	webhostingCmd := &cobra.Command{
		Use:   "webhosting",
		Short: "Retrieve information and manage your WebHosting services",
	}

	// Command to list WebHosting services
	webhostingListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your WebHosting services",
		Run:     webhosting.ListWebHosting,
	}
	webhostingCmd.AddCommand(withFilterFlag(webhostingListCmd))

	// Command to get a single WebHosting
	webhostingCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific WebHosting",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.GetWebHosting,
	})

	// Command to update a single WebHosting
	webhostingEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given WebHosting",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.EditWebHosting,
	}
	webhostingEditCmd.Flags().StringVar(&webhosting.WebHostingSpec.DisplayName, "display-name", "", "Display name of the WebHosting")
	webhostingEditCmd.Flags().BoolVar(&webhosting.WebHostingClearDisplayName, "clear-display-name", false, "Clear the display name (set default value)")
	addInteractiveEditorFlag(webhostingEditCmd)
	webhostingCmd.AddCommand(webhostingEditCmd)

	// Attached domains
	attachedDomainCmd := &cobra.Command{Use: "domain", Short: "Manage attached domains"}
	attachedDomainListCmd := &cobra.Command{
		Use:   "list <service_name>",
		Short: "List attached domains",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListAttachedDomains,
	}
	attachedDomainCmd.AddCommand(withFilterFlag(attachedDomainListCmd))

	attachedDomainGetCmd := &cobra.Command{
		Use:   "get <service_name> <domain>",
		Short: "Get an attached domain",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetAttachedDomain,
	}
	attachedDomainCmd.AddCommand(attachedDomainGetCmd)
	attachedDomainDigCmd := &cobra.Command{
		Use:   "dig-status <service_name> <domain>",
		Short: "Check DNS status for an attached domain",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetAttachedDomainDigStatus,
	}
	attachedDomainCmd.AddCommand(attachedDomainDigCmd)

	attachedDomainAddCmd := &cobra.Command{
		Use:   "add <service_name>",
		Short: "Attach a domain",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.AddAttachedDomain,
	}
	attachedDomainAddCmd.Flags().StringVar(&webhosting.AttachedDomainDomain, "domain", "", "Domain to link")
	attachedDomainAddCmd.Flags().StringVar(&webhosting.AttachedDomainPath, "path", "", "Path of the attached domain")
	attachedDomainAddCmd.Flags().IntVar(&webhosting.AttachedDomainRuntimeID, "runtime-id", 0, "Runtime configuration ID used on this domain")
	attachedDomainAddCmd.Flags().BoolVar(&webhosting.AttachedDomainEnableSSL, "enable-ssl", false, "Whether to put the attached domain in the SSL certificate")
	attachedDomainAddCmd.Flags().BoolVar(&webhosting.AttachedDomainDisableSSL, "disable-ssl", false, "Exclude the attached domain from the SSL certificate")
	attachedDomainAddCmd.Flags().StringVar(&webhosting.AttachedDomainCDN, "cdn", "", "Whether the attached domain is linked to the hosting CDN (allowed: active, none)")
	attachedDomainAddCmd.Flags().StringVar(&webhosting.AttachedDomainFirewall, "firewall", "", "Whether the firewall is active for this domain (allowed: active, none)")
	attachedDomainAddCmd.Flags().StringVar(&webhosting.AttachedDomainIPLocation, "ip-location", "", "Change attached domain's DNS to the IP of the country (allowed: BE, CA, CZ, DE, ES, FI, FR, IE, IT, LT, NL, PL, PT, UK)")
	attachedDomainAddCmd.Flags().StringVar(&webhosting.AttachedDomainOwnLog, "own-log", "", "Domain to separate the logs on")
	attachedDomainAddCmd.Flags().BoolVar(&webhosting.AttachedDomainBypassDNS, "bypass-dns", false, "If set to true, DNS zone will not be updated by the operation")
	addFromFileFlag(attachedDomainAddCmd)
	addInteractiveEditorFlag(attachedDomainAddCmd)
	attachedDomainCmd.AddCommand(attachedDomainAddCmd)

	attachedDomainUpdateCmd := &cobra.Command{
		Use:   "update <service_name> <domain>",
		Short: "Update an attached domain",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.UpdateAttachedDomain,
	}
	attachedDomainUpdateCmd.Flags().StringVar(&webhosting.AttachedDomainPath, "path", "", "Path of the attached domain")
	attachedDomainUpdateCmd.Flags().IntVar(&webhosting.AttachedDomainRuntimeID, "runtime-id", 0, "Runtime configuration ID used on this domain")
	attachedDomainUpdateCmd.Flags().BoolVar(&webhosting.AttachedDomainEnableSSL, "enable-ssl", false, "Whether to put the attached domain in the SSL certificate")
	attachedDomainUpdateCmd.Flags().BoolVar(&webhosting.AttachedDomainDisableSSL, "disable-ssl", false, "Exclude the attached domain from the SSL certificate")
	attachedDomainUpdateCmd.Flags().StringVar(&webhosting.AttachedDomainCDN, "cdn", "", "Whether the attached domain is linked to the hosting CDN (allowed: active, none)")
	attachedDomainUpdateCmd.Flags().StringVar(&webhosting.AttachedDomainFirewall, "firewall", "", "Whether the firewall is active for this domain (allowed: active, none)")
	attachedDomainUpdateCmd.Flags().StringVar(&webhosting.AttachedDomainIPLocation, "ip-location", "", "Change attached domain's DNS to the IP of the country (allowed: BE, CA, CZ, DE, ES, FI, FR, IE, IT, LT, NL, PL, PT, UK)")
	attachedDomainUpdateCmd.Flags().StringVar(&webhosting.AttachedDomainOwnLog, "own-log", "", "Domain to separate the logs on")
	attachedDomainUpdateCmd.Flags().BoolVar(&webhosting.AttachedDomainBypassDNS, "bypass-dns", false, "If set to true, DNS zone will not be updated by the operation")
	addInteractiveEditorFlag(attachedDomainUpdateCmd)
	attachedDomainCmd.AddCommand(attachedDomainUpdateCmd)
	attachedDomainFindCmd := &cobra.Command{
		Use:   "find <domain>",
		Short: "Find hosting service linked to a domain",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.FindHostingByDomain,
	}
	attachedDomainCmd.AddCommand(attachedDomainFindCmd)
	attachedDomainOfferCmd := &cobra.Command{
		Use:   "available-offer <domain>",
		Short: "List hosting offers available for a domain",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListAvailableHostingOffers,
	}
	attachedDomainCmd.AddCommand(attachedDomainOfferCmd)

	attachedDomainDeleteCmd := &cobra.Command{
		Use:   "delete <service_name> <domain>",
		Short: "Delete an attached domain",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.DeleteAttachedDomain,
	}
	attachedDomainCmd.AddCommand(attachedDomainDeleteCmd)

	attachedDomainPurgeCmd := &cobra.Command{
		Use:   "purge-cache <service_name> <domain>",
		Short: "Purge CDN cache for an attached domain",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.PurgeAttachedDomainCache,
	}
	attachedDomainCmd.AddCommand(attachedDomainPurgeCmd)

	attachedDomainRestartCmd := &cobra.Command{
		Use:   "restart <service_name> <domain>",
		Short: "Restart an attached domain",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.RestartAttachedDomain,
	}
	attachedDomainCmd.AddCommand(attachedDomainRestartCmd)

	webhostingCmd.AddCommand(attachedDomainCmd)

	incidentCmd := &cobra.Command{
		Use:   "incident",
		Short: "List current incidents",
		Args:  cobra.NoArgs,
		Run:   webhosting.ListHostingIncidents,
	}
	webhostingCmd.AddCommand(withFilterFlag(incidentCmd))

	// Cron
	cronCmd := &cobra.Command{Use: "cron", Short: "Manage cron tasks"}
	cronListCmd := &cobra.Command{
		Use:   "list <service_name>",
		Short: "List cron tasks",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListCrons,
	}
	cronCmd.AddCommand(withFilterFlag(cronListCmd))
	cronGetCmd := &cobra.Command{
		Use:   "get <service_name> <id>",
		Short: "Get a cron task",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetCron,
	}
	cronCmd.AddCommand(cronGetCmd)

	cronCreateCmd := &cobra.Command{
		Use:   "create <service_name>",
		Short: "Create a cron task",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.CreateCron,
	}
	cronCreateCmd.Flags().StringVar(&webhosting.CronCommand, "command", "", "Command to execute")
	cronCreateCmd.Flags().StringVar(&webhosting.CronFrequency, "frequency", "", "Frequency (crontab format)")
	cronCreateCmd.Flags().StringVar(&webhosting.CronLanguage, "language", "", "Language")
	cronCreateCmd.Flags().StringVar(&webhosting.CronEmail, "email", "", "Email for stderr")
	cronCreateCmd.Flags().StringVar(&webhosting.CronDesc, "description", "", "Description")
	addFromFileFlag(cronCreateCmd)
	addInteractiveEditorFlag(cronCreateCmd)
	cronCmd.AddCommand(cronCreateCmd)

	cronUpdateCmd := &cobra.Command{
		Use:   "update <service_name> <id>",
		Short: "Update a cron task",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.UpdateCron,
	}
	cronUpdateCmd.Flags().StringVar(&webhosting.CronCommand, "command", "", "Command to execute")
	cronUpdateCmd.Flags().StringVar(&webhosting.CronFrequency, "frequency", "", "Frequency (crontab format)")
	cronUpdateCmd.Flags().StringVar(&webhosting.CronLanguage, "language", "", "Language")
	cronUpdateCmd.Flags().StringVar(&webhosting.CronEmail, "email", "", "Email for stderr")
	cronUpdateCmd.Flags().StringVar(&webhosting.CronDesc, "description", "", "Description")
	addInteractiveEditorFlag(cronUpdateCmd)
	cronCmd.AddCommand(cronUpdateCmd)

	cronDeleteCmd := &cobra.Command{
		Use:   "delete <service_name> <id>",
		Short: "Delete a cron task",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.DeleteCron,
	}
	cronCmd.AddCommand(cronDeleteCmd)
	cronAvailableLangCmd := &cobra.Command{
		Use:   "available-languages <service_name>",
		Short: "List available cron languages",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListCronAvailableLanguages,
	}
	cronCmd.AddCommand(cronAvailableLangCmd)
	webhostingCmd.AddCommand(cronCmd)

	// Databases
	dbCmd := &cobra.Command{Use: "db", Short: "Manage databases"}
	dbListCmd := &cobra.Command{
		Use:   "list <service_name>",
		Short: "List databases",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListDatabases,
	}
	dbCmd.AddCommand(withFilterFlag(dbListCmd))
	dbGetCmd := &cobra.Command{
		Use:   "get <service_name> <name>",
		Short: "Get a database",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetDatabase,
	}
	dbCmd.AddCommand(dbGetCmd)

	dbCapabilitiesCmd := &cobra.Command{
		Use:   "capabilities <service_name> <name>",
		Short: "Get database capabilities",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetDatabaseCapabilities,
	}
	dbCmd.AddCommand(dbCapabilitiesCmd)

	dbAvailableTypeCmd := &cobra.Command{
		Use:   "available-type <service_name>",
		Short: "List database types available for creation",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListDatabaseAvailableTypes,
	}
	dbCmd.AddCommand(withFilterFlag(dbAvailableTypeCmd))

	dbAvailableVersionCmd := &cobra.Command{
		Use:   "available-version <service_name>",
		Short: "List available versions for a database type",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListDatabaseAvailableVersions,
	}
	dbAvailableVersionCmd.Flags().StringVar(&webhosting.DatabaseVersionQueryType, "type", "", "Database type (required)")
	dbCmd.AddCommand(dbAvailableVersionCmd)

	dbCreationCapabilitiesCmd := &cobra.Command{
		Use:   "creation-capabilities <service_name>",
		Short: "List database creation capabilities",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListDatabaseCreationCapabilities,
	}
	dbCmd.AddCommand(withFilterFlag(dbCreationCapabilitiesCmd))

	dbCreateCmd := &cobra.Command{
		Use:   "create <service_name>",
		Short: "Create a database",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.CreateDatabase,
	}
	dbCreateCmd.Flags().StringVar(&webhosting.DatabaseCapability, "capability", "", "Database capability (allowed: extraSqlPerso, local, privateDatabase, sqlLocal, sqlPerso, sqlPro)")
	dbCreateCmd.Flags().StringVar(&webhosting.DatabaseType, "type", "", "Database type (allowed: mariadb, mysql, postgresql, redis)")
	dbCreateCmd.Flags().StringVar(&webhosting.DatabaseUser, "user", "", "Database user (must start with hosting login, lowercase)")
	dbCreateCmd.Flags().StringVar(&webhosting.DatabasePassword, "password", "", "Database password")
	dbCreateCmd.Flags().StringVar(&webhosting.DatabaseVersion, "version", "", "Database version (allowed: 10, 10.1, 10.11, 10.2, 10.3, 10.4, 10.5, 10.6, 11, 12, 13, 15, 3.2, 3.4, 4.0, 5.1, 5.5, 5.6, 5.7, 6.0, 7.0, 8.0, 8.4, 9.4, 9.5, 9.6)")
	dbCreateCmd.Flags().StringVar(&webhosting.DatabaseQuota, "quota", "", "Database quota (allowed: 25, 100, 200, 256, 400, 512, 800, 1024)")
	addFromFileFlag(dbCreateCmd)
	addInteractiveEditorFlag(dbCreateCmd)
	dbCmd.AddCommand(dbCreateCmd)

	dbDeleteCmd := &cobra.Command{
		Use:   "delete <service_name> <name>",
		Short: "Delete a database",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.DeleteDatabase,
	}
	dbCmd.AddCommand(dbDeleteCmd)

	dbChangePasswordCmd := &cobra.Command{
		Use:   "change-password <service_name> <name>",
		Short: "Change database password",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.ChangeDatabasePassword,
	}
	dbChangePasswordCmd.Flags().StringVar(&webhosting.DatabasePassword, "password", "", "New password")
	dbCmd.AddCommand(dbChangePasswordCmd)

	dbCopyCmd := &cobra.Command{Use: "copy", Short: "Manage database copies"}
	dbCopyListCmd := &cobra.Command{
		Use:   "list <service_name> <name>",
		Short: "List database copies",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.ListDatabaseCopies,
	}
	dbCopyCmd.AddCommand(withFilterFlag(dbCopyListCmd))
	dbCopyGetCmd := &cobra.Command{
		Use:   "get <service_name> <name> <id>",
		Short: "Get a database copy",
		Args:  cobra.ExactArgs(3),
		Run:   webhosting.GetDatabaseCopy,
	}
	dbCopyCmd.AddCommand(dbCopyGetCmd)
	dbCopyCreateCmd := &cobra.Command{
		Use:   "create <service_name> <name>",
		Short: "Create a database copy",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.CreateDatabaseCopy,
	}
	dbCopyCmd.AddCommand(dbCopyCreateCmd)
	dbCopyDeleteCmd := &cobra.Command{
		Use:   "delete <service_name> <name> <id>",
		Short: "Delete a database copy",
		Args:  cobra.ExactArgs(3),
		Run:   webhosting.DeleteDatabaseCopy,
	}
	dbCopyCmd.AddCommand(dbCopyDeleteCmd)
	dbCopyRestoreCmd := &cobra.Command{
		Use:   "restore <service_name> <name>",
		Short: "Restore a database copy",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.RestoreDatabaseCopy,
	}
	dbCopyRestoreCmd.Flags().StringVar(&webhosting.DatabaseCopyID, "copy-id", "", "Copy ID to restore")
	dbCopyRestoreCmd.Flags().BoolVar(&webhosting.DatabaseFlush, "flush", false, "Flush database before restore")
	addFromFileFlag(dbCopyRestoreCmd)
	addInteractiveEditorFlag(dbCopyRestoreCmd)
	dbCopyCmd.AddCommand(dbCopyRestoreCmd)
	dbCmd.AddCommand(dbCopyCmd)

	dbDumpCmd := &cobra.Command{Use: "dump", Short: "Manage database dumps"}
	dbDumpListCmd := &cobra.Command{
		Use:   "list <service_name> <name>",
		Short: "List database dumps",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.ListDatabaseDumps,
	}
	dbDumpCmd.AddCommand(withFilterFlag(dbDumpListCmd))
	dbDumpGetCmd := &cobra.Command{
		Use:   "get <service_name> <name> <id>",
		Short: "Get a database dump",