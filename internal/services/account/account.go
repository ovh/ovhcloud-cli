// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package account

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
	//go:embed parameter-samples/oauth2-client-create.json
	Oauth2ClientCreateSample string

	//go:embed templates/me.tmpl
	meTemplate string

	sshKeysColumnsToDisplay = []string{"keyName name", "key"}

	Oauth2ClientSpec struct {
		CallbackUrls []string `json:"callbackUrls,omitempty"`
		Description  string   `json:"description,omitempty"`
		Flow         string   `json:"flow,omitempty"`
		Name         string   `json:"name,omitempty"`
	}
)

func GetMe(_ *cobra.Command, _ []string) {
	common.ManageObjectRequest("/me", "", meTemplate)
}

func ListSSHKeys(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/me/sshKey", "", sshKeysColumnsToDisplay, flags.GenericFilters)
}

func CreateOAuth2Client(cmd *cobra.Command, args []string) {
	client, err := common.CreateResource(
		cmd,
		"/me/api/oauth2/client",
		"/v1/me/api/oauth2/client",
		Oauth2ClientCreateSample,
		Oauth2ClientSpec,
		assets.MeOpenapiSchema,
		[]string{"name", "description", "flow"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create OAuth2 client: %s", err)
		return
	}

	display.OutputInfo(
		&flags.OutputFormatConfig,
		client,
		"✅ OAuth2 client created successfully (client ID: %s, client secret: %s)",
		client["clientId"],
		client["clientSecret"],
	)
}

func ListOAuth2Clients(_ *cobra.Command, _ []string) {
	endpoint := "/v1/me/api/oauth2/client"
	common.ManageListRequest(endpoint, "", []string{"clientId", "name", "description", "flow", "createdAt"}, flags.GenericFilters)
}

func GetOauth2Client(cmd *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/me/api/oauth2/client", args[0], "")
}

func DeleteOauth2Client(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/me/api/oauth2/client/%s", url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete OAuth2 client: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ OAuth2 client '%s' deleted successfully", args[0])
}

func EditOauth2Client(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/me/api/oauth2/client/{clientId}",
		fmt.Sprintf("/v1/me/api/oauth2/client/%s", url.PathEscape(args[0])),
		Oauth2ClientSpec,
		assets.MeOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
