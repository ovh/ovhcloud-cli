// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/editor"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
	"github.com/ovh/ovhcloud-cli/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// nonContentEditFlags lists the flags that do not describe a change to apply:
// resource/project selectors, global flags, and --wait. They must not, on their
// own, trigger an update. Content-providing flags like --editor and --from-file
// are deliberately absent so that they DO count as content.
var nonContentEditFlags = map[string]bool{
	"cloud-project": true,
	"wait":          true,
	"output":        true,
	"debug":         true,
	"ignore-errors": true,
	"profile":       true,
	"help":          true,
}

// editHasContentFlags reports whether the user provided at least one flag that
// actually carries edit content, ignoring selector/global flags and --wait.
func editHasContentFlags(cmd *cobra.Command) bool {
	found := false
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if !nonContentEditFlags[f.Name] {
			found = true
		}
	})
	return found
}

var (
	//go:embed templates/service_info.tmpl
	ServiceInfoTemplate string

	ServiceInfoSpec struct {
		Renew struct {
			Automatic          bool `json:"automatic"`
			DeleteAtExpiration bool `json:"deleteAtExpiration"`
			Forced             bool `json:"forced"`
			ManualPayment      bool `json:"manualPayment"`
			Period             int  `json:"period"`
		} `json:"renew"`
	}
)

func ManageListRequest(path, idField string, columnsToDisplay, filters []string) {
	body, err := httpLib.FetchExpandedArray(path, idField)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch results: %s", err)
		return
	}

	body, err = filtersLib.FilterLines(body, filters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(body, columnsToDisplay, &flags.OutputFormatConfig)
}

func ManageListRequestNoExpand(path string, columnsToDisplay, filters []string) {
	body, err := httpLib.FetchArray(path, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch results: %s", err)
		return
	}

	var objects []map[string]any
	for _, object := range body {
		objects = append(objects, object.(map[string]any))
	}

	objects, err = filtersLib.FilterLines(objects, filters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(objects, columnsToDisplay, &flags.OutputFormatConfig)
}

func ManageObjectRequest(path, objectID, templateContent string) {
	url := fmt.Sprintf("%s/%s", path, url.PathEscape(objectID))

	var object map[string]any
	if err := httpLib.Client.Get(url, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching %s: %s", url, err)
		return
	}

	display.OutputObject(object, objectID, templateContent, &flags.OutputFormatConfig)
}

func CreateResource(cmd *cobra.Command, path, endpoint, defaultExample string,
	cliParams any, openapiSpec []byte, mandatoryFields []string) (map[string]any, error) {
	// Create object from parameters given on command line
	jsonCliParameters, err := json.Marshal(cliParams)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare arguments from command line: %w", err)
	}
	var cliParameters map[string]any
	if err := json.Unmarshal(jsonCliParameters, &cliParameters); err != nil {
		return nil, fmt.Errorf("failed to parse arguments from command line: %w", err)
	}

	parameters := make(map[string]any)

	switch {
	case utils.IsInputFromPipe(): // Data given through a pipe
		var stdin []byte
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			stdin = append(stdin, scanner.Bytes()...)
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(stdin, &parameters); err != nil {
			return nil, fmt.Errorf("failed to parse given data: %w", err)
		}

	case flags.ParametersViaEditor: // Data given through an editor
		log.Print("Flag --editor used, all other flags will override the example values")

		examples, err := openapi.GetOperationRequestExamples(openapiSpec, path, "post", defaultExample, cliParameters)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch API call examples: %w", err)
		}

		_, choice, err := display.RunGenericChoicePicker("Please select a creation example", examples, 0)
		if err != nil {
			return nil, err
		}

		if choice == "" {
			return nil, errors.New("no example selected, exiting…")
		}

		newValue, err := editor.EditValueWithEditor([]byte(choice))
		if err != nil {
			return nil, fmt.Errorf("failed to edit parameters using editor: %w", err)
		}

		if err := json.Unmarshal(newValue, &parameters); err != nil {
			return nil, fmt.Errorf("failed to parse given parameters: %w", err)
		}

	case flags.ParametersFile != "": // Data given in a file
		log.Print("Flag --from-file used, all other flags will override the file values")

		fd, err := os.Open(flags.ParametersFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open given file: %w", err)
		}
		defer fd.Close()

		if err := json.NewDecoder(fd).Decode(&parameters); err != nil {
			return nil, fmt.Errorf("failed to parse given file: %w", err)
		}
	}

	// Only merge CLI parameters with other ones if not in --editor mode.
	// In this case, the CLI parameters have already been merged with the
	// request examples coming from API schemas.
	if !flags.ParametersViaEditor {
		if err := utils.MergeMaps(parameters, cliParameters); err != nil {
			return nil, fmt.Errorf("failed to merge replace values into example: %w", err)
		}
	}

	// Check if mandatory fields are present
	for _, field := range mandatoryFields {
		if _, ok := parameters[field]; !ok {
			return nil, fmt.Errorf("mandatory field %q is missing in the parameters\n\n%s", field, cmd.UsageString())
		}
	}

	out, err := json.MarshalIndent(parameters, "", " ")
	if err != nil {
		return nil, fmt.Errorf("parameters cannot be marshalled: %w", err)
	}

	log.Println("Final parameters: \n" + string(out))

	var createdResource map[string]any
	if err := httpLib.Client.Post(endpoint, parameters, &createdResource); err != nil {
		return nil, fmt.Errorf("error creating resource: %w", err)
	}

	return createdResource, nil
}

// prepareEditBody resolves the request body for an edit (PUT) operation from the
// CLI parameters, an optional --from-file and an optional --editor. It performs
// no API call and prints no success message. The returned boolean is false when
// there is nothing to edit (an informational message has then already been
// printed).
func prepareEditBody(cmd *cobra.Command, path, url string, cliParams any, openapiSpec []byte) (any, bool, error) {
	if !editHasContentFlags(cmd) {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No parameters given, nothing to edit")
		return nil, false, nil
	}

	// Create object from parameters given on command line
	jsonCliParameters, err := json.Marshal(cliParams)
	if err != nil {
		return nil, false, fmt.Errorf("failed to prepare arguments from command line: %w", err)
	}
	var cliParameters map[string]any
	if err := json.Unmarshal(jsonCliParameters, &cliParameters); err != nil {
		return nil, false, fmt.Errorf("failed to parse arguments from command line: %w", err)
	}

	// Handle data from file if --from-file flag is used
	if flags.ParametersFile != "" {
		log.Print("Flag --from-file used, all other flags will override the file values")

		fd, err := os.Open(flags.ParametersFile)
		if err != nil {
			return nil, false, fmt.Errorf("failed to open given file: %w", err)
		}
		defer fd.Close()

		var fileParameters map[string]any
		if err := json.NewDecoder(fd).Decode(&fileParameters); err != nil {
			return nil, false, fmt.Errorf("failed to parse given file: %w", err)
		}

		// Merge CLI parameters with file parameters (CLI takes precedence)
		if err := utils.MergeMaps(fileParameters, cliParameters); err != nil {
			return nil, false, fmt.Errorf("failed to merge CLI parameters with file parameters: %w", err)
		}

		cliParameters = fileParameters
	}

	// Fetch resource
	var object map[string]any
	if err := httpLib.Client.Get(url, &object); err != nil {
		return nil, false, fmt.Errorf("error fetching resource %s: %w", url, err)
	}

	// Merge CLI parameters with the fetched object
	if err := utils.MergeMaps(object, cliParameters); err != nil {
		return nil, false, fmt.Errorf("failed to merge CLI parameters into example: %w", err)
	}

	// Filter editable fields from OpenAPI spec
	editableBody, err := openapi.FilterEditableFields(
		openapiSpec,
		path,
		"put",
		object,
	)
	if err != nil {
		return nil, false, fmt.Errorf("failed to extract writable properties: %w", err)
	}

	// If editor not needed, use the filtered body directly
	if !flags.ParametersViaEditor {
		return editableBody, true, nil
	}

	// Format editable body
	editableOutput, err := json.MarshalIndent(editableBody, "", "  ")
	if err != nil {
		return nil, false, fmt.Errorf("failed to marshal writable body: %w", err)
	}

	// Edit value
	updatedBody, err := editor.EditValueWithEditor(editableOutput)
	if err != nil {
		return nil, false, fmt.Errorf("failed to edit properties: %w", err)
	}

	return json.RawMessage(updatedBody), true, nil
}

func EditResource(cmd *cobra.Command, path, url string, cliParams any, openapiSpec []byte) error {
	body, ok, err := prepareEditBody(cmd, path, url, cliParams, openapiSpec)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	if err := httpLib.Client.Put(url, body, nil); err != nil {
		return fmt.Errorf("failed to update resource: %w", err)
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Resource updated successfully")

	return nil
}

// EditResourceAndReturn behaves like EditResource but returns the resource as
// sent back by the API and prints no message, letting the caller craft its own
// output (for example to wait for an asynchronous resource to become ready).
// The returned boolean is false when there was nothing to edit.
func EditResourceAndReturn(cmd *cobra.Command, path, url string, cliParams any, openapiSpec []byte) (map[string]any, bool, error) {
	body, ok, err := prepareEditBody(cmd, path, url, cliParams, openapiSpec)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}

	var updated map[string]any
	if err := httpLib.Client.Put(url, body, &updated); err != nil {
		return nil, false, fmt.Errorf("failed to update resource: %w", err)
	}

	return updated, true, nil
}
