package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectUserColumnsToDisplay = []string{"id", "username", "description", "status"}

	//go:embed templates/cloud_user.tmpl
	cloudUserTemplate string

	//go:embed parameter-samples/user-create.json
	UserCreateExample string

	UserSpec struct {
		Description string   `json:"description,omitempty"`
		Roles       []string `json:"roles,omitempty"`
	}
)

func ListCloudUsers(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}
	path := fmt.Sprintf("/cloud/project/%s/user", projectID)

	var body []map[string]any
	if err := httpLib.Client.Get(path, &body); err != nil {
		display.ExitError("failed to fetch SSH keys: %s", err)
		return
	}

	body, err = filtersLib.FilterLines(body, flags.GenericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
		return
	}

	display.RenderTable(body, cloudprojectUserColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetCloudUser(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/user", projectID), args[0], cloudUserTemplate)
}

func CreateCloudUser(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	client, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/user",
		fmt.Sprintf("/cloud/project/%s/user", projectID),
		UserCreateExample,
		UserSpec,
		assets.CloudOpenapiSchema,
		nil,
	)
	if err != nil {
		display.ExitError("failed to create user: %s", err)
		return
	}

	fmt.Printf("✅ User '%s' created successfully\n", client["id"])
}

func DeleteCloudUser(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/user/%s", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("failed to delete user: %s", err)
		return
	}

	fmt.Printf("✅ User '%s' deleted successfully\n", args[0])
}
