package emailpro

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
	emailproColumnsToDisplay = []string{"domain", "displayName", "state", "offer"}

	//go:embed templates/emailpro.tmpl
	emailproTemplate string

	//go:embed api-schemas/emailpro.json
	emailproOpenapiSchema []byte

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
	common.ManageListRequest("/email/pro", "", emailproColumnsToDisplay, flags.GenericFilters)
}

func GetEmailPro(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/email/pro", args[0], emailproTemplate)
}

func EditEmailPro(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/email/pro/{service}",
		fmt.Sprintf("/email/pro/%s", url.PathEscape(args[0])),
		EmailProSpec,
		emailproOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
