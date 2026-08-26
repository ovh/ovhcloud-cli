// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectSSHKeyColumnsToDisplay = []string{"name", "createdAt", "updatedAt"}

	//go:embed templates/cloud_ssh_key.tmpl
	cloudSSHKeyTemplate string

	//go:embed parameter-samples/ssh-key-create.json
	SSHKeyCreationExample string

	// SSHKeyCreationParameters holds the parameters for creating a new SSH key.
	SSHKeyCreationParameters struct {
		Name      string `json:"name,omitempty"`
		PublicKey string `json:"publicKey,omitempty"`
	}
)

func ListCloudSSHKeys(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(fmt.Sprintf("/v2/publicCloud/project/%s/sshKey", projectID), cloudprojectSSHKeyColumnsToDisplay, flags.GenericFilters)
}

func GetCloudSSHKey(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v2/publicCloud/project/%s/sshKey", projectID), args[0], cloudSSHKeyTemplate)
}

func CreateCloudSSHKey(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/sshKey", projectID)
	_, err = common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/sshKey",
		endpoint,
		SSHKeyCreationExample,
		SSHKeyCreationParameters,
		assets.CloudV2OpenapiSchema,
		[]string{"name", "publicKey"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create SSH key: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ SSH key successfully created")
}

func DeleteCloudSSHKey(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/sshKey/%s", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error deleting SSH key %q: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ SSH key successfully deleted")
}
