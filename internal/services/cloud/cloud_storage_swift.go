// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	swiftContainerColumnsToDisplay = []string{
		"id",
		"targetSpec.name name",
		"targetSpec.location.region region",
		"targetSpec.type type",
		"resourceStatus status",
	}

	swiftObjectColumnsToDisplay = []string{
		"key",
		"size",
		"contentType",
		"lastModifiedAt",
	}

	//go:embed templates/cloud_storage_swift.tmpl
	cloudStorageSwiftTemplate string

	//go:embed templates/cloud_storage_swift_object.tmpl
	cloudStorageSwiftObjectTemplate string

	//go:embed parameter-samples/storage-swift-container-create.json
	SwiftContainerCreateExample string

	SwiftContainerSpec struct {
		TargetSpec struct {
			Name     string `json:"name,omitempty"`
			Location struct {
				Region string `json:"region,omitempty"`
			} `json:"location,omitzero"`
			Archive bool     `json:"archive,omitempty"`
			Type    string   `json:"type,omitempty"`
			Cors    []string `json:"cors,omitempty"`
		} `json:"targetSpec"`
	}

	SwiftContainerEditSpec struct {
		TargetSpec struct {
			Type string   `json:"type,omitempty"`
			Cors []string `json:"cors,omitempty"`
		} `json:"targetSpec,omitzero"`
	}

	SwiftContainerPresignSpec struct {
		Key       string `json:"key"`
		ExpiresIn int    `json:"expiresIn,omitempty"`
	}

	SwiftObjectListPrefix string
)

func swiftContainerV2Endpoint(projectID string) string {
	return fmt.Sprintf("/v2/publicCloud/project/%s/storage/swift/container", projectID)
}

func swiftContainerEndpoint(projectID, containerID string) string {
	return fmt.Sprintf("%s/%s", swiftContainerV2Endpoint(projectID), url.PathEscape(containerID))
}

func ListCloudStorageSwift(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(swiftContainerV2Endpoint(projectID), swiftContainerColumnsToDisplay, flags.GenericFilters)
}

func GetStorageSwift(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var container map[string]any
	if err := httpLib.Client.Get(swiftContainerEndpoint(projectID, args[0]), &container); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch container: %s", err)
		return
	}

	// formatByteSize in the template takes a float64
	if state, ok := container["currentState"].(map[string]any); ok {
		if size, ok := state["objectsSize"].(json.Number); ok {
			sizeFloat, err := size.Float64()
			if err != nil {
				display.OutputError(&flags.OutputFormatConfig, "error parsing used storage: %s", err)
				return
			}
			state["objectsSize"] = sizeFloat
		}
	}

	display.OutputObject(container, args[0], cloudStorageSwiftTemplate, &flags.OutputFormatConfig)
}

func CreateStorageSwift(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		display.OutputError(&flags.OutputFormatConfig, "region argument is required\n\n%s", cmd.UsageString())
		return
	}

	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	SwiftContainerSpec.TargetSpec.Location.Region = args[0]
	container, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/swift/container",
		swiftContainerV2Endpoint(projectID),
		SwiftContainerCreateExample,
		SwiftContainerSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec.name"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create container: %s", err)
		return
	}

	containerID, _ := container["id"].(string)

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, container, "⚡️ Container %s creation started", containerID)
		return
	}

	ready, err := waitForCloudResourceReady(swiftContainerEndpoint(projectID, containerID), 10*time.Minute)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for container creation: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, ready, "✅ Container %s created successfully", containerID)
}

func EditStorageSwift(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := swiftContainerEndpoint(projectID, args[0])
	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/storage/swift/container/{containerId}",
		endpoint,
		&SwiftContainerEditSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if !flags.WaitForTask {
		return
	}

	ready, err := waitForCloudResourceReady(endpoint, 10*time.Minute)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for container update: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, ready, "✅ Container %s updated successfully", args[0])
}

func DeleteStorageSwift(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(swiftContainerEndpoint(projectID, args[0]), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete container: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Container %s is being deleted", args[0])
}

func PresignStorageSwift(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	SwiftContainerPresignSpec.Key = args[1]

	var response map[string]any
	if err := httpLib.Client.Post(swiftContainerEndpoint(projectID, args[0])+"/presign", SwiftContainerPresignSpec, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to generate presigned URL: %s", err)
		return
	}

	var sb strings.Builder
	sb.WriteString("✅ Presigned URL generated successfully:\n")
	sb.WriteString(fmt.Sprintf("-> %s %s\n", response["method"], response["url"]))
	sb.WriteString(fmt.Sprintf("-> Expires at: %s", response["expiresAt"]))

	display.OutputInfo(&flags.OutputFormatConfig, response, "%s", &sb)
}

func ListStorageSwiftObjects(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := swiftContainerEndpoint(projectID, args[0]) + "/object"
	if SwiftObjectListPrefix != "" {
		endpoint += "?" + url.Values{"prefix": {SwiftObjectListPrefix}}.Encode()
	}

	common.ManageListRequestNoExpand(endpoint, swiftObjectColumnsToDisplay, flags.GenericFilters)
}

func GetStorageSwiftObject(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(swiftContainerEndpoint(projectID, args[0])+"/object", args[1], cloudStorageSwiftObjectTemplate)
}

func DeleteStorageSwiftObject(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("%s/object/%s", swiftContainerEndpoint(projectID, args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete object: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Object %s deleted from container %s", args[1], args[0])
}
