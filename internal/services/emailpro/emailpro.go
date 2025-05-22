package emailpro

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
	emailproColumnsToDisplay = []string{"domain", "displayName", "state", "offer"}

	//go:embed templates/emailpro.tmpl
	emailproTemplate string

	//go:embed api-schemas/emailpro.json
	emailproOpenapiSchema []byte
)

func ListEmailPro(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/email/pro", "", emailproColumnsToDisplay, flags.GenericFilters)
}

func GetEmailPro(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/email/pro", args[0], emailproTemplate)
}

func EditEmailPro(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/email/pro/%s", url.PathEscape(args[0]))
	editor.EditResource(httpLib.Client, "/email/pro/{service}", endpoint, emailproOpenapiSchema)
}
