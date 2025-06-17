package domainname

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	domainnameColumnsToDisplay = []string{"domain", "state", "whoisOwner", "expirationDate", "renewalDate"}

	//go:embed templates/domainname.tmpl
	domainnameTemplate string

	//go:embed api-schemas/domain.json
	domainOpenapiSchema []byte
)

func ListDomainName(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/domain", "", domainnameColumnsToDisplay, flags.GenericFilters)
}

func GetDomainName(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/domain", args[0], domainnameTemplate)
}

func EditDomainName(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/domain/%s", url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/domain/{serviceName}", endpoint, domainOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}
