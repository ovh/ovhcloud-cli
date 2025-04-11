package cmd

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"stash.ovh.net/api/ovh-cli/internal/display"
	filtersLib "stash.ovh.net/api/ovh-cli/internal/filters"
)

func fetchExpandedArray(path string) ([]map[string]any, error) {
	req, err := client.NewRequest(http.MethodGet, path, nil, true)
	if err != nil {
		return nil, fmt.Errorf("error crafting request: %s", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching %s: %s", path, err)
	}

	var body []map[string]any
	if err := client.UnmarshalResponse(resp, &body); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %s", err)
	}

	return body, nil
}

func manageListRequest(path string, columnsToDisplay, filters []string) {
	body, err := fetchExpandedArray(path)
	if err != nil {
		log.Fatalf("failed to fetch results: %s", err)
	}

	body, err = filtersLib.FilterLines(body, filters)
	if err != nil {
		log.Fatalf("failed to filter results: %s", err)
	}

	display.RenderTable(body, columnsToDisplay, jsonOutput, yamlOutput)
}

func manageObjectRequest(path, objectID, templateContent string) {
	url := fmt.Sprintf("%s/%s", path, url.PathEscape(objectID))

	var object map[string]any
	if err := client.Get(url, &object); err != nil {
		log.Fatalf("error fetching %s: %s", url, err)
	}

	display.OutputObject(object, objectID, templateContent, jsonOutput, yamlOutput, interactiveOutput)
}
