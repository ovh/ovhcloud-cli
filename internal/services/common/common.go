package common

import (
	"fmt"
	"net/url"

	"stash.ovh.net/api/ovh-cli/internal/display"
	filtersLib "stash.ovh.net/api/ovh-cli/internal/filters"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
)

func ManageListRequest(path, idField string, columnsToDisplay, filters []string) {
	body, err := httpLib.FetchExpandedArray(path, idField)
	if err != nil {
		display.ExitError("failed to fetch results: %s", err)
		return
	}

	body, err = filtersLib.FilterLines(body, filters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
		return
	}

	display.RenderTable(body, columnsToDisplay, &flags.OutputFormatConfig)
}

func ManageListRequestNoExpand(path string, columnsToDisplay, filters []string) {
	body, err := httpLib.FetchArray(path, "")
	if err != nil {
		display.ExitError("failed to fetch results: %s", err)
		return
	}

	var objects []map[string]any
	for _, object := range body {
		objects = append(objects, object.(map[string]any))
	}

	objects, err = filtersLib.FilterLines(objects, filters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
		return
	}

	display.RenderTable(objects, columnsToDisplay, &flags.OutputFormatConfig)
}

func ManageObjectRequest(path, objectID, templateContent string) {
	url := fmt.Sprintf("%s/%s", path, url.PathEscape(objectID))

	var object map[string]any
	if err := httpLib.Client.Get(url, &object); err != nil {
		display.ExitError("error fetching %s: %s", url, err)
		return
	}

	display.OutputObject(object, objectID, templateContent, &flags.OutputFormatConfig)
}
