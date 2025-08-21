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

	sshKeysColumnsToDisplay = []string{"keyName name", "key"}

	Oauth2ClientSpec struct {
		CallbackUrls []string `json:"callbackUrls,omitempty"`
		Description  string   `json:"description,omitempty"`
		Flow         string   `json:"flow,omitempty"`
		Name         string   `json:"name,omitempty"`
	}
)

func ListSSHKeys(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/me/sshKey", "", sshKeysColumnsToDisplay, flags.GenericFilters)
}

func CreateOAuth2Client(cmd *cobra.Command, args []string) {
	client, err := common.CreateResource(
		cmd,
		"/me/api/oauth2/client",
		"/me/api/oauth2/client",
		Oauth2ClientCreateSample,
		Oauth2ClientSpec,
		assets.MeOpenapiSchema,
		[]string{"name", "description", "flow"},
	)
	if err != nil {
		display.ExitError("failed to create OAuth2 client: %s", err)
		return
	}

	fmt.Println("✅ OAuth2 client created successfully")
	fmt.Printf("Client ID: %s\n", client["clientId"].(string))
	fmt.Printf("Client Secret: %s\n", client["clientSecret"].(string))
}

func ListOAuth2Clients(_ *cobra.Command, _ []string) {
	endpoint := "/me/api/oauth2/client"
	common.ManageListRequest(endpoint, "", []string{"clientId", "name", "description", "flow", "createdAt"}, flags.GenericFilters)
}

func GetOauth2Client(cmd *cobra.Command, args []string) {
	common.ManageObjectRequest("/me/api/oauth2/client", args[0], "")
}

func DeleteOauth2Client(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/me/api/oauth2/client/%s", url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("failed to delete OAuth2 client: %s", err)
		return
	}

	fmt.Printf("✅ OAuth2 client '%s' deleted successfully\n", args[0])
}

func EditOauth2Client(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/me/api/oauth2/client/{clientId}",
		fmt.Sprintf("/me/api/oauth2/client/%s", url.PathEscape(args[0])),
		Oauth2ClientSpec,
		assets.MeOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
