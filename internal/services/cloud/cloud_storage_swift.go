// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectStorageSwiftColumnsToDisplay = []string{"id", "name", "region"}

	//go:embed templates/cloud_storage_swift.tmpl
	cloudStorageSwiftTemplate string

	// CloudSwiftContainerType is the type of the SWIFT storage container
	CloudSwiftContainerType string
)

func ListCloudStorageSwift(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(fmt.Sprintf("/v1/cloud/project/%s/storage", projectID), cloudprojectStorageSwiftColumnsToDisplay, flags.GenericFilters)
}

func GetStorageSwift(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v1/cloud/project/%s/storage", projectID), args[0], cloudStorageSwiftTemplate)
}

func EditStorageSwift(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/storage/{containerId}",
		fmt.Sprintf("/v1/cloud/project/%s/storage/%s", projectID, url.PathEscape(args[0])),
		map[string]any{
			"containerType": CloudSwiftContainerType,
		},
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
