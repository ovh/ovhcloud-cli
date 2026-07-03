// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package completion

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/config"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
)

func resolveCloudProject(cmd *cobra.Command) string {
	if p, _ := cmd.Flags().GetString("cloud-project"); p != "" {
		return p
	}
	config.ActiveProfileOverride = flags.Profile
	if p := os.Getenv("OVH_CLOUD_PROJECT_SERVICE"); p != "" {
		return p
	}
	if p := os.Getenv("OS_TENANT_ID"); p != "" {
		return p
	}
	if p, err := config.GetConfigValue(flags.CliConfig, "", "default_cloud_project"); err == nil {
		return p
	}
	return ""
}

// completionTimeout bounds how long a completion may wait for the API. Some list
// endpoints are slow (e.g. instances are aggregated across all regions), and a
// synchronous call would freeze the user's shell while <tab> is pressed. Past
// this delay we give up and return no suggestion rather than blocking.
const completionTimeout = 3 * time.Second

func fetchSuggestions(endpoint, labelField, cacheKey string) ([]string, cobra.ShellCompDirective) {
	if cacheKey != "" {
		if cached, ok := readCachedSuggestions(cacheKey); ok {
			return cached, cobra.ShellCompDirectiveNoFileComp
		}
	}

	if httpLib.Client == nil {
		// Initialize the client the same way commands do (the completion path
		// skips PersistentPreRun): honour the config file and the active profile,
		// otherwise profile-based auth would be ignored and the API call would
		// hit the wrong account / no credentials.
		config.ActiveProfileOverride = flags.Profile
		httpLib.InitClientWithProfile(flags.CliConfig, flags.Profile)
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

		// Start from a non-nil slice so a valid empty result is still cached
		// (nil is reserved for the error path above, which must not be cached).
		suggestions := make([]string, 0, len(ids))
		for _, raw := range ids {
			switch v := raw.(type) {
			case string:
				suggestions = append(suggestions, v)
			case float64:
				suggestions = append(suggestions, strconv.FormatFloat(v, 'f', -1, 64))
			case map[string]any:
				id, ok := v["id"].(string)
				if !ok || id == "" {
					continue
				}
				if labelField != "" {
					if label, ok := v[labelField].(string); ok && label != "" {
						suggestions = append(suggestions, id+"\t"+label)
						continue
					}
				}
				suggestions = append(suggestions, id)
			}
		}
		result <- suggestions
	}()

	select {
	case suggestions := <-result:
		if cacheKey != "" && suggestions != nil {
			writeCachedSuggestions(cacheKey, suggestions)
		}
		return suggestions, cobra.ShellCompDirectiveNoFileComp
	case <-time.After(completionTimeout):
		return nil, cobra.ShellCompDirectiveNoFileComp
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
		return fetchSuggestions(endpoint, "", "")
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
		project := resolveCloudProject(cmd)
		if project == "" {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return fetchSuggestions(fmt.Sprintf(pathTemplate, url.PathEscape(project)), "", "")
	}
}

// CloudResourceWithChild returns a completion function for a command taking two
// positional arguments: a parent resource id then one of its sub-resource ids.
// It completes the first argument from parentTemplate (one %s: project id) and the
// second from childTemplate (two %s: project id and the parent id given as the
// first argument). The project id is read from the --cloud-project flag.
func CloudResourceWithChild(parentTemplate, childTemplate string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		project := resolveCloudProject(cmd)
		if project == "" {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		switch len(args) {
		case 0:
			return fetchSuggestions(fmt.Sprintf(parentTemplate, url.PathEscape(project)), "", "")
		case 1:
			return fetchSuggestions(fmt.Sprintf(childTemplate, url.PathEscape(project), url.PathEscape(args[0])), "", "")
		default:
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}
}

func CloudProjects(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return fetchSuggestions("/v2/publicCloud/project", "name", "cloud-projects")
}
