package editor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"

	"github.com/ovh/go-ovh/ovh"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/openapi"
)

const defaultEditor = "vi"

func EditValueWithEditor(value []byte) ([]byte, error) {
	editor := defaultEditor
	if s := os.Getenv("EDITOR"); s != "" {
		editor = s
	}

	// Create temp file
	f, err := os.CreateTemp("", "ovh-cli-edit")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(value); err != nil {
		return nil, fmt.Errorf("failed to write input file: %w", err)
	}

	// Open editor
	cmd := exec.Command("sh", "-c", editor+" "+f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to edit file: %w", err)
	}

	// Read updated file
	b, err := os.ReadFile(f.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read updated file: %w", err)
	}

	return b, nil
}

func EditResource(client *ovh.Client, path, url string, openapiSpec []byte) {
	// Fetch resource
	var object map[string]any
	if err := client.Get(url, &object); err != nil {
		display.ExitError("error fetching resource %s: %s\n", url, err)
	}

	// Filter editable fields
	editableBody, err := openapi.FilterEditableFields(
		openapiSpec,
		path,
		"put",
		object,
	)
	if err != nil {
		display.ExitError("failed to extract writable properties: %s", err)
	}

	// Format editable body
	editableOutput, err := json.MarshalIndent(editableBody, "", "  ")
	if err != nil {
		display.ExitError("failed to marshal writable body: %s", err)
	}

	// Edit value
	updatedBody, err := EditValueWithEditor(editableOutput)
	if err != nil {
		display.ExitError("failed to edit properties: %s", err)
	}

	if slices.Equal(updatedBody, editableOutput) {
		fmt.Println("\n🟠 Resource not updated, exiting")
		return
	}

	// Update API call
	if err := client.Put(url, json.RawMessage(updatedBody), nil); err != nil {
		display.ExitError("failed to update resource: %s", err)
	}

	fmt.Println("\n✅ Resource updated succesfully ...")
}
