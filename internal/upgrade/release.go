// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package upgrade

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const latestReleaseURL = "https://github.com/ovh/ovhcloud-cli/releases/latest"

// LatestTag returns the tag_name of the latest GitHub release of ovhcloud-cli.
func LatestTag(ctx context.Context) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	return latestTag(ctx, client, latestReleaseURL)
}

func latestTag(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch release metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("release metadata: HTTP %d", resp.StatusCode)
	}

	var data struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode release metadata: %w", err)
	}
	if data.TagName == "" {
		return "", errors.New("release metadata: empty tag_name")
	}
	return data.TagName, nil
}
