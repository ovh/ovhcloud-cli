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