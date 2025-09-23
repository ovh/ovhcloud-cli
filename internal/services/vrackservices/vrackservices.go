// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package vrackservices

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
	vrackservicesColumnsToDisplay = []string{"id", "currentState.region", "currentState. productStatus", "resourceStatus"}

	//go:embed templates/vrackservices.tmpl
	vrackservicesTemplate string
)

func ListVrackServices(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v2/vrackServices/resource", "id", vrackservicesColumnsToDisplay, flags.GenericFilters)
}

func GetVrackServices(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v2/vrackServices/resource", args[0], vrackservicesTemplate)
}

func EditVrackServices(cmd *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v2/vrackServices/resource/%s", url.PathEscape(args[0]))
	if err := common.EditResource(
		cmd,
		"/vrackServices/resource/{vrackServicesId}",
		endpoint,
		nil,
		assets.VrackservicesOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
	}
}
