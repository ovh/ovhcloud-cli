package common

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	filtersLib "stash.ovh.net/api/ovh-cli/internal/filters"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/openapi"
	"stash.ovh.net/api/ovh-cli/internal/utils"
)

func ManageListRequest(path, idField string, columnsToDisplay, filters []string) {
	body, err := httpLib.FetchExpandedArray(path, idField)
	if err != nil {
		display.ExitError("failed to fetch results: %s", err)
		return
	}

	body, err = filtersLib.FilterLines(body, filters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
		return
	}

	display.RenderTable(body, columnsToDisplay, &flags.OutputFormatConfig)
}

func ManageListRequestNoExpand(path string, columnsToDisplay, filters []string) {
	body, err := httpLib.FetchArray(path, "")
	if err != nil {
		display.ExitError("failed to fetch results: %s", err)
		return
	}

	var objects []map[string]any
	for _, object := range body {
		objects = append(objects, object.(map[string]any))
	}

	objects, err = filtersLib.FilterLines(objects, filters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
		return
	}

	display.RenderTable(objects, columnsToDisplay, &flags.OutputFormatConfig)
}

func ManageObjectRequest(path, objectID, templateContent string) {
	url := fmt.Sprintf("%s/%s", path, url.PathEscape(objectID))

	var object map[string]any
	if err := httpLib.Client.Get(url, &object); err != nil {
		display.ExitError("error fetching %s: %s", url, err)
		return
	}

	display.OutputObject(object, objectID, templateContent, &flags.OutputFormatConfig)
}

func CreateResource(path, endpoint, defaultExample string, cliParams any, openapiSpec []byte, mandatoryFields []string) (map[string]any, error) {
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
			return nil, errors.New("No example selected, exiting...")
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
			return nil, fmt.Errorf("mandatory field %q is missing in the parameters", field)
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

func EditResource(cmd *cobra.Command, path, url string, cliParams any, openapiSpec []byte) error {
	if cmd.Flags().NFlag() == 0 {
		fmt.Println("🟠 No parameters given, nothing to edit")
		return nil
	}

	// Create object from parameters given on command line
	jsonCliParameters, err := json.Marshal(cliParams)
	if err != nil {
		return fmt.Errorf("failed to prepare arguments from command line: %w", err)
	}
	var cliParameters map[string]any
	if err := json.Unmarshal(jsonCliParameters, &cliParameters); err != nil {
		return fmt.Errorf("failed to parse arguments from command line: %w", err)
	}

	// Fetch resource
	var object map[string]any
	if err := httpLib.Client.Get(url, &object); err != nil {
		return fmt.Errorf("error fetching resource %s: %w", url, err)
	}

	// Merge CLI parameters with the fetched object
	if err := utils.MergeMaps(object, cliParameters); err != nil {
		return fmt.Errorf("failed to merge CLI parameters into example: %w", err)
	}

	// Filter editable fields from OpenAPI spec
	editableBody, err := openapi.FilterEditableFields(
		openapiSpec,
		path,
		"put",
		object,
	)
	if err != nil {
		return fmt.Errorf("failed to extract writable properties: %w", err)
	}

	// If editor not needed, update the resource directly
	if !flags.ParametersViaEditor {
		if err := httpLib.Client.Put(url, editableBody, nil); err != nil {
			return fmt.Errorf("failed to update resource: %w", err)
		}

		fmt.Println("\n✅ Resource updated succesfully ...")

		return nil
	}

	// Format editable body
	editableOutput, err := json.MarshalIndent(editableBody, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal writable body: %w", err)
	}

	// Edit value
	updatedBody, err := editor.EditValueWithEditor(editableOutput)
	if err != nil {
		return fmt.Errorf("failed to edit properties: %w", err)
	}

	// Update API call
	if err := httpLib.Client.Put(url, json.RawMessage(updatedBody), nil); err != nil {
		return fmt.Errorf("failed to update resource: %w", err)
	}

	fmt.Println("\n✅ Resource updated succesfully ...")

	return nil
}
