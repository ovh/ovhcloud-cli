package sms

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
	smsColumnsToDisplay = []string{"name", "status"}

	//go:embed templates/sms.tmpl
	smsTemplate string

	//go:embed api-schemas/sms.json
	smsOpenapiSchema []byte

	SmsSpec struct {
		AutomaticRecreditAmount             int    `json:"automaticRecreditAmount,omitempty"`
		CallBack                            string `json:"callBack,omitempty"`
		CreditThresholdForAutomaticRecredit int    `json:"creditThresholdForAutomaticRecredit,omitempty"`
		Description                         string `json:"description,omitempty"`
		SmsResponse                         struct {
			CgiUrl                   string `json:"cgiUrl,omitempty"`
			ResponseType             string `json:"responseType,omitempty"`
			Text                     string `json:"text,omitempty"`
			TrackingDefaultSmsSender string `json:"trackingDefaultSmsSender,omitempty"`
		} `json:"smsResponse,omitzero"`
		StopCallBack string `json:"stopCallBack,omitempty"`
		Templates    struct {
			CustomizedEmailMode        bool   `json:"customizedEmailMode,omitempty"`
			CustomizedSmsMode          bool   `json:"customizedSmsMode,omitempty"`
			EmailBody                  string `json:"emailBody,omitempty"`
			EmailFrom                  string `json:"emailFrom,omitempty"`
			EmailSubject               string `json:"emailSubject,omitempty"`
			SmsBody                    string `json:"smsBody,omitempty"`
			Time2chatAutomaticResponse string `json:"time2chatAutomaticResponse,omitempty"`
		} `json:"templates,omitzero"`
	}
)

func ListSms(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/sms", "", smsColumnsToDisplay, flags.GenericFilters)
}

func GetSms(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/sms", args[0], smsTemplate)
}

func EditSms(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/sms/{serviceName}",
		fmt.Sprintf("/sms/%s", url.PathEscape(args[0])),
		SmsSpec,
		smsOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
