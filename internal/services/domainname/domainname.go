// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package domainname

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
	domainnameColumnsToDisplay = []string{"domain", "state", "whoisOwner", "expirationDate", "renewalDate"}

	//go:embed templates/domainname.tmpl
	domainnameTemplate string

	DomainSpec struct {
		NameServerType    string `json:"nameServerType,omitempty"`
		TranferLockStatus string `json:"transferLockStatus,omitempty"`
	}
)

func ListDomainName(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/domain", "", domainnameColumnsToDisplay, flags.GenericFilters)
}

func GetDomainName(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/domain", args[0], domainnameTemplate)
}

func EditDomainName(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/domain/{serviceName}",
		fmt.Sprintf("/v1/domain/%s", url.PathEscape(args[0])),
		DomainSpec,
		assets.DomainOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
