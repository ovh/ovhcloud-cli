// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
)

var (
	//go:embed templates/cloud_quota.tmpl
	cloudQuotaTemplate string
)

func GetCloudQuota(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	url := fmt.Sprintf("/v1/cloud/project/%s/region/%s/quota", projectID, url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(url, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching quotas for region %s: %s", args[0], err)
		return
	}

	display.OutputObject(object, args[0], cloudQuotaTemplate, &flags.OutputFormatConfig)
}
