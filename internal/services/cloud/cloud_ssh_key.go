// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectSSHKeyColumnsToDisplay = []string{"id", "name", "regions"}

	//go:embed templates/cloud_ssh_key.tmpl
	cloudSSHKeyTemplate string

	//go:embed parameter-samples/ssh-key-create.json
	SSHKeyCreationExample string

	// sshKeyCreationParameters holds the parameters for creating a new SSH key.
	SSHKeyCreationParameters struct {
		Name      string `json:"name,omitempty"`
		PublicKey string `json:"publicKey,omitempty"`
		Region    string `json:"region,omitempty"`
	}
)

func ListCloudSSHKeys(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	path := fmt.Sprintf("/v1/cloud/project/%s/sshkey", projectID)

	var body []map[string]any
	if err := httpLib.Client.Get(path, &body); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch SSH keys: %s", err)
		return
	}

	body, err = filtersLib.FilterLines(body, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(body, cloudprojectSSHKeyColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetCloudSSHKey(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v1/cloud/project/%s/sshkey", projectID), args[0], cloudSSHKeyTemplate)
}

func CreateCloudSSHKey(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/sshkey", projectID)
	_, err = common.CreateResource(
		cmd,
		"/v1/cloud/project/{serviceName}/sshkey",
		endpoint,
		SSHKeyCreationExample,
		SSHKeyCreationParameters,
		assets.CloudOpenapiSchema,
		[]string{"name", "publicKey"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create SSH key: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ SSH key successfully created")
}
