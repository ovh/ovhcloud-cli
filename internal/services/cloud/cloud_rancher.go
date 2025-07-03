package cloud

import (
	_ "embed"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	cloudprojectRancherColumnsToDisplay = []string{"id", "currentState.name name", "currentState.region region", "currentState.version version", "resourceStatus"}

	//go:embed templates/cloud_rancher.tmpl
	cloudRancherTemplate string

	RancherTargetSpec struct {
		Name              string                 `json:"name,omitempty"`
		Plan              string                 `json:"plan,omitempty"`
		Version           string                 `json:"version,omitempty"`
		IPRestrictions    []rancherIPRestriction `json:"ipRestrictions,omitempty"`
		CLIIPRestrictions []string               `json:"-"`
	}
)

type rancherIPRestriction struct {
	CIDRBlock   string `json:"cidrBlock"`
	Description string `json:"description"`
}

func ListCloudRanchers(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageListRequestNoExpand(fmt.Sprintf("/v2/publicCloud/project/%s/rancher", projectID), cloudprojectRancherColumnsToDisplay, flags.GenericFilters)
}

func GetRancher(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v2/publicCloud/project/%s/rancher", projectID), args[0], cloudRancherTemplate)
}

func EditRancher(cmd *cobra.Command, args []string) {
	for _, ipRestriction := range RancherTargetSpec.CLIIPRestrictions {
		parts := strings.Split(ipRestriction, ",")
		if len(parts) != 2 {
			display.ExitError("Invalid IP restriction format: %s. Expected format: '<cidrBlock>,<description>'", ipRestriction)
			return
		}
		RancherTargetSpec.IPRestrictions = append(RancherTargetSpec.IPRestrictions, rancherIPRestriction{
			CIDRBlock:   parts[0],
			Description: parts[1],
		})
	}

	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/rancher/{rancherId}",
		fmt.Sprintf("/v2/publicCloud/project/%s/rancher/%s", projectID, url.PathEscape(args[0])),
		map[string]any{
			"targetSpec": RancherTargetSpec,
		},
		cloudV2OpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
