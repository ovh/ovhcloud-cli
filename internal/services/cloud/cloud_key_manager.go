// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
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
	keyManagerSecretColumnsToDisplay    = []string{"id", "currentState.name name", "currentState.location.region region", "currentState.secretType type", "resourceStatus"}
	keyManagerContainerColumnsToDisplay = []string{"id", "currentState.name name", "currentState.location.region region", "currentState.type type", "resourceStatus"}
	keyManagerConsumerColumnsToDisplay  = []string{"id", "resourceId", "resourceType", "service"}

	//go:embed templates/cloud_key_manager_secret.tmpl
	keyManagerSecretTemplate string

	//go:embed templates/cloud_key_manager_container.tmpl
	keyManagerContainerTemplate string

	//go:embed parameter-samples/key-manager-secret-create.json
	KeyManagerSecretCreateExample string

	//go:embed parameter-samples/key-manager-container-create.json
	KeyManagerContainerCreateExample string

	// Secret creation parameters
	KeyManagerSecretSpec struct {
		TargetSpec struct {
			Algorithm          string              `json:"algorithm,omitempty"`
			BitLength          int                 `json:"bitLength,omitempty"`
			Expiration         string              `json:"expiration,omitempty"`
			Location           *keyManagerLocation `json:"location,omitempty"`
			Metadata           map[string]string   `json:"metadata,omitempty"`
			Mode               string              `json:"mode,omitempty"`
			Name               string              `json:"name,omitempty"`
			Payload            string              `json:"payload,omitempty"`
			PayloadContentType string              `json:"payloadContentType,omitempty"`
			SecretType         string              `json:"secretType,omitempty"`
		} `json:"targetSpec"`
	}
	KeyManagerSecretCreateRegion           string
	KeyManagerSecretCreateAvailabilityZone string

	// Secret edit parameters (only metadata is mutable)
	KeyManagerSecretEditSpec struct {
		TargetSpec struct {
			Metadata map[string]string `json:"metadata,omitempty"`
		} `json:"targetSpec"`
	}

	// Container creation parameters
	KeyManagerContainerSpec struct {
		TargetSpec struct {
			Location   *keyManagerLocation   `json:"location,omitempty"`
			Name       string                `json:"name,omitempty"`
			SecretRefs []keyManagerSecretRef `json:"secretRefs,omitempty"`
			Type       string                `json:"type,omitempty"`
		} `json:"targetSpec"`
	}
	KeyManagerContainerCreateRegion           string
	KeyManagerContainerCreateAvailabilityZone string
	KeyManagerContainerSecretRefs             []string

	// Container edit parameters (only secretRefs are mutable)
	KeyManagerContainerEditSpec struct {
		TargetSpec struct {
			SecretRefs []keyManagerSecretRef `json:"secretRefs"`
		} `json:"targetSpec"`
	}

	// Consumer register parameters (shared by secret and container)
	KeyManagerConsumerSpec struct {
		ResourceId   string `json:"resourceId,omitempty"`
		ResourceType string `json:"resourceType,omitempty"`
		Service      string `json:"service,omitempty"`
	}
)

type (
	keyManagerLocation struct {
		Region           string `json:"region,omitempty"`
		AvailabilityZone string `json:"availabilityZone,omitempty"`
	}

	keyManagerSecretRef struct {
		Name   string `json:"name"`
		Secret struct {
			Id string `json:"id"`
		} `json:"secret"`
	}
)

// parseSecretRefs converts "role=secretId" (or "role:secretId") strings into
// keyManagerSecretRef structures.
func parseSecretRefs(refs []string) ([]keyManagerSecretRef, error) {
	result := make([]keyManagerSecretRef, 0, len(refs))
	for _, ref := range refs {
		sep := strings.IndexAny(ref, "=:")
		if sep <= 0 || sep == len(ref)-1 {
			return nil, fmt.Errorf("invalid secret reference %q, expected format '<name>=<secretId>'", ref)
		}
		var r keyManagerSecretRef
		r.Name = ref[:sep]
		r.Secret.Id = ref[sep+1:]
		result = append(result, r)
	}
	return result, nil
}

//
// Secret commands
//

func ListKeyManagerSecrets(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/secret", projectID), keyManagerSecretColumnsToDisplay, flags.GenericFilters)
}

func GetKeyManagerSecret(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/secret", projectID), args[0], keyManagerSecretTemplate)
}

func CreateKeyManagerSecret(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if KeyManagerSecretCreateRegion != "" || KeyManagerSecretCreateAvailabilityZone != "" {
		KeyManagerSecretSpec.TargetSpec.Location = &keyManagerLocation{
			Region:           KeyManagerSecretCreateRegion,
			AvailabilityZone: KeyManagerSecretCreateAvailabilityZone,
		}
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/secret", projectID)
	secret, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/keyManager/secret",
		endpoint,
		KeyManagerSecretCreateExample,
		KeyManagerSecretSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create Key Manager secret: %s", err)
		return
	}

	secretID, _ := secret["id"].(string)

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, secret, "✅ Key Manager secret created successfully (id: %s)", secretID)
		return
	}

	if err := waitForCloudResourceReady(fmt.Sprintf("%s/%s", endpoint, url.PathEscape(secretID)), 20*time.Minute); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for Key Manager secret to be ready: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, secret, "✅ Key Manager secret %s is now ready", secretID)
}

func EditKeyManagerSecret(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/keyManager/secret/{secretId}",
		fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/secret/%s", projectID, url.PathEscape(args[0])),
		KeyManagerSecretEditSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteKeyManagerSecret(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/secret/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete Key Manager secret: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Key Manager secret %s deleted successfully", args[0])
}

// GetKeyManagerSecretPayload fetches the payload (sensitive material) of a secret.
// The API exposes this as a POST that returns the payload in its response body.
func GetKeyManagerSecretPayload(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/secret/%s/payload", projectID, url.PathEscape(args[0]))
	var response map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch Key Manager secret payload: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, response, "%s", response["payload"])
}

//
// Secret consumer commands
//

func ListKeyManagerSecretConsumers(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/secret/%s/consumer", projectID, url.PathEscape(args[0])),
		keyManagerConsumerColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetKeyManagerSecretConsumer(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/secret/%s/consumer", projectID, url.PathEscape(args[0])),
		args[1],
		"",
	)
}

func RegisterKeyManagerSecretConsumer(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/secret/%s/consumer", projectID, url.PathEscape(args[0]))
	var response map[string]any
	if err := httpLib.Client.Post(endpoint, KeyManagerConsumerSpec, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to register consumer for Key Manager secret: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, response, "✅ Consumer registered successfully for secret %s (id: %s)", args[0], response["id"])
}

func DeleteKeyManagerSecretConsumer(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/secret/%s/consumer/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete consumer from Key Manager secret: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Consumer %s deleted successfully from secret %s", args[1], args[0])
}

//
// Container commands
//

func ListKeyManagerContainers(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/container", projectID), keyManagerContainerColumnsToDisplay, flags.GenericFilters)
}

func GetKeyManagerContainer(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/container", projectID), args[0], keyManagerContainerTemplate)
}

func CreateKeyManagerContainer(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if KeyManagerContainerCreateRegion != "" || KeyManagerContainerCreateAvailabilityZone != "" {
		KeyManagerContainerSpec.TargetSpec.Location = &keyManagerLocation{
			Region:           KeyManagerContainerCreateRegion,
			AvailabilityZone: KeyManagerContainerCreateAvailabilityZone,
		}
	}

	if len(KeyManagerContainerSecretRefs) > 0 {
		refs, err := parseSecretRefs(KeyManagerContainerSecretRefs)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}
		KeyManagerContainerSpec.TargetSpec.SecretRefs = refs
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/container", projectID)
	container, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/keyManager/container",
		endpoint,
		KeyManagerContainerCreateExample,
		KeyManagerContainerSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create Key Manager container: %s", err)
		return
	}

	containerID, _ := container["id"].(string)

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, container, "✅ Key Manager container created successfully (id: %s)", containerID)
		return
	}

	if err := waitForCloudResourceReady(fmt.Sprintf("%s/%s", endpoint, url.PathEscape(containerID)), 20*time.Minute); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for Key Manager container to be ready: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, container, "✅ Key Manager container %s is now ready", containerID)
}

func EditKeyManagerContainer(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if len(KeyManagerContainerSecretRefs) > 0 {
		refs, err := parseSecretRefs(KeyManagerContainerSecretRefs)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}
		KeyManagerContainerEditSpec.TargetSpec.SecretRefs = refs
	}

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/keyManager/container/{containerId}",
		fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/container/%s", projectID, url.PathEscape(args[0])),
		KeyManagerContainerEditSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteKeyManagerContainer(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/container/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete Key Manager container: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Key Manager container %s deleted successfully", args[0])
}

//
// Container consumer commands
//

func ListKeyManagerContainerConsumers(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/container/%s/consumer", projectID, url.PathEscape(args[0])),
		keyManagerConsumerColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetKeyManagerContainerConsumer(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/container/%s/consumer", projectID, url.PathEscape(args[0])),
		args[1],
		"",
	)
}

func RegisterKeyManagerContainerConsumer(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/container/%s/consumer", projectID, url.PathEscape(args[0]))
	var response map[string]any
	if err := httpLib.Client.Post(endpoint, KeyManagerConsumerSpec, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to register consumer for Key Manager container: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, response, "✅ Consumer registered successfully for container %s (id: %s)", args[0], response["id"])
}

func DeleteKeyManagerContainerConsumer(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/keyManager/container/%s/consumer/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete consumer from Key Manager container: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Consumer %s deleted successfully from container %s", args[1], args[0])
}
