// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"gopkg.in/ini.v1"
)

var (
	ConfigPaths = []string{
		// System wide configuration
		"/etc/ovh.conf",
		// Configuration in user's home
		"~/.ovh.conf",
		// Configuration in local folder
		"./ovh.conf",
	}

	ConfigurableFields = map[string]string{
		"endpoint":              "default",
		"default_cloud_project": "ovh-cli",
	}
)

// currentUserHome attempts to get current user's home directory.
func currentUserHome() (string, error) {
	usr, err := user.Current()
	if err != nil {
		// Fallback by trying to read $HOME
		if userHome := os.Getenv("HOME"); userHome != "" {
			return userHome, nil
		}
		return "", err
	}

	return usr.HomeDir, nil
}

// configPaths returns configPaths, with ~/ prefix expanded.
func ExpandConfigPaths() []string {
	paths := []string{}

	// Will be initialized on first use
	var home string
	var homeErr error

	for _, path := range ConfigPaths {
		if strings.HasPrefix(path, "~/") {
			// Find home if needed
			if home == "" && homeErr == nil {
				home, homeErr = currentUserHome()
			}
			// Ignore file in HOME if we cannot find it
			if homeErr != nil {
				continue
			}

			path = home + path[1:]
		}

		paths = append(paths, path)
	}

	return paths
}

// loadINI builds a ini.File from the configuration paths provided in configPaths.
// It's a helper for loadConfig.
func LoadINI() (*ini.File, string) {
	paths := ExpandConfigPaths()
	if len(paths) == 0 {
		return ini.Empty(), ""
	}

	for _, path := range paths {
		if cfg, err := ini.Load(path); err == nil {
			return cfg, path
		}
	}

	return ini.Empty(), ""
}

// getConfigValue returns the value of OVH_<NAME> or "name" value from "section". If
// the value could not be read from either env or any configuration files, return 'def'
func getConfigValue(cfg *ini.File, section, name, defaultValue string) string {
	// Attempt to load from environment
	fromEnv := os.Getenv("OVH_" + strings.ToUpper(name))
	if fromEnv != "" {
		return fromEnv
	}

	// Attempt to load from configuration
	fromSection := cfg.Section(section)
	if fromSection == nil {
		return defaultValue
	}

	fromSectionKey := fromSection.Key(name)
	if fromSectionKey == nil {
		return defaultValue
	}

	return fromSectionKey.String()
}

func GetConfigValue(cfg *ini.File, sectionName, keyName string) (string, error) {
	// In profile mode, read from the active profile section only (no fallback to legacy)
	if sectionName == "" {
		if profileName := GetActiveProfileName(cfg, ActiveProfileOverride); profileName != "" && !IsDefaultProfile(profileName) {
			return GetProfileConfigValue(cfg, profileName, keyName), nil
		}
	}

	if sectionName == "" {
		sectionName = ConfigurableFields[keyName]
		if sectionName == "" {
			return "", fmt.Errorf("unknown configuration field %q", keyName)
		}
	}

	return getConfigValue(cfg, sectionName, keyName, ""), nil
}

func SetConfigValue(cfg *ini.File, path, sectionName, keyName, value string) error {
	if path == "" {
		path = ConfigPaths[0]
	}

	if sectionName == "" {
		sectionName = ConfigurableFields[keyName]
		if sectionName == "" {
			return fmt.Errorf("unknown configuration field %q", keyName)
		}
	}

	var err error

	section := cfg.Section(sectionName)
	if section == nil {
		section, err = cfg.NewSection("ovh-cli")
		if err != nil {
			return err
		}
	}

	key := section.Key(keyName)
	if section == nil {
		_, err = section.NewKey(keyName, value)
		if err != nil {
			return err
		}
	} else {
		key.SetValue(value)
	}

	return cfg.SaveTo(path)
}

// ActiveProfileOverride is set from the --profile CLI flag by the root command's
// PersistentPreRun. It allows GetConfigValue to respect the --profile flag without
// the config package needing to import the flags package.
var ActiveProfileOverride string

const profileSectionPrefix = "profile:"

// DefaultProfileName is the reserved name for the legacy/default configuration.
// It does not map to a [profile:default] section — instead it represents the
// standard go-ovh configuration from [default] + [ovh-<endpoint>] sections.
const DefaultProfileName = "default"

// IsDefaultProfile returns true if the given name is the reserved default profile.
func IsDefaultProfile(name string) bool {
	return name == DefaultProfileName
}

// GetActiveProfileName returns the active profile name by checking (in order):
// 1. The flagOverride parameter (from --profile CLI flag)
// 2. The OVH_PROFILE environment variable
// 3. The "profile" key in the [default] INI section
// Returns "" if no profile is configured (legacy mode).
func GetActiveProfileName(cfg *ini.File, flagOverride string) string {
	if flagOverride != "" {
		return flagOverride
	}
	if p := os.Getenv("OVH_PROFILE"); p != "" {
		return p
	}
	if cfg != nil {
		section := cfg.Section("default")
		if section != nil && section.HasKey("profile") {
			return section.Key("profile").String()
		}
	}
	return ""
}

// IsProfileMode returns true if a profile is active (either from flag, env, or config).
func IsProfileMode(cfg *ini.File, flagOverride string) bool {
	return GetActiveProfileName(cfg, flagOverride) != ""
}

// GetProfileSection returns the [profile:<name>] INI section for the given profile.
func GetProfileSection(cfg *ini.File, profileName string) (*ini.Section, error) {
	sectionName := profileSectionPrefix + profileName
	section := cfg.Section(sectionName)
	if section == nil || len(section.Keys()) == 0 {
		return nil, fmt.Errorf("profile %q not found in configuration", profileName)
	}
	return section, nil
}

// GetProfileCredentials reads OVHcloud API credentials from a profile section.
// Environment variables (OVH_ENDPOINT, OVH_APPLICATION_KEY, etc.) take precedence.
func GetProfileCredentials(cfg *ini.File, profileName string) (endpoint, appKey, appSecret, consumerKey string, err error) {
	section, err := GetProfileSection(cfg, profileName)
	if err != nil {
		return "", "", "", "", err
	}

	readKey := func(envName, iniKey string) string {
		if v := os.Getenv(envName); v != "" {
			return v
		}
		if section.HasKey(iniKey) {
			return section.Key(iniKey).String()
		}
		return ""
	}

	endpoint = readKey("OVH_ENDPOINT", "endpoint")
	appKey = readKey("OVH_APPLICATION_KEY", "application_key")
	appSecret = readKey("OVH_APPLICATION_SECRET", "application_secret")
	consumerKey = readKey("OVH_CONSUMER_KEY", "consumer_key")

	return endpoint, appKey, appSecret, consumerKey, nil
}

// GetProfileConfigValue reads a configuration value from a profile section.
func GetProfileConfigValue(cfg *ini.File, profileName, keyName string) string {
	section, err := GetProfileSection(cfg, profileName)
	if err != nil {
		return ""
	}
	if section.HasKey(keyName) {
		return section.Key(keyName).String()
	}
	return ""
}

// ListProfiles returns the names of all profiles found in the INI config.
func ListProfiles(cfg *ini.File) []string {
	var profiles []string
	for _, section := range cfg.Sections() {
		if name, ok := strings.CutPrefix(section.Name(), profileSectionPrefix); ok {
			profiles = append(profiles, name)
		}
	}
	return profiles
}

// SetActiveProfile sets the active profile in the [default] section.
func SetActiveProfile(cfg *ini.File, path, profileName string) error {
	if path == "" {
		path = ConfigPaths[0]
	}
	section := cfg.Section("default")
	if section == nil {
		var err error
		section, err = cfg.NewSection("default")
		if err != nil {
			return err
		}
	}
	section.Key("profile").SetValue(profileName)
	return cfg.SaveTo(path)
}

// DeleteProfile removes a profile section from the config. If the deleted profile
// was the active one, the "profile" key is removed from [default] (falling back to legacy mode).
func DeleteProfile(cfg *ini.File, path, profileName string) error {
	if path == "" {
		path = ConfigPaths[0]
	}

	sectionName := profileSectionPrefix + profileName
	section := cfg.Section(sectionName)
	if section == nil || len(section.Keys()) == 0 {
		return fmt.Errorf("profile %q not found in configuration", profileName)
	}

	cfg.DeleteSection(sectionName)

	// If the deleted profile was the active one, clear the active profile
	if GetActiveProfileName(cfg, "") == profileName {
		defaultSection := cfg.Section("default")
		if defaultSection != nil {
			defaultSection.DeleteKey("profile")
		}
	}

	return cfg.SaveTo(path)
}

// SetProfileConfigValue sets a configuration value in a profile section,
// creating the section if it does not exist.
func SetProfileConfigValue(cfg *ini.File, path, profileName, keyName, value string) error {
	if path == "" {
		path = ConfigPaths[0]
	}

	sectionName := profileSectionPrefix + profileName
	section := cfg.Section(sectionName)
	if section == nil {
		var err error
		section, err = cfg.NewSection(sectionName)
		if err != nil {
			return err
		}
	}

	section.Key(keyName).SetValue(value)
	return cfg.SaveTo(path)
}
