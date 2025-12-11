// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/display"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

type CloudProjectOperation struct {
	Id            string                     `json:"id"`
	Action        string                     `json:"action"`
	CreateAt      string                     `json:"createdAt"`
	StartedAt     string                     `json:"startedAt"`
	CompletedAt   *string                    `json:"completedAt"`
	Progress      int                        `json:"progress"`
	Regions       []string                   `json:"regions"`
	ResourceId    *string                    `json:"resourceId"`
	Status        string                     `json:"status"`
	SubOperations []CloudProjectSubOperation `json:"subOperations"`
}

type CloudProjectSubOperation struct {
	ResourceId *string `json:"resourceId"`
	Action     string  `json:"action"`
}

func getAvailableImages(projectID string, region string) (map[string]string, error) {
	// Fetch available images for the project
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/image", projectID)
	if region != "" {
		endpoint += "?region=" + url.QueryEscape(region)
	}

	var images []map[string]any
	if err := httpLib.Client.Get(endpoint, &images); err != nil {
		return nil, fmt.Errorf("failed to fetch available images: %w", err)
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no image found for project %s", projectID)
	}

	imageChoices := make(map[string]string, len(images))
	for _, img := range images {
		name := img["name"].(string)
		imageChoices[name] = img["id"].(string)
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
	// Fetch available flavors for the project
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/flavor", projectID)
	if region != "" {
		endpoint += "?region=" + url.QueryEscape(region)
	}

	var flavors []map[string]any
	if err := httpLib.Client.Get(endpoint, &flavors); err != nil {
		return nil, fmt.Errorf("failed to fetch available flavors: %w", err)
	}

	flavorChoices := make(map[string]string, len(flavors))
	for _, flavor := range flavors {
		name := flavor["name"].(string)
		flavorChoices[name] = flavor["id"].(string)
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

func waitForCloudOperation(projectID, operationID, action string, retryDuration time.Duration) (string, error) {
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/operation/%s", url.PathEscape(projectID), url.PathEscape(operationID))
	resourceID := ""

	ctx, cancel := context.WithTimeout(context.Background(), retryDuration)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		var operation CloudProjectOperation
		if err := httpLib.Client.Get(endpoint, &operation); err != nil {
			return "", fmt.Errorf("error fetching operation: %w", err)
		}

		switch operation.Status {
		case "in-error":
			return "", fmt.Errorf("operation %q ended in error", operation.Action)
		case "completed":
			if operation.ResourceId != nil {
				resourceID = *operation.ResourceId
			} else if len(operation.SubOperations) > 0 {
				for _, subOp := range operation.SubOperations {
					if subOp.Action == action && subOp.ResourceId != nil {
						resourceID = *subOp.ResourceId
						break
					}
				}
			}
			return resourceID, nil
		}

		select {
		case <-ctx.Done():
			return "", errors.New("timeout waiting for operation to complete")
		case <-ticker.C:
			log.Printf("Still waiting for operation to complete (status=%s)…", operation.Status)
			continue
		}
	}
}
