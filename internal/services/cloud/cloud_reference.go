package cloud

import (
	"fmt"
	"net/url"

	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

func GetFlavors(region string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/flavor", projectID)
	if region != "" {
		endpoint += "?region=" + url.QueryEscape(region)
	}

	common.ManageListRequestNoExpand(endpoint, []string{"id", "name", "region", "osType", "available"}, flags.GenericFilters)
}

func GetImages(region, osType string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	query := url.Values{}
	if region != "" {
		query.Add("region", region)
	}
	if osType != "" {
		query.Add("osType", osType)
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/image?%s", projectID, query.Encode())

	common.ManageListRequestNoExpand(endpoint, []string{"id", "name", "region", "type", "status"}, flags.GenericFilters)
}
