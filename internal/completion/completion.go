// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package completion

import (
	"fmt"

	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
)

// CloudProjects returns completion suggestions for the --cloud-project flag.
// Each suggestion is "projectID\tDescription" so shells can display the description alongside.
func CloudProjects(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	if httpLib.Client == nil {
		httpLib.InitClient()
	}
	if httpLib.Client == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var projectIDs []string
	if err := httpLib.Client.Get("/v1/cloud/project", &projectIDs); err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, id := range projectIDs {
		var project struct {
			Description string `json:"description"`
		}
		if err := httpLib.Client.Get(fmt.Sprintf("/v1/cloud/project/%s", id), &project); err == nil && project.Description != "" {
			suggestions = append(suggestions, id+"\t"+project.Description)
		} else {
			suggestions = append(suggestions, id)
		}
	}
	return suggestions, cobra.ShellCompDirectiveNoFileComp
}
