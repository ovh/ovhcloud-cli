package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	cloudprojectStorageSwiftColumnsToDisplay = []string{"id", "name", "region"}

	//go:embed templates/cloud_storage_swift.tmpl
	cloudStorageSwiftTemplate string
)

func ListCloudStorageSwift(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageListRequestNoExpand(fmt.Sprintf("/cloud/project/%s/storage", projectID), cloudprojectStorageSwiftColumnsToDisplay, flags.GenericFilters)
}

func GetStorageSwift(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/storage", projectID), args[0], cloudStorageSwiftTemplate)
}

func EditStorageSwift(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/storage/%s", projectID, url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/cloud/project/{serviceName}/storage/{containerId}", endpoint, CloudOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}
