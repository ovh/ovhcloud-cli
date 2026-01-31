// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package emaildomain

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	emaildomainColumnsToDisplay = []string{"domain", "status", "offer"}

	//go:embed templates/emaildomain.tmpl
	emaildomainTemplate string

	//go:embed templates/redirection.tmpl
	redirectionTemplate string

	//go:embed parameter-samples/redirection-create.json
	RedirectionCreateExample string

	RedirectionSpec struct {
		From      string `json:"from"`
		To        string `json:"to"`
		LocalCopy bool   `json:"localCopy"`
	}
)

func ListEmailDomain(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/email/domain", "", emaildomainColumnsToDisplay, flags.GenericFilters)
}

func GetEmailDomain(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/email/domain", args[0], emaildomainTemplate)
}

func ListRedirections(_ *cobra.Command, args []string) {
	serviceName := args[0]
	path := fmt.Sprintf("/email/domain/%s/redirection", url.PathEscape(serviceName))
	columnsToDisplay := []string{"id", "from", "to", "localCopy"}
	common.ManageListRequest(path, "", columnsToDisplay, flags.GenericFilters)
}

func GetRedirection(_ *cobra.Command, args []string) {
	serviceName := args[0]
	redirectionID := args[1]
	path := fmt.Sprintf("/email/domain/%s/redirection", url.PathEscape(serviceName))
	common.ManageObjectRequest(path, redirectionID, redirectionTemplate)
}

func CreateRedirection(cmd *cobra.Command, args []string) {
	serviceName := args[0]

	redirection, err := common.CreateResource(
		cmd,
		"/email/domain/{serviceName}/redirection",
		fmt.Sprintf("/email/domain/%s/redirection", url.PathEscape(serviceName)),
		RedirectionCreateExample,
		RedirectionSpec,
		assets.EmaildomainOpenapiSchema,
		[]string{"from", "to"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error creating redirection: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, redirection, "✅ Email redirection created successfully (ID: %v)", redirection["id"])
}

func DeleteRedirection(_ *cobra.Command, args []string) {
	serviceName := args[0]
	redirectionID := args[1]
	path := fmt.Sprintf("/email/domain/%s/redirection/%s", url.PathEscape(serviceName), url.PathEscape(redirectionID))

	if err := httpLib.Client.Delete(path, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete redirection: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Email redirection %s deleted successfully", redirectionID)
}
