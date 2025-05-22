package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	cloudprojectKubeColumnsToDisplay = []string{"id", "name", "region", "version", "status"}

	//go:embed templates/cloud_kube.tmpl
	cloudKubeTemplate string
)

func ListKubes(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	common.ManageListRequest(fmt.Sprintf("/cloud/project/%s/kube", projectID), "", cloudprojectKubeColumnsToDisplay, flags.GenericFilters)
}

func GetKube(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/kube", projectID), args[0], cloudKubeTemplate)
}

func EditKube(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s", projectID, url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/cloud/project/{serviceName}/kube/{kubeId}", endpoint, cloudOpenapiSchema)
}
