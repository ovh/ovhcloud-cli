package storagenetapp

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	storagenetappColumnsToDisplay = []string{"id", "name", "region", "status"}

	//go:embed templates/storagenetapp.tmpl
	storagenetappTemplate string

	//go:embed api-schemas/storagenetapp.json
	storagenetappOpenapiSchema []byte
)

func ListStorageNetApp(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/storage/netapp", "", storagenetappColumnsToDisplay, flags.GenericFilters)
}

func GetStorageNetApp(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/storage/netapp", args[0], storagenetappTemplate)
}

func EditStorageNetApp(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/storage/netapp/%s", url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/storage/netapp/{serviceName}", endpoint, storagenetappOpenapiSchema)
}
