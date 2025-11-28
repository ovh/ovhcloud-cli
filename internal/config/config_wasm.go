// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build js && wasm

package config

import (
	"encoding/json"
	"fmt"
	"net/url"

	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"gopkg.in/ini.v1"
)

type (
	configParameters struct {
		// Mapping between a configuration field and its equivalents in
		// the /me/preferences/manager API. The order of the slice defines the priority when
		// looking for a value.
		APIFields []string

		// AccessFunc is a function that, given the raw value from the API,
		// extracts and returns the actual configuration value.
		AccessFunc func(string) (string, error)

		// SetterFunc is a function that, given a value to set, returns the string
		// representation to be sent to the API.
		SetterFunc func(any) (string, error)
	}

	configAPIResponse struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
)

var (
	ConfigurableFields = map[string]configParameters{
		"default_cloud_project": {
			APIFields: []string{"INTERNAL_CLOUD_SHELL_DEFAULT_PROJECT", "PUBLIC_CLOUD_DEFAULT_PROJECT"},
			AccessFunc: func(value string) (string, error) {
				var parsedValue map[string]string
				err := json.Unmarshal([]byte(value), &parsedValue)
				if err != nil {
					return "", fmt.Errorf("failed to parse configuration value: %w", err)
				}
				return parsedValue["projectId"], nil
			},
			SetterFunc: func(value any) (string, error) {
				projectID, ok := value.(string)
				if !ok {
					return "", fmt.Errorf("invalid value type for default_cloud_project, expected string")
				}
				payload := map[string]string{
					"projectId": projectID,
				}
				marshalledPayload, err := json.Marshal(payload)
				if err != nil {
					return "", fmt.Errorf("failed to marshal configuration value: %w", err)
				}
				return string(marshalledPayload), nil
			},
		},
	}

	ConfigPaths []string
)

func LoadINI() (*ini.File, string) {
	return ini.Empty(), ""
}

func ExpandConfigPaths() []string {
	return ConfigPaths
}

func GetConfigValue(_ *ini.File, _, keyName string) (string, error) {
	configParams, ok := ConfigurableFields[keyName]
	if !ok {
		return "", fmt.Errorf("unknown configuration field %q", keyName)
	}

	for _, apiKey := range configParams.APIFields {
		var response configAPIResponse
		err := httpLib.Client.Get("/me/preferences/manager/"+url.PathEscape(apiKey), &response)
		if err != nil {
			continue
		}

		if value, err := configParams.AccessFunc(response.Value); err == nil {
			return value, nil
		}
	}

	return "", nil
}

func SetConfigValue(_ *ini.File, _, _, keyName string, value any) error {
	configParams, ok := ConfigurableFields[keyName]
	if !ok {
		return fmt.Errorf("unknown configuration field %q", keyName)
	}

	// Ensure the key can be set via API
	if len(configParams.APIFields) == 0 {
		return fmt.Errorf("configuration field %q cannot be set in WASM mode", keyName)
	}

	// Value can be anything, so apply SetterFunc to prepare it
	marshalledValue, err := configParams.SetterFunc(value)
	if err != nil {
		return fmt.Errorf("failed to prepare configuration value: %w", err)
	}

	payload := map[string]any{
		"key":   configParams.APIFields[0],
		"value": marshalledValue,
	}

	err = httpLib.Client.Post("/me/preferences/manager", payload, nil)
	if err != nil {
		return fmt.Errorf("failed to set configuration value via API: %w", err)
	}

	return nil
}
