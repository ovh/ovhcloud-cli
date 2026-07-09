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
	cloudSecurityGroupColumnsToDisplay = []string{"id", "currentState.name name", "currentState.location.region region", "resourceStatus"}

	//go:embed templates/cloud_security_group.tmpl
	cloudSecurityGroupTemplate string

	//go:embed parameter-samples/security-group-create.json
	SecurityGroupCreationExample string

	// SecurityGroupSpec holds the parameters used to create or edit a security group.
	SecurityGroupSpec struct {
		TargetSpec struct {
			Name        string                    `json:"name,omitempty"`
			Description string                    `json:"description,omitempty"`
			Location    securityGroupLocation     `json:"location,omitzero"`
			Rules       []securityGroupTargetRule `json:"rules,omitempty"`
		} `json:"targetSpec"`
	}
)

type (
	securityGroupLocation struct {
		Region           string `json:"region,omitempty"`
		AvailabilityZone string `json:"availabilityZone,omitempty"`
	}

	securityGroupTargetRule struct {
		Description    string `json:"description,omitempty"`
		Direction      string `json:"direction,omitempty"`
		EthernetType   string `json:"ethernetType,omitempty"`
		PortRangeMin   int    `json:"portRangeMin,omitempty"`
		PortRangeMax   int    `json:"portRangeMax,omitempty"`
		Protocol       string `json:"protocol,omitempty"`
		RemoteIpPrefix string `json:"remoteIpPrefix,omitempty"`
	}
)

func ListSecurityGroups(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		fmt.Sprintf("/v2/publicCloud/project/%s/securityGroup", projectID),
		cloudSecurityGroupColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetSecurityGroup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v2/publicCloud/project/%s/securityGroup", projectID),
		args[0],
		cloudSecurityGroupTemplate,
	)
}

func CreateSecurityGroup(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/securityGroup", projectID)
	securityGroup, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/securityGroup",
		endpoint,
		SecurityGroupCreationExample,
		SecurityGroupSpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create security group: %s", err)
		return
	}

	securityGroupID, _ := securityGroup["id"].(string)

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, securityGroup, "✅ Security group created successfully (id: %s)", securityGroupID)
		return
	}

	if err := waitForCloudResourceReady(fmt.Sprintf("%s/%s", endpoint, url.PathEscape(securityGroupID)), 30*time.Minute); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for security group creation: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, securityGroup, "✅ Security group %s created successfully", securityGroupID)
}

func EditSecurityGroup(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/securityGroup/{securityGroupId}",
		fmt.Sprintf("/v2/publicCloud/project/%s/securityGroup/%s", projectID, url.PathEscape(args[0])),
		SecurityGroupSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteSecurityGroup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/securityGroup/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete security group: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Security group %s is being deleted…", args[0])
}
