package sms

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	smsColumnsToDisplay = []string{"name", "status"}

	//go:embed templates/sms.tmpl
	smsTemplate string

	//go:embed api-schemas/sms.json
	smsOpenapiSchema []byte
)

func ListSms(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/sms", "", smsColumnsToDisplay, flags.GenericFilters)
}

func GetSms(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/sms", args[0], smsTemplate)
}

func EditSms(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/sms/%s", url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/sms/{serviceName}", endpoint, smsOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}
