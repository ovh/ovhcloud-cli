package editor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"

	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
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

func EditResource(client *ovh.Client, path, url string, openapiSpec []byte) error {
	// Fetch resource
	var object map[string]any
	if err := client.Get(url, &object); err != nil {
		return fmt.Errorf("error fetching resource %s: %w", url, err)
	}

	// Filter editable fields
	editableBody, err := openapi.FilterEditableFields(
		openapiSpec,
		path,
		"put",
		object,
	)
	if err != nil {
		return fmt.Errorf("failed to extract writable properties: %w", err)
	}

	// Format editable body
	editableOutput, err := json.MarshalIndent(editableBody, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal writable body: %w", err)
	}

	// Edit value
	updatedBody, err := EditValueWithEditor(editableOutput)
	if err != nil {
		return fmt.Errorf("failed to edit properties: %w", err)
	}

	if slices.Equal(updatedBody, editableOutput) {
		fmt.Println("\n🟠 Resource not updated, exiting")
		return nil
	}

	// Update API call
	if err := client.Put(url, json.RawMessage(updatedBody), nil); err != nil {
		return fmt.Errorf("failed to update resource: %w", err)
	}

	fmt.Println("✅ Resource updated succesfully ...")

	return nil
}
