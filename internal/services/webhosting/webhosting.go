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