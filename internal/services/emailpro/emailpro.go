// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package emailpro

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	emailproColumnsToDisplay = []string{"domain", "displayName", "state", "offer"}

	//go:embed templates/emailpro.tmpl
	emailproTemplate string

	EmailProSpec struct {
		ComplexityEnabled         bool   `json:"complexityEnabled,omitempty"`
		DisplayName               string `json:"displayName,omitempty"`
		LockoutDuration           int    `json:"lockoutDuration,omitempty"`
		LockoutObservationWindow  int    `json:"lockoutObservationWindow,omitempty"`
		LockoutThreshold          int    `json:"lockoutThreshold,omitempty"`
		MaxPasswordAge            int    `json:"maxPasswordAge,omitempty"`
		MaxReceiveSize            int    `json:"maxReceiveSize,omitempty"`
		MaxSendSize               int    `json:"maxSendSize,omitempty"`
		MinPasswordAge            int    `json:"minPasswordAge,omitempty"`
		MinPasswordLength         int    `json:"minPasswordLength,omitempty"`
		SpamAndVirusConfiguration struct {
			CheckDKIM   bool `json:"checkDKIM,omitempty"`
			CheckSPF    bool `json:"checkSPF,omitempty"`
			DeleteSpam  bool `json:"deleteSpam,omitempty"`
			DeleteVirus bool `json:"deleteVirus,omitempty"`
			PutInJunk   bool `json:"putInJunk,omitempty"`
			TagSpam     bool `json:"tagSpam,omitempty"`
			TagVirus    bool `json:"tagVirus,omitempty"`
		} `json:"spamAndVirusConfiguration,omitzero"`
	}
)

func ListEmailPro(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/email/pro", "", emailproColumnsToDisplay, flags.GenericFilters)
}

func GetEmailPro(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/email/pro", args[0], emailproTemplate)
}

func EditEmailPro(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/email/pro/{service}",
		fmt.Sprintf("/v1/email/pro/%s", url.PathEscape(args[0])),
		EmailProSpec,
		assets.EmailproOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
