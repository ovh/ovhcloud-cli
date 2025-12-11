// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/editor"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/ovh/ovhcloud-cli/internal/utils"
	"github.com/spf13/cobra"
)

var (
	cloudprojectUserColumnsToDisplay = []string{"id", "username", "description", "status"}

	//go:embed templates/cloud_user.tmpl
	cloudUserTemplate string

	//go:embed parameter-samples/user-create.json
	UserCreateExample string

	//go:embed parameter-samples/storage-s3-policy.json
	CloudStorageS3ContainerPolicyExample string

	UserSpec struct {
		Description string   `json:"description,omitempty"`
		Roles       []string `json:"roles,omitempty"`
	}

	StorageS3ContainerPolicySpec struct {
		Policy string `json:"policy,omitempty"`
	}
)

func ListCloudUsers(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	path := fmt.Sprintf("/v1/cloud/project/%s/user", projectID)

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

	display.RenderTable(body, cloudprojectUserColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetCloudUser(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v1/cloud/project/%s/user", projectID), args[0], cloudUserTemplate)
}

func CreateCloudUser(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	client, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/user",
		fmt.Sprintf("/v1/cloud/project/%s/user", projectID),
		UserCreateExample,
		UserSpec,
		assets.CloudOpenapiSchema,
		nil,
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create user: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, client, "✅ User '%s' created successfully", client["id"])
}

func DeleteCloudUser(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/user/%s", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete user: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ User '%s' deleted successfully", args[0])
}

func CreateUserS3Policy(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	parameters := make(map[string]any)

	jsonCliParameters, err := json.Marshal(StorageS3ContainerPolicySpec)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to prepare arguments from command line: %s", err)
		return
	}
	if err := json.Unmarshal(jsonCliParameters, &parameters); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to parse arguments from command line: %s", err)
		return
	}

	switch {
	case utils.IsInputFromPipe(): // Data given through a pipe
		var stdin []byte
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			stdin = append(stdin, scanner.Bytes()...)
		}
		if err := scanner.Err(); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to read from stdin: %s", err)
			return
		}

		parameters["policy"] = string(stdin)

	case flags.ParametersViaEditor: // Data given through an editor
		log.Print("Flag --editor used, all other flags will override the example values")

		examples, err := openapi.GetOperationRequestExamples(
			assets.CloudOpenapiSchema,
			"/cloud/project/{serviceName}/user/{userId}/policy",
			"post",
			CloudStorageS3ContainerPolicyExample,
			map[string]any{},
		)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to fetch API call examples: %s", err)
			return
		}

		_, choice, err := display.RunGenericChoicePicker("Please select a creation example", examples, 0)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to run choice picker: %s", err)
			return
		}

		if choice == "" {
			display.OutputError(&flags.OutputFormatConfig, "no example selected, exiting…")
			return
		}

		newValue, err := editor.EditValueWithEditor([]byte(choice))
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to edit parameters using editor: %s", err)
			return
		}

		parameters["policy"] = string(newValue)

	case flags.ParametersFile != "": // Data given in a file
		log.Print("Flag --from-file used, all other flags will override the file values")

		fileContent, err := os.ReadFile(flags.ParametersFile)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to open given file: %s", err)
			return
		}
		parameters["policy"] = string(fileContent)
	}

	if policy, ok := parameters["policy"]; !ok || policy == "" {
		display.OutputError(&flags.OutputFormatConfig, "A policy must be provided\n\n%s", cmd.UsageString())
		return
	}

	out, err := json.MarshalIndent(parameters, "", " ")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "parameters cannot be marshalled: %s", err)
		return
	}

	log.Println("Final parameters: \n" + string(out))

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/user/%s/policy", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Post(endpoint, parameters, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error creating resource: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Policy created successfully for user %s", args[0])
}

func GetUserS3Policy(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v1/cloud/project/%s/user/%s/policy", projectID, args[0]), "", "")
}
