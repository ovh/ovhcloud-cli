// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	//go:embed templates/cloud_quota.tmpl
	cloudQuotaTemplate string

	//go:embed parameter-samples/quota-edit.json
	CloudQuotaEditExample string

	QuotaEditSpec struct {
		TargetSpec struct {
			PreventAutomaticQuotaUpgrade *bool `json:"preventAutomaticQuotaUpgrade,omitempty"`
		} `json:"targetSpec"`
	}
)

func GetCloudQuota(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/quota", projectID)

	var object map[string]any
	if err := httpLib.Client.Get(endpoint, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching quota for project %s: %s", projectID, err)
		return
	}

	display.OutputObject(object, projectID, cloudQuotaTemplate, &flags.OutputFormatConfig)
}

func EditCloudQuota(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/quota",
		fmt.Sprintf("/v2/publicCloud/project/%s/quota", projectID),
		QuotaEditSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
