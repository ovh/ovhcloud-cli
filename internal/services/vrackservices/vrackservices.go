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
	"github.com/ovh/ovhcloud-cli/internal/editor"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
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

func EditVrackServices(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v2/vrackServices/resource/%s", url.PathEscape(args[0]))
	if err := editor.EditResource(
		httpLib.Client,
		"/vrackServices/resource/{vrackServicesId}",
		endpoint,
		assets.VrackservicesOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
	}
}
