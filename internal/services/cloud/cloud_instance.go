package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	cloudprojectInstanceColumnsToDisplay = []string{"id", "name", "region", "flavor.name", "status"}

	//go:embed templates/cloud_instance.tmpl
	cloudInstanceTemplate string
)

func ListInstances(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}
	common.ManageListRequest(fmt.Sprintf("/cloud/project/%s/instance", projectID), "id", cloudprojectInstanceColumnsToDisplay, flags.GenericFilters)
}

func GetInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}
	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/instance", projectID), args[0], cloudInstanceTemplate)
}

func StartInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/start", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.ExitError("error starting instance %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Instance starting ...")
}

func StopInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/stop", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.ExitError("error stopping instance %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Instance stopping ...")
}

func ShelveInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/shelve", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.ExitError("error shelving instance %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Instance is being shelved ...")
}

func UnshelveInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/unshelve", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.ExitError("error unshelving instance %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Instance is being unshelved ...")
}

func ResumeInstance(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s/resume", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.ExitError("error resuming instance %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Instance is being resumed ...")
}
