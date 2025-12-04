// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package domainzone

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	domainzoneColumnsToDisplay = []string{"name", "dnssecSupported", "hasDnsAnycast"}

	//go:embed templates/domainzone.tmpl
	domainzoneTemplate string

	//go:embed parameter-samples/record-update.json
	RecordUpdateExample string

	UpdateRecordSpec struct {
		SubDomain string `json:"subDomain,omitempty"`
		Target    string `json:"target,omitempty"`
		TTL       int    `json:"ttl"`
	}
)

func ListDomainZone(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/domain/zone", "", domainzoneColumnsToDisplay, flags.GenericFilters)
}

func GetDomainZone(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v1/domain/zone/%s", url.PathEscape(args[0]))

	// Fetch domain zone
	var object map[string]any
	if err := httpLib.Client.Get(path, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching %s: %s", path, err)
		return
	}

	// Fetch running tasks
	path = fmt.Sprintf("/v1/domain/zone/%s/record", url.PathEscape(args[0]))
	records, err := httpLib.FetchExpandedArray(path, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching records for %s: %s", args[0], err)
		return
	}
	object["records"] = records

	display.OutputObject(object, args[0], domainzoneTemplate, &flags.OutputFormatConfig)
}

func GetRecord(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v1/domain/zone/%s/record", url.PathEscape(args[0]))
	common.ManageObjectRequest(path, args[1], "")
}

func RefreshZone(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v1/domain/zone/%s/refresh", url.PathEscape(args[0]))

	if err := httpLib.Client.Post(path, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error refreshing zone %s: %s", path, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Zone %s refreshed!", args[0])
}

func UpdateRecord(cmd *cobra.Command, args []string) {
	if UpdateRecordSpec.TTL < 1 {
		UpdateRecordSpec.TTL = 0
	}

	if err := common.EditResource(
		cmd,
		"/domain/zone/{zoneName}/record/{id}",
		fmt.Sprintf("/v1/domain/zone/%s/record/%s", url.PathEscape(args[0]), url.PathEscape(args[1])),
		UpdateRecordSpec,
		assets.DomainOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ record %s in %s updated, don't forget to refresh the associated zone!", args[1], args[0])
}
