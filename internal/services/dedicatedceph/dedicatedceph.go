package dedicatedceph

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	dedicatedcephColumnsToDisplay = []string{"serviceName", "region", "state", "status"}

	//go:embed templates/dedicatedceph.tmpl
	dedicatedcephTemplate string
)

func ListDedicatedCeph(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/dedicated/ceph", "", dedicatedcephColumnsToDisplay, flags.GenericFilters)
}

func GetDedicatedCeph(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/dedicated/ceph", args[0], dedicatedcephTemplate)
}
