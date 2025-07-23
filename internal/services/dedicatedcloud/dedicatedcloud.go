package dedicatedcloud

import (
	_ "embed"

	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	dedicatedcloudColumnsToDisplay = []string{"serviceName", "location", "state", "description"}

	//go:embed templates/dedicatedcloud.tmpl
	dedicatedcloudTemplate string
)

func ListDedicatedCloud(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/dedicatedCloud", "", dedicatedcloudColumnsToDisplay, flags.GenericFilters)
}

func GetDedicatedCloud(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/dedicatedCloud", args[0], dedicatedcloudTemplate)
}
