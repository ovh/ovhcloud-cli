// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudExportTerraformCommand(cloudCmd *cobra.Command) {
	exportCmd := &cobra.Command{
		Use:   "export-terraform",
		Short: "Export the project's resources as Terraform import blocks",
		Long: `Generate Terraform "import" blocks for the resources of a Public Cloud project,
to adopt Infrastructure-as-Code on an existing project without retyping everything.

The command writes an "imports.tf" file; Terraform then generates the matching
configuration for you:

  ovhcloud cloud export-terraform --cloud-project <id>
  terraform plan -generate-config-out=generated.tf

Only resources supported by the OVHcloud Terraform provider are exported. This
command reads your infrastructure but never modifies it.`,
		Example: `  # Export all supported resources of a project
  ovhcloud cloud export-terraform --cloud-project <project_id> --output-dir ./tf

  # Restrict to some resource types
  ovhcloud cloud export-terraform --cloud-project <project_id> --resources network,user`,
		Run:  cloud.ExportTerraform,
		Args: cobra.NoArgs,
	}

	exportCmd.Flags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")
	exportCmd.Flags().StringVar(&cloud.TerraformExportOutputDir, "output-dir", ".", "Directory where imports.tf is written")
	exportCmd.Flags().StringSliceVar(&cloud.TerraformExportResources, "resources", nil, "Restrict export to these resource types (e.g. network,user)")

	cloudCmd.AddCommand(exportCmd)
}
