package storagenetapp

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
	storagenetappColumnsToDisplay = []string{"id", "name", "region", "status"}

	//go:embed templates/storagenetapp.tmpl
	storagenetappTemplate string

	StorageNetAppSpec struct {
		Name string `json:"name,omitempty"`
	}
)

func ListStorageNetApp(_ *cobra.Command, _ []string) {
	common.ManageListRequestNoExpand("/storage/netapp", storagenetappColumnsToDisplay, flags.GenericFilters)
}

func GetStorageNetApp(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/storage/netapp", args[0], storagenetappTemplate)
}

func EditStorageNetApp(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/storage/netapp/{serviceName}",
		fmt.Sprintf("/storage/netapp/%s", url.PathEscape(args[0])),
		StorageNetAppSpec,
		assets.SmsOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
