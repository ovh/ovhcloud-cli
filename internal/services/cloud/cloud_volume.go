package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectVolumeColumnsToDisplay = []string{"id", "name", "region", "type", "status"}

	//go:embed templates/cloud_volume.tmpl
	cloudVolumeTemplate string

	CloudVolume struct {
		AttachedTo       []string `json:"attachedTo,omitempty"`
		AvailabilityZone string   `json:"availabilityZone,omitempty"`
		Bootable         bool     `json:"bootable,omitempty"`
		CreationDate     string   `json:"creationDate,omitempty"`
		Description      string   `json:"description,omitempty"`
		Name             string   `json:"name,omitempty"`
		PlanCode         string   `json:"planCode,omitempty"`
		Region           string   `json:"region,omitempty"`
		Size             int      `json:"size,omitempty"`
		Status           string   `json:"status,omitempty"`
		Type             string   `json:"type,omitempty"`
	}
)

func ListCloudVolumes(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageListRequestNoExpand(fmt.Sprintf("/cloud/project/%s/volume", projectID), cloudprojectVolumeColumnsToDisplay, flags.GenericFilters)
}

func GetVolume(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/volume", projectID), args[0], cloudVolumeTemplate)
}

func EditVolume(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/volume/{volumeId}",
		fmt.Sprintf("/cloud/project/%s/volume/%s", projectID, url.PathEscape(args[0])),
		CloudVolume,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
