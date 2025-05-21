package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	cloudprojectRegionColumnsToDisplay = []string{"name", "type", "status"}

	//go:embed templates/cloud_region.tmpl
	cloudRegionTemplate string
)

func ListCloudRegions(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	common.ManageListRequest(fmt.Sprintf("/cloud/project/%s/region", projectID), "", cloudprojectRegionColumnsToDisplay, flags.GenericFilters)
}

func GetCloudRegion(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/region", projectID), args[0], cloudRegionTemplate)
}
