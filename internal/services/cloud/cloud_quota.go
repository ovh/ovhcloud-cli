package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
)

var (
	//go:embed templates/cloud_quota.tmpl
	cloudQuotaTemplate string
)

func GetCloudQuota(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}
	url := fmt.Sprintf("/cloud/project/%s/region/%s/quota", projectID, url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(url, &object); err != nil {
		display.ExitError("error fetching quotas for region %s: %s", args[0], err)
		return
	}

	display.OutputObject(object, args[0], cloudQuotaTemplate, &flags.OutputFormatConfig)
}
