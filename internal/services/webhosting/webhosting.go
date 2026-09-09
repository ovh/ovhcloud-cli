// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0
// Crafted with Codex

package webhosting

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/editor"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/ovh/ovhcloud-cli/internal/utils"
	"github.com/spf13/cobra"
)

var (
	webhostingColumnsToDisplay = []string{"serviceName", "displayName", "datacenter", "state"}

	//go:embed templates/webhosting.tmpl
	webhostingTemplate string
	//go:embed templates/attached_domain.tmpl
	attachedDomainTemplate string
	//go:embed templates/attached_domain_dig_status.tmpl
	attachedDomainDigStatusTemplate string
	//go:embed templates/cron.tmpl
	cronTemplate string
	//go:embed templates/database.tmpl
	databaseTemplate string
	//go:embed templates/database_dump.tmpl
	databaseDumpTemplate string
	//go:embed templates/database_available_versions.tmpl
	databaseAvailableVersionsTemplate string
	//go:embed templates/database_copy.tmpl
	databaseCopyTemplate string
	//go:embed templates/database_capabilities.tmpl
	databaseCapabilitiesTemplate string
	//go:embed templates/email.tmpl
	emailTemplate string
	//go:embed templates/email_option.tmpl
	emailOptionTemplate string
	//go:embed templates/email_option_service_info.tmpl
	emailOptionServiceInfoTemplate string
	//go:embed templates/extra_sql_option.tmpl
	extraSqlOptionTemplate string
	//go:embed templates/extra_sql_service_info.tmpl
	extraSqlServiceInfoTemplate string
	//go:embed templates/env_var.tmpl
	envVarTemplate string
	//go:embed templates/module.tmpl
	moduleTemplate string
	//go:embed templates/module_catalog.tmpl
	moduleCatalogTemplate string
	//go:embed templates/offer_capabilities.tmpl
	offerCapabilitiesTemplate string
	//go:embed templates/abuse_state.tmpl
	abuseStateTemplate string
	//go:embed templates/ssl_resource_certificates.tmpl
	sslResourceCertificatesTemplate string
	//go:embed templates/runtime.tmpl
	runtimeTemplate string
	//go:embed templates/website.tmpl
	websiteTemplate string
	//go:embed templates/website_deployment.tmpl
	websiteDeploymentTemplate string
	//go:embed templates/website_creation_capabilities.tmpl
	websiteCreationCapabilitiesTemplate string
	//go:embed templates/cdn_domain.tmpl
	cdnDomainTemplate string
	//go:embed templates/cdn_domain_option.tmpl
	cdnDomainOptionTemplate string
	//go:embed templates/cdn_domain_statistics.tmpl
	cdnDomainStatisticsTemplate string
	//go:embed templates/cdn_operation.tmpl
	cdnOperationTemplate string
	//go:embed templates/cdn.tmpl
	cdnInfoTemplate string
	//go:embed templates/cdn_available_options.tmpl
	cdnAvailableOptionsTemplate string
	//go:embed templates/user.tmpl
	userTemplate string
	//go:embed templates/task.tmpl
	taskTemplate string
	//go:embed templates/boost_history.tmpl
	boostHistoryTemplate string
	//go:embed templates/ovh_config.tmpl
	ovhConfigTemplate string
	//go:embed templates/ovh_config_capabilities.tmpl
	ovhConfigCapabilitiesTemplate string
	//go:embed templates/ovh_config_recommended.tmpl
	ovhConfigRecommendedTemplate string
	//go:embed templates/own_log.tmpl
	ownLogTemplate string
	//go:embed templates/own_log_user.tmpl
	ownLogUserTemplate string
	//go:embed templates/runtime_available_types.tmpl
	runtimeAvailableTypesTemplate string
	//go:embed templates/local_seo_directories.tmpl
	localSeoDirectoriesTemplate string
	//go:embed templates/local_seo_email.tmpl
	localSeoEmailTemplate string
	//go:embed templates/local_seo_visibility_check.tmpl
	localSeoVisibilityCheckTemplate string
	//go:embed templates/local_seo_visibility_result.tmpl
	localSeoVisibilityResultTemplate string
	//go:embed templates/local_seo_account.tmpl
	localSeoAccountTemplate string
	//go:embed templates/local_seo_location.tmpl
	localSeoLocationTemplate string
	//go:embed templates/ssh_key.tmpl
	sshKeyTemplate string
	//go:embed templates/token.tmpl
	tokenTemplate string
	//go:embed templates/vcs_webhooks.tmpl
	vcsWebhooksTemplate string
	//go:embed templates/ssl.tmpl
	sslTemplate string
	//go:embed templates/ssl_service.tmpl
	sslServiceTemplate string
	//go:embed templates/ssl_report.tmpl
	sslReportTemplate string

	WebHostingSpec struct {
		DisplayName *string `json:"displayName,omitempty"`
	}

	WebHostingDisplayName      string
	WebHostingClearDisplayName bool

	// Attached domains
	AttachedDomainDomain     string
	AttachedDomainPath       string
	AttachedDomainRuntimeID  int
	AttachedDomainEnableSSL  bool
	AttachedDomainDisableSSL bool
	AttachedDomainCDN        string
	AttachedDomainFirewall   string
	AttachedDomainIPLocation string
	AttachedDomainOwnLog     string
	AttachedDomainBypassDNS  bool

	// Cron
	CronCommand   string
	CronFrequency string
	CronLanguage  string
	CronEmail     string
	CronDesc      string
	CronStatus    string

	// Database
	DatabaseType             string
	DatabaseCapability       string
	DatabaseUser             string
	DatabasePassword         string
	DatabaseVersion          string
	DatabaseQuota            string
	DatabaseAction           string
	DatabaseDumpDate         string
	DatabaseSendEmail        bool
	DatabaseFlush            bool
	DatabaseDocumentID       string
	DatabaseCopyID           string
	DatabaseStatsPeriod      string
	DatabaseStatsType        string
	DatabaseVersionQueryType string

	// Env vars
	EnvVarKey   string
	EnvVarType  string
	EnvVarValue string

	// Email
	EmailContactAddress string
	EmailBounceLimit    int
	EmailRequestAction  string

	// Modules
	ModuleID                  int
	ModuleName                string
	ModuleDomain              string
	ModulePath                string
	ModuleLanguage            string
	ModuleAdmin               string
	ModulePassword            string
	ModuleCatalogBranch       string
	ModuleCatalogActiveFilter string
	ModuleCatalogLatestFilter string
	OfferName                 string

	// Runtime
	RuntimeName         string
	RuntimeType         string
	RuntimePublicDir    string
	RuntimeAppEnv       string
	RuntimeAppBootstrap string
	RuntimeIsDefault    bool
	RuntimeDomains      []string
	RuntimeLanguage     string

	// Websites
	WebsitePath        string
	WebsiteVcsURL      string
	WebsiteBranch      string
	WebsiteDeleteFiles bool
	WebsiteDeployReset bool

	// Users
	UserLogin          string
	UserPassword       string
	UserHome           string
	UserSSHState       string
	VcsWebhookPath     string
	VcsWebhookPlatform string

	// Local SEO
	LocalSEOCountry            string
	LocalSEOOffer              string
	LocalSEOEmail              string
	LocalSEOName               string
	LocalSEOStreet             string
	LocalSEOZip                string
	LocalSEODirectory          string
	LocalSEOToken              string
	LocalSEOAccountEmailFilter string

	// Misc
	BoostOffer       string
	RestoreBackup    string
	StatisticsPeriod string

	// SSL service-level
	SSLCertificate string
	SSLChain       string
	SSLKey         string
	StatisticsType string

	// OVH Config
	OvhConfigPathFilter     string
	OvhConfigHistoricalOnly bool
	OvhConfigEngineName     string
	OvhConfigEngineVersion  string
	OvhConfigEnvironment    string
	OvhConfigHTTPFirewall   string
	OvhConfigContainer      string
	OvhConfigRollbackID     int

	// Own Logs
	OwnLogUserLogin       string
	OwnLogUserPassword    string
	OwnLogUserDescription string
	OwnLogUserNewPassword string
	RequestAction         string

	SupportedBoostOffers = []string{
		"KS",
		"PERFORMANCE_1",
		"PERFORMANCE_2",
		"PERFORMANCE_3",
		"PERFORMANCE_4",
		"PERSO",
		"PRO",
		"START",
	}

	boostOfferChoiceSet = func() map[string]struct{} {
		m := make(map[string]struct{}, len(SupportedBoostOffers))
		for _, offer := range SupportedBoostOffers {
			m[offer] = struct{}{}
		}
		return m
	}()

	CdnOptionName                 string
	CdnOptionType                 string
	CdnOptionEnabled              bool
	CdnOptionPattern              string
	CdnOptionConfigDestination    string
	CdnOptionConfigFollowURI      bool
	CdnOptionConfigOrigins        string
	CdnOptionConfigPatternType    string
	CdnOptionConfigPriority       int
	CdnOptionConfigQueryParameter string
	CdnOptionConfigResources      []string
	CdnOptionConfigStatusCode     int
	CdnOptionConfigTTL            int

	CdnStatisticPeriod string
)

const defaultRenewExample = `{
  "renew": {
    "automatic": false,
    "deleteAtExpiration": false,
    "forced": false,
    "manualPayment": false,
    "period": 0
  }
}`

const localSeoVisibilityExample = `{
  "country": "",
  "name": "",
  "street": "",
  "zip": ""
}`

const ovhConfigChangeExample = `{
  "engineName": "",
  "engineVersion": "",
  "environment": "",
  "httpFirewall": "",
  "container": ""
}`

const ownLogUserCreateExample = `{
  "login": "",
  "password": "",
  "description": ""
}`

func ListWebHosting(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/hosting/web", "", webhostingColumnsToDisplay, flags.GenericFilters)
}

func GetWebHosting(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/hosting/web", args[0], webhostingTemplate)
}

func EditWebHosting(cmd *cobra.Command, args []string) {
	switch {
	case WebHostingClearDisplayName:
		empty := ""
		WebHostingSpec.DisplayName = &empty
	case cmd.Flags().Changed("display-name"):
		WebHostingSpec.DisplayName = &WebHostingDisplayName
	}

	if err := common.EditResource(
		cmd,
		"/hosting/web/{serviceName}",
		fmt.Sprintf("/v1/hosting/web/%s", url.PathEscape(args[0])),
		WebHostingSpec,
		assets.WebhostingOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func serviceEndpoint(serviceName, suffix string) string {
	if suffix == "" {
		return fmt.Sprintf("/v1/hosting/web/%s", url.PathEscape(serviceName))
	}
	if strings.HasPrefix(suffix, "/") {
		return fmt.Sprintf("/v1/hosting/web/%s%s", url.PathEscape(serviceName), suffix)
	}
	return fmt.Sprintf("/v1/hosting/web/%s/%s", url.PathEscape(serviceName), suffix)
}

func renderDetails(body any) {
	display.OutputWithFormat(&display.OutputMessage{Details: body}, &flags.OutputFormatConfig)
}

// Attached domains
func ListAttachedDomains(cmd *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "attachedDomain")
	common.ManageListRequest(endpoint, "", []string{"domain", "path", "runtimeId", "ssl", "cdn", "firewall"}, flags.GenericFilters)
}

func FindHostingByDomain(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/hosting/web/attachedDomain?domain=%s", url.QueryEscape(args[0]))
	var services []string
	if err := httpLib.Client.Get(endpoint, &services); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to find hosting for domain: %s", err)
		return
	}
	rows := make([]map[string]any, 0, len(services))
	for _, service := range services {
		rows = append(rows, map[string]any{"domain": args[0], "serviceName": service})
	}
	if len(rows) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "ℹ️ No hosting found for domain %s", args[0])
		return
	}
	display.RenderTable(rows, []string{"domain", "serviceName"}, &flags.OutputFormatConfig)
}

func ListAvailableHostingOffers(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/hosting/web/availableOffer?domain=%s", url.QueryEscape(args[0]))
	var offers []string
	if err := httpLib.Client.Get(endpoint, &offers); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch available offers: %s", err)
		return
	}
	if len(offers) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "ℹ️ No offer available for domain %s", args[0])
		return
	}
	rows := make([]map[string]any, 0, len(offers))
	for _, offer := range offers {
		rows = append(rows, map[string]any{"domain": args[0], "offer": offer})
	}
	display.RenderTable(rows, []string{"domain", "offer"}, &flags.OutputFormatConfig)
}

func ListHostingIncidents(_ *cobra.Command, _ []string) {
	var incidents []string
	if err := httpLib.Client.Get("/v1/hosting/web/incident", &incidents); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch incidents: %s", err)
		return
	}
	if len(incidents) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "ℹ️ No ongoing incidents")
		return
	}
	rows := make([]map[string]any, 0, len(incidents))
	for _, incident := range incidents {
		rows = append(rows, map[string]any{"incident": incident})
	}
	common.RenderFilteredTable(rows, []string{"incident"})
}

func GetAttachedDomain(cmd *cobra.Command, args []string) {
	common.ManageObjectRequest(serviceEndpoint(args[0], "attachedDomain"), args[1], attachedDomainTemplate)
}

func GetAttachedDomainDigStatus(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("attachedDomain/%s/digStatus", url.PathEscape(args[1])))
	var body map[string]any
	if err := httpLib.Client.Get(endpoint, &body); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch DNS status: %s", err)
		return
	}
	display.OutputObject(body, args[1], attachedDomainDigStatusTemplate, &flags.OutputFormatConfig)
}

func AddAttachedDomain(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if AttachedDomainDomain != "" {
		params["domain"] = AttachedDomainDomain
	}
	if AttachedDomainPath != "" {
		params["path"] = AttachedDomainPath
	}
	if cmd.Flags().Changed("runtime-id") {
		params["runtimeId"] = AttachedDomainRuntimeID
	}
	if cmd.Flags().Changed("enable-ssl") {
		params["ssl"] = AttachedDomainEnableSSL
	}
	if cmd.Flags().Changed("disable-ssl") {
		params["ssl"] = false
	}
	if AttachedDomainCDN != "" {
		params["cdn"] = AttachedDomainCDN
	}
	if AttachedDomainFirewall != "" {
		params["firewall"] = AttachedDomainFirewall
	}
	if AttachedDomainIPLocation != "" {
		params["ipLocation"] = AttachedDomainIPLocation
	}
	if AttachedDomainOwnLog != "" {
		params["ownLog"] = AttachedDomainOwnLog
	}
	if cmd.Flags().Changed("bypass-dns") {
		params["bypassDNSConfiguration"] = AttachedDomainBypassDNS
	}

	endpoint := serviceEndpoint(args[0], "attachedDomain")
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/attachedDomain",
		endpoint,
		"",
		params,
		assets.WebhostingOpenapiSchema,
		[]string{"domain", "path"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to attach domain: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Domain attached")
}

func UpdateAttachedDomain(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if AttachedDomainPath != "" {
		params["path"] = AttachedDomainPath
	}
	if cmd.Flags().Changed("runtime-id") {
		params["runtimeId"] = AttachedDomainRuntimeID
	}
	if cmd.Flags().Changed("enable-ssl") {
		params["ssl"] = AttachedDomainEnableSSL
	}
	if cmd.Flags().Changed("disable-ssl") {
		params["ssl"] = false
	}
	if AttachedDomainCDN != "" {
		params["cdn"] = AttachedDomainCDN
	}
	if AttachedDomainFirewall != "" {
		params["firewall"] = AttachedDomainFirewall
	}
	if AttachedDomainIPLocation != "" {
		params["ipLocation"] = AttachedDomainIPLocation
	}
	if cmd.Flags().Changed("own-log") {
		params["ownLog"] = AttachedDomainOwnLog
	}
	if cmd.Flags().Changed("bypass-dns") {
		params["bypassDNSConfiguration"] = AttachedDomainBypassDNS
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("attachedDomain/%s", url.PathEscape(args[1])))

	if err := updateResource(cmd, "/hosting/web/{serviceName}/attachedDomain/{domain}", endpoint, params, nil); err != nil {
		if errors.Is(err, errNothingToEdit) {
			display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
			return
		}
		display.OutputError(&flags.OutputFormatConfig, "failed to update attached domain: %s", err)
		return
	}
}

func DeleteAttachedDomain(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("attachedDomain/%s", url.PathEscape(args[1])))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete attached domain: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Domain deleted")
}

func PurgeAttachedDomainCache(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("attachedDomain/%s/purgeCache", url.PathEscape(args[1])))
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to purge domain cache: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Purge triggered")
}

func RestartAttachedDomain(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("attachedDomain/%s/restart", url.PathEscape(args[1])))
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to restart domain: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Restart triggered")
}

// Cron
func ListCrons(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "cron")
	common.ManageListRequest(endpoint, "", []string{"id", "command", "frequency", "language", "status"}, flags.GenericFilters)
}

func GetCron(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(serviceEndpoint(args[0], "cron"), args[1], cronTemplate)
}

func CreateCron(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if CronCommand != "" {
		params["command"] = CronCommand
	}
	if CronFrequency != "" {
		params["frequency"] = CronFrequency
	}
	if CronLanguage != "" {
		params["language"] = CronLanguage
	}
	if CronEmail != "" {
		params["email"] = CronEmail
	}
	if CronDesc != "" {
		params["description"] = CronDesc
	}
	if CronStatus != "" {
		params["status"] = CronStatus
	}

	endpoint := serviceEndpoint(args[0], "cron")
	const defaultCronExample = `{
  "command": "",
  "frequency": "",
  "language": "",
  "email": "",
  "description": "",
  "status": ""
}`
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/cron",
		endpoint,
		defaultCronExample,
		params,
		assets.WebhostingOpenapiSchema,
		[]string{"command", "frequency", "language"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create cron: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Cron created")
}

func UpdateCron(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if CronCommand != "" {
		params["command"] = CronCommand
	}
	if CronFrequency != "" {
		params["frequency"] = CronFrequency
	}
	if CronLanguage != "" {
		params["language"] = CronLanguage
	}
	if CronEmail != "" {
		params["email"] = CronEmail
	}
	if CronDesc != "" {
		params["description"] = CronDesc
	}
	if CronStatus != "" {
		params["status"] = CronStatus
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("cron/%s", url.PathEscape(args[1])))
	if err := updateResource(cmd, "/hosting/web/{serviceName}/cron/{id}", endpoint, params, nil); err != nil {
		if errors.Is(err, errNothingToEdit) {
			display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
			return
		}
		display.OutputError(&flags.OutputFormatConfig, "failed to update cron: %s", err)
		return
	}
}

func DeleteCron(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("cron/%s", url.PathEscape(args[1])))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete cron: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Cron deleted")
}

// Databases
func ListDatabases(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "database")
	body, err := httpLib.FetchExpandedArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch results: %s", err)
		return
	}

	for _, db := range body {
		if formatted, ok := formatQuota(db["quotaUsed"]); ok {
			db["quotaUsed"] = formatted
		}
	}

	body, err = filtersLib.FilterLines(body, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(body, []string{"name", "type", "mode", "state", "quotaUsed"}, &flags.OutputFormatConfig)
}

func ListDatabaseAvailableTypes(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "databaseAvailableType")
	var types []string
	if err := httpLib.Client.Get(endpoint, &types); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch database types: %s", err)
		return
	}

	rows := make([]map[string]any, 0, len(types))
	for _, t := range types {
		rows = append(rows, map[string]any{"type": t})
	}

	common.RenderFilteredTable(rows, []string{"type"})
}

func ListDatabaseAvailableVersions(_ *cobra.Command, args []string) {
	if DatabaseVersionQueryType == "" {
		display.OutputError(&flags.OutputFormatConfig, "flag --type is required to list versions")
		return
	}

	endpoint := fmt.Sprintf(
		"%s?type=%s",
		serviceEndpoint(args[0], "databaseAvailableVersion"),
		url.QueryEscape(DatabaseVersionQueryType),
	)

	var info map[string]any
	if err := httpLib.Client.Get(endpoint, &info); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch versions: %s", err)
		return
	}

	display.OutputObject(info, args[0], databaseAvailableVersionsTemplate, &flags.OutputFormatConfig)
}

func ListDatabaseCreationCapabilities(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "databaseCreationCapabilities")
	var capabilities []map[string]any
	if err := httpLib.Client.Get(endpoint, &capabilities); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch database capabilities: %s", err)
		return
	}

	for _, cap := range capabilities {
		if quota, ok := formatQuota(cap["quota"]); ok {
			cap["quotaDisplay"] = quota
		}
		if engines, ok := stringifyList(cap["engines"]); ok {
			cap["enginesDisplay"] = engines
		}
	}

	common.RenderFilteredTable(capabilities, []string{"type", "isolation", "available", "quotaDisplay quota", "enginesDisplay engines"})
}

func GetDatabase(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(serviceEndpoint(args[0], "database"), args[1], databaseTemplate)
}

func GetDatabaseCapabilities(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("database/%s/capabilities", url.PathEscape(args[1])))
	var capabilities map[string]any
	if err := httpLib.Client.Get(endpoint, &capabilities); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch database capabilities: %s", err)
		return
	}

	display.OutputObject(capabilities, args[1], databaseCapabilitiesTemplate, &flags.OutputFormatConfig)
}

// Email management
func GetEmailInfo(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "email")
	var info map[string]any
	if err := httpLib.Client.Get(endpoint, &info); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch email settings: %s", err)
		return
	}
	display.OutputObject(info, args[0], emailTemplate, &flags.OutputFormatConfig)
}

func UpdateEmail(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if cmd.Flags().Changed("contact-email") {
		params["email"] = EmailContactAddress
	}

	endpoint := serviceEndpoint(args[0], "email")
	if err := updateResource(cmd, "/hosting/web/{serviceName}/email", endpoint, params, nil); err != nil {
		if errors.Is(err, errNothingToEdit) {
			display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
			return
		}
		display.OutputError(&flags.OutputFormatConfig, "failed to update email settings: %s", err)
		return
	}
}

func ListEmailBounces(_ *cobra.Command, args []string) {
	if EmailBounceLimit <= 0 {
		display.OutputError(&flags.OutputFormatConfig, "limit must be greater than zero")
		return
	}
	if EmailBounceLimit > 100 {
		display.OutputError(&flags.OutputFormatConfig, "limit cannot exceed 100")
		return
	}

	endpoint := fmt.Sprintf("%s?limit=%d", serviceEndpoint(args[0], "email/bounces"), EmailBounceLimit)
	var bounces []map[string]any
	if err := httpLib.Client.Get(endpoint, &bounces); err != nil {
		if ovhErr, ok := err.(*ovh.APIError); ok && ovhErr.Code == 404 {
			bounces = []map[string]any{}
		} else {
			display.OutputError(&flags.OutputFormatConfig, "failed to fetch bounces: %s", err)
			return
		}
	}

	for _, bounce := range bounces {
		if val, ok := bounce["date"]; ok {
			bounce["date"] = formatTimestamp(val)
		}
	}

	display.RenderTable(bounces, []string{"date", "to", "message"}, &flags.OutputFormatConfig)
}

func RequestEmailAction(_ *cobra.Command, args []string) {
	if EmailRequestAction == "" {
		display.OutputError(&flags.OutputFormatConfig, "flag --action is required")
		return
	}

	body := map[string]any{"action": EmailRequestAction}
	endpoint := serviceEndpoint(args[0], "email/request")
	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to request email action: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Email action requested")
}

func ListEmailVolumes(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "email/volumes")
	var volumes []map[string]any
	if err := httpLib.Client.Get(endpoint, &volumes); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch email volumes: %s", err)
		return
	}

	for _, entry := range volumes {
		if val, ok := entry["date"]; ok {
			entry["date"] = formatTimestamp(val)
		}
	}

	common.RenderFilteredTable(volumes, []string{"date", "volume"})
}

func ListEmailOptions(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "emailOption")
	ids, err := httpLib.FetchArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch email options: %s", err)
		return
	}

	var rows []map[string]any
	for _, id := range ids {
		rows = append(rows, map[string]any{"id": fmt.Sprintf("%v", id)})
	}

	common.RenderFilteredTable(rows, []string{"id"})
}

func GetEmailOption(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(serviceEndpoint(args[0], "emailOption"), args[1], emailOptionTemplate)
}

func GetEmailOptionServiceInfo(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("emailOption/%s/serviceInfos", url.PathEscape(args[1])))
	var info map[string]any
	if err := httpLib.Client.Get(endpoint, &info); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch service info: %s", err)
		return
	}
	display.OutputObject(info, args[1], emailOptionServiceInfoTemplate, &flags.OutputFormatConfig)
}

func TerminateEmailOption(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("emailOption/%s/terminate", url.PathEscape(args[1])))
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to terminate email option: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Email option termination requested")
}

func ListExtraSqlOptions(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "extraSqlPerso")
	names, err := httpLib.FetchArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch extra SQL options: %s", err)
		return
	}

	var rows []map[string]any
	for _, value := range names {
		rows = append(rows, map[string]any{"id": fmt.Sprintf("%v", value)})
	}

	common.RenderFilteredTable(rows, []string{"id"})
}

func GetExtraSqlOption(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(serviceEndpoint(args[0], "extraSqlPerso"), args[1], extraSqlOptionTemplate)
}

func ListExtraSqlDatabases(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("extraSqlPerso/%s/databases", url.PathEscape(args[1])))
	databases, err := httpLib.FetchArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch linked databases: %s", err)
		return
	}

	var rows []map[string]any
	for _, value := range databases {
		rows = append(rows, map[string]any{"database": fmt.Sprintf("%v", value)})
	}

	common.RenderFilteredTable(rows, []string{"database"})
}

func GetExtraSqlServiceInfo(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("extraSqlPerso/%s/serviceInfos", url.PathEscape(args[1])))
	var info map[string]any
	if err := httpLib.Client.Get(endpoint, &info); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch service info: %s", err)
		return
	}
	display.OutputObject(info, args[1], extraSqlServiceInfoTemplate, &flags.OutputFormatConfig)
}

func UpdateExtraSqlServiceInfo(cmd *cobra.Command, args []string) {
	payload := buildServiceInfoRenewPayload(cmd)
	if len(payload) == 0 && !flags.ParametersViaEditor && flags.ParametersFile == "" {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
		return
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("extraSqlPerso/%s/serviceInfosUpdate", url.PathEscape(args[1])))
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/extraSqlPerso/{name}/serviceInfosUpdate",
		endpoint,
		defaultRenewExample,
		payload,
		assets.WebhostingOpenapiSchema,
		[]string{"renew"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to update extra SQL service info: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Extra SQL service info updated")
}

func TerminateExtraSqlOption(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("extraSqlPerso/%s/terminate", url.PathEscape(args[1])))
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to terminate extra SQL option: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Extra SQL option termination requested")
}

func CreateDatabase(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if DatabaseCapability != "" {
		params["capabilitie"] = DatabaseCapability
	}
	if DatabaseType != "" {
		params["type"] = DatabaseType
	}
	if DatabaseUser != "" {
		params["user"] = DatabaseUser
	}
	if DatabasePassword != "" {
		params["password"] = DatabasePassword
	}
	if DatabaseVersion != "" {
		params["version"] = DatabaseVersion
	}
	if DatabaseQuota != "" {
		params["quota"] = DatabaseQuota
	}

	endpoint := serviceEndpoint(args[0], "database")
	const defaultDatabaseExample = `{
  "capabilitie": "",
  "type": "",
  "user": "",
  "password": "",
  "version": "",
  "quota": ""
}`
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/database",
		endpoint,
		defaultDatabaseExample,
		params,
		assets.WebhostingOpenapiSchema,
		[]string{"capabilitie", "type", "user"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create database: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Database creation requested")
}

func DeleteDatabase(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("database/%s", url.PathEscape(args[1])))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete database: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Database deletion requested")
}

func ChangeDatabasePassword(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("database/%s/changePassword", url.PathEscape(args[1])))
	body := map[string]any{"password": DatabasePassword}
	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to change password: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Password updated")
}

func ListDatabaseCopies(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("database/%s/copy", url.PathEscape(args[1])))
	common.ManageListRequest(endpoint, "", []string{"id", "status"}, flags.GenericFilters)
}

func GetDatabaseCopy(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(
		serviceEndpoint(args[0], fmt.Sprintf("database/%s/copy", url.PathEscape(args[1]))),
		args[2],
		databaseCopyTemplate,
	)
}

func CreateDatabaseCopy(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("database/%s/copy", url.PathEscape(args[1])))
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to request database copy: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Copy requested")
}

func DeleteDatabaseCopy(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("database/%s/copy/%s", url.PathEscape(args[1]), url.PathEscape(args[2])))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete database copy: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Copy deleted")
}

func RestoreDatabaseCopy(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if DatabaseCopyID != "" {
		params["copyId"] = DatabaseCopyID
	}
	if cmd.Flags().Changed("flush") {
		params["flushDatabase"] = DatabaseFlush
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("database/%s/copyRestore", url.PathEscape(args[1])))
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/database/{name}/copyRestore",
		endpoint,
		"",
		params,
		assets.WebhostingOpenapiSchema,
		[]string{"copyId"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to restore copy: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Restore requested")
}

func ListDatabaseDumps(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("database/%s/dump", url.PathEscape(args[1])))
	dumps, err := httpLib.FetchExpandedArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch results: %s", err)
		return
	}

	var downloadLinks []string
	for _, dump := range dumps {
		if val, ok := dump["creationDate"]; ok {
			dump["date"] = val
		}
		if link, ok := dump["url"].(string); ok && len(link) > 0 {
			downloadLinks = append(downloadLinks, link)
			if len(link) > 60 {
				dump["urlPreview"] = safeTableString(link[:60] + "…")
			} else {
				dump["urlPreview"] = safeTableString(link)
			}
		}
	}

	dumps, err = filtersLib.FilterLines(dumps, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(dumps, []string{"id", "date", "type", "status", "urlPreview preview"}, &flags.OutputFormatConfig)

	if len(downloadLinks) > 0 && !flags.OutputFormatConfig.IsJson() && !flags.OutputFormatConfig.IsYaml() && flags.OutputFormatConfig.CustomFormat() == "" {
		var builder strings.Builder
		builder.WriteString("\nFull download URLs:\n")
		for i, link := range downloadLinks {
			builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, link))
		}
		display.OutputInfo(&flags.OutputFormatConfig, nil, "%s", safeMessageString(builder.String()))
	}
}

func GetDatabaseDump(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(
		args[0],
		fmt.Sprintf("database/%s/dump/%s", url.PathEscape(args[1]), url.PathEscape(args[2])),
	)

	var dump map[string]any
	if err := httpLib.Client.Get(endpoint, &dump); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch dump: %s", err)
		return
	}

	var fullURL string
	if link, ok := dump["url"].(string); ok && link != "" {
		fullURL = link
		if len(link) > 60 {
			dump["urlShort"] = safeTableString(link[:60] + "…")
		} else {
			dump["urlShort"] = safeTableString(link)
		}
	}

	display.OutputObject(dump, args[2], databaseDumpTemplate, &flags.OutputFormatConfig)

	if fullURL != "" && !flags.OutputFormatConfig.IsJson() && !flags.OutputFormatConfig.IsYaml() && flags.OutputFormatConfig.CustomFormat() == "" {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "%s", safeMessageString("Download URL: "+fullURL))
	}
}

func GetDatabaseStatistics(cmd *cobra.Command, args []string) {
	if DatabaseStatsPeriod == "" || DatabaseStatsType == "" {
		display.OutputError(&flags.OutputFormatConfig, "both --period and --type are required")
		return
	}

	endpoint := serviceEndpoint(
		args[0],
		fmt.Sprintf("database/%s/statistics", url.PathEscape(args[1])),
	)
	endpoint = fmt.Sprintf(
		"%s?period=%s&type=%s",
		endpoint,
		url.QueryEscape(DatabaseStatsPeriod),
		url.QueryEscape(DatabaseStatsType),
	)

	var stats []map[string]any
	if err := httpLib.Client.Get(endpoint, &stats); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch statistics: %s", err)
		return
	}

	rows := make([]map[string]any, 0, len(stats))
	for _, point := range stats {
		row := map[string]any{}
		if ts, ok := point["timestamp"]; ok {
			row["timestamp"] = formatTimestamp(ts)
		}
		if val, ok := point["value"]; ok {
			row["value"] = val
		}
		rows = append(rows, row)
	}

	common.RenderFilteredTable(rows, []string{"timestamp", "value"})
}

func RequestDatabaseDump(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if DatabaseDumpDate != "" {
		params["date"] = DatabaseDumpDate
	}
	if cmd.Flags().Changed("send-email") {
		params["sendEmail"] = DatabaseSendEmail
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("database/%s/dump", url.PathEscape(args[1])))
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/database/{name}/dump",
		endpoint,
		"",
		params,
		assets.WebhostingOpenapiSchema,
		[]string{"date"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to request dump: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Dump requested")
}

func DeleteDatabaseDump(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("database/%s/dump/%s", url.PathEscape(args[1]), url.PathEscape(args[2])))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete dump: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Dump deleted")
}

func RestoreDatabaseDump(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("database/%s/dump/%s/restore", url.PathEscape(args[1]), url.PathEscape(args[2])))
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to request dump restore: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Restore requested")
}

func RestoreDatabaseFromDate(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if DatabaseDumpDate != "" {
		params["date"] = DatabaseDumpDate
	}
	if cmd.Flags().Changed("send-email") {
		params["sendEmail"] = DatabaseSendEmail
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("database/%s/restore", url.PathEscape(args[1])))
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/database/{name}/restore",
		endpoint,
		"",
		params,
		assets.WebhostingOpenapiSchema,
		[]string{"date"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to request restore: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Restore requested")
}

func ImportDatabaseDump(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if DatabaseDocumentID != "" {
		params["documentId"] = DatabaseDocumentID
	}
	if cmd.Flags().Changed("flush") {
		params["flushDatabase"] = DatabaseFlush
	}
	if cmd.Flags().Changed("send-email") {
		params["sendEmail"] = DatabaseSendEmail
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("database/%s/import", url.PathEscape(args[1])))
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/database/{name}/import",
		endpoint,
		"",
		params,
		assets.WebhostingOpenapiSchema,
		[]string{"documentId"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to import dump: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Import requested")
}

func RequestDatabaseAction(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if DatabaseAction != "" {
		params["action"] = DatabaseAction
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("database/%s/request", url.PathEscape(args[1])))
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/database/{name}/request",
		endpoint,
		"",
		params,
		assets.WebhostingOpenapiSchema,
		[]string{"action"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to request action: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Action requested")
}

// Environment variables
func ListEnvVars(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "envVar")
	common.ManageListRequest(endpoint, "", []string{"key", "type", "value"}, flags.GenericFilters)
}

func GetEnvVar(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(serviceEndpoint(args[0], "envVar"), args[1], envVarTemplate)
}

func CreateEnvVar(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if EnvVarKey != "" {
		params["key"] = EnvVarKey
	}
	if EnvVarType != "" {
		params["type"] = EnvVarType
	}
	if EnvVarValue != "" {
		params["value"] = EnvVarValue
	}

	endpoint := serviceEndpoint(args[0], "envVar")
	const defaultEnvVarExample = `{
  "key": "",
  "type": "",
  "value": ""
}`
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/envVar",
		endpoint,
		defaultEnvVarExample,
		params,
		assets.WebhostingOpenapiSchema,
		[]string{"key", "type", "value"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create env var: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Env var created")
}

func UpdateEnvVar(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if EnvVarType != "" {
		params["type"] = EnvVarType
	}
	if EnvVarValue != "" {
		params["value"] = EnvVarValue
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("envVar/%s", url.PathEscape(args[1])))
	if err := updateResource(cmd, "/hosting/web/{serviceName}/envVar/{key}", endpoint, params, nil); err != nil {
		if errors.Is(err, errNothingToEdit) {
			display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
			return
		}
		display.OutputError(&flags.OutputFormatConfig, "failed to update env var: %s", err)
		return
	}
}

func DeleteEnvVar(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("envVar/%s", url.PathEscape(args[1])))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete env var: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Env var deleted")
}

// Modules
func ListModules(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "module")
	body, err := httpLib.FetchExpandedArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch modules: %s", err)
		return
	}

	body, err = filtersLib.FilterLines(body, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	moduleIDs := collectModuleIDs(body)
	catalog, err := fetchModuleInfos(moduleIDs)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch module catalog: %s", err)
		return
	}

	for _, module := range body {
		enrichModuleWithCatalog(module, catalog)
	}

	columns := []string{
		"id Install ID",
		"moduleName Module",
		"moduleId Module ID",
		"status Status",
		"targetUrl Target",
		"path Path",
		"language Language",
		"adminName Admin",
	}
	display.RenderTable(body, columns, &flags.OutputFormatConfig)
}

func GetModule(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("module/%s", url.PathEscape(args[1])))

	var module map[string]any
	if err := httpLib.Client.Get(endpoint, &module); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch module: %s", err)
		return
	}

	if moduleID, ok := toInt(module["moduleId"]); ok {
		if details, err := fetchModuleInfos([]int{moduleID}); err == nil {
			enrichModuleWithCatalog(module, details)
		}
	}

	display.OutputObject(module, args[1], moduleTemplate, &flags.OutputFormatConfig)
}

func InstallModule(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	moduleID, err := resolveModuleID()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	params["moduleId"] = moduleID
	if ModuleDomain != "" {
		params["domain"] = ModuleDomain
	}
	if ModulePath != "" {
		params["path"] = ModulePath
	}
	if ModuleLanguage != "" {
		params["language"] = ModuleLanguage
	}
	if ModuleAdmin != "" {
		params["adminName"] = ModuleAdmin
	}
	if ModulePassword != "" {
		params["adminPassword"] = ModulePassword
	}

	endpoint := serviceEndpoint(args[0], "module")
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/module",
		endpoint,
		"",
		params,
		assets.WebhostingOpenapiSchema,
		[]string{"moduleId"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to install module: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Module installation requested")
}

func DeleteModule(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("module/%s", url.PathEscape(args[1])))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete module: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Module deletion requested")
}

func ListModuleCatalog(_ *cobra.Command, _ []string) {
	query := url.Values{}
	if ModuleCatalogBranch != "" {
		query.Set("branch", ModuleCatalogBranch)
	}
	if ModuleCatalogActiveFilter != "" {
		query.Set("active", ModuleCatalogActiveFilter)
	} else {
		query.Set("active", "true")
	}
	if ModuleCatalogLatestFilter != "" {
		query.Set("latest", ModuleCatalogLatestFilter)
	}

	modules, err := fetchModuleCatalogWithFilters(query)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to list available modules: %s", err)
		return
	}

	common.RenderFilteredTable(modules, []string{"id", "name", "branch", "version", "active", "latest"})
}

func GetModuleCatalog(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/hosting/web/moduleList/%s", url.PathEscape(args[0]))
	var module map[string]any
	if err := httpLib.Client.Get(endpoint, &module); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch module catalog entry: %s", err)
		return
	}
	display.OutputObject(module, args[0], moduleCatalogTemplate, &flags.OutputFormatConfig)
}

func GetOfferCapabilities(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/hosting/web/offerCapabilities?offer=%s", url.QueryEscape(args[0]))
	var capabilities map[string]any
	if err := httpLib.Client.Get(endpoint, &capabilities); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch offer capabilities: %s", err)
		return
	}
	display.OutputObject(capabilities, args[0], offerCapabilitiesTemplate, &flags.OutputFormatConfig)
}

func ListSupportedVcs(_ *cobra.Command, _ []string) {
	var vcs []string
	if err := httpLib.Client.Get("/v1/hosting/web/vcs/supported", &vcs); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch supported VCS platforms: %s", err)
		return
	}
	rows := make([]map[string]any, 0, len(vcs))
	for _, platform := range vcs {
		rows = append(rows, map[string]any{"platform": platform})
	}
	if len(rows) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "ℹ️ No VCS platform returned")
		return
	}
	common.RenderFilteredTable(rows, []string{"platform"})
}

func GetAbuseState(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "abuseState")
	var state map[string]any
	if err := httpLib.Client.Get(endpoint, &state); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch abuse state: %s", err)
		return
	}
	display.OutputObject(state, args[0], abuseStateTemplate, &flags.OutputFormatConfig)
}

func collectModuleIDs(modules []map[string]any) []int {
	var ids []int
	for _, module := range modules {
		if id, ok := toInt(module["moduleId"]); ok {
			ids = append(ids, id)
		}
	}
	return ids
}

func fetchModuleInfos(moduleIDs []int) (map[int]map[string]any, error) {
	catalog := make(map[int]map[string]any, len(moduleIDs))
	seen := make(map[int]struct{}, len(moduleIDs))
	for _, id := range moduleIDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}

		endpoint := fmt.Sprintf("/v1/hosting/web/moduleList/%d", id)
		var moduleInfo map[string]any
		if err := httpLib.Client.Get(endpoint, &moduleInfo); err != nil {
			return nil, fmt.Errorf("failed to fetch module catalog entry %d: %w", id, err)
		}
		catalog[id] = moduleInfo
	}
	return catalog, nil
}

func fetchFullModuleCatalog() (map[int]map[string]any, error) {
	items, err := httpLib.FetchExpandedArray("/v1/hosting/web/moduleList", "")
	if err != nil {
		return nil, err
	}

	catalog := make(map[int]map[string]any, len(items))
	for _, entry := range items {
		if id, ok := toInt(entry["id"]); ok {
			catalog[id] = entry
		}
	}
	return catalog, nil
}

func fetchModuleCatalogWithFilters(query url.Values) ([]map[string]any, error) {
	path := "/v1/hosting/web/moduleList"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	ids, err := httpLib.FetchArray(path, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch module ids: %w", err)
	}

	objects, err := httpLib.FetchObjectsParallel[map[string]any]("/v1/hosting/web/moduleList/%s", ids, flags.IgnoreErrors)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch module catalog entries: %w", err)
	}

	results := make([]map[string]any, 0, len(objects))
	for _, obj := range objects {
		if obj != nil {
			results = append(results, obj)
		}
	}
	return results, nil
}

func enrichModuleWithCatalog(module map[string]any, catalog map[int]map[string]any) {
	id, ok := toInt(module["moduleId"])
	if !ok {
		return
	}
	info, ok := catalog[id]
	if !ok {
		return
	}

	if name, exists := info["name"]; exists {
		module["moduleName"] = name
	}
	if version, exists := info["version"]; exists {
		module["moduleVersion"] = version
	}
	if branch, exists := info["branch"]; exists {
		module["moduleBranch"] = branch
	}
}

func resolveModuleID() (int, error) {
	if ModuleID != 0 {
		return ModuleID, nil
	}
	if ModuleName == "" {
		return 0, errors.New("either --module-id or --module-name must be provided")
	}

	catalog, err := fetchFullModuleCatalog()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch module catalog: %w", err)
	}

	var chosen map[string]any
	for _, entry := range catalog {
		name, _ := entry["name"].(string)
		if strings.EqualFold(name, ModuleName) {
			chosen = entry
			if latest, ok := entry["latest"].(bool); ok && latest {
				break
			}
		}
	}

	if chosen == nil {
		return 0, fmt.Errorf("module %q not found. Please provide a valid --module-id or another name", ModuleName)
	}

	id, ok := toInt(chosen["id"])
	if !ok {
		return 0, fmt.Errorf("invalid module id for %q", ModuleName)
	}

	return id, nil
}

func toInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float32:
		return int(v), true
	case float64:
		return int(v), true
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return int(i), true
	default:
		return 0, false
	}
}

// OVH Config helpers and commands
func buildOvhConfigQuery() url.Values {
	values := url.Values{}
	if OvhConfigPathFilter != "" {
		values.Set("path", OvhConfigPathFilter)
	}
	if OvhConfigHistoricalOnly {
		values.Set("historical", strconv.FormatBool(true))
	}
	return values
}

func ListOvhConfigs(_ *cobra.Command, args []string) {
	baseEndpoint := serviceEndpoint(args[0], "ovhConfig")
	listEndpoint := baseEndpoint
	if query := buildOvhConfigQuery(); len(query) > 0 {
		listEndpoint = fmt.Sprintf("%s?%s", baseEndpoint, query.Encode())
	}

	ids, err := httpLib.FetchArray(listEndpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch configurations: %s", err)
		return
	}

	objects, err := httpLib.FetchObjectsParallel[map[string]any](baseEndpoint+"/%s", ids, flags.IgnoreErrors)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch configuration details: %s", err)
		return
	}

	objects, err = filtersLib.FilterLines(objects, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	columns := []string{"id", "path", "engineName Engine", "engineVersion Version", "environment", "httpFirewall", "status"}
	sort.Slice(objects, func(i, j int) bool {
		idI, okI := toInt(objects[i]["id"])
		idJ, okJ := toInt(objects[j]["id"])
		if !okI || !okJ {
			return i < j
		}
		return idI < idJ
	})
	display.RenderTable(objects, columns, &flags.OutputFormatConfig)
}

func GetOvhConfig(_ *cobra.Command, args []string) {
	cfg, err := fetchOvhConfig(args[0], args[1])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch configuration: %s", err)
		return
	}
	display.OutputObject(cfg, args[1], ovhConfigTemplate, &flags.OutputFormatConfig)
}

func fetchOvhConfig(serviceName, configID string) (map[string]any, error) {
	endpoint := serviceEndpoint(serviceName, fmt.Sprintf("ovhConfig/%s", url.PathEscape(configID)))
	var cfg map[string]any
	if err := httpLib.Client.Get(endpoint, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func ChangeOvhConfig(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if OvhConfigEngineName != "" {
		params["engineName"] = OvhConfigEngineName
	}
	if OvhConfigEngineVersion != "" {
		params["engineVersion"] = OvhConfigEngineVersion
	}
	if OvhConfigEnvironment != "" {
		params["environment"] = OvhConfigEnvironment
	}
	if OvhConfigHTTPFirewall != "" {
		params["httpFirewall"] = OvhConfigHTTPFirewall
	}
	if OvhConfigContainer != "" {
		params["container"] = OvhConfigContainer
	}

	if len(params) == 0 && !flags.ParametersViaEditor && flags.ParametersFile == "" && !utils.IsInputFromPipe() {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to apply")
		return
	}

	defaultExample := ovhConfigChangeExample
	if flags.ParametersViaEditor {
		if cfg, err := fetchOvhConfig(args[0], args[1]); err == nil {
			editorPreset := map[string]any{}
			copyField := func(key string) {
				if val, ok := cfg[key]; ok && val != nil {
					editorPreset[key] = val
				}
			}
			copyField("engineName")
			copyField("engineVersion")
			copyField("environment")
			copyField("httpFirewall")
			copyField("container")

			if len(editorPreset) > 0 {
				if content, err := json.MarshalIndent(editorPreset, "", "  "); err == nil {
					defaultExample = string(content)
				}
			}
		}
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("ovhConfig/%s/changeConfiguration", url.PathEscape(args[1])))
	result, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/ovhConfig/{id}/changeConfiguration",
		endpoint,
		defaultExample,
		params,
		assets.WebhostingOpenapiSchema,
		nil,
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to change configuration: %s", err)
		return
	}
	display.OutputObject(result, args[1], taskTemplate, &flags.OutputFormatConfig)
}

func RollbackOvhConfig(_ *cobra.Command, args []string) {
	if OvhConfigRollbackID == 0 {
		display.OutputError(&flags.OutputFormatConfig, "flag --rollback-id is required")
		return
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("ovhConfig/%s/rollback", url.PathEscape(args[1])))
	body := map[string]any{"rollbackId": OvhConfigRollbackID}

	var task map[string]any
	if err := httpLib.Client.Post(endpoint, body, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to rollback configuration: %s", err)
		return
	}
	display.OutputObject(task, args[1], taskTemplate, &flags.OutputFormatConfig)
}

func GetOvhConfigCapabilities(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "ovhConfigCapabilities")
	var capabilities []map[string]any
	if err := httpLib.Client.Get(endpoint, &capabilities); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch capabilities: %s", err)
		return
	}

	// Filtered before the wrapping, not after: what --filter selects is an
	// entry, and once the list is inside an object there are no rows left to
	// select from.
	filtered, ok := common.FilteredRows(capabilities)
	if !ok {
		return
	}

	payload := map[string]any{"entries": filtered}
	display.OutputObject(payload, args[0], ovhConfigCapabilitiesTemplate, &flags.OutputFormatConfig)
}

func GetOvhConfigRecommendedValues(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "ovhConfigRecommendedValues")
	var values map[string]any
	if err := httpLib.Client.Get(endpoint, &values); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch recommended values: %s", err)
		return
	}
	display.OutputObject(values, args[0], ovhConfigRecommendedTemplate, &flags.OutputFormatConfig)
}

func RefreshOvhConfig(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "ovhConfigRefresh")
	var task map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to refresh configuration cache: %s", err)
		return
	}
	display.OutputObject(task, args[0], taskTemplate, &flags.OutputFormatConfig)
}

func RequestHostingAction(_ *cobra.Command, args []string) {
	if RequestAction == "" {
		display.OutputError(&flags.OutputFormatConfig, "flag --action is required (allowed: CHECK_QUOTA, FLUSH_CACHE, SCAN_ANTIHACK)")
		return
	}

	endpoint := serviceEndpoint(args[0], "request")
	body := map[string]any{"action": RequestAction}
	var task map[string]any
	if err := httpLib.Client.Post(endpoint, body, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to request hosting action: %s", err)
		return
	}
	display.OutputObject(task, args[0], taskTemplate, &flags.OutputFormatConfig)
}

// Own logs
func ListOwnLogsEntries(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "ownLogs")
	ids, err := httpLib.FetchArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch own logs: %s", err)
		return
	}

	objects, err := httpLib.FetchObjectsParallel[map[string]any](endpoint+"/%s", ids, flags.IgnoreErrors)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch own log details: %s", err)
		return
	}

	objects, err = filtersLib.FilterLines(objects, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	sort.Slice(objects, func(i, j int) bool {
		idI, okI := toInt(objects[i]["id"])
		idJ, okJ := toInt(objects[j]["id"])
		if !okI || !okJ {
			return i < j
		}
		return idI < idJ
	})

	display.RenderTable(objects, []string{"id", "fqdn", "status", "logs", "stats"}, &flags.OutputFormatConfig)
}

func GetOwnLog(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "ownLogs")
	common.ManageObjectRequest(endpoint, args[1], ownLogTemplate)
}

func ListOwnLogUsers(_ *cobra.Command, args []string) {
	baseEndpoint := serviceEndpoint(args[0], fmt.Sprintf("ownLogs/%s/userLogs", url.PathEscape(args[1])))
	logins, err := httpLib.FetchArray(baseEndpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch user logs: %s", err)
		return
	}

	objects, err := httpLib.FetchObjectsParallel[map[string]any](baseEndpoint+"/%s", logins, flags.IgnoreErrors)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch user log details: %s", err)
		return
	}

	objects, err = filtersLib.FilterLines(objects, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	sort.Slice(objects, func(i, j int) bool {
		loginI, _ := objects[i]["login"].(string)
		loginJ, _ := objects[j]["login"].(string)
		return loginI < loginJ
	})

	display.RenderTable(objects, []string{"login", "description", "status", "creationDate"}, &flags.OutputFormatConfig)
}

func GetOwnLogUser(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("ownLogs/%s/userLogs", url.PathEscape(args[1])))
	common.ManageObjectRequest(endpoint, args[2], ownLogUserTemplate)
}

func CreateOwnLogUser(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if OwnLogUserLogin != "" {
		params["login"] = OwnLogUserLogin
	}
	if OwnLogUserPassword != "" {
		params["password"] = OwnLogUserPassword
	}
	if OwnLogUserDescription != "" {
		params["description"] = OwnLogUserDescription
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("ownLogs/%s/userLogs", url.PathEscape(args[1])))
	result, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/ownLogs/{id}/userLogs",
		endpoint,
		ownLogUserCreateExample,
		params,
		assets.WebhostingOpenapiSchema,
		[]string{"login", "password"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create user log: %s", err)
		return
	}
	display.OutputObject(result, args[1], ownLogUserTemplate, &flags.OutputFormatConfig)
}

func UpdateOwnLogUser(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if OwnLogUserDescription != "" {
		params["description"] = OwnLogUserDescription
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("ownLogs/%s/userLogs/%s", url.PathEscape(args[1]), url.PathEscape(args[2])))
	if err := updateResource(cmd, "/hosting/web/{serviceName}/ownLogs/{id}/userLogs/{login}", endpoint, params, nil); err != nil {
		if errors.Is(err, errNothingToEdit) {
			display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
			return
		}
		display.OutputError(&flags.OutputFormatConfig, "failed to update user log: %s", err)
		return
	}
}

func DeleteOwnLogUser(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("ownLogs/%s/userLogs/%s", url.PathEscape(args[1]), url.PathEscape(args[2])))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete user log: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ User log deleted")
}

func ChangeOwnLogUserPassword(_ *cobra.Command, args []string) {
	if OwnLogUserPassword == "" {
		display.OutputError(&flags.OutputFormatConfig, "flag --password is required")
		return
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("ownLogs/%s/userLogs/%s/changePassword", url.PathEscape(args[1]), url.PathEscape(args[2])))
	body := map[string]any{"password": OwnLogUserPassword}
	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to change password: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Password change requested")
}

// Runtime
func ListRuntimes(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "runtime")
	common.ManageListRequest(endpoint, "", []string{"id", "type", "name", "isDefault"}, flags.GenericFilters)
}

func ListRuntimeAvailableTypes(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "runtimeAvailableTypes")
	if RuntimeLanguage != "" {
		endpoint = fmt.Sprintf("%s?language=%s", endpoint, url.QueryEscape(RuntimeLanguage))
	}
	types, err := httpLib.FetchArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch runtime types: %s", err)
		return
	}

	var rows []map[string]any
	for _, entry := range types {
		obj, ok := entry.(map[string]any)
		if !ok {
			rows = append(rows, map[string]any{"type": fmt.Sprintf("%v", entry)})
			continue
		}
		if t, exists := obj["type"]; exists {
			rows = append(rows, map[string]any{"type": fmt.Sprintf("%v", t)})
		} else {
			rows = append(rows, obj)
		}
	}

	rows, err = filtersLib.FilterLines(rows, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	payload := map[string]any{"types": rows}
	display.OutputObject(payload, args[0], runtimeAvailableTypesTemplate, &flags.OutputFormatConfig)
}

func GetRuntime(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(serviceEndpoint(args[0], "runtime"), args[1], runtimeTemplate)
}

func CreateRuntime(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if RuntimeName != "" {
		params["name"] = RuntimeName
	}
	if RuntimeType != "" {
		params["type"] = RuntimeType
	}
	if RuntimePublicDir != "" {
		params["publicDir"] = RuntimePublicDir
	}
	if RuntimeAppEnv != "" {
		params["appEnv"] = RuntimeAppEnv
	}
	if RuntimeAppBootstrap != "" {
		params["appBootstrap"] = RuntimeAppBootstrap
	}
	if cmd.Flags().Changed("runtime-default") {
		params["isDefault"] = RuntimeIsDefault
	}
	if len(RuntimeDomains) > 0 {
		params["attachedDomains"] = RuntimeDomains
	}

	endpoint := serviceEndpoint(args[0], "runtime")
	const runtimeCreateExample = `{
  "name": "",
  "type": "",
  "publicDir": "",
  "appEnv": "",
  "appBootstrap": "",
  "isDefault": false,
  "attachedDomains": []
}`
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/runtime",
		endpoint,
		runtimeCreateExample,
		params,
		assets.WebhostingOpenapiSchema,
		nil,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create runtime: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Runtime created")
}

func UpdateRuntime(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if RuntimeName != "" {
		params["name"] = RuntimeName
	}
	if RuntimePublicDir != "" {
		params["publicDir"] = RuntimePublicDir
	}
	if RuntimeAppEnv != "" {
		params["appEnv"] = RuntimeAppEnv
	}
	if RuntimeAppBootstrap != "" {
		params["appBootstrap"] = RuntimeAppBootstrap
	}
	if cmd.Flags().Changed("runtime-default") {
		params["isDefault"] = RuntimeIsDefault
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("runtime/%s", url.PathEscape(args[1])))
	if err := updateResource(cmd, "/hosting/web/{serviceName}/runtime/{id}", endpoint, params, nil); err != nil {
		if errors.Is(err, errNothingToEdit) {
			display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
			return
		}
		display.OutputError(&flags.OutputFormatConfig, "failed to update runtime: %s", err)
		return
	}
}

func DeleteRuntime(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("runtime/%s", url.PathEscape(args[1])))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete runtime: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Runtime deleted")
}

func ListRuntimeDomains(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("runtime/%s/attachedDomains", url.PathEscape(args[1])))
	domains, err := httpLib.FetchArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch runtime domains: %s", err)
		return
	}

	rows := make([]map[string]any, 0, len(domains))
	for _, entry := range domains {
		rows = append(rows, map[string]any{"domain": fmt.Sprintf("%v", entry)})
	}

	rows, err = filtersLib.FilterLines(rows, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(rows, []string{"domain"}, &flags.OutputFormatConfig)
}

// Websites
func ListWebsites(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "website")
	common.ManageListRequest(endpoint, "", []string{"id", "path", "vcsUrl", "vcsBranch"}, flags.GenericFilters)
}

func GetWebsiteCreationCapabilities(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "websiteCreationCapabilities")
	var capabilities map[string]any
	if err := httpLib.Client.Get(endpoint, &capabilities); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch website creation capabilities: %s", err)
		return
	}
	display.OutputObject(capabilities, args[0], websiteCreationCapabilitiesTemplate, &flags.OutputFormatConfig)
}

func GetWebsite(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(serviceEndpoint(args[0], "website"), args[1], websiteTemplate)
}

func CreateWebsite(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if WebsitePath != "" {
		params["path"] = WebsitePath
	}
	if WebsiteVcsURL != "" {
		params["vcsUrl"] = WebsiteVcsURL
	}
	if WebsiteBranch != "" {
		params["vcsBranch"] = WebsiteBranch
	}

	const createWebsiteExample = `{
  "path": "",
  "vcsUrl": "",
  "vcsBranch": ""
}`

	endpoint := serviceEndpoint(args[0], "website")
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/website",
		endpoint,
		createWebsiteExample,
		params,
		assets.WebhostingOpenapiSchema,
		[]string{"path", "vcsBranch", "vcsUrl"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create website: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Website created")
}

func UpdateWebsite(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if cmd.Flags().Changed("branch") {
		params["vcsBranch"] = WebsiteBranch
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("website/%s", url.PathEscape(args[1])))
	if err := updateResource(cmd, "/hosting/web/{serviceName}/website/{id}", endpoint, params, nil); err != nil {
		if errors.Is(err, errNothingToEdit) {
			display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
			return
		}
		display.OutputError(&flags.OutputFormatConfig, "failed to update website: %s", err)
		return
	}
}

func DeleteWebsite(_ *cobra.Command, args []string) {
	base := serviceEndpoint(args[0], fmt.Sprintf("website/%s", url.PathEscape(args[1])))
	query := url.Values{}
	query.Set("deleteFiles", strconv.FormatBool(WebsiteDeleteFiles))
	endpoint := fmt.Sprintf("%s?%s", base, query.Encode())
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete website: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Website deleted")
}

func DeployWebsite(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("website/%s/deploy", url.PathEscape(args[1])))
	body := map[string]any{}
	if WebsiteDeployReset {
		body["reset"] = true
	}
	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to trigger deployment: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Deployment triggered")
}

func ListWebsiteDeployments(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("website/%s/deployment", url.PathEscape(args[1])))
	common.ManageListRequest(endpoint, "", []string{
		"id",
		"status",
		"source",
		"vcsBranch VCS Branch",
		"vcsCommitId Commit",
		"date",
	}, flags.GenericFilters)
}

func GetWebsiteDeployment(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(serviceEndpoint(args[0], fmt.Sprintf("website/%s/deployment", url.PathEscape(args[1]))), args[2], websiteDeploymentTemplate)
}

func GetWebsiteDeploymentLogs(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("website/%s/deployment/%s/logs", url.PathEscape(args[1]), url.PathEscape(args[2])))
	var logs []map[string]any
	if err := httpLib.Client.Get(endpoint, &logs); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch logs: %s", err)
		return
	}
	if len(logs) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "ℹ️ No logs returned")
		return
	}
	display.RenderTable(logs, []string{"date", "message"}, &flags.OutputFormatConfig)
}

// VCS
func GetVcsWebhooks(_ *cobra.Command, args []string) {
	if VcsWebhookPath == "" {
		display.OutputError(&flags.OutputFormatConfig, "flag --path is required")
		return
	}
	if VcsWebhookPlatform == "" {
		display.OutputError(&flags.OutputFormatConfig, "flag --vcs is required")
		return
	}

	query := url.Values{}
	query.Set("path", VcsWebhookPath)
	query.Set("vcs", VcsWebhookPlatform)

	endpoint := fmt.Sprintf("%s?%s", serviceEndpoint(args[0], "vcs/webhooks"), query.Encode())
	var webhooks map[string]any
	if err := httpLib.Client.Get(endpoint, &webhooks); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch VCS webhooks: %s", err)
		return
	}
	if webhooks == nil {
		webhooks = map[string]any{}
	}
	webhooks["path"] = VcsWebhookPath
	webhooks["vcs"] = VcsWebhookPlatform

	display.OutputObject(webhooks, args[0], vcsWebhooksTemplate, &flags.OutputFormatConfig)
}

// SSL
func GetSSL(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("attachedDomain/%s/ssl", url.PathEscape(args[1])))
	var body map[string]any
	if err := httpLib.Client.Get(endpoint, &body); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get SSL: %s", err)
		return
	}
	if body == nil {
		body = map[string]any{}
	}
	display.OutputObject(body, fmt.Sprintf("%s (%s)", args[1], args[0]), sslTemplate, &flags.OutputFormatConfig)
}

func CreateSSL(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("attachedDomain/%s/ssl", url.PathEscape(args[1])))
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create SSL: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ SSL creation requested")
}

func DeleteSSL(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("attachedDomain/%s/ssl", url.PathEscape(args[1])))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete SSL: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ SSL deleted")
}

func ListSSLAttachedDomains(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v2/webhosting/resource/%s/certificate", url.PathEscape(args[0]))
	var certificates []map[string]any
	if err := httpLib.Client.Get(endpoint, &certificates); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch SSL certificates: %s", err)
		return
	}
	filtered, ok := common.FilteredRows(certificates)
	if !ok {
		return
	}

	payload := map[string]any{"certificates": filtered}
	display.OutputObject(payload, args[0], sslResourceCertificatesTemplate, &flags.OutputFormatConfig)
}

// SSL service-level endpoints

func GetServiceSSL(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "ssl")
	var body map[string]any
	if err := httpLib.Client.Get(endpoint, &body); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get SSL: %s", err)
		return
	}
	if body == nil {
		body = map[string]any{}
	}
	display.OutputObject(body, args[0], sslServiceTemplate, &flags.OutputFormatConfig)
}

func CreateServiceSSL(_ *cobra.Command, args []string) {
	body := map[string]any{}
	if SSLCertificate != "" {
		body["certificate"] = SSLCertificate
	}
	if SSLChain != "" {
		body["chain"] = SSLChain
	}
	if SSLKey != "" {
		body["key"] = SSLKey
	}
	endpoint := serviceEndpoint(args[0], "ssl")
	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create SSL: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ SSL creation requested")
}

func DeleteServiceSSL(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "ssl")
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete SSL: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ SSL deleted")
}

func RegenerateServiceSSL(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "ssl/regenerate")
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to regenerate SSL: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ SSL regeneration requested")
}

func GetSSLReport(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "ssl/report")
	var body map[string]any
	if err := httpLib.Client.Get(endpoint, &body); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get SSL report: %s", err)
		return
	}
	if body == nil {
		body = map[string]any{}
	}
	display.OutputObject(body, args[0], sslReportTemplate, &flags.OutputFormatConfig)
}

// CDN
func GetCdn(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "cdn")
	var body map[string]any
	if err := httpLib.Client.Get(endpoint, &body); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch CDN info: %s", err)
		return
	}
	display.OutputObject(body, args[0], cdnInfoTemplate, &flags.OutputFormatConfig)
}

func ListCdnDomains(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "cdn/domain")
	common.ManageListRequestNoExpand(endpoint, []string{"name domain", "status"}, flags.GenericFilters)
}

func GetCdnDomain(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(serviceEndpoint(args[0], "cdn/domain"), args[1], cdnDomainTemplate)
}

func PurgeCdnDomain(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("cdn/domain/%s/purge", url.PathEscape(args[1])))
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to purge CDN cache: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ CDN purge requested")
}

func RefreshCdnDomain(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("cdn/domain/%s/refresh", url.PathEscape(args[1])))
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to refresh CDN: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ CDN refresh requested")
}

func ListCdnOperations(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "cdn/operation")
	common.ManageListRequestNoExpand(endpoint, []string{"id", "function", "status"}, flags.GenericFilters)
}

func GetCdnOperation(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(serviceEndpoint(args[0], "cdn/operation"), args[1], cdnOperationTemplate)
}

func GetCdnServiceInfo(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "cdn/serviceInfos")
	var info map[string]any
	if err := httpLib.Client.Get(endpoint, &info); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch CDN service info: %s", err)
		return
	}
	display.OutputObject(info, args[0], common.ServiceInfoTemplate, &flags.OutputFormatConfig)
}

func UpdateCdnServiceInfo(cmd *cobra.Command, args []string) {
	payload := buildServiceInfoRenewPayload(cmd)
	if len(payload) == 0 && !flags.ParametersViaEditor && flags.ParametersFile == "" {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
		return
	}

	endpoint := serviceEndpoint(args[0], "cdn/serviceInfosUpdate")
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/cdn/serviceInfosUpdate",
		endpoint,
		defaultRenewExample,
		payload,
		assets.WebhostingOpenapiSchema,
		[]string{"renew"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to update CDN service info: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ CDN service info updated")
}

func GetCdnDomainStatistics(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("cdn/domain/%s/statistics", url.PathEscape(args[1])))
	if CdnStatisticPeriod != "" {
		endpoint = fmt.Sprintf("%s?period=%s", endpoint, url.QueryEscape(CdnStatisticPeriod))
	}

	var stats []map[string]any
	if err := httpLib.Client.Get(endpoint, &stats); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch CDN statistics: %s", err)
		return
	}

	filtered, err := filtersLib.FilterLines(stats, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to apply filters: %s", err)
		return
	}

	payload := map[string]any{"series": filtered}
	display.OutputObject(payload, args[1], cdnDomainStatisticsTemplate, &flags.OutputFormatConfig)
}

func ListCdnAvailableOptions(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "cdn/availableOptions")
	var options []map[string]any
	if err := httpLib.Client.Get(endpoint, &options); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch CDN options: %s", err)
		return
	}

	filtered, err := filtersLib.FilterLines(options, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to apply filters: %s", err)
		return
	}

	payload := map[string]any{"options": filtered}
	display.OutputObject(payload, args[0], cdnAvailableOptionsTemplate, &flags.OutputFormatConfig)
}

func ListCronAvailableLanguages(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "cronAvailableLanguage")
	var languages []string
	if err := httpLib.Client.Get(endpoint, &languages); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch cron languages: %s", err)
		return
	}

	rows := make([]map[string]any, 0, len(languages))
	for _, lang := range languages {
		rows = append(rows, map[string]any{"language": lang})
	}

	display.RenderTable(rows, []string{"language"}, &flags.OutputFormatConfig)
}

func ListCdnDomainOptions(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("cdn/domain/%s/option", url.PathEscape(args[1])))
	common.ManageListRequestNoExpand(endpoint, []string{"name", "type", "enabled"}, flags.GenericFilters)
}

func AddCdnDomainOption(cmd *cobra.Command, args []string) {
	if !flags.ParametersViaEditor && flags.ParametersFile == "" {
		if !cmd.Flags().Changed("name") || CdnOptionName == "" {
			display.OutputError(&flags.OutputFormatConfig, "option name is required")
			return
		}
		if !cmd.Flags().Changed("type") || CdnOptionType == "" {
			display.OutputError(&flags.OutputFormatConfig, "option type is required")
			return
		}
		if !cmd.Flags().Changed("enabled") {
			display.OutputError(&flags.OutputFormatConfig, "option enabled flag must be provided")
			return
		}
	}

	body := buildCdnOptionBody(cmd, true)
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("cdn/domain/%s/option", url.PathEscape(args[1])))

	created, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/cdn/domain/{domainName}/option",
		endpoint,
		"",
		body,
		assets.WebhostingOpenapiSchema,
		[]string{"name", "type", "enabled"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create CDN option: %s", err)
		return
	}

	display.OutputObject(created, args[1], cdnDomainOptionTemplate, &flags.OutputFormatConfig)
}

func GetCdnDomainOption(_ *cobra.Command, args []string) {
	base := serviceEndpoint(args[0], fmt.Sprintf("cdn/domain/%s/option", url.PathEscape(args[1])))
	common.ManageObjectRequest(base, args[2], cdnDomainOptionTemplate)
}

func UpdateCdnDomainOption(cmd *cobra.Command, args []string) {
	params := buildCdnOptionBody(cmd, false)
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("cdn/domain/%s/option/%s", url.PathEscape(args[1]), url.PathEscape(args[2])))

	if err := updateResource(cmd, "/hosting/web/{serviceName}/cdn/domain/{domainName}/option/{optionName}", endpoint, params, nil); err != nil {
		if errors.Is(err, errNothingToEdit) {
			display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
			return
		}
		display.OutputError(&flags.OutputFormatConfig, "failed to update CDN option: %s", err)
		return
	}
}

func DeleteCdnDomainOption(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("cdn/domain/%s/option/%s", url.PathEscape(args[1]), url.PathEscape(args[2])))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete CDN option: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ CDN option deleted")
}

func buildCdnOptionBody(cmd *cobra.Command, includeName bool) map[string]any {
	body := map[string]any{}
	if includeName && cmd.Flags().Changed("name") {
		body["name"] = CdnOptionName
	}
	if cmd.Flags().Changed("type") {
		body["type"] = CdnOptionType
	}
	if cmd.Flags().Changed("enabled") {
		body["enabled"] = CdnOptionEnabled
	}
	if cmd.Flags().Changed("pattern") {
		body["pattern"] = CdnOptionPattern
	}

	if config := buildCdnOptionConfig(cmd); len(config) > 0 {
		body["config"] = config
	}

	return body
}

func buildCdnOptionConfig(cmd *cobra.Command) map[string]any {
	config := map[string]any{}

	if cmd.Flags().Changed("destination") {
		config["destination"] = CdnOptionConfigDestination
	}
	if cmd.Flags().Changed("follow-uri") {
		config["followUri"] = CdnOptionConfigFollowURI
	}
	if cmd.Flags().Changed("origins") {
		config["origins"] = CdnOptionConfigOrigins
	}
	if cmd.Flags().Changed("pattern-type") {
		config["patternType"] = CdnOptionConfigPatternType
	}
	if cmd.Flags().Changed("priority") {
		config["priority"] = CdnOptionConfigPriority
	}
	if cmd.Flags().Changed("query-parameters") {
		config["queryParameters"] = CdnOptionConfigQueryParameter
	}
	if cmd.Flags().Changed("resource") {
		config["resources"] = CdnOptionConfigResources
	}
	if cmd.Flags().Changed("status-code") {
		config["statusCode"] = CdnOptionConfigStatusCode
	}
	if cmd.Flags().Changed("ttl") {
		config["ttl"] = CdnOptionConfigTTL
	}

	return config
}

func buildServiceInfoRenewPayload(cmd *cobra.Command) map[string]any {
	renew := map[string]any{}
	if cmd.Flags().Changed("renew-automatic") {
		renew["automatic"] = common.ServiceInfoSpec.Renew.Automatic
	}
	if cmd.Flags().Changed("renew-delete-at-expiration") {
		renew["deleteAtExpiration"] = common.ServiceInfoSpec.Renew.DeleteAtExpiration
	}
	if cmd.Flags().Changed("renew-forced") {
		renew["forced"] = common.ServiceInfoSpec.Renew.Forced
	}
	if cmd.Flags().Changed("renew-manual-payment") {
		renew["manualPayment"] = common.ServiceInfoSpec.Renew.ManualPayment
	}
	if cmd.Flags().Changed("renew-period") {
		renew["period"] = common.ServiceInfoSpec.Renew.Period
	}

	if len(renew) == 0 {
		return map[string]any{}
	}

	return map[string]any{"renew": renew}
}

func formatQuota(value any) (string, bool) {
	quotaMap, ok := value.(map[string]any)
	if !ok {
		return "", false
	}

	val, valOK := quotaMap["value"]
	unit, unitOK := quotaMap["unit"]
	if !valOK || !unitOK {
		return "", false
	}

	var amount float64
	switch v := val.(type) {
	case float64:
		amount = v
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return "", false
		}
		amount = f
	default:
		return "", false
	}

	return fmt.Sprintf("%.2f %v", amount, unit), true
}

func safeTableString(value string) string {
	return strings.ReplaceAll(value, "%", "%%")
}

func safeMessageString(value string) string {
	return strings.ReplaceAll(value, "%", "%%")
}

func formatTimestamp(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		sec := int64(v)
		if v > 1e12 {
			sec = int64(v / 1000)
		}
		return time.Unix(sec, 0).UTC().Format(time.RFC3339)
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return v.String()
		}
		return formatTimestamp(f)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func stringifyList(value any) (string, bool) {
	array, ok := value.([]any)
	if !ok {
		stringsSlice, ok := value.([]string)
		if ok {
			if len(stringsSlice) == 0 {
				return "", false
			}
			return strings.Join(stringsSlice, ", "), true
		}
		return "", false
	}

	var parts []string
	for _, entry := range array {
		if s, ok := entry.(string); ok {
			parts = append(parts, s)
		}
	}
	if len(parts) == 0 {
		return "", false
	}
	return strings.Join(parts, ", "), true
}

// Users
func ListUsers(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "user")
	common.ManageListRequest(endpoint, "", []string{
		"login",
		"home",
		"state",
		"status",
		"sshState SSH State",
		"serviceManagementCredentials.ftp.url FTP Host",
		"serviceManagementCredentials.ftp.port FTP Port",
		"serviceManagementCredentials.ssh.url SSH Host",
		"serviceManagementCredentials.ssh.port SSH Port",
	}, flags.GenericFilters)
}

func GetUser(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(serviceEndpoint(args[0], "user"), args[1], userTemplate)
}

func CreateUser(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if UserHome != "" {
		params["home"] = UserHome
	}
	if UserLogin != "" {
		params["login"] = UserLogin
	}
	if UserPassword != "" {
		params["password"] = UserPassword
	}
	if UserSSHState != "" {
		params["sshState"] = UserSSHState
	}

	const createUserExample = `{
  "home": "",
  "login": "",
  "password": "",
  "sshState": ""
}`

	endpoint := serviceEndpoint(args[0], "user")
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/user",
		endpoint,
		createUserExample,
		params,
		assets.WebhostingOpenapiSchema,
		[]string{"home", "login", "password"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create user: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ User created")
}

func UpdateUser(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if UserHome != "" {
		params["home"] = UserHome
	}
	if UserSSHState != "" {
		params["sshState"] = UserSSHState
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("user/%s", url.PathEscape(args[1])))
	if err := updateResource(cmd, "/hosting/web/{serviceName}/user/{login}", endpoint, params, nil); err != nil {
		if errors.Is(err, errNothingToEdit) {
			display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
			return
		}
		display.OutputError(&flags.OutputFormatConfig, "failed to update user: %s", err)
		return
	}
}

func DeleteUser(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("user/%s", url.PathEscape(args[1])))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete user: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ User deleted")
}

func ChangeUserPassword(_ *cobra.Command, args []string) {
	if UserPassword == "" {
		display.OutputError(&flags.OutputFormatConfig, "password required, use --password to provide the new value")
		return
	}
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("user/%s/changePassword", url.PathEscape(args[1])))
	body := map[string]any{"password": UserPassword}
	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to change password: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Password updated")
}

// SSH keys
func GetSSHKey(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "key/ssh")
	var key map[string]any
	if err := httpLib.Client.Get(endpoint, &key); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch SSH key: %s", err)
		return
	}
	display.OutputObject(key, args[0], sshKeyTemplate, &flags.OutputFormatConfig)
}

func CreateSSHKey(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "key/ssh")
	var key map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &key); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create SSH key: %s", err)
		return
	}
	if key == nil {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ SSH key generated")
		return
	}
	display.OutputObject(key, args[0], sshKeyTemplate, &flags.OutputFormatConfig)
}

func GetServiceInfo(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "serviceInfos")
	var info map[string]any
	if err := httpLib.Client.Get(endpoint, &info); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch service info: %s", err)
		return
	}
	display.OutputObject(info, args[0], common.ServiceInfoTemplate, &flags.OutputFormatConfig)
}

func UpdateServiceInfo(cmd *cobra.Command, args []string) {
	payload := buildServiceInfoRenewPayload(cmd)
	if len(payload) == 0 && !flags.ParametersViaEditor && flags.ParametersFile == "" && !utils.IsInputFromPipe() {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
		return
	}

	body, err := prepareServiceInfoPayload(cmd, payload, defaultRenewExample)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to prepare payload: %s", err)
		return
	}
	if len(body) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
		return
	}

	endpoint := serviceEndpoint(args[0], "serviceInfos")
	if err := httpLib.Client.Put(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to update service info: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Service info updated")
}

func prepareServiceInfoPayload(cmd *cobra.Command, params map[string]any, example string) (map[string]any, error) {
	switch {
	case utils.IsInputFromPipe():
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		var body map[string]any
		if err := json.Unmarshal(data, &body); err != nil {
			return nil, err
		}
		return body, nil
	case flags.ParametersFile != "":
		log.Print("Flag --from-file used, all other flags will override the file values")
		fd, err := os.Open(flags.ParametersFile)
		if err != nil {
			return nil, err
		}
		defer fd.Close()
		var body map[string]any
		if err := json.NewDecoder(fd).Decode(&body); err != nil {
			return nil, err
		}
		return body, nil
	case flags.ParametersViaEditor:
		log.Print("Flag --editor used, all other flags will override the example values")
		base := map[string]any{}
		if example != "" {
			_ = json.Unmarshal([]byte(example), &base)
		}
		if len(params) > 0 {
			if err := utils.MergeMaps(base, params); err != nil {
				return nil, err
			}
		}
		content, err := json.MarshalIndent(base, "", "  ")
		if err != nil {
			return nil, err
		}
		edited, err := editor.EditValueWithEditor(content)
		if err != nil {
			return nil, err
		}
		var body map[string]any
		if err := json.Unmarshal(edited, &body); err != nil {
			return nil, err
		}
		return body, nil
	default:
		return params, nil
	}
}

// Local SEO
func ListLocalSeoDirectories(_ *cobra.Command, _ []string) {
	if LocalSEOCountry == "" || LocalSEOOffer == "" {
		display.OutputError(&flags.OutputFormatConfig, "flags --country and --offer are required")
		return
	}

	query := url.Values{}
	query.Set("country", LocalSEOCountry)
	query.Set("offer", LocalSEOOffer)

	endpoint := fmt.Sprintf("/v1/hosting/web/localSeo/directoriesList?%s", query.Encode())
	var directories map[string]any
	if err := httpLib.Client.Get(endpoint, &directories); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch directories: %s", err)
		return
	}

	display.OutputObject(directories, fmt.Sprintf("%s • %s", LocalSEOCountry, LocalSEOOffer), localSeoDirectoriesTemplate, &flags.OutputFormatConfig)
}

func CheckLocalSeoEmailAvailability(_ *cobra.Command, args []string) {
	if LocalSEOEmail == "" {
		display.OutputError(&flags.OutputFormatConfig, "flag --email is required")
		return
	}

	query := url.Values{}
	query.Set("email", LocalSEOEmail)

	var endpoint string
	label := LocalSEOEmail
	if len(args) == 0 {
		endpoint = fmt.Sprintf("/v1/hosting/web/localSeo/emailAvailability?%s", query.Encode())
	} else {
		endpoint = fmt.Sprintf("%s?%s", serviceEndpoint(args[0], "localSeo/emailAvailability"), query.Encode())
		label = fmt.Sprintf("%s • %s", args[0], LocalSEOEmail)
	}

	var availability map[string]any
	if err := httpLib.Client.Get(endpoint, &availability); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to check email availability: %s", err)
		return
	}
	display.OutputObject(availability, label, localSeoEmailTemplate, &flags.OutputFormatConfig)
}

func RunLocalSeoVisibilityCheck(cmd *cobra.Command, _ []string) {
	params := map[string]any{}
	if LocalSEOCountry != "" {
		params["country"] = LocalSEOCountry
	}
	if LocalSEOName != "" {
		params["name"] = LocalSEOName
	}
	if LocalSEOStreet != "" {
		params["street"] = LocalSEOStreet
	}
	if LocalSEOZip != "" {
		params["zip"] = LocalSEOZip
	}

	result, err := common.CreateResource(
		cmd,
		"/hosting/web/localSeo/visibilityCheck",
		"/v1/hosting/web/localSeo/visibilityCheck",
		localSeoVisibilityExample,
		params,
		assets.WebhostingOpenapiSchema,
		[]string{"country", "name", "street", "zip"},
	)
	if err != nil {
		var apiErr *ovh.APIError
		if errors.As(err, &apiErr) && apiErr.Code == 400 {
			message := apiErr.Message
			if message == "" {
				message = "Location not found"
			}
			payload := map[string]any{
				"notFound": true,
				"message":  message,
			}
			display.OutputObject(payload, fmt.Sprintf("%s • %s", LocalSEOName, LocalSEOCountry), localSeoVisibilityCheckTemplate, &flags.OutputFormatConfig)
			return
		}

		display.OutputError(&flags.OutputFormatConfig, "failed to run visibility check: %s", err)
		return
	}

	display.OutputObject(result, fmt.Sprintf("%s • %s", LocalSEOName, LocalSEOCountry), localSeoVisibilityCheckTemplate, &flags.OutputFormatConfig)
}

func GetLocalSeoVisibilityResult(_ *cobra.Command, args []string) {
	if LocalSEODirectory == "" || LocalSEOToken == "" {
		display.OutputError(&flags.OutputFormatConfig, "flags --directory and --token are required")
		return
	}

	query := url.Values{}
	query.Set("directory", LocalSEODirectory)
	query.Set("id", args[0])
	query.Set("token", LocalSEOToken)

	endpoint := fmt.Sprintf("/v1/hosting/web/localSeo/visibilityCheckResult?%s", query.Encode())
	var results []map[string]any
	if err := httpLib.Client.Get(endpoint, &results); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch visibility results: %s", err)
		return
	}

	payload := map[string]any{
		"results":   results,
		"directory": LocalSEODirectory,
	}
	display.OutputObject(payload, args[0], localSeoVisibilityResultTemplate, &flags.OutputFormatConfig)
}

func ListLocalSeoAccounts(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "localSeo/account")
	if LocalSEOAccountEmailFilter != "" {
		endpoint = fmt.Sprintf("%s?email=%s", endpoint, url.QueryEscape(LocalSEOAccountEmailFilter))
	}

	accountIDs, err := httpLib.FetchArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch accounts: %s", err)
		return
	}

	var rows []map[string]any
	for _, id := range accountIDs {
		rows = append(rows, map[string]any{"id": fmt.Sprintf("%v", id)})
	}

	common.RenderFilteredTable(rows, []string{"id"})
}

func GetLocalSeoAccount(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(serviceEndpoint(args[0], "localSeo/account"), args[1], localSeoAccountTemplate)
}

func LoginLocalSeoAccount(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("localSeo/account/%s/login", url.PathEscape(args[1])))
	var loginURL string
	if err := httpLib.Client.Post(endpoint, nil, &loginURL); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to generate SSO URL: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"url": loginURL}, "🔐 SSO link: %s", loginURL)
}

func ListLocalSeoLocations(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "localSeo/location")
	locationIDs, err := httpLib.FetchArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch locations: %s", err)
		return
	}

	var rows []map[string]any
	for _, id := range locationIDs {
		rows = append(rows, map[string]any{"id": fmt.Sprintf("%v", id)})
	}

	common.RenderFilteredTable(rows, []string{"id"})
}

func GetLocalSeoLocation(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(serviceEndpoint(args[0], "localSeo/location"), args[1], localSeoLocationTemplate)
}

func GetLocalSeoLocationServiceInfo(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("localSeo/location/%s/serviceInfos", url.PathEscape(args[1])))
	var info map[string]any
	if err := httpLib.Client.Get(endpoint, &info); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch service info: %s", err)
		return
	}
	display.OutputObject(info, fmt.Sprintf("Location %s", args[1]), common.ServiceInfoTemplate, &flags.OutputFormatConfig)
}

func UpdateLocalSeoLocationServiceInfo(cmd *cobra.Command, args []string) {
	payload := buildServiceInfoRenewPayload(cmd)
	if len(payload) == 0 && !flags.ParametersViaEditor && flags.ParametersFile == "" {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
		return
	}

	endpoint := serviceEndpoint(args[0], fmt.Sprintf("localSeo/location/%s/serviceInfosUpdate", url.PathEscape(args[1])))
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/localSeo/location/{id}/serviceInfosUpdate",
		endpoint,
		defaultRenewExample,
		payload,
		assets.WebhostingOpenapiSchema,
		[]string{"renew"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to update service info: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Local SEO service info updated")
}

func TerminateLocalSeoLocation(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], fmt.Sprintf("localSeo/location/%s/terminate", url.PathEscape(args[1])))
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to request termination: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Local SEO location termination requested")
}

func ListTasks(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "tasks")
	common.ManageListRequest(endpoint, "", []string{"id", "function", "status"}, flags.GenericFilters)
}

func GetTask(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(serviceEndpoint(args[0], "tasks"), args[1], taskTemplate)
}

func RequestBoost(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	var chosenOffer string
	if BoostOffer != "" {
		if _, ok := boostOfferChoiceSet[BoostOffer]; !ok {
			display.OutputError(&flags.OutputFormatConfig, "unsupported boost offer %q. Allowed values: %s", BoostOffer, strings.Join(SupportedBoostOffers, ", "))
			return
		}
		params["offer"] = BoostOffer
		chosenOffer = BoostOffer
	} else if !flags.ParametersViaEditor && flags.ParametersFile == "" {
		display.OutputError(&flags.OutputFormatConfig, "boost offer is required. Allowed values: %s", strings.Join(SupportedBoostOffers, ", "))
		return
	}

	endpoint := serviceEndpoint(args[0], "requestBoost")
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/requestBoost",
		endpoint,
		"",
		params,
		assets.WebhostingOpenapiSchema,
		[]string{"offer"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to request boost: %s", err)
		return
	}
	if chosenOffer != "" {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Boost %s requested", chosenOffer)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Boost requested")
}

func RestoreSnapshot(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if RestoreBackup != "" {
		params["backup"] = RestoreBackup
	}

	const restoreSnapshotExample = `{
  "backup": ""
}`

	endpoint := serviceEndpoint(args[0], "restoreSnapshot")
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/restoreSnapshot",
		endpoint,
		restoreSnapshotExample,
		params,
		assets.WebhostingOpenapiSchema,
		[]string{"backup"},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to request restore: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Snapshot restore requested")
}

func GetToken(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "token")
	var raw json.RawMessage
	if err := httpLib.Client.Get(endpoint, &raw); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch token: %s", err)
		return
	}

	payload := map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil || len(payload) == 0 {
		var token string
		if err := json.Unmarshal(raw, &token); err == nil {
			payload = map[string]any{"token": token}
		} else {
			payload = map[string]any{"raw": strings.TrimSpace(string(raw))}
		}
	}

	display.OutputObject(payload, args[0], tokenTemplate, &flags.OutputFormatConfig)
}

func ListBoostHistory(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "boostHistory")
	var dates []string
	if err := httpLib.Client.Get(endpoint, &dates); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch boost history: %s", err)
		return
	}

	var histories []map[string]any
	for _, date := range dates {
		detailEndpoint := serviceEndpoint(args[0], fmt.Sprintf("boostHistory/%s", url.PathEscape(date)))
		var detail map[string]any
		if err := httpLib.Client.Get(detailEndpoint, &detail); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to fetch boost history for %s: %s", date, err)
			return
		}
		histories = append(histories, detail)
	}

	filtered, err := filtersLib.FilterLines(histories, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to apply filters: %s", err)
		return
	}

	renderPayload := map[string]any{
		"entries": filtered,
	}

	display.OutputObject(renderPayload, args[0], boostHistoryTemplate, &flags.OutputFormatConfig)
}

func TerminateService(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "terminate")
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to request termination: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Termination requested")
}

func UnblockTCPOut(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "unblockTCPOut")
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to unblock outgoing TCP: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Outgoing TCP unblocked")
}

func ConfirmTermination(cmd *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "confirmTermination")
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/confirmTermination",
		endpoint,
		"",
		map[string]any{},
		assets.WebhostingOpenapiSchema,
		nil,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to confirm termination: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Termination confirmed")
}

// Generic API call for advanced cases
func CallWebHostingAPI(cmd *cobra.Command, args []string) {
	method := strings.ToUpper(args[0])
	path := args[1]
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasPrefix(path, "/v1/") {
		path = "/v1" + path
	}

	var (
		body   any
		result any
	)

	if method == "POST" || method == "PUT" {
		// Prepare payload from file or editor, if provided
		if flags.ParametersFile != "" {
			content, err := os.ReadFile(flags.ParametersFile)
			if err != nil {
				display.OutputError(&flags.OutputFormatConfig, "failed to read parameters file: %s", err)
				return
			}
			if err := json.Unmarshal(content, &body); err != nil {
				display.OutputError(&flags.OutputFormatConfig, "failed to parse parameters file: %s", err)
				return
			}
		} else if flags.ParametersViaEditor {
			edited, err := editor.EditValueWithEditor([]byte("{}"))
			if err != nil {
				display.OutputError(&flags.OutputFormatConfig, "failed to edit payload: %s", err)
				return
			}
			if err := json.Unmarshal(edited, &body); err != nil {
				display.OutputError(&flags.OutputFormatConfig, "failed to parse edited payload: %s", err)
				return
			}
		}
	}

	var err error
	switch method {
	case "GET":
		err = httpLib.Client.Get(path, &result)
	case "POST":
		err = httpLib.Client.Post(path, body, &result)
	case "PUT":
		err = httpLib.Client.Put(path, body, &result)
	case "DELETE":
		err = httpLib.Client.Delete(path, &result)
	default:
		display.OutputError(&flags.OutputFormatConfig, "unsupported method %s", method)
		return
	}
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "request failed: %s", err)
		return
	}

	if result != nil {
		renderDetails(result)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Request completed")
}

var errNothingToEdit = errors.New("nothing to edit")

func updateResource(cmd *cobra.Command, pathSpec, endpoint string, params map[string]any, allowNil map[string]bool) error {
	cleaned := map[string]any{}
	for k, v := range params {
		if v == nil && (allowNil == nil || !allowNil[k]) {
			continue
		}
		cleaned[k] = v
	}

	if len(cleaned) == 0 && !flags.ParametersViaEditor {
		return errNothingToEdit
	}

	source := cleaned
	if flags.ParametersViaEditor {
		var current map[string]any
		if err := httpLib.Client.Get(endpoint, &current); err != nil {
			return fmt.Errorf("error fetching %s: %w", endpoint, err)
		}
		for k, v := range cleaned {
			current[k] = v
		}
		source = current
	}

	editableBody, err := openapi.FilterEditableFields(
		assets.WebhostingOpenapiSchema,
		pathSpec,
		"put",
		source,
	)
	if err != nil {
		return err
	}

	if !flags.ParametersViaEditor {
		if err := httpLib.Client.Put(endpoint, editableBody, nil); err != nil {
			return err
		}
		display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Resource updated successfully")
		return nil
	}

	editableOutput, err := json.MarshalIndent(editableBody, "", "  ")
	if err != nil {
		return err
	}

	updatedBody, err := editor.EditValueWithEditor(editableOutput)
	if err != nil {
		return err
	}

	if err := httpLib.Client.Put(endpoint, json.RawMessage(updatedBody), nil); err != nil {
		return err
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Resource updated successfully")
	return nil
}
