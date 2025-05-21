package domainname

import (
	_ "embed"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	domainnameColumnsToDisplay = []string{"domain", "state", "whoisOwner", "expirationDate", "renewalDate"}

	//go:embed templates/domainname.tmpl
	domainnameTemplate string
)

func ListDomainName(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/domain", "", domainnameColumnsToDisplay, flags.GenericFilters)
}

func GetDomainName(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/domain", args[0], domainnameTemplate)
}
