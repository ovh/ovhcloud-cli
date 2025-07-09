package ip

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	ipColumnsToDisplay = []string{"ip", "rir", "routedTo.serviceName", "country", "description"}

	//go:embed templates/ip.tmpl
	ipTemplate string

	//go:embed api-schemas/ip.json
	ipOpenapiSchema []byte

	IPSpec struct {
		Description string `json:"description,omitempty"`
	}
)

func ListIp(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/ip", "", ipColumnsToDisplay, flags.GenericFilters)
}

func GetIp(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/ip", args[0], ipTemplate)
}

func EditIp(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/ip/{ip}",
		fmt.Sprintf("/ip/%s", url.PathEscape(args[0])),
		IPSpec,
		ipOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}

func IpSetReverse(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/ip/%s/reverse", url.PathEscape(args[0]))
	if err := httpLib.Client.Post(url, map[string]string{
		"ipReverse": args[1],
		"reverse":   args[2],
	}, nil); err != nil {
		display.ExitError(err.Error())
		return
	}

	fmt.Println("\n⚡️ Reverse correctly set")
}

func IpGetReverse(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/ip/%s/reverse", url.PathEscape(args[0]))
	common.ManageListRequest(url, "", []string{"ipReverse", "reverse"}, flags.GenericFilters)
}

func IpDeleteReverse(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/ip/%s/reverse/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(url, nil); err != nil {
		display.ExitError(err.Error())
		return
	}

	fmt.Println("\n⚡️ Reverse correctly deleted")
}
