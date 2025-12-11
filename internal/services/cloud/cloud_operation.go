// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"

	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectOperationColumnsToDisplay = []string{"id", "action", "progress", "status"}

	//go:embed templates/cloud_operation.tmpl
	cloudOperationTemplate string
)

func ListCloudOperations(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var operations []map[string]any
	err = httpLib.Client.Get(fmt.Sprintf("/v1/cloud/project/%s/operation", projectID), &operations)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch results: %s", err)
		return
	}

	operations, err = filtersLib.FilterLines(operations, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(operations, cloudprojectOperationColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetCloudOperation(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v1/cloud/project/%s/operation", projectID), args[0], cloudOperationTemplate)
}
