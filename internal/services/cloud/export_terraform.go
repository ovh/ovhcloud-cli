// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
)

// Flags of the "cloud export-terraform" command.
var (
	TerraformExportOutputDir string
	TerraformExportResources []string
)

// exportedResource is a single resource to import into Terraform.
type exportedResource struct {
	Name     string // Terraform-safe local resource name (e.g. "my_network")
	ImportID string // provider import id (e.g. "projectID/networkID")
}

// terraformExporter describes how to export one OVH resource type as Terraform
// import blocks. Adding support for a new resource = adding one entry to the
// terraformExporters registry below.
type terraformExporter struct {
	// Key is the short name used by --resources and in logs (e.g. "network").
	Key string
	// TerraformType is the resource type in the OVHcloud provider
	// (e.g. "ovh_cloud_project_network_private").
	TerraformType string
	// List returns the resources of this type existing in the given project.
	List func(projectID string) ([]exportedResource, error)
}

// terraformExporters is the registry of supported resource types. It is the
// single maintenance surface: each entry maps an OVH resource to its Terraform
// resource type and its provider import-id format.
var terraformExporters = []terraformExporter{
	{
		Key:           "network",
		TerraformType: "ovh_cloud_project_network_private",
		// Import id format (provider): "service_name/network_id".
		List: func(projectID string) ([]exportedResource, error) {
			return listServiceScopedResources(fmt.Sprintf("/v1/cloud/project/%s/network/private", projectID), projectID, "name")
		},
	},
	{
		Key:           "user",
		TerraformType: "ovh_cloud_project_user",
		// Import id format (provider): "service_name/id".
		List: func(projectID string) ([]exportedResource, error) {
			return listServiceScopedResources(fmt.Sprintf("/v1/cloud/project/%s/user", projectID), projectID, "username")
		},
	},
}

// listServiceScopedResources lists resources at endpoint and builds import
// entries for resources whose provider import id is "service_name/id".
// nameField selects the object field used to derive a readable resource name
// (it falls back to the id when absent).
func listServiceScopedResources(endpoint, projectID, nameField string) ([]exportedResource, error) {
	var objects []map[string]any
	if err := httpLib.Client.Get(endpoint, &objects); err != nil {
		return nil, err
	}

	resources := make([]exportedResource, 0, len(objects))
	for _, o := range objects {
		id := stringifyID(o["id"])
		if id == "" {
			continue
		}
		label, _ := o[nameField].(string)
		resources = append(resources, exportedResource{
			Name:     hclName(label, id),
			ImportID: fmt.Sprintf("%s/%s", projectID, id),
		})
	}
	return resources, nil
}

// stringifyID normalises an id that may be a string or a number in the JSON
// (the HTTP client decodes numbers as json.Number).
func stringifyID(raw any) string {
	switch v := raw.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return ""
	}
}

// hclName turns a resource label/id into a valid Terraform identifier
// ([a-zA-Z_][a-zA-Z0-9_-]*), lower-cased.
func hclName(label, id string) string {
	base := label
	if base == "" {
		base = id
	}

	var sb strings.Builder
	for _, r := range strings.ToLower(base) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_', r == '-':
			sb.WriteRune(r)
		default:
			sb.WriteRune('_')
		}
	}

	name := sb.String()
	if name == "" {
		name = "resource"
	}
	if name[0] >= '0' && name[0] <= '9' {
		name = "r_" + name
	}
	return name
}

// selectExporters returns the exporters matching the given --resources filter
// (all of them when the filter is empty).
func selectExporters(filter []string) []terraformExporter {
	if len(filter) == 0 {
		return terraformExporters
	}

	wanted := make(map[string]bool, len(filter))
	for _, f := range filter {
		wanted[strings.ToLower(strings.TrimSpace(f))] = true
	}

	var selected []terraformExporter
	for _, e := range terraformExporters {
		if wanted[e.Key] {
			selected = append(selected, e)
		}
	}
	return selected
}

// uniqueName ensures Terraform resource names do not collide by suffixing
// duplicates (_2, _3, …).
func uniqueName(seen map[string]int, name string) string {
	seen[name]++
	if seen[name] == 1 {
		return name
	}
	return fmt.Sprintf("%s_%d", name, seen[name])
}

// ExportTerraform generates Terraform "import" blocks for the resources of the
// configured Public Cloud project, so an existing project can be adopted as
// Infrastructure-as-Code. Terraform then generates the matching configuration
// via "terraform plan -generate-config-out=...".
func ExportTerraform(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	exporters := selectExporters(TerraformExportResources)
	if len(exporters) == 0 {
		display.OutputError(&flags.OutputFormatConfig, "no supported resource type matches --resources %v", TerraformExportResources)
		return
	}

	var (
		builder strings.Builder
		seen    = map[string]int{}
		total   int
	)

	builder.WriteString("# Generated by 'ovhcloud cloud export-terraform'.\n")
	builder.WriteString("# Generate the matching configuration with:\n")
	builder.WriteString("#   terraform plan -generate-config-out=generated.tf\n\n")

	for _, exporter := range exporters {
		resources, err := exporter.List(projectID)
		if err != nil {
			// Non-fatal: skip this resource type and keep exporting the others.
			log.Printf("skipping %q: %s", exporter.Key, err)
			continue
		}

		for _, resource := range resources {
			name := uniqueName(seen, resource.Name)
			fmt.Fprintf(&builder, "import {\n  to = %s.%s\n  id = %q\n}\n\n", exporter.TerraformType, name, resource.ImportID)
			total++
		}
	}

	if total == 0 {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "🟠 No exportable resource found in project %s", projectID)
		return
	}

	if err := os.MkdirAll(TerraformExportOutputDir, 0o755); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create output directory %s: %s", TerraformExportOutputDir, err)
		return
	}

	outputPath := filepath.Join(TerraformExportOutputDir, "imports.tf")
	if err := os.WriteFile(outputPath, []byte(builder.String()), 0o644); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to write %s: %s", outputPath, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil,
		"✅ Exported %d resource(s) to %s\n\nNext step — let Terraform generate the configuration:\n  terraform plan -generate-config-out=generated.tf",
		total, outputPath)
}
