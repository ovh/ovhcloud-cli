// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package vps

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/display"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

func waitForVpsTask(serviceName string, taskInput map[string]any, retryDuration time.Duration) (map[string]any, error) {
	id, ok := taskInput["id"]
	if !ok {
		return nil, errors.New("task input does not contain 'id'")
	}

	taskID, ok := id.(json.Number)
	if !ok {
		return nil, fmt.Errorf("task ID is not a valid JSON number: %v", id)
	}

	endpoint := fmt.Sprintf("/v1/vps/%s/tasks/%s", url.PathEscape(serviceName), url.PathEscape(string(taskID)))

	ctx, cancel := context.WithTimeout(context.Background(), retryDuration)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		var task map[string]any
		if err := httpLib.Client.Get(endpoint, &task); err != nil {
			return nil, fmt.Errorf("error fetching operation: %w", err)
		}

		switch task["state"].(string) {
		case "blocked", "cancelled", "error":
			return task, fmt.Errorf("task %s ended in error state %q", taskID, task["state"])
		case "done":
			return task, nil
		}

		select {
		case <-ctx.Done():
			return nil, errors.New("timeout waiting for task to complete")
		case <-ticker.C:
			log.Printf("Still waiting for task to complete (status=%s)…", task["state"])
			continue
		}
	}
}

func getAvailableImages(serviceName string) (map[string]string, error) {
	endpoint := fmt.Sprintf("/v1/vps/%s/images/available", url.PathEscape(serviceName))

	images, err := httpLib.FetchExpandedArray(endpoint, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch available images: %w", err)
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no image found for service %s", serviceName)
	}

	imageChoices := make(map[string]string, len(images))
	for _, img := range images {
		name := img["name"].(string)
		imageChoices[name] = img["id"].(string)
	}

	return imageChoices, nil
}

func runImageSelector(serviceName string) (string, string, error) {
	imageChoices, err := getAvailableImages(serviceName)
	if err != nil {
		return "", "", fmt.Errorf("failed to get available images: %w", err)
	}

	selectedImage, selectedID, err := display.RunGenericChoicePicker("Please select an image", imageChoices, 30)
	if err != nil {
		return "", "", err
	}

	return selectedImage, selectedID, nil
}

func runSSHKeySelector() (string, string, error) {
	endpoint := "/v1/me/sshKey"

	keys, err := httpLib.FetchExpandedArray(endpoint, "")
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch account ssh keys: %w", err)
	}

	if len(keys) == 0 {
		return "", "", errors.New("no SSH keys found in your account")
	}

	keyChoices := make(map[string]string, len(keys))
	for _, key := range keys {
		name := key["keyName"].(string)
		keyChoices[name] = key["key"].(string)
	}

	return display.RunGenericChoicePicker("Please select an SSH key", keyChoices, 10)
}
