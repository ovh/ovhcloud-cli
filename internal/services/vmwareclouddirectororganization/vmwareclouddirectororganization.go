// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package vmwareclouddirectororganization

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
	vmwareclouddirectororganizationColumnsToDisplay = []string{"id", "currentState.fullName", "currentState.region", "resourceStatus"}

	//go:embed templates/vmwareclouddirectororganization.tmpl
	vmwareclouddirectororganizationTemplate string

	VmwareCloudDirectorOrganizationSpec struct {
		TargetSpec struct {
			Description string `json:"description,omitempty"`
			FullName    string `json:"fullName,omitempty"`
		} `json:"targetSpec,omitzero"`
	}
)

func ListVmwareCloudDirectorOrganization(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v2/vmwareCloudDirector/organization", "id", vmwareclouddirectororganizationColumnsToDisplay, flags.GenericFilters)
}

func GetVmwareCloudDirectorOrganization(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v2/vmwareCloudDirector/organization", args[0], vmwareclouddirectororganizationTemplate)
}

func EditVmwareCloudDirectorOrganization(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/vmwareCloudDirector/organization/{organizationId}",
		fmt.Sprintf("/v2/vmwareCloudDirector/organization/%s", url.PathEscape(args[0])),
		VmwareCloudDirectorOrganizationSpec,
		assets.VmwareclouddirectororganizationOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
