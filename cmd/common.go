package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"stash.ovh.net/api/ovh-cli/internal/display"
	filtersLib "stash.ovh.net/api/ovh-cli/internal/filters"
)

func fetchObjectsParallel(path string, ids []any) ([]map[string]any, error) {
	var (
		parallelRequests = 10
		sem              = semaphore.NewWeighted(int64(parallelRequests))
		objects          = make([]map[string]any, len(ids))
		g, ctx           = errgroup.WithContext(context.Background())
	)

	for i, id := range ids {
		if err := sem.Acquire(ctx, 1); err != nil {
			return nil, fmt.Errorf("failed to acquire semaphore: %w", err)
		}

		g.Go(func() error {
			defer sem.Release(1)
			url := fmt.Sprintf("%s/%s", path, url.PathEscape(fmt.Sprint(id)))

			var object map[string]any
			if err := client.Get(url, &object); err != nil {
				return fmt.Errorf("failed to fetch object %q: %w", fmt.Sprint(id), err)
			}

			objects[i] = object

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return objects, nil
}

func fetchIDs(path, idField string) ([]any, error) {
	req, err := client.NewRequest(http.MethodGet, path, nil, true)
	if err != nil {
		return nil, fmt.Errorf("error crafting request: %s", err)
	}

	var (
		allIDs     []any
		nextCursor string
	)

	for {
		if nextCursor != "" {
			req.Header.Set("X-Pagination-Cursor", nextCursor)
		}

		response, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error fetching %s: %s", path, err)
		}

		var pageIDs []any
		if err := client.UnmarshalResponse(response, &pageIDs); err != nil {
			return nil, fmt.Errorf("failed to parse ids: %s", err)
		}

		if idField != "" {
			for _, item := range pageIDs {
				object, ok := item.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("failed to extract ID from object, value %q is not an object", item)
				}
				allIDs = append(allIDs, object[idField])
			}
		} else {
			allIDs = append(allIDs, pageIDs...)
		}

		nextCursor = response.Header.Get("X-Pagination-Cursor-Next")
		if nextCursor == "" {
			break
		}
	}

	return allIDs, nil
}

func fetchExpandedArray(path, idField string) ([]map[string]any, error) {
	ids, err := fetchIDs(path, idField)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ids: %w", err)
	}

	objects, err := fetchObjectsParallel(path, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch objects: %w", err)
	}

	return objects, nil
}

func manageListRequest(path, idField string, columnsToDisplay, filters []string) {
	body, err := fetchExpandedArray(path, idField)
	if err != nil {
		display.ExitError("failed to fetch results: %s", err)
	}

	body, err = filtersLib.FilterLines(body, filters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
	}

	display.RenderTable(body, columnsToDisplay, &outputFormatConfig)
}

func manageObjectRequest(path, objectID, templateContent string) {
	url := fmt.Sprintf("%s/%s", path, url.PathEscape(objectID))

	var object map[string]any
	if err := client.Get(url, &object); err != nil {
		display.ExitError("error fetching %s: %s", url, err)
	}

	display.OutputObject(object, objectID, templateContent, &outputFormatConfig)
}

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}
