// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	_ "embed"

	"github.com/ovh/ovhcloud-cli/internal/config"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/spf13/cobra"
)

//go:embed templates/profile.tmpl
var profileTemplate string

func ListProfiles(_ *cobra.Command, _ []string) {
	activeProfile := config.GetActiveProfileName(flags.CliConfig, "")

	// The "default" profile always exists — it represents the legacy configuration
	defaultEndpoint, _ := config.GetConfigValue(flags.CliConfig, "default", "endpoint")
	defaultActive := ""
	if activeProfile == "" || config.IsDefaultProfile(activeProfile) {
		defaultActive = "*"
	}

	rows := []map[string]any{
		{
			"active":   defaultActive,
			"name":     config.DefaultProfileName,
			"endpoint": defaultEndpoint,
		},
	}

	// Append named profiles
	for _, name := range config.ListProfiles(flags.CliConfig) {
		active := ""
		if name == activeProfile {
			active = "*"
		}
		endpoint := config.GetProfileConfigValue(flags.CliConfig, name, "endpoint")
		rows = append(rows, map[string]any{
			"active":   active,
			"name":     name,
			"endpoint": endpoint,
		})
	}

	display.RenderTable(rows, []string{"active", "name", "endpoint"}, &flags.OutputFormatConfig)
}

func ShowProfile(_ *cobra.Command, args []string) {
	var profileName string
	if len(args) > 0 {
		profileName = args[0]
	} else {
		profileName = config.GetActiveProfileName(flags.CliConfig, flags.Profile)
	}

	// Show default/legacy profile info
	if profileName == "" || config.IsDefaultProfile(profileName) {
		endpoint, _ := config.GetConfigValue(flags.CliConfig, "default", "endpoint")
		defaultProject, _ := config.GetConfigValue(flags.CliConfig, "ovh-cli", "default_cloud_project")
		result := map[string]any{
			"profile":  config.DefaultProfileName,
			"endpoint": endpoint,
		}
		if defaultProject != "" {
			result["default_cloud_project"] = defaultProject
		}
		display.OutputObject(result, "profile", profileTemplate, &flags.OutputFormatConfig)
		return
	}

	section, err := config.GetProfileSection(flags.CliConfig, profileName)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "profile %q not found in configuration", profileName)
		return
	}

	result := map[string]any{
		"profile": profileName,
	}
	for _, key := range section.Keys() {
		// Mask secrets in display
		value := key.Value()
		if key.Name() == "application_secret" || key.Name() == "consumer_key" {
			if len(value) > 4 {
				value = value[:4] + "****"
			}
		}
		result[key.Name()] = value
	}

	display.OutputObject(result, "profile", profileTemplate, &flags.OutputFormatConfig)
}

func SwitchProfile(_ *cobra.Command, args []string) {
	profileName := args[0]

	// Switching to "default" means reverting to legacy mode
	if config.IsDefaultProfile(profileName) {
		defaultSection := flags.CliConfig.Section("default")
		if defaultSection != nil {
			defaultSection.DeleteKey("profile")
		}
		if flags.CliConfigPath != "" {
			if err := flags.CliConfig.SaveTo(flags.CliConfigPath); err != nil {
				display.OutputError(&flags.OutputFormatConfig, "failed to switch profile: %s", err)
				return
			}
		}
		display.OutputInfo(&flags.OutputFormatConfig, nil, "Switched to default profile")
		return
	}

	// Validate profile exists
	if _, err := config.GetProfileSection(flags.CliConfig, profileName); err != nil {
		profiles := config.ListProfiles(flags.CliConfig)
		display.OutputError(&flags.OutputFormatConfig, "profile %q not found. Available profiles: %v", profileName, profiles)
		return
	}

	if err := config.SetActiveProfile(flags.CliConfig, flags.CliConfigPath, profileName); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to switch profile: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "Switched to profile %q", profileName)
}

func DeleteProfile(_ *cobra.Command, args []string) {
	profileName := args[0]

	if config.IsDefaultProfile(profileName) {
		display.OutputError(&flags.OutputFormatConfig, "cannot delete the default profile")
		return
	}

	if err := config.DeleteProfile(flags.CliConfig, flags.CliConfigPath, profileName); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete profile: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "Profile %q deleted", profileName)
}
