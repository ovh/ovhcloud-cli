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
		Args:  cobra.ExactArgs(3),
		Run:   webhosting.GetDatabaseDump,
	}
	dbDumpCmd.AddCommand(dbDumpGetCmd)
	dbDumpCreateCmd := &cobra.Command{
		Use:   "create <service_name> <name>",
		Short: "Request a database dump",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.RequestDatabaseDump,
	}
	dbDumpCreateCmd.Flags().StringVar(&webhosting.DatabaseDumpDate, "date", "", "Dump type (allowed: daily.1, now, weekly.1)")
	dbDumpCreateCmd.Flags().BoolVar(&webhosting.DatabaseSendEmail, "send-email", true, "Send email when dump is ready")
	addFromFileFlag(dbDumpCreateCmd)
	addInteractiveEditorFlag(dbDumpCreateCmd)
	dbDumpCmd.AddCommand(dbDumpCreateCmd)
	dbDumpDeleteCmd := &cobra.Command{
		Use:   "delete <service_name> <name> <id>",
		Short: "Delete a database dump",
		Args:  cobra.ExactArgs(3),
		Run:   webhosting.DeleteDatabaseDump,
	}
	dbDumpCmd.AddCommand(dbDumpDeleteCmd)
	dbDumpRestoreCmd := &cobra.Command{
		Use:   "restore <service_name> <name> <id>",
		Short: "Restore from a dump",
		Args:  cobra.ExactArgs(3),
		Run:   webhosting.RestoreDatabaseDump,
	}
	dbDumpCmd.AddCommand(dbDumpRestoreCmd)
	dbCmd.AddCommand(dbDumpCmd)

	dbRestoreCmd := &cobra.Command{
		Use:   "restore <service_name> <name>",
		Short: "Restore database from snapshot date",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.RestoreDatabaseFromDate,
	}
	dbRestoreCmd.Flags().StringVar(&webhosting.DatabaseDumpDate, "date", "", "Dump type to restore (allowed: daily.1, now, weekly.1)")
	dbRestoreCmd.Flags().BoolVar(&webhosting.DatabaseSendEmail, "send-email", false, "Send email when restore completes")
	addFromFileFlag(dbRestoreCmd)
	addInteractiveEditorFlag(dbRestoreCmd)
	dbCmd.AddCommand(dbRestoreCmd)

	dbImportCmd := &cobra.Command{
		Use:   "import <service_name> <name>",
		Short: "Import a database dump",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.ImportDatabaseDump,
	}
	dbImportCmd.Flags().StringVar(&webhosting.DatabaseDocumentID, "document-id", "", "Document ID from /me/documents")
	dbImportCmd.Flags().BoolVar(&webhosting.DatabaseFlush, "flush", false, "Flush database before import")
	dbImportCmd.Flags().BoolVar(&webhosting.DatabaseSendEmail, "send-email", false, "Send email when done")
	addFromFileFlag(dbImportCmd)
	addInteractiveEditorFlag(dbImportCmd)
	dbCmd.AddCommand(dbImportCmd)

	dbStatsCmd := &cobra.Command{
		Use:   "stats <service_name> <name>",
		Short: "Get database statistics",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetDatabaseStatistics,
	}
	dbStatsCmd.Flags().StringVar(&webhosting.DatabaseStatsPeriod, "period", "", "Statistics period (allowed: daily, monthly, weekly, yearly)")
	dbStatsCmd.Flags().StringVar(&webhosting.DatabaseStatsType, "type", "", "Statistics type (allowed: statement, statementMeanTime)")
	dbCmd.AddCommand(withFilterFlag(dbStatsCmd))

	dbActionCmd := &cobra.Command{
		Use:   "request-action <service_name> <name>",
		Short: "Request an action on a database",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.RequestDatabaseAction,
	}
	dbActionCmd.Flags().StringVar(&webhosting.DatabaseAction, "action", "", "Action to request (allowed: CHECK_QUOTA)")
	addFromFileFlag(dbActionCmd)
	addInteractiveEditorFlag(dbActionCmd)
	dbCmd.AddCommand(dbActionCmd)

	webhostingCmd.AddCommand(dbCmd)

	// Email
	emailCmd := &cobra.Command{Use: "email", Short: "Manage automated emails"}
	emailInfoCmd := &cobra.Command{
		Use:   "info <service_name>",
		Short: "Get email sending settings",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.GetEmailInfo,
	}
	emailCmd.AddCommand(emailInfoCmd)
	emailUpdateCmd := &cobra.Command{
		Use:   "update <service_name>",
		Short: "Update email sending settings",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.UpdateEmail,
	}
	emailUpdateCmd.Flags().StringVar(&webhosting.EmailContactAddress, "contact-email", "", "Email used to receive error notifications")
	addInteractiveEditorFlag(emailUpdateCmd)
	emailCmd.AddCommand(emailUpdateCmd)
	emailBouncesCmd := &cobra.Command{
		Use:   "bounces <service_name>",
		Short: "List recent email bounces",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListEmailBounces,
	}
	emailBouncesCmd.Flags().IntVar(&webhosting.EmailBounceLimit, "limit", 20, "Maximum number of bounces to fetch (1-100)")
	emailCmd.AddCommand(emailBouncesCmd)
	emailRequestCmd := &cobra.Command{
		Use:   "request-action <service_name>",
		Short: "Request an email action",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.RequestEmailAction,
	}
	emailRequestCmd.Flags().StringVar(&webhosting.EmailRequestAction, "action", "", "Action to request (allowed: BLOCK, PURGE, UNBLOCK)")
	emailCmd.AddCommand(emailRequestCmd)
	emailVolumesCmd := &cobra.Command{
		Use:   "volumes <service_name>",
		Short: "List email sending volumes",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListEmailVolumes,
	}
	emailCmd.AddCommand(withFilterFlag(emailVolumesCmd))
	webhostingCmd.AddCommand(emailCmd)

	// Email options
	emailOptionCmd := &cobra.Command{Use: "email-option", Short: "Manage email options"}
	emailOptionListCmd := &cobra.Command{
		Use:   "list <service_name>",
		Short: "List email options",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListEmailOptions,
	}
	emailOptionCmd.AddCommand(withFilterFlag(emailOptionListCmd))
	emailOptionGetCmd := &cobra.Command{
		Use:   "get <service_name> <id>",
		Short: "Get an email option",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetEmailOption,
	}
	emailOptionCmd.AddCommand(emailOptionGetCmd)
	emailOptionServiceInfoCmd := &cobra.Command{
		Use:   "service-info <service_name> <id>",
		Short: "Get email option service info",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetEmailOptionServiceInfo,
	}
	emailOptionCmd.AddCommand(emailOptionServiceInfoCmd)
	emailOptionTerminateCmd := &cobra.Command{
		Use:   "terminate <service_name> <id>",
		Short: "Terminate an email option",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.TerminateEmailOption,
	}
	emailOptionCmd.AddCommand(emailOptionTerminateCmd)
	webhostingCmd.AddCommand(emailOptionCmd)

	// Extra SQL options
	extraSQLCmd := &cobra.Command{Use: "extra-sql", Short: "Manage extra SQL options"}
	extraSQLListCmd := &cobra.Command{
		Use:   "list <service_name>",
		Short: "List extra SQL options",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListExtraSqlOptions,
	}
	extraSQLCmd.AddCommand(withFilterFlag(extraSQLListCmd))
	extraSQLGetCmd := &cobra.Command{
		Use:   "get <service_name> <id>",
		Short: "Get an extra SQL option",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetExtraSqlOption,
	}
	extraSQLCmd.AddCommand(extraSQLGetCmd)
	extraSQLDatabasesCmd := &cobra.Command{
		Use:   "databases <service_name> <id>",
		Short: "List databases linked to an extra SQL option",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.ListExtraSqlDatabases,
	}
	extraSQLCmd.AddCommand(withFilterFlag(extraSQLDatabasesCmd))
	extraSQLServiceInfoCmd := &cobra.Command{Use: "service-info", Short: "Manage extra SQL service info"}
	extraSQLServiceInfoGetCmd := &cobra.Command{
		Use:   "get <service_name> <id>",
		Short: "Get extra SQL service information",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetExtraSqlServiceInfo,
	}
	extraSQLServiceInfoCmd.AddCommand(extraSQLServiceInfoGetCmd)
	extraSQLServiceInfoUpdateCmd := &cobra.Command{
		Use:   "update <service_name> <id>",
		Short: "Update extra SQL service information",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.UpdateExtraSqlServiceInfo,
	}
	extraSQLServiceInfoUpdateCmd.Flags().BoolVar(&common.ServiceInfoSpec.Renew.Automatic, "renew-automatic", false, "Enable automatic renewal")
	extraSQLServiceInfoUpdateCmd.Flags().BoolVar(&common.ServiceInfoSpec.Renew.DeleteAtExpiration, "renew-delete-at-expiration", false, "Delete service at expiration")
	extraSQLServiceInfoUpdateCmd.Flags().BoolVar(&common.ServiceInfoSpec.Renew.Forced, "renew-forced", false, "Force renewal")
	extraSQLServiceInfoUpdateCmd.Flags().BoolVar(&common.ServiceInfoSpec.Renew.ManualPayment, "renew-manual-payment", false, "Enable manual payment for renewal")
	extraSQLServiceInfoUpdateCmd.Flags().IntVar(&common.ServiceInfoSpec.Renew.Period, "renew-period", 0, "Renewal period in months")
	addFromFileFlag(extraSQLServiceInfoUpdateCmd)
	addInteractiveEditorFlag(extraSQLServiceInfoUpdateCmd)
	extraSQLServiceInfoCmd.AddCommand(extraSQLServiceInfoUpdateCmd)
	extraSQLCmd.AddCommand(extraSQLServiceInfoCmd)
	extraSQLTerminateCmd := &cobra.Command{
		Use:   "terminate <service_name> <id>",
		Short: "Terminate an extra SQL option",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.TerminateExtraSqlOption,
	}
	extraSQLCmd.AddCommand(extraSQLTerminateCmd)
	webhostingCmd.AddCommand(extraSQLCmd)

	// Env vars
	envCmd := &cobra.Command{Use: "env", Short: "Manage environment variables"}
	envListCmd := &cobra.Command{
		Use:   "list <service_name>",
		Short: "List env vars",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListEnvVars,
	}
	envCmd.AddCommand(withFilterFlag(envListCmd))
	envGetCmd := &cobra.Command{
		Use:   "get <service_name> <key>",
		Short: "Get an env var",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetEnvVar,
	}
	envCmd.AddCommand(envGetCmd)
	envCreateCmd := &cobra.Command{
		Use:   "create <service_name>",
		Short: "Create an env var",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.CreateEnvVar,
	}
	envCreateCmd.Flags().StringVar(&webhosting.EnvVarKey, "key", "", "Variable name")
	envCreateCmd.Flags().StringVar(&webhosting.EnvVarType, "type", "", "Variable type (allowed: integer, password, string)")
	envCreateCmd.Flags().StringVar(&webhosting.EnvVarValue, "value", "", "Variable value")
	addFromFileFlag(envCreateCmd)
	addInteractiveEditorFlag(envCreateCmd)
	envCmd.AddCommand(envCreateCmd)
	envUpdateCmd := &cobra.Command{
		Use:   "update <service_name> <key>",
		Short: "Update an env var",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.UpdateEnvVar,
	}
	envUpdateCmd.Flags().StringVar(&webhosting.EnvVarType, "type", "", "Variable type (allowed: integer, password, string)")
	envUpdateCmd.Flags().StringVar(&webhosting.EnvVarValue, "value", "", "Variable value")
	addInteractiveEditorFlag(envUpdateCmd)
	envCmd.AddCommand(envUpdateCmd)
	envDeleteCmd := &cobra.Command{
		Use:   "delete <service_name> <key>",
		Short: "Delete an env var",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.DeleteEnvVar,
	}
	envCmd.AddCommand(envDeleteCmd)
	webhostingCmd.AddCommand(envCmd)

	// Modules
	moduleCmd := &cobra.Command{Use: "module", Short: "Manage one-click modules"}
	moduleListCmd := &cobra.Command{
		Use:   "list <service_name>",
		Short: "List modules",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListModules,
	}
	moduleCmd.AddCommand(withFilterFlag(moduleListCmd))
	moduleGetCmd := &cobra.Command{
		Use:   "get <service_name> <id>",
		Short: "Get a module",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetModule,
	}
	moduleCmd.AddCommand(moduleGetCmd)
	moduleInstallCmd := &cobra.Command{
		Use:   "install <service_name>",
		Short: "Install a module",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.InstallModule,
	}
	moduleInstallCmd.Flags().IntVar(&webhosting.ModuleID, "module-id", 0, "Module ID")
	moduleInstallCmd.Flags().StringVar(&webhosting.ModuleName, "module-name", "", "Module name (latest version will be selected)")
	moduleInstallCmd.Flags().StringVar(&webhosting.ModuleDomain, "domain", "", "Domain")
	moduleInstallCmd.Flags().StringVar(&webhosting.ModulePath, "path", "", "Install path")
	moduleInstallCmd.Flags().StringVar(&webhosting.ModuleLanguage, "language", "", "Language")
	moduleInstallCmd.Flags().StringVar(&webhosting.ModuleAdmin, "admin", "", "Admin login")
	moduleInstallCmd.Flags().StringVar(&webhosting.ModulePassword, "admin-password", "", "Admin password")
	addFromFileFlag(moduleInstallCmd)
	addInteractiveEditorFlag(moduleInstallCmd)
	moduleCmd.AddCommand(moduleInstallCmd)
	moduleDeleteCmd := &cobra.Command{
		Use:   "delete <service_name> <id>",
		Short: "Delete a module",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.DeleteModule,
	}
	moduleCmd.AddCommand(moduleDeleteCmd)

	moduleCatalogCmd := &cobra.Command{Use: "catalog", Short: "Browse available one-click modules"}
	moduleCatalogListCmd := &cobra.Command{
		Use:   "list",
		Short: "List available modules",
		Args:  cobra.NoArgs,
		Run:   webhosting.ListModuleCatalog,
	}
	moduleCatalogListCmd.Flags().StringVar(&webhosting.ModuleCatalogBranch, "branch", "", "Filter by branch (allowed: old, stable, testing)")
	moduleCatalogListCmd.Flags().StringVar(&webhosting.ModuleCatalogActiveFilter, "active", "", "Filter by active flag (true/false)")
	moduleCatalogListCmd.Flags().StringVar(&webhosting.ModuleCatalogLatestFilter, "latest", "", "Filter by latest flag (true/false)")
	moduleCatalogCmd.AddCommand(withFilterFlag(moduleCatalogListCmd))
	moduleCatalogGetCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get available module details",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.GetModuleCatalog,
	}
	moduleCatalogCmd.AddCommand(moduleCatalogGetCmd)
	moduleCmd.AddCommand(moduleCatalogCmd)
	webhostingCmd.AddCommand(moduleCmd)

	offerCmd := &cobra.Command{Use: "offer", Short: "Inspect hosting offers"}
	offerCapabilitiesCmd := &cobra.Command{
		Use:   "capabilities <offer>",
		Short: "Get offer capabilities",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.GetOfferCapabilities,
	}
	offerCmd.AddCommand(offerCapabilitiesCmd)
	vcsSupportedCmd := &cobra.Command{
		Use:   "vcs-supported",
		Short: "List supported VCS platforms",
		Args:  cobra.NoArgs,
		Run:   webhosting.ListSupportedVcs,
	}
	offerCmd.AddCommand(withFilterFlag(vcsSupportedCmd))
	webhostingCmd.AddCommand(offerCmd)

	abuseCmd := &cobra.Command{
		Use:   "abuse-state <service_name>",
		Short: "Get abuse state",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.GetAbuseState,
	}
	webhostingCmd.AddCommand(abuseCmd)

	ovhConfigCmd := &cobra.Command{Use: "ovh-config", Short: "Manage .ovhconfig settings"}
	ovhConfigListCmd := &cobra.Command{
		Use:   "list <service_name>",
		Short: "List .ovhconfig entries",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListOvhConfigs,
	}
	ovhConfigListCmd.Flags().StringVar(&webhosting.OvhConfigPathFilter, "path", "", "Filter configurations by path")
	ovhConfigListCmd.Flags().BoolVar(&webhosting.OvhConfigHistoricalOnly, "historical", false, "Show only historical configurations")
	ovhConfigCmd.AddCommand(withFilterFlag(ovhConfigListCmd))
	ovhConfigGetCmd := &cobra.Command{
		Use:   "get <service_name> <id>",
		Short: "Get a .ovhconfig entry",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetOvhConfig,
	}
	ovhConfigCmd.AddCommand(ovhConfigGetCmd)
	ovhConfigChangeCmd := &cobra.Command{
		Use:   "change <service_name> <id>",
		Short: "Change a .ovhconfig entry",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.ChangeOvhConfig,
	}
	ovhConfigChangeCmd.Flags().StringVar(&webhosting.OvhConfigEngineName, "engine-name", "", "Engine name")
	ovhConfigChangeCmd.Flags().StringVar(&webhosting.OvhConfigEngineVersion, "engine-version", "", "Engine version")
	ovhConfigChangeCmd.Flags().StringVar(&webhosting.OvhConfigEnvironment, "environment", "", "Environment (production, development, ...)")
	ovhConfigChangeCmd.Flags().StringVar(&webhosting.OvhConfigHTTPFirewall, "http-firewall", "", "HTTP firewall mode (none, security, ...)")
	ovhConfigChangeCmd.Flags().StringVar(&webhosting.OvhConfigContainer, "container", "", "Container image")
	addFromFileFlag(ovhConfigChangeCmd)
	addInteractiveEditorFlag(ovhConfigChangeCmd)
	ovhConfigCmd.AddCommand(ovhConfigChangeCmd)
	ovhConfigRollbackCmd := &cobra.Command{
		Use:   "rollback <service_name> <id>",
		Short: "Rollback a .ovhconfig entry",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.RollbackOvhConfig,
	}
	ovhConfigRollbackCmd.Flags().IntVar(&webhosting.OvhConfigRollbackID, "rollback-id", 0, "Configuration ID to rollback to")
	ovhConfigCmd.AddCommand(ovhConfigRollbackCmd)
	ovhConfigCapabilitiesCmd := &cobra.Command{
		Use:   "capabilities <service_name>",
		Short: "List available versions and containers",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.GetOvhConfigCapabilities,
	}
	ovhConfigCmd.AddCommand(withFilterFlag(ovhConfigCapabilitiesCmd))
	ovhConfigRecommendedCmd := &cobra.Command{
		Use:   "recommended <service_name>",
		Short: "Show recommended values",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.GetOvhConfigRecommendedValues,
	}
	ovhConfigCmd.AddCommand(ovhConfigRecommendedCmd)
	ovhConfigRefreshCmd := &cobra.Command{
		Use:   "refresh <service_name>",
		Short: "Refresh cached .ovhconfig data",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.RefreshOvhConfig,
	}
	ovhConfigCmd.AddCommand(ovhConfigRefreshCmd)
	webhostingCmd.AddCommand(ovhConfigCmd)

	ownLogCmd := &cobra.Command{Use: "own-log", Short: "Manage own logs"}
	ownLogListCmd := &cobra.Command{
		Use:   "list <service_name>",
		Short: "List own logs",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListOwnLogsEntries,
	}
	ownLogCmd.AddCommand(withFilterFlag(ownLogListCmd))
	ownLogGetCmd := &cobra.Command{
		Use:   "get <service_name> <id>",
		Short: "Get an own log entry",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetOwnLog,
	}
	ownLogCmd.AddCommand(ownLogGetCmd)

	ownLogUserCmd := &cobra.Command{Use: "user", Short: "Manage own log users"}
	ownLogUserListCmd := &cobra.Command{
		Use:   "list <service_name> <ownlog_id>",
		Short: "List users for an own log",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.ListOwnLogUsers,
	}
	ownLogUserCmd.AddCommand(withFilterFlag(ownLogUserListCmd))
	ownLogUserGetCmd := &cobra.Command{
		Use:   "get <service_name> <ownlog_id> <login>",
		Short: "Get an own log user",
		Args:  cobra.ExactArgs(3),
		Run:   webhosting.GetOwnLogUser,
	}
	ownLogUserCmd.AddCommand(ownLogUserGetCmd)
	ownLogUserCreateCmd := &cobra.Command{
		Use:   "create <service_name> <ownlog_id>",
		Short: "Create an own log user",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.CreateOwnLogUser,
	}
	ownLogUserCreateCmd.Flags().StringVar(&webhosting.OwnLogUserLogin, "login", "", "User login used to connect to logs.ovh.net")
	ownLogUserCreateCmd.Flags().StringVar(&webhosting.OwnLogUserPassword, "password", "", "User password (required)")
	ownLogUserCreateCmd.Flags().StringVar(&webhosting.OwnLogUserDescription, "description", "", "Description for this user (required)")
	addFromFileFlag(ownLogUserCreateCmd)
	addInteractiveEditorFlag(ownLogUserCreateCmd)
	ownLogUserCmd.AddCommand(ownLogUserCreateCmd)
	ownLogUserUpdateCmd := &cobra.Command{
		Use:   "update <service_name> <ownlog_id> <login>",
		Short: "Update an own log user",
		Args:  cobra.ExactArgs(3),
		Run:   webhosting.UpdateOwnLogUser,
	}
	ownLogUserUpdateCmd.Flags().StringVar(&webhosting.OwnLogUserDescription, "description", "", "User description")
	addInteractiveEditorFlag(ownLogUserUpdateCmd)
	ownLogUserCmd.AddCommand(ownLogUserUpdateCmd)
	ownLogUserDeleteCmd := &cobra.Command{
		Use:   "delete <service_name> <ownlog_id> <login>",
		Short: "Delete an own log user",
		Args:  cobra.ExactArgs(3),
		Run:   webhosting.DeleteOwnLogUser,
	}
	ownLogUserCmd.AddCommand(ownLogUserDeleteCmd)
	ownLogUserPasswordCmd := &cobra.Command{
		Use:   "change-password <service_name> <ownlog_id> <login>",
		Short: "Change an own log user password",
		Args:  cobra.ExactArgs(3),
		Run:   webhosting.ChangeOwnLogUserPassword,
	}
	ownLogUserPasswordCmd.Flags().StringVar(&webhosting.OwnLogUserPassword, "password", "", "New password")
	ownLogUserCmd.AddCommand(ownLogUserPasswordCmd)
	ownLogCmd.AddCommand(ownLogUserCmd)
	webhostingCmd.AddCommand(ownLogCmd)

	requestCmd := &cobra.Command{
		Use:   "request-action <service_name>",
		Short: "Request a hosting operation",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.RequestHostingAction,
	}
	requestCmd.Flags().StringVar(&webhosting.RequestAction, "action", "", "Action to request (allowed: CHECK_QUOTA, FLUSH_CACHE, SCAN_ANTIHACK)")
	webhostingCmd.AddCommand(requestCmd)

	// Runtime
	runtimeCmd := &cobra.Command{Use: "runtime", Short: "Manage runtimes"}
	runtimeListCmd := &cobra.Command{
		Use:   "list <service_name>",
		Short: "List runtimes",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListRuntimes,
	}
	runtimeCmd.AddCommand(withFilterFlag(runtimeListCmd))
	runtimeAvailableTypesCmd := &cobra.Command{
		Use:   "available-types <service_name>",
		Short: "List available runtime backend types",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListRuntimeAvailableTypes,
	}
	runtimeAvailableTypesCmd.Flags().StringVar(&webhosting.RuntimeLanguage, "language", "", "Filter by programming language")
	runtimeCmd.AddCommand(withFilterFlag(runtimeAvailableTypesCmd))
	runtimeGetCmd := &cobra.Command{
		Use:   "get <service_name> <id>",
		Short: "Get a runtime",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetRuntime,
	}
	runtimeCmd.AddCommand(runtimeGetCmd)
	runtimeCreateCmd := &cobra.Command{
		Use:   "create <service_name>",
		Short: "Create a runtime",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.CreateRuntime,
	}
	runtimeCreateCmd.Flags().StringVar(&webhosting.RuntimeName, "name", "", "Runtime name")
	runtimeCreateCmd.Flags().StringVar(&webhosting.RuntimeType, "type", "", "Runtime backend type")
	runtimeCreateCmd.Flags().StringVar(&webhosting.RuntimePublicDir, "public-dir", "", "Public directory")
	runtimeCreateCmd.Flags().StringVar(&webhosting.RuntimeAppEnv, "app-env", "", "Application environment")
	runtimeCreateCmd.Flags().StringVar(&webhosting.RuntimeAppBootstrap, "app-bootstrap", "", "Application bootstrap script")
	runtimeCreateCmd.Flags().BoolVar(&webhosting.RuntimeIsDefault, "runtime-default", false, "Set as default runtime")
	runtimeCreateCmd.Flags().StringSliceVar(&webhosting.RuntimeDomains, "domain", []string{}, "Domains to attach")
	addFromFileFlag(runtimeCreateCmd)
	addInteractiveEditorFlag(runtimeCreateCmd)
	runtimeCmd.AddCommand(runtimeCreateCmd)
	runtimeUpdateCmd := &cobra.Command{
		Use:   "update <service_name> <id>",
		Short: "Update a runtime",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.UpdateRuntime,
	}
	runtimeUpdateCmd.Flags().StringVar(&webhosting.RuntimeName, "name", "", "Runtime name")
	runtimeUpdateCmd.Flags().StringVar(&webhosting.RuntimePublicDir, "public-dir", "", "Public directory")
	runtimeUpdateCmd.Flags().StringVar(&webhosting.RuntimeAppEnv, "app-env", "", "Application environment")
	runtimeUpdateCmd.Flags().StringVar(&webhosting.RuntimeAppBootstrap, "app-bootstrap", "", "Application bootstrap script")
	runtimeUpdateCmd.Flags().BoolVar(&webhosting.RuntimeIsDefault, "runtime-default", false, "Set as default runtime")
	addInteractiveEditorFlag(runtimeUpdateCmd)
	runtimeCmd.AddCommand(runtimeUpdateCmd)
	runtimeDeleteCmd := &cobra.Command{
		Use:   "delete <service_name> <id>",
		Short: "Delete a runtime",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.DeleteRuntime,
	}
	runtimeCmd.AddCommand(runtimeDeleteCmd)
	runtimeDomainsCmd := &cobra.Command{
		Use:   "domains <service_name> <id>",
		Short: "List domains attached to a runtime",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.ListRuntimeDomains,
	}
	runtimeCmd.AddCommand(withFilterFlag(runtimeDomainsCmd))
	webhostingCmd.AddCommand(runtimeCmd)

	// Websites
	websiteCmd := &cobra.Command{Use: "website", Short: "Manage websites deployments"}
	websiteListCmd := &cobra.Command{
		Use:   "list <service_name>",
		Short: "List websites",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListWebsites,
	}
	websiteCmd.AddCommand(withFilterFlag(websiteListCmd))
	websiteGetCmd := &cobra.Command{
		Use:   "get <service_name> <id>",
		Short: "Get a website",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetWebsite,
	}
	websiteCmd.AddCommand(websiteGetCmd)
	websiteCreateCmd := &cobra.Command{
		Use:   "create <service_name>",
		Short: "Create a website",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.CreateWebsite,
	}
	websiteCreateCmd.Flags().StringVar(&webhosting.WebsitePath, "path", "", "Deployment path")
	websiteCreateCmd.Flags().StringVar(&webhosting.WebsiteVcsURL, "vcs-url", "", "Repository URL")
	websiteCreateCmd.Flags().StringVar(&webhosting.WebsiteBranch, "branch", "", "Branch to deploy")
	addFromFileFlag(websiteCreateCmd)
	addInteractiveEditorFlag(websiteCreateCmd)
	websiteCmd.AddCommand(websiteCreateCmd)
	websiteCapabilitiesCmd := &cobra.Command{
		Use:   "creation-capabilities <service_name>",
		Short: "Show website creation capabilities",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.GetWebsiteCreationCapabilities,
	}
	websiteCmd.AddCommand(websiteCapabilitiesCmd)
	websiteUpdateCmd := &cobra.Command{
		Use:   "update <service_name> <id>",
		Short: "Update a website",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.UpdateWebsite,
	}
	websiteUpdateCmd.Flags().StringVar(&webhosting.WebsitePath, "path", "", "Deployment path")
	websiteUpdateCmd.Flags().StringVar(&webhosting.WebsiteVcsURL, "vcs-url", "", "Repository URL")
	websiteUpdateCmd.Flags().StringVar(&webhosting.WebsiteBranch, "branch", "", "Branch to deploy")
	addInteractiveEditorFlag(websiteUpdateCmd)
	websiteCmd.AddCommand(websiteUpdateCmd)
	websiteDeleteCmd := &cobra.Command{
		Use:   "delete <service_name> <id>",
		Short: "Delete a website",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.DeleteWebsite,
	}
	websiteDeleteCmd.Flags().BoolVar(&webhosting.WebsiteDeleteFiles, "delete-files", false, "Also delete files in the website path")
	websiteCmd.AddCommand(websiteDeleteCmd)
	websiteDeployCmd := &cobra.Command{
		Use:   "deploy <service_name> <id>",
		Short: "Trigger a deployment",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.DeployWebsite,
	}
	websiteCmd.AddCommand(websiteDeployCmd)
	websiteDeploymentCmd := &cobra.Command{Use: "deployment", Short: "Manage website deployments"}
	websiteDeploymentListCmd := &cobra.Command{
		Use:   "list <service_name> <id>",
		Short: "List deployments",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.ListWebsiteDeployments,
	}
	websiteDeploymentCmd.AddCommand(withFilterFlag(websiteDeploymentListCmd))
	websiteDeploymentGetCmd := &cobra.Command{
		Use:   "get <service_name> <id> <deployment_id>",
		Short: "Get a deployment",
		Args:  cobra.ExactArgs(3),
		Run:   webhosting.GetWebsiteDeployment,
	}
	websiteDeploymentCmd.AddCommand(websiteDeploymentGetCmd)
	websiteDeploymentLogsCmd := &cobra.Command{
		Use:   "logs <service_name> <id> <deployment_id>",
		Short: "Get deployment logs",
		Args:  cobra.ExactArgs(3),
		Run:   webhosting.GetWebsiteDeploymentLogs,
	}
	websiteDeploymentCmd.AddCommand(websiteDeploymentLogsCmd)
	websiteCmd.AddCommand(websiteDeploymentCmd)
	webhostingCmd.AddCommand(websiteCmd)

	vcsCmd := &cobra.Command{Use: "vcs", Short: "Manage VCS integrations"}
	vcsWebhooksCmd := &cobra.Command{
		Use:   "webhooks <service_name>",
		Short: "Get VCS webhook URLs",
		Long:  "Retrieve webhook URLs to configure on your VCS provider (supported platforms: github).",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.GetVcsWebhooks,
	}
	vcsWebhooksCmd.Flags().StringVar(&webhosting.VcsWebhookPath, "path", "", "Hosting path to filter on (required)")
	vcsWebhooksCmd.Flags().StringVar(&webhosting.VcsWebhookPlatform, "vcs", "", "VCS platform (allowed: github)")
	vcsCmd.AddCommand(vcsWebhooksCmd)
	webhostingCmd.AddCommand(vcsCmd)

	// SSL
	sslCmd := &cobra.Command{Use: "ssl", Short: "Manage SSL"}
	sslGetCmd := &cobra.Command{
		Use:   "get <service_name>",
		Short: "Get SSL info",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.GetSSL,
	}
	sslCmd.AddCommand(sslGetCmd)
	sslCreateCmd := &cobra.Command{
		Use:   "create <service_name>",
		Short: "Create SSL",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.CreateSSL,
	}
	sslCreateCmd.Flags().StringVar(&webhosting.SSLCertificate, "certificate", "", "Certificate")
	sslCreateCmd.Flags().StringVar(&webhosting.SSLChain, "chain", "", "Certificate chain")
	sslCreateCmd.Flags().StringVar(&webhosting.SSLKey, "key", "", "Private key")
	addFromFileFlag(sslCreateCmd)
	addInteractiveEditorFlag(sslCreateCmd)
	sslCmd.AddCommand(sslCreateCmd)
	sslDeleteCmd := &cobra.Command{
		Use:   "delete <service_name>",
		Short: "Delete SSL",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.DeleteSSL,
	}
	sslCmd.AddCommand(sslDeleteCmd)
	sslRegenerateCmd := &cobra.Command{
		Use:   "regenerate <service_name>",
		Short: "Regenerate SSL",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.RegenerateSSL,
	}
	sslCmd.AddCommand(sslRegenerateCmd)
	sslDomainsCmd := &cobra.Command{
		Use:   "domains <service_name>",
		Short: "List SSL domains",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListSSLAttachedDomains,
	}
	sslCmd.AddCommand(withFilterFlag(sslDomainsCmd))
	sslReportCmd := &cobra.Command{
		Use:   "report <service_name>",
		Short: "Get SSL report",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.GetSSLReport,
	}
	sslCmd.AddCommand(sslReportCmd)
	webhostingCmd.AddCommand(sslCmd)

	// CDN
	cdnCmd := &cobra.Command{Use: "cdn", Short: "Manage CDN"}
	cdnGetCmd := &cobra.Command{
		Use:   "get <service_name>",
		Short: "Get CDN info",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.GetCdn,
	}
	cdnCmd.AddCommand(cdnGetCmd)
	cdnDomainCmd := &cobra.Command{Use: "domain", Short: "Manage CDN domains"}
	cdnDomainListCmd := &cobra.Command{
		Use:   "list <service_name>",
		Short: "List CDN domains",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListCdnDomains,
	}
	cdnDomainCmd.AddCommand(withFilterFlag(cdnDomainListCmd))
	cdnDomainGetCmd := &cobra.Command{
		Use:   "get <service_name> <domain>",
		Short: "Get a CDN domain",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetCdnDomain,
	}
	cdnDomainCmd.AddCommand(cdnDomainGetCmd)
	cdnPurgeCmd := &cobra.Command{
		Use:   "purge <service_name> <domain>",
		Short: "Purge CDN domain cache",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.PurgeCdnDomain,
	}
	cdnDomainCmd.AddCommand(cdnPurgeCmd)
	cdnRefreshCmd := &cobra.Command{
		Use:   "refresh <service_name> <domain>",
		Short: "Refresh CDN domain",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.RefreshCdnDomain,
	}
	cdnDomainCmd.AddCommand(cdnRefreshCmd)
	cdnDomainOptionCmd := &cobra.Command{Use: "option", Short: "Manage CDN domain options"}
	cdnDomainOptionListCmd := &cobra.Command{
		Use:   "list <service_name> <domain>",
		Short: "List CDN domain options",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.ListCdnDomainOptions,
	}
	cdnDomainOptionCmd.AddCommand(withFilterFlag(cdnDomainOptionListCmd))
	cdnDomainOptionAddCmd := &cobra.Command{
		Use:   "add <service_name> <domain>",
		Short: "Add a CDN domain option",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.AddCdnDomainOption,
	}
	cdnDomainOptionAddCmd.Flags().StringVar(&webhosting.CdnOptionName, "name", "", "Option name")
	cdnDomainOptionAddCmd.Flags().StringVar(&webhosting.CdnOptionType, "type", "", "Option type")
	cdnDomainOptionAddCmd.Flags().BoolVar(&webhosting.CdnOptionEnabled, "enabled", false, "Enable or disable the option")
	cdnDomainOptionAddCmd.Flags().StringVar(&webhosting.CdnOptionPattern, "pattern", "", "URL pattern for the option")
	cdnDomainOptionAddCmd.Flags().StringVar(&webhosting.CdnOptionConfigDestination, "destination", "", "Destination URL for redirects")
	cdnDomainOptionAddCmd.Flags().BoolVar(&webhosting.CdnOptionConfigFollowURI, "follow-uri", false, "Follow URI on redirects")
	cdnDomainOptionAddCmd.Flags().StringVar(&webhosting.CdnOptionConfigOrigins, "origins", "", "Authorized origins (comma separated)")
	cdnDomainOptionAddCmd.Flags().StringVar(&webhosting.CdnOptionConfigPatternType, "pattern-type", "", "Pattern type")
	cdnDomainOptionAddCmd.Flags().IntVar(&webhosting.CdnOptionConfigPriority, "priority", 0, "Cache rule priority")
	cdnDomainOptionAddCmd.Flags().StringVar(&webhosting.CdnOptionConfigQueryParameter, "query-parameters", "", "Action on query parameters")
	cdnDomainOptionAddCmd.Flags().StringSliceVar(&webhosting.CdnOptionConfigResources, "resource", nil, "Resource URI (repeatable)")
	cdnDomainOptionAddCmd.Flags().IntVar(&webhosting.CdnOptionConfigStatusCode, "status-code", 0, "Redirection HTTP status code")
	cdnDomainOptionAddCmd.Flags().IntVar(&webhosting.CdnOptionConfigTTL, "ttl", 0, "Cache time in seconds")
	addFromFileFlag(cdnDomainOptionAddCmd)
	addInteractiveEditorFlag(cdnDomainOptionAddCmd)
	cdnDomainOptionCmd.AddCommand(cdnDomainOptionAddCmd)
	cdnDomainOptionGetCmd := &cobra.Command{
		Use:   "get <service_name> <domain> <option>",
		Short: "Get CDN domain option details",
		Args:  cobra.ExactArgs(3),
		Run:   webhosting.GetCdnDomainOption,
	}
	cdnDomainOptionCmd.AddCommand(cdnDomainOptionGetCmd)
	cdnDomainOptionUpdateCmd := &cobra.Command{
		Use:   "update <service_name> <domain> <option>",
		Short: "Update a CDN domain option",
		Args:  cobra.ExactArgs(3),
		Run:   webhosting.UpdateCdnDomainOption,
	}
	cdnDomainOptionUpdateCmd.Flags().BoolVar(&webhosting.CdnOptionEnabled, "enabled", false, "Enable or disable the option")
	cdnDomainOptionUpdateCmd.Flags().StringVar(&webhosting.CdnOptionType, "type", "", "Option type")
	cdnDomainOptionUpdateCmd.Flags().StringVar(&webhosting.CdnOptionPattern, "pattern", "", "URL pattern for the option")
	cdnDomainOptionUpdateCmd.Flags().StringVar(&webhosting.CdnOptionConfigDestination, "destination", "", "Destination URL for redirects")
	cdnDomainOptionUpdateCmd.Flags().BoolVar(&webhosting.CdnOptionConfigFollowURI, "follow-uri", false, "Follow URI on redirects")
	cdnDomainOptionUpdateCmd.Flags().StringVar(&webhosting.CdnOptionConfigOrigins, "origins", "", "Authorized origins (comma separated)")
	cdnDomainOptionUpdateCmd.Flags().StringVar(&webhosting.CdnOptionConfigPatternType, "pattern-type", "", "Pattern type")
	cdnDomainOptionUpdateCmd.Flags().IntVar(&webhosting.CdnOptionConfigPriority, "priority", 0, "Cache rule priority")
	cdnDomainOptionUpdateCmd.Flags().StringVar(&webhosting.CdnOptionConfigQueryParameter, "query-parameters", "", "Action on query parameters")
	cdnDomainOptionUpdateCmd.Flags().StringSliceVar(&webhosting.CdnOptionConfigResources, "resource", nil, "Resource URI (repeatable)")
	cdnDomainOptionUpdateCmd.Flags().IntVar(&webhosting.CdnOptionConfigStatusCode, "status-code", 0, "Redirection HTTP status code")
	cdnDomainOptionUpdateCmd.Flags().IntVar(&webhosting.CdnOptionConfigTTL, "ttl", 0, "Cache time in seconds")
	addFromFileFlag(cdnDomainOptionUpdateCmd)
	addInteractiveEditorFlag(cdnDomainOptionUpdateCmd)
	cdnDomainOptionCmd.AddCommand(cdnDomainOptionUpdateCmd)
	cdnDomainOptionDeleteCmd := &cobra.Command{
		Use:   "delete <service_name> <domain> <option>",
		Short: "Delete a CDN domain option",
		Args:  cobra.ExactArgs(3),
		Run:   webhosting.DeleteCdnDomainOption,
	}
	cdnDomainOptionCmd.AddCommand(cdnDomainOptionDeleteCmd)
	cdnDomainCmd.AddCommand(cdnDomainOptionCmd)
	cdnDomainStatsCmd := &cobra.Command{
		Use:   "statistics <service_name> <domain>",
		Short: "Get CDN domain statistics",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetCdnDomainStatistics,
	}
	cdnDomainStatsCmd.Flags().StringVar(&webhosting.CdnStatisticPeriod, "period", "", "Statistics period (default day)")
	cdnDomainCmd.AddCommand(withFilterFlag(cdnDomainStatsCmd))
	cdnCmd.AddCommand(cdnDomainCmd)
	cdnServiceInfoCmd := &cobra.Command{Use: "service-info", Short: "Manage CDN service info"}
	cdnServiceInfoGetCmd := &cobra.Command{
		Use:   "get <service_name>",
		Short: "Get CDN service information",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.GetCdnServiceInfo,
	}
	cdnServiceInfoCmd.AddCommand(cdnServiceInfoGetCmd)
	cdnServiceInfoUpdateCmd := &cobra.Command{
		Use:   "update <service_name>",
		Short: "Update CDN service information",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.UpdateCdnServiceInfo,
	}
	cdnServiceInfoUpdateCmd.Flags().BoolVar(&common.ServiceInfoSpec.Renew.Automatic, "renew-automatic", false, "Enable automatic renewal")
	cdnServiceInfoUpdateCmd.Flags().BoolVar(&common.ServiceInfoSpec.Renew.DeleteAtExpiration, "renew-delete-at-expiration", false, "Delete service at expiration")
	cdnServiceInfoUpdateCmd.Flags().BoolVar(&common.ServiceInfoSpec.Renew.Forced, "renew-forced", false, "Force renewal")
	cdnServiceInfoUpdateCmd.Flags().BoolVar(&common.ServiceInfoSpec.Renew.ManualPayment, "renew-manual-payment", false, "Enable manual payment for renewal")
	cdnServiceInfoUpdateCmd.Flags().IntVar(&common.ServiceInfoSpec.Renew.Period, "renew-period", 0, "Renewal period in months")
	addFromFileFlag(cdnServiceInfoUpdateCmd)
	addInteractiveEditorFlag(cdnServiceInfoUpdateCmd)
	cdnServiceInfoCmd.AddCommand(cdnServiceInfoUpdateCmd)
	cdnCmd.AddCommand(cdnServiceInfoCmd)
	cdnOperationCmd := &cobra.Command{Use: "operation", Short: "Manage CDN operations"}
	cdnOpListCmd := &cobra.Command{
		Use:   "list <service_name>",
		Short: "List CDN operations",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListCdnOperations,
	}
	cdnOperationCmd.AddCommand(withFilterFlag(cdnOpListCmd))
	cdnOpGetCmd := &cobra.Command{
		Use:   "get <service_name> <id>",
		Short: "Get a CDN operation",
		Args:  cobra.ExactArgs(2),
		Run:   webhosting.GetCdnOperation,
	}
	cdnOperationCmd.AddCommand(cdnOpGetCmd)
	cdnCmd.AddCommand(cdnOperationCmd)
	cdnAvailableOptionsCmd := &cobra.Command{
		Use:   "available-options <service_name>",
		Short: "List available CDN options",
		Args:  cobra.ExactArgs(1),
		Run:   webhosting.ListCdnAvailableOptions,
	}
	cdnCmd.AddCommand(withFilterFlag(cdnAvailableOptionsCmd))
	webhostingCmd.AddCommand(cdnCmd)

	// Users
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manage FTP/SSH users",
		Long:  "Create and manage the FTP/SSH users allowed to access your web hosting space.",
	}
	userListCmd := &cobra.Command{