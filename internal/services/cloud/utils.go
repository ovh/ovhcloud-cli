// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/display"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// getInstanceReferenceChoices fetches every page of the v2 instance flavors or
// images ("flavor" or "image") and maps their names to their IDs.
func getInstanceReferenceChoices(projectID, kind, region string) (map[string]string, error) {
	items, err := httpLib.FetchArray(instanceReferenceV2Endpoint(projectID, kind, region), "")
	if err != nil {
		return nil, err
	}

	choices := make(map[string]string, len(items))
	for _, item := range items {
		object, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name, _ := object["name"].(string)
		id, _ := object["id"].(string)
		if name != "" && id != "" {
			choices[name] = id
		}
	}

	return choices, nil
}

func getAvailableImages(projectID string, region string) (map[string]string, error) {
	imageChoices, err := getInstanceReferenceChoices(projectID, "image", region)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch available images: %w", err)
	}

	if len(imageChoices) == 0 {
		return nil, fmt.Errorf("no image found for project %s", projectID)
	}

	return imageChoices, nil
}

func runImageSelector(projectID string, region string) (string, string, error) {
	imageChoices, err := getAvailableImages(projectID, region)
	if err != nil {
		return "", "", fmt.Errorf("failed to get available images: %w", err)
	}

	selectedImage, selectedID, err := display.RunGenericChoicePicker("Please select an image", imageChoices, 30)
	if err != nil {
		return "", "", err
	}

	if selectedImage == "" {
		return "", "", nil
	}

	return selectedImage, selectedID, nil
}

func getAvailableFlavors(projectID string, region string) (map[string]string, error) {
	flavorChoices, err := getInstanceReferenceChoices(projectID, "flavor", region)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch available flavors: %w", err)
	}

	return flavorChoices, nil
}

func runFlavorSelector(projectID string, region string) (string, string, error) {
	flavorChoices, err := getAvailableFlavors(projectID, region)
	if err != nil {
		return "", "", fmt.Errorf("failed to get available flavors: %w", err)
	}

	selectedFlavor, selectedID, err := display.RunGenericChoicePicker("Please select a flavor", flavorChoices, 30)
	if err != nil {
		return "", "", err
	}

	if selectedFlavor == "" {
		return "", "", nil
	}

	return selectedFlavor, selectedID, nil
}

// waitForCloudResourceReady polls a v2 asynchronous resource at the given full
// endpoint (e.g. ".../storage/block/volume/{id}") until its "resourceStatus"
// reaches "READY". It fails if the resource reports an "ERROR" status and times
// out after retryDuration. Tasks reported in error are logged but not treated as
// fatal, as they can be transient. The last fetched resource is returned.
func waitForCloudResourceReady(endpoint string, retryDuration time.Duration) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), retryDuration)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		var resource map[string]any
		if err := httpLib.Client.Get(endpoint, &resource); err != nil {
			return nil, fmt.Errorf("error fetching resource: %w", err)
		}

		status, _ := resource["resourceStatus"].(string)
		switch status {
		case "READY":
			return resource, nil
		case "ERROR":
			return nil, errors.New("resource ended in error state (resourceStatus=ERROR)")
		}

		// Log tasks reported in error but keep waiting: such errors can be
		// temporary and get resolved by the backend.
		if tasks, ok := resource["currentTasks"].([]any); ok {
			for _, t := range tasks {
				task, ok := t.(map[string]any)
				if !ok {
					continue
				}
				if taskStatus, _ := task["status"].(string); taskStatus == "ERROR" {
					var messages []string
					if taskErrors, ok := task["errors"].([]any); ok {
						for _, e := range taskErrors {
							if em, ok := e.(map[string]any); ok {
								if msg, ok := em["message"].(string); ok && msg != "" {
									messages = append(messages, msg)
								}
							}
						}
					}
					if len(messages) > 0 {
						log.Printf("resource task %v is in error state: %s (still waiting…)", task["type"], strings.Join(messages, "; "))
					} else {
						log.Printf("resource task %v is in error state, still waiting…", task["type"])
					}
				}
			}
		}

		select {
		case <-ctx.Done():
			return nil, errors.New("timeout waiting for resource to be ready")
		case <-ticker.C:
			log.Printf("Still waiting for resource to be ready (resourceStatus=%s)…", status)
			continue
		}
	}
}
