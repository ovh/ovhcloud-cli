// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package vmwareclouddirectorbackup

import (
	_ "embed"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	vmwareclouddirectorbackupColumnsToDisplay = []string{"id", "iam.displayName", "currentState.azName", "resourceStatus"}

	//go:embed templates/vmwareclouddirectorbackup.tmpl
	vmwareclouddirectorbackupTemplate string

	VmwareCloudDirectorBackupSpec struct {
		TargetSpec struct {
			Offers    []VmwareCloudDirectorBackupOffer `json:"offers,omitempty"`
			CliOffers []string                         `json:"-"`
		} `json:"targetSpec,omitzero"`
	}
)

type VmwareCloudDirectorBackupOffer struct {
	Name      string `json:"name,omitempty"`
	QuotaInTB int    `json:"quotaInTB,omitempty"`
}

func ListVmwareCloudDirectorBackup(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v2/vmwareCloudDirector/backup", "id", vmwareclouddirectorbackupColumnsToDisplay, flags.GenericFilters)
}

func GetVmwareCloudDirectorBackup(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v2/vmwareCloudDirector/backup", args[0], vmwareclouddirectorbackupTemplate)
}

func EditVmwareCloudDirectorBackup(cmd *cobra.Command, args []string) {
	for _, offer := range VmwareCloudDirectorBackupSpec.TargetSpec.CliOffers {
		offerParts := strings.Split(offer, ":")
		if len(offerParts) != 2 {
			display.OutputError(&flags.OutputFormatConfig, "Invalid offer format: %s. Expected format is '<name>:<quotaInTB>'", offer)
			return
		}

		intQuota, err := strconv.Atoi(offerParts[1])
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "Invalid quota value '%s' for offer '%s'. It should be an integer.", offerParts[1], offerParts[0])
			return
		}

		VmwareCloudDirectorBackupSpec.TargetSpec.Offers = append(VmwareCloudDirectorBackupSpec.TargetSpec.Offers,
			VmwareCloudDirectorBackupOffer{
				Name:      offerParts[0],
				QuotaInTB: intQuota,
			},
		)
	}

	if err := common.EditResource(
		cmd,
		"/vmwareCloudDirector/backup/{backupId}",
		fmt.Sprintf("/v2/vmwareCloudDirector/backup/%s", url.PathEscape(args[0])),
		VmwareCloudDirectorBackupSpec,
		assets.VmwareclouddirectorbackupOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
