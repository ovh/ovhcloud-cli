// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// GetAIAuthorization returns the AI Endpoints authorization status of the project
func GetAIAuthorization(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var authorization map[string]any
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/ai/authorization", projectID)
	if err := httpLib.Client.Get(endpoint, &authorization); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching AI Endpoints authorization: %s", err)
		return
	}

	display.OutputObject(authorization, projectID, "", &flags.OutputFormatConfig)
}

// CreateAIAuthorization authorizes AI Endpoints usage on the project
func CreateAIAuthorization(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/ai/authorization", projectID)
	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to authorize AI Endpoints: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ AI Endpoints authorized successfully")
}
