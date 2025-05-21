package sms

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	smsColumnsToDisplay = []string{"name", "status"}

	//go:embed templates/sms.tmpl
	smsTemplate string
)

func ListSms(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/sms", "", smsColumnsToDisplay, flags.GenericFilters)
}

func GetSms(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/sms", args[0], smsTemplate)
}
