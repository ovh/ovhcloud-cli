// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package ip

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	ipColumnsToDisplay = []string{"ip", "rir", "routedTo.serviceName", "country", "description"}

	//go:embed templates/ip.tmpl
	ipTemplate string

	IPSpec struct {
		Description string `json:"description,omitempty"`
	}
)

func ListIp(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/ip", "", ipColumnsToDisplay, flags.GenericFilters)
}

func GetIp(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/ip", args[0], ipTemplate)
}

func EditIp(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/ip/{ip}",
		fmt.Sprintf("/v1/ip/%s", url.PathEscape(args[0])),
		IPSpec,
		assets.IpOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func IpSetReverse(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/v1/ip/%s/reverse", url.PathEscape(args[0]))
	if err := httpLib.Client.Post(url, map[string]string{
		"ipReverse": args[1],
		"reverse":   args[2],
	}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Reverse correctly set")
}

func IpGetReverse(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/v1/ip/%s/reverse", url.PathEscape(args[0]))
	common.ManageListRequest(url, "", []string{"ipReverse", "reverse"}, flags.GenericFilters)
}

func IpDeleteReverse(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/v1/ip/%s/reverse/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(url, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Reverse correctly deleted")
}
