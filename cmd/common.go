package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"stash.ovh.net/api/ovh-cli/internal/display"
)

func manageListRequest(path string, columnsToDisplay []string) {
	req, err := client.NewRequest(http.MethodGet, path, nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching %s: %s\n", path, err)
	}

	if jsonOutput || yamlOutput {
		var body []map[string]any
		if err := client.UnmarshalResponse(resp, &body); err != nil {
			log.Fatalf("error unmarshalling response: %s\n", err)
		}
		display.RenderTableRaw(body, jsonOutput, yamlOutput)
	} else {
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("error reading response body: %s", err)
		}
		display.RenderTable(bodyBytes, columnsToDisplay)
	}
}

func manageObjectRequest(path, objectID, idKey string) {
	url := fmt.Sprintf("%s/%s", path, url.PathEscape(objectID))

	var object map[string]any
	if err := client.Get(url, &object); err != nil {
		log.Fatalf("error fetching %s: %s\n", url, err)
	}

	display.OutputObject(object, idKey, jsonOutput, yamlOutput)
}
