package emailmxplan

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
	emailmxplanColumnsToDisplay = []string{"domain", "displayName", "state", "offer"}

	//go:embed templates/emailmxplan.tmpl
	emailmxplanTemplate string

	EmailMXPlanSpec struct {
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

func ListEmailMXPlan(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/email/mxplan", "", emailmxplanColumnsToDisplay, flags.GenericFilters)
}

func GetEmailMXPlan(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/email/mxplan", args[0], emailmxplanTemplate)
}

func EditEmailMXPlan(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/email/mxplan/{service}",
		fmt.Sprintf("/email/mxplan/%s", url.PathEscape(args[0])),
		EmailMXPlanSpec,
		assets.EmailmxplanOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
