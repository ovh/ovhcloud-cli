package telephony

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	telephonyColumnsToDisplay = []string{"billingAccount", "description", "status"}

	//go:embed templates/telephony.tmpl
	telephonyTemplate string

	//go:embed api-schemas/telephony.json
	telephonyOpenapiSchema []byte

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
	common.ManageListRequest("/telephony", "", telephonyColumnsToDisplay, flags.GenericFilters)
}

func GetTelephony(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/telephony", args[0], telephonyTemplate)
}

func EditTelephony(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/telephony/{billingAccount}",
		fmt.Sprintf("/telephony/%s", url.PathEscape(args[0])),
		TelephonySpec,
		telephonyOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
