// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package emaildomain

import (
	_ "embed"

	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	emaildomainColumnsToDisplay = []string{"domain", "status", "offer"}

	//go:embed templates/emaildomain.tmpl
	emaildomainTemplate string
)

func ListEmailDomain(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/email/domain", "", emaildomainColumnsToDisplay, flags.GenericFilters)
}

func GetEmailDomain(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/email/domain", args[0], emaildomainTemplate)
}
