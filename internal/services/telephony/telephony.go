// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package telephony

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
	telephonyColumnsToDisplay = []string{"billingAccount", "description", "status"}

	//go:embed templates/telephony.tmpl
	telephonyTemplate string

	TelephonySpec struct {
		CreditThreshold struct {
			CurrencyCode string `json:"currencyCode,omitempty"`
			Text         string `json:"text,omitempty"`
			Value        int    `json:"value,omitempty"`
		}
		Description             string `json:"description,omitempty"`
		HiddenExternalNumber    bool   `json:"hiddenExternalNumber,omitempty"`
		OverrideDisplayedNumber bool   `json:"overrideDisplayedNumber,omitempty"`
	}
)

func ListTelephony(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/telephony", "", telephonyColumnsToDisplay, flags.GenericFilters)
}

func GetTelephony(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/telephony", args[0], telephonyTemplate)
}

func EditTelephony(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/telephony/{billingAccount}",
		fmt.Sprintf("/v1/telephony/%s", url.PathEscape(args[0])),
		TelephonySpec,
		assets.TelephonyOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
