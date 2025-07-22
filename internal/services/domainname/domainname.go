package domainname

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/assets"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	domainnameColumnsToDisplay = []string{"domain", "state", "whoisOwner", "expirationDate", "renewalDate"}

	//go:embed templates/domainname.tmpl
	domainnameTemplate string

	DomainSpec struct {
		NameServerType    string `json:"nameServerType,omitempty"`
		TranferLockStatus string `json:"transferLockStatus,omitempty"`
	}
)

func ListDomainName(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/domain", "", domainnameColumnsToDisplay, flags.GenericFilters)
}

func GetDomainName(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/domain", args[0], domainnameTemplate)
}

func EditDomainName(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/domain/{serviceName}",
		fmt.Sprintf("/domain/%s", url.PathEscape(args[0])),
		DomainSpec,
		assets.DomainOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}
