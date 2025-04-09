package config

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"gopkg.in/ini.v1"
)

var (
	configPaths = []string{
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
func expandConfigPaths() []string {
	paths := []string{}

	// Will be initialized on first use
	var home string
	var homeErr error

	for _, path := range configPaths {
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
	paths := expandConfigPaths()
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
		path = configPaths[0]
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
