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

	WebHostingSpec struct {
		DisplayName string `json:"displayName,omitempty"`
	}

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

	// SSL
	SSLCertificate string
	SSLChain       string
	SSLKey         string

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
	StatisticsType   string

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
  "description": "",
  "ownLogsId": 0
}`

func ListWebHosting(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/hosting/web", "", webhostingColumnsToDisplay, flags.GenericFilters)
}

func GetWebHosting(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/hosting/web", args[0], webhostingTemplate)
}

func EditWebHosting(cmd *cobra.Command, args []string) {
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
		return fmt.Sprintf("/hosting/web/%s", url.PathEscape(serviceName))
	}
	if strings.HasPrefix(suffix, "/") {
		return fmt.Sprintf("/hosting/web/%s%s", url.PathEscape(serviceName), suffix)
	}
	return fmt.Sprintf("/hosting/web/%s/%s", url.PathEscape(serviceName), suffix)
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
	endpoint := fmt.Sprintf("/hosting/web/attachedDomain?domain=%s", url.QueryEscape(args[0]))
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
	endpoint := fmt.Sprintf("/hosting/web/availableOffer?domain=%s", url.QueryEscape(args[0]))
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
	if err := httpLib.Client.Get("/hosting/web/incident", &incidents); err != nil {
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
	display.RenderTable(rows, []string{"incident"}, &flags.OutputFormatConfig)
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

	endpoint := serviceEndpoint(args[0], "cron")
	const defaultCronExample = `{
  "command": "",
  "frequency": "",
  "language": "",
  "email": "",
  "description": ""
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

	display.RenderTable(rows, []string{"type"}, &flags.OutputFormatConfig)
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

	display.RenderTable(capabilities, []string{"type", "isolation", "available", "quotaDisplay quota", "enginesDisplay engines"}, &flags.OutputFormatConfig)
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
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch bounces: %s", err)
		return
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

	display.RenderTable(volumes, []string{"date", "volume"}, &flags.OutputFormatConfig)
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

	display.RenderTable(rows, []string{"id"}, &flags.OutputFormatConfig)
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

	display.RenderTable(rows, []string{"id"}, &flags.OutputFormatConfig)
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

	display.RenderTable(rows, []string{"database"}, &flags.OutputFormatConfig)
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

	if len(downloadLinks) > 0 && !flags.OutputFormatConfig.JsonOutput && !flags.OutputFormatConfig.YamlOutput && flags.OutputFormatConfig.CustomFormat == "" {
		var builder strings.Builder
		builder.WriteString("\nFull download URLs:\n")
		for i, link := range downloadLinks {
			builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, link))
		}
		display.OutputInfo(&flags.OutputFormatConfig, nil, safeMessageString(builder.String()))
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

	if fullURL != "" && !flags.OutputFormatConfig.JsonOutput && !flags.OutputFormatConfig.YamlOutput && flags.OutputFormatConfig.CustomFormat == "" {
		display.OutputInfo(&flags.OutputFormatConfig, nil, safeMessageString("Download URL: "+fullURL))
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

	display.RenderTable(rows, []string{"timestamp", "value"}, &flags.OutputFormatConfig)
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

	display.RenderTable(modules, []string{"id", "name", "branch", "version", "active", "latest"}, &flags.OutputFormatConfig)
}

func GetModuleCatalog(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/hosting/web/moduleList/%s", url.PathEscape(args[0]))
	var module map[string]any
	if err := httpLib.Client.Get(endpoint, &module); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch module catalog entry: %s", err)
		return
	}
	display.OutputObject(module, args[0], moduleCatalogTemplate, &flags.OutputFormatConfig)
}

func GetOfferCapabilities(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/hosting/web/offerCapabilities?offer=%s", url.QueryEscape(args[0]))
	var capabilities map[string]any
	if err := httpLib.Client.Get(endpoint, &capabilities); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch offer capabilities: %s", err)
		return
	}
	display.OutputObject(capabilities, args[0], offerCapabilitiesTemplate, &flags.OutputFormatConfig)
}

func ListSupportedVcs(_ *cobra.Command, _ []string) {
	var vcs []string
	if err := httpLib.Client.Get("/hosting/web/vcs/supported", &vcs); err != nil {
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
	display.RenderTable(rows, []string{"platform"}, &flags.OutputFormatConfig)
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

		endpoint := fmt.Sprintf("/hosting/web/moduleList/%d", id)
		var moduleInfo map[string]any
		if err := httpLib.Client.Get(endpoint, &moduleInfo); err != nil {
			return nil, fmt.Errorf("failed to fetch module catalog entry %d: %w", id, err)
		}
		catalog[id] = moduleInfo
	}
	return catalog, nil
}

func fetchFullModuleCatalog() (map[int]map[string]any, error) {
	items, err := httpLib.FetchExpandedArray("/hosting/web/moduleList", "")
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
	path := "/hosting/web/moduleList"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	ids, err := httpLib.FetchArray(path, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch module ids: %w", err)
	}

	objects, err := httpLib.FetchObjectsParallel[map[string]any]("/hosting/web/moduleList/%s", ids, flags.IgnoreErrors)
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

	payload := map[string]any{"entries": capabilities}
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
		rows = append(rows, map[string]any{"type": fmt.Sprintf("%v", entry)})
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
	if WebsitePath != "" {
		params["path"] = WebsitePath
	}
	if WebsiteVcsURL != "" {
		params["vcsUrl"] = WebsiteVcsURL
	}
	if WebsiteBranch != "" {
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
	endpoint := serviceEndpoint(args[0], "ssl")
	var body map[string]any
	if err := httpLib.Client.Get(endpoint, &body); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get SSL: %s", err)
		return
	}
	renderDetails(body)
}

func CreateSSL(cmd *cobra.Command, args []string) {
	params := map[string]any{}
	if SSLCertificate != "" {
		params["certificate"] = SSLCertificate
	}
	if SSLChain != "" {
		params["chain"] = SSLChain
	}
	if SSLKey != "" {
		params["key"] = SSLKey
	}

	endpoint := serviceEndpoint(args[0], "ssl")
	if _, err := common.CreateResource(
		cmd,
		"/hosting/web/{serviceName}/ssl",
		endpoint,
		"",
		params,
		assets.WebhostingOpenapiSchema,
		nil,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create SSL: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ SSL creation requested")
}

func DeleteSSL(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "ssl")
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete SSL: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ SSL deleted")
}

func RegenerateSSL(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "ssl/regenerate")
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to regenerate SSL: %s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Regeneration requested")
}

func ListSSLAttachedDomains(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "ssl/domains")
	common.ManageListRequestNoExpand(endpoint, []string{"domain"}, flags.GenericFilters)
}

func GetSSLReport(_ *cobra.Command, args []string) {
	endpoint := serviceEndpoint(args[0], "ssl/report")
	var body map[string]any
	if err := httpLib.Client.Get(endpoint, &body); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch SSL report: %s", err)
		return
	}
	renderDetails(body)
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