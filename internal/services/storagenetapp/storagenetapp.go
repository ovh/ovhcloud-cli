package storagenetapp

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	storagenetappColumnsToDisplay = []string{"id", "name", "region", "status"}

	//go:embed templates/storagenetapp.tmpl
	storagenetappTemplate string
)

func ListStorageNetApp(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/storage/netapp", "", storagenetappColumnsToDisplay, flags.GenericFilters)
}

func GetStorageNetApp(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/storage/netapp", args[0], storagenetappTemplate)
}
