// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package iploadbalancing

import (
	_ "embed"
	"fmt"
	"net/url"
	"strconv"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	iploadbalancingSSLColumnsToDisplay = []string{"id"}

	//go:embed templates/iploadbalancing_ssl.tmpl
	iploadbalancingSSLTemplate string

	// IPLoadbalancingSSLCreateExample is the default parameters example used by --init-file and --editor
	//go:embed parameter-samples/ssl-create.json
	IPLoadbalancingSSLCreateExample string

	// IPLoadbalancingSSLSpec holds the parameters given to the SSL certificate creation command
	IPLoadbalancingSSLSpec struct {
		Certificate string `json:"certificate,omitempty"`
		Chain       string `json:"chain,omitempty"`
		DisplayName string `json:"displayName,omitempty"`
		Key         string `json:"key,omitempty"`
	}
)

// ListIpLoadbalancingSSL lists the SSL certificate IDs of the given IpLoadbalancing service
func ListIpLoadbalancingSSL(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/ipLoadbalancing/%s/ssl", url.PathEscape(args[0]))

	var sslIDs []int
	if err := httpLib.Client.Get(endpoint, &sslIDs); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch SSL certificates: %s", err)
		return
	}

	rows := make([]map[string]any, 0, len(sslIDs))
	for _, id := range sslIDs {
		rows = append(rows, map[string]any{"id": id})
	}

	display.RenderTable(rows, iploadbalancingSSLColumnsToDisplay, &flags.OutputFormatConfig)
}

// GetIpLoadbalancingSSL retrieves the given SSL certificate of an IpLoadbalancing service
func GetIpLoadbalancingSSL(_ *cobra.Command, args []string) {
	if _, err := strconv.Atoi(args[1]); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "invalid SSL ID given, expected a number")
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v1/ipLoadbalancing/%s/ssl", url.PathEscape(args[0])), args[1], iploadbalancingSSLTemplate)
}

// CreateIpLoadbalancingSSL adds a new custom SSL certificate on the given IpLoadbalancing service
func CreateIpLoadbalancingSSL(cmd *cobra.Command, args []string) {
	ssl, err := common.CreateResource(
		cmd,
		"/ipLoadbalancing/{serviceName}/ssl",
		fmt.Sprintf("/v1/ipLoadbalancing/%s/ssl", url.PathEscape(args[0])),
		IPLoadbalancingSSLCreateExample,
		IPLoadbalancingSSLSpec,
		assets.IploadbalancingOpenapiSchema,
		[]string{"certificate", "key"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create SSL certificate: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, ssl, "✅ SSL certificate %v created successfully", ssl["id"])
}

// DeleteIpLoadbalancingSSL deletes the given custom SSL certificate of an IpLoadbalancing service
func DeleteIpLoadbalancingSSL(_ *cobra.Command, args []string) {
	if _, err := strconv.Atoi(args[1]); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "invalid SSL ID given, expected a number")
		return
	}

	endpoint := fmt.Sprintf("/v1/ipLoadbalancing/%s/ssl/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete SSL certificate %s: %s", args[1], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ SSL certificate %s deleted successfully", args[1])
}
