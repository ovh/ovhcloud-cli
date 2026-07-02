// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package completion

import (
	"fmt"
	"strconv"
	"time"

	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
)

// completionTimeout bounds how long a completion may wait for the API. Some list
// endpoints are slow (e.g. instances are aggregated across all regions), and a
// synchronous call would freeze the user's shell while <tab> is pressed. Past
// this delay we give up and return no suggestion rather than blocking.
const completionTimeout = 3 * time.Second

// fetchIDSuggestions lists the identifiers exposed at the given API endpoint and
// returns them as completion suggestions. v1 list endpoints return an array of
// identifiers (strings or numbers), while some endpoints return an array of
// objects: in that case the "id" field is used as the suggested value.
//
// The request is bounded by completionTimeout so a slow endpoint never freezes
// the shell: on timeout, no suggestion is returned (the in-flight request is
// abandoned, which is harmless as the completion process exits right after).
func fetchIDSuggestions(endpoint string) ([]string, cobra.ShellCompDirective) {
	if httpLib.Client == nil {
		httpLib.InitClient()
	}
	if httpLib.Client == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	result := make(chan []string, 1)
	go func() {
		ids, err := httpLib.FetchArray(endpoint, "")
		if err != nil {
			result <- nil
			return
		}

		var suggestions []string
		for _, raw := range ids {
			switch v := raw.(type) {
			case string:
				suggestions = append(suggestions, v)
			case float64:
				suggestions = append(suggestions, strconv.FormatFloat(v, 'f', -1, 64))
			case map[string]any:
				if id, ok := v["id"].(string); ok && id != "" {
					suggestions = append(suggestions, id)
				}
			}
		}
		result <- suggestions
	}()

	select {
	case suggestions := <-result:
		return suggestions, cobra.ShellCompDirectiveNoFileComp
	case <-time.After(completionTimeout):
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// StaticValues returns a completion function suggesting a fixed set of values.
// It is meant for positional arguments (or flags) that accept a known enum.
func StaticValues(values ...string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		// Only complete the first positional argument.
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return values, cobra.ShellCompDirectiveNoFileComp
	}
}

// ServiceList returns a completion function suggesting the identifiers listed at
// the given API endpoint (e.g. "/v1/vps"). It is meant for positional arguments
// that expect a service/resource identifier.
func ServiceList(endpoint string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		// Only complete the first positional argument.
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return fetchIDSuggestions(endpoint)
	}
}

// CloudResources returns a completion function for a positional argument that is
// the identifier of a project-scoped cloud resource. pathTemplate must contain a
// single %s placeholder for the project ID (e.g. "/v1/cloud/project/%s/instance").
// The project ID is read from the --cloud-project flag; without it, no suggestion
// is returned.
func CloudResources(pathTemplate string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		// Only complete the first positional argument.
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		project, _ := cmd.Flags().GetString("cloud-project")
		if project == "" {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return fetchIDSuggestions(fmt.Sprintf(pathTemplate, project))
	}
}

// CloudResourceWithChild returns a completion function for a command taking two
// positional arguments: a parent resource id then one of its sub-resource ids.
// It completes the first argument from parentTemplate (one %s: project id) and the
// second from childTemplate (two %s: project id and the parent id given as the
// first argument). The project id is read from the --cloud-project flag.
func CloudResourceWithChild(parentTemplate, childTemplate string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		project, _ := cmd.Flags().GetString("cloud-project")
		if project == "" {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		switch len(args) {
		case 0:
			return fetchIDSuggestions(fmt.Sprintf(parentTemplate, project))
		case 1:
			return fetchIDSuggestions(fmt.Sprintf(childTemplate, project, args[0]))
		default:
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}
}

// CloudProjects returns completion suggestions for the --cloud-project flag.
// Each suggestion is "projectID\tName" so shells can display the project name alongside.
func CloudProjects(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	const cacheKey = "cloud-projects"

	// Completion runs on every <tab>: serve from the on-disk cache when it is
	// still fresh to avoid an API call each time.
	if cached, ok := readCachedSuggestions(cacheKey); ok {
		return cached, cobra.ShellCompDirectiveNoFileComp
	}

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

	suggestions := make([]string, 0, len(projects))
	for _, project := range projects {
		if project.Name != "" {
			suggestions = append(suggestions, project.ID+"\t"+project.Name)
		} else {
			suggestions = append(suggestions, project.ID)
		}
	}

	writeCachedSuggestions(cacheKey, suggestions)
	return suggestions, cobra.ShellCompDirectiveNoFileComp
}
