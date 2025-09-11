// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectLoadbalancerColumnsToDisplay = []string{"id", "openstackRegion", "size", "status"}

	//go:embed templates/cloud_loadbalancer.tmpl
	cloudLoadbalancerTemplate string

	CloudLoadbalancerUpdateFields struct {
		Description string `json:"description,omitempty"`
		Name        string `json:"name,omitempty"`
		Size        string `json:"size,omitempty"`
	}
)

func ListCloudLoadbalancers(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageListRequest(fmt.Sprintf("/cloud/project/%s/loadbalancer", projectID), "", cloudprojectLoadbalancerColumnsToDisplay, flags.GenericFilters)
}

func GetCloudLoadbalancer(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/loadbalancer", projectID), args[0], cloudLoadbalancerTemplate)
}

func EditCloudLoadbalancer(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/loadbalancer/{loadBalancerId}",
		fmt.Sprintf("/cloud/project/%s/loadbalancer/%s", projectID, url.PathEscape(args[0])),
		CloudLoadbalancerUpdateFields,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
