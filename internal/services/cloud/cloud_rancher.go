// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"
	"net/url"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectRancherColumnsToDisplay = []string{"id", "currentState.name name", "currentState.region region", "currentState.version version", "resourceStatus"}

	//go:embed templates/cloud_rancher.tmpl
	cloudRancherTemplate string

	//go:embed parameter-samples/rancher-create.json
	CloudRancherCreationExample string

	RancherSpec struct {
		TargetSpec struct {
			IAMAuthEnabled    bool                   `json:"iamAuthEnabled,omitempty"`
			Name              string                 `json:"name,omitempty"`
			Plan              string                 `json:"plan,omitempty"`
			Version           string                 `json:"version,omitempty"`
			IPRestrictions    []rancherIPRestriction `json:"ipRestrictions,omitempty"`
			CLIIPRestrictions []string               `json:"-"`
		} `json:"targetSpec"`
	}
)

type (
	rancherIPRestriction struct {
		CIDRBlock   string `json:"cidrBlock"`
		Description string `json:"description"`
	}

	rancherUser struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
)

func ListCloudRanchers(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(fmt.Sprintf("/v2/publicCloud/project/%s/rancher", projectID), cloudprojectRancherColumnsToDisplay, flags.GenericFilters)
}

func GetRancher(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v2/publicCloud/project/%s/rancher", projectID), args[0], cloudRancherTemplate)
}

func EditRancher(cmd *cobra.Command, args []string) {
	for _, ipRestriction := range RancherSpec.TargetSpec.CLIIPRestrictions {
		parts := strings.Split(ipRestriction, ",")
		if len(parts) != 2 {
			display.OutputError(&flags.OutputFormatConfig, "Invalid IP restriction format: %s. Expected format: '<cidrBlock>,<description>'", ipRestriction)
			return
		}
		RancherSpec.TargetSpec.IPRestrictions = append(RancherSpec.TargetSpec.IPRestrictions, rancherIPRestriction{
			CIDRBlock:   parts[0],
			Description: parts[1],
		})
	}

	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/rancher/{rancherId}",
		fmt.Sprintf("/v2/publicCloud/project/%s/rancher/%s", projectID, url.PathEscape(args[0])),
		RancherSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func CreateRancher(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/rancher", projectID)
	rancher, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/rancher",
		endpoint,
		CloudRancherCreationExample,
		RancherSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create Rancher service: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, rancher, "✅ Rancher %s created successfully (id: %s)", RancherSpec.TargetSpec.Name, rancher["id"])
}

func ResetRancherAdminCredentials(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var user rancherUser
	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/rancher/%s/adminCredentials", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Post(endpoint, nil, &user); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to reset admin credentials for Rancher service: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, user, "✅ New Rancher service password for user %s: %s", user.Username, user.Password)
}

func DeleteRancher(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/rancher/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete Rancher service: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Rancher service is being deleted…")
}
