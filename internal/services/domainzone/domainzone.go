package domainzone

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
	domainzoneColumnsToDisplay = []string{"name", "dnssecSupported", "hasDnsAnycast"}

	//go:embed templates/domainzone.tmpl
	domainzoneTemplate string
)

func ListDomainZone(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/domain/zone", "", domainzoneColumnsToDisplay, flags.GenericFilters)
}

func GetDomainZone(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/domain/zone/%s", url.PathEscape(args[0]))

	// Fetch domain zone
	var object map[string]any
	if err := httpLib.Client.Get(path, &object); err != nil {
		display.ExitError("error fetching %s: %s\n", path, err)
		return
	}

	// Fetch running tasks
	path = fmt.Sprintf("/domain/zone/%s/record", url.PathEscape(args[0]))
	records, err := httpLib.FetchExpandedArray(path, "")
	if err != nil {
		display.ExitError("error fetching records for %s: %s", args[0], err)
		return
	}
	object["records"] = records

	display.OutputObject(object, args[0], domainzoneTemplate, &flags.OutputFormatConfig)
}
