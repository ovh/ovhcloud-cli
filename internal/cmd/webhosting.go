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