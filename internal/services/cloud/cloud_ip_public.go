// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"
	"net/url"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	// Columns displayed for the various public IP kinds (API v2).
	cloudPublicIPColumnsToDisplay           = []string{"ip", "type"}
	cloudPublicIPFloatingColumnsToDisplay   = []string{"id", "currentState.status status", "currentState.location.region region", "resourceStatus"}
	cloudPublicIPAdditionalColumnsToDisplay = []string{"id", "currentState.associatedResource.id associatedResource", "resourceStatus"}
	cloudPublicIPExtNetColumnsToDisplay     = []string{"id", "currentState.location.region region", "resourceStatus"}

	//go:embed templates/cloud_ip_public_floating.tmpl
	cloudPublicIPFloatingTemplate string

	//go:embed templates/cloud_ip_public_additional.tmpl
	cloudPublicIPAdditionalTemplate string

	//go:embed templates/cloud_ip_public_extnet.tmpl
	cloudPublicIPExtNetTemplate string

	//go:embed parameter-samples/public-ip-floating-create.json
	PublicIPFloatingCreationExample string

	// PublicIPFloatingCreationSpec holds the parameters for creating a floating IP.
	PublicIPFloatingCreationSpec struct {
		TargetSpec struct {
			Description string `json:"description,omitempty"`
			Location    struct {
				Region           string `json:"region,omitempty"`
				AvailabilityZone string `json:"availabilityZone,omitempty"`
			} `json:"location,omitzero"`
		} `json:"targetSpec"`
	}

	// PublicIPFloatingUpdateSpec holds the mutable parameters for a floating IP.
	PublicIPFloatingUpdateSpec struct {
		TargetSpec struct {
			Description string `json:"description,omitempty"`
		} `json:"targetSpec,omitzero"`
	}
)

// ---------------------------------------------------------------------------
// All public IPs (summary)
// ---------------------------------------------------------------------------

func ListAllPublicIPs(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		fmt.Sprintf("/v2/publicCloud/project/%s/publicIp", projectID),
		cloudPublicIPColumnsToDisplay,
		flags.GenericFilters,
	)
}

// ---------------------------------------------------------------------------
// Floating IPs (full CRUD)
// ---------------------------------------------------------------------------

func ListPublicIPFloating(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		fmt.Sprintf("/v2/publicCloud/project/%s/publicIp/floating", projectID),
		cloudPublicIPFloatingColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetPublicIPFloating(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v2/publicCloud/project/%s/publicIp/floating", projectID),
		args[0],
		cloudPublicIPFloatingTemplate,
	)
}

func CreatePublicIPFloating(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/publicIp/floating", projectID)
	floatingIP, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/publicIp/floating",
		endpoint,
		PublicIPFloatingCreationExample,
		PublicIPFloatingCreationSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create floating IP: %s", err)
		return
	}

	id, _ := floatingIP["id"].(string)

	// Wait for the floating IP to be ready if --wait flag is set
	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, floatingIP, "✅ Floating IP creation started successfully (id: %s)", id)
		return
	}

	resourceEndpoint := fmt.Sprintf("%s/%s", endpoint, url.PathEscape(id))
	if _, err := waitForCloudResourceReady(resourceEndpoint, 20*time.Minute); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for floating IP creation: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Floating IP %s created successfully", id)
}

func EditPublicIPFloating(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/publicIp/floating/{id}",
		fmt.Sprintf("/v2/publicCloud/project/%s/publicIp/floating/%s", projectID, url.PathEscape(args[0])),
		PublicIPFloatingUpdateSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeletePublicIPFloating(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/publicIp/floating/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete floating IP: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Floating IP %s deleted successfully", args[0])
}

// ---------------------------------------------------------------------------
// Additional IPs (read-only)
// ---------------------------------------------------------------------------

func ListPublicIPAdditional(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		fmt.Sprintf("/v2/publicCloud/project/%s/publicIp/additional", projectID),
		cloudPublicIPAdditionalColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetPublicIPAdditional(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v2/publicCloud/project/%s/publicIp/additional", projectID),
		args[0],
		cloudPublicIPAdditionalTemplate,
	)
}

// AttachPublicIPAdditional attaches an additional IP to an instance. Attach is
// not available on the v2 publicIp API yet, so it still uses the v1 endpoint.
func AttachPublicIPAdditional(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/ip/failover/%s/attach", projectID, url.PathEscape(args[0]))

	var result map[string]any
	if err := httpLib.Client.Post(endpoint, map[string]string{"instanceId": args[1]}, &result); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to attach additional IP %q to instance %q: %s", args[0], args[1], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Additional IP %s attached to instance %s successfully", args[0], args[1])
}

// ---------------------------------------------------------------------------
// Ext-Net IPs (list, get, delete)
// ---------------------------------------------------------------------------

func ListPublicIPExtNet(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		fmt.Sprintf("/v2/publicCloud/project/%s/publicIp/extNet", projectID),
		cloudPublicIPExtNetColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetPublicIPExtNet(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v2/publicCloud/project/%s/publicIp/extNet", projectID),
		args[0],
		cloudPublicIPExtNetTemplate,
	)
}

func DeletePublicIPExtNet(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/publicIp/extNet/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete Ext-Net IP: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Ext-Net IP %s deleted successfully", args[0])
}
