// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package completion

import (
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

	projects, err := httpLib.FetchExpandedArray("/v1/cloud/project", "")
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, project := range projects {
		id, ok := project["project_id"].(string)
		if !ok {
			continue
		}
		if description, ok := project["description"].(string); ok && description != "" {
			suggestions = append(suggestions, id+"\t"+description)
		} else {
			suggestions = append(suggestions, id)
		}
	}
	return suggestions, cobra.ShellCompDirectiveNoFileComp
}
