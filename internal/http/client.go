// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"

	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/ovhcloud-cli/internal/config"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/version"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"gopkg.in/ini.v1"
)

// OVH API client
var Client *ovh.Client

func InitClient() {
	InitClientWithProfile(nil, "")
}

func InitClientWithProfile(cfg *ini.File, profileOverride string) {
	var err error

	// Init API client
	if runtime.GOARCH == "wasm" && runtime.GOOS == "js" {
		// In WASM mode, we use an unauthenticated client
		Client = &ovh.Client{
			Client: &http.Client{},
		}
		Client.UserAgent = os.Getenv("OVH_USER_AGENT")
		Client.SetEndpoint(os.Getenv("OVH_ENDPOINT"))
	} else {
		profileName := config.GetActiveProfileName(cfg, profileOverride)
		if profileName != "" && !config.IsDefaultProfile(profileName) {
			// Profile mode: read credentials from the profile section
			var endpoint, appKey, appSecret, consumerKey string
			endpoint, appKey, appSecret, consumerKey, err = config.GetProfileCredentials(cfg, profileName)
			if err == nil {
				Client, err = ovh.NewClient(endpoint, appKey, appSecret, consumerKey)
			}
		} else {
			// Legacy mode: let go-ovh read from env/config files
			Client, err = ovh.NewDefaultClient()
		}
		if Client != nil {
			Client.UserAgent = "ovh-cli/" + version.Version
		}
	}
	if err != nil {
		log.Printf(`OVHcloud API client not initialized, please run "ovhcloud login" to authenticate (%s)`, err)
	} else if Client != nil {
		// Chain transports: schemasVersion (adds X-Schemas-Version for /v2/ paths)
		// → debug logging (logs request/response) → default transport (sends over the wire).
		Client.Client.Transport = newSchemasVersionTransport(NewTransport("OVH", http.DefaultTransport))
	}
}

// FetchObjectsParallel fetches one object per id, ten requests at a time.
//
// The result is INDEX-ALIGNED with ids, and stays aligned when ignoreErrors is
// set: an id whose fetch failed keeps its slot, holding the zero value of T. That
// is the contract, not an oversight, and compacting the result would be a silent
// breaking change — seven call sites in internal/services/browser pair
// objects[i] with a name they hold in a parallel slice, so dropping a hole there
// would attach one region's details to another region's name.
//
// What that leaves each caller to do depends on T, and measured across the 30
// call sites all three cases are already handled:
//
//   - ignoreErrors false: a failure returns an error and no holes exist (2 sites).
//   - T is a map or a struct: the hole is nil and has to be skipped explicitly
//     (20 sites do, several of them through FetchExpandedArray, which filters).
//   - T is a slice: the hole is a nil slice, which ranges as empty (8 sites).
//
// So it is safe to leave alone, and worth saying so: the padding looks like a bug
// until you know who depends on it.
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
					// The caller opted to ignore per-item errors (e.g. a feature
					// that is not available in some regions, such as floating IPs
					// on local zones). These errors are expected, so only surface
					// them in debug mode to avoid polluting normal output (#173).
					if flags.Debug {
						log.Printf("error fetching %s: %s", url, err)
					}
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

	objects, err := FetchObjectsParallel[map[string]any](path+"/%s", ids, flags.IgnoreErrors)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch objects: %w", err)
	}

	// If we ignore errors, we can have nil entries in the output, so filter them out
	if flags.IgnoreErrors {
		withNilFiltered := make([]map[string]any, 0, len(objects))
		for _, obj := range objects {
			if obj != nil {
				withNilFiltered = append(withNilFiltered, obj)
			}
		}

		return withNilFiltered, nil
	}

	return objects, nil
}
