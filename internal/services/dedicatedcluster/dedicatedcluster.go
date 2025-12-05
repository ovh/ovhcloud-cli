// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package dedicatedcluster

import (
	_ "embed"

	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	dedicatedclusterColumnsToDisplay = []string{"id", "region", "model", "status"}

	//go:embed templates/dedicatedcluster.tmpl
	dedicatedclusterTemplate string
)

func ListDedicatedCluster(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/dedicated/cluster", "", dedicatedclusterColumnsToDisplay, flags.GenericFilters)
}

func GetDedicatedCluster(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/dedicated/cluster", args[0], dedicatedclusterTemplate)
}
