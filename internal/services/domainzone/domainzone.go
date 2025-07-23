package domainzone

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
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
