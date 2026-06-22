// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package completion

import (
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
)

// CloudProjects returns completion suggestions for the --cloud-project flag.
// Each suggestion is "projectID\tName" so shells can display the project name alongside.
func CloudProjects(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	if httpLib.Client == nil {
		httpLib.InitClient()
	}
	if httpLib.Client == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// The v2 endpoint returns the full project objects (including the name) in a
	// single call, so there is no need to fetch each project individually.
	var projects []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := httpLib.Client.Get("/v2/publicCloud/project", &projects); err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, project := range projects {
		if project.Name != "" {
			suggestions = append(suggestions, project.ID+"\t"+project.Name)
		} else {
			suggestions = append(suggestions, project.ID)
		}
	}
	return suggestions, cobra.ShellCompDirectiveNoFileComp
}
