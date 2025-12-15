// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package sms

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
	smsColumnsToDisplay = []string{"name", "status"}

	//go:embed templates/sms.tmpl
	smsTemplate string

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
			CustomizedEmailMode bool   `json:"customizedEmailMode,omitempty"`
			CustomizedSmsMode   bool   `json:"customizedSmsMode,omitempty"`
			EmailBody           string `json:"emailBody,omitempty"`
			EmailFrom           string `json:"emailFrom,omitempty"`
			EmailSubject        string `json:"emailSubject,omitempty"`
			SmsBody             string `json:"smsBody,omitempty"`
		} `json:"templates,omitzero"`
	}
)

func ListSms(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/sms", "", smsColumnsToDisplay, flags.GenericFilters)
}

func GetSms(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/sms", args[0], smsTemplate)
}

func EditSms(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/sms/{serviceName}",
		fmt.Sprintf("/v1/sms/%s", url.PathEscape(args[0])),
		SmsSpec,
		assets.SmsOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
