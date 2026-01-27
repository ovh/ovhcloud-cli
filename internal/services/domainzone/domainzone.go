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

	recordColumnsToDisplay = []string{"id", "subDomain", "fieldType", "target", "ttl"}

	//go:embed templates/domainzone.tmpl
	domainzoneTemplate string

	//go:embed parameter-samples/record-create.json
	RecordCreateExample string

	//go:embed parameter-samples/record-update.json
	RecordUpdateExample string

	CreateRecordSpec struct {
		FieldType string `json:"fieldType,omitempty"`
		SubDomain string `json:"subDomain,omitempty"`
		Target    string `json:"target,omitempty"`
		TTL       int    `json:"ttl"`
	}

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

func ListRecords(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v1/domain/zone/%s/record", url.PathEscape(args[0]))
	common.ManageListRequest(path, "", recordColumnsToDisplay, flags.GenericFilters)
}

func RefreshZone(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v1/domain/zone/%s/refresh", url.PathEscape(args[0]))

	if err := httpLib.Client.Post(path, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error refreshing zone %s: %s", path, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Zone %s refreshed!", args[0])
}

func CreateRecord(cmd *cobra.Command, args []string) {
	if CreateRecordSpec.TTL < 1 {
		CreateRecordSpec.TTL = 0
	}

	record, err := common.CreateResource(
		cmd,
		"/domain/zone/{zoneName}/record",
		fmt.Sprintf("/v1/domain/zone/%s/record", url.PathEscape(args[0])),
		RecordCreateExample,
		CreateRecordSpec,
		assets.DomainOpenapiSchema,
		[]string{"fieldType", "target"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error creating record %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ record %s created in %s, don't forget to refresh the associated zone!", record["id"], args[0])
}

func DeleteRecord(cmd *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/domain/zone/%s/record/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error deleting record %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ record %s deleted successfully from %s", args[1], args[0])
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
