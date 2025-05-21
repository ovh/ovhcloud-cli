package dedicatedcluster

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	dedicatedclusterColumnsToDisplay = []string{"id", "region", "model", "status"}

	//go:embed templates/dedicatedcluster.tmpl
	dedicatedclusterTemplate string
)

func ListDedicatedCluster(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/dedicated/cluster", "", dedicatedclusterColumnsToDisplay, flags.GenericFilters)
}

func GetDedicatedCluster(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/dedicated/cluster", args[0], dedicatedclusterTemplate)
}
