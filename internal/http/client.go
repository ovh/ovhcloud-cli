package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/ovh/go-ovh/ovh"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// OVH API client
var Client *ovh.Client

func InitClient() {
	var err error

	// Init API client
	Client, err = ovh.NewDefaultClient()
	if err != nil {
		log.Print(`OVHcloud API client not initialized, please run "ovh-cli login" to authenticate`)
	} else {
		Client.Client.Transport = NewTransport("OVH", http.DefaultTransport)
	}
}

func FetchObjectsParallel[T any](path string, ids []any, ignoreErrors bool) ([]T, error) {
	var (
		parallelRequests = 10
		sem              = semaphore.NewWeighted(int64(parallelRequests))
		objects          = make([]T, len(ids))
		g, ctx           = errgroup.WithContext(context.Background())
	)

	for i, id := range ids {
		if err := sem.Acquire(ctx, 1); err != nil {
			// Here the error is ctx.Err(), so just log it and
			// let the g.Wait() return the "real" error
			log.Printf("failed to acquire semaphore: %s", err)
			break
		}

		g.Go(func() error {
			defer sem.Release(1)
			url := fmt.Sprintf(path, url.PathEscape(fmt.Sprint(id)))

			var object T
			if err := Client.Get(url, &object); err != nil {
				if ignoreErrors {
					log.Printf("error fetching %s: %s", url, err)
					return nil
				}
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

// fetchArray calls the given path (and expects it to return an array), and
// paginates to fetch all the results.
// If "idField" given, it tries to extract the given field from the objects returned
// by the API call.
func FetchArray(path, idField string) ([]any, error) {
	req, err := Client.NewRequest(http.MethodGet, path, nil, true)
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

		response, err := Client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error fetching %s: %s", path, err)
		}

		var pageIDs []any
		if err := Client.UnmarshalResponse(response, &pageIDs); err != nil {
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

func FetchExpandedArray(path, idField string) ([]map[string]any, error) {
	ids, err := FetchArray(path, idField)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ids: %w", err)
	}

	objects, err := FetchObjectsParallel[map[string]any](path+"/%s", ids, false)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch objects: %w", err)
	}

	return objects, nil
}
