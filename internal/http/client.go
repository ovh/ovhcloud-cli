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
	"strings"

	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/ovhcloud-cli/internal/config"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/version"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"gopkg.in/ini.v1"
)

// APIClient wraps a *ovh.Client to work around a go-ovh convenience (see
// getTarget in ovh-go's ovh.go): a request path starting with "/v1/" (this
// CLI never issues "/v2/" paths outside genuine cloud/IAM-style products) is
// aliased to the endpoint's bare domain root instead of being appended under
// its "/1.0" base. That aliasing only holds on the official ovh-eu/ovh-ca/
// ovh-us gateways. Every other endpoint (Kimsufi, SoYouStart, or a custom
// URL) only serves the API under "/1.0", so on those endpoints the "/v1/"
// path must be rewritten before go-ovh builds and signs the request.
type APIClient struct {
	*ovh.Client
	rewriteV1Paths bool
}

// NewAPIClient wraps an already-configured *ovh.Client.
func NewAPIClient(raw *ovh.Client) *APIClient {
	return &APIClient{Client: raw, rewriteV1Paths: !isOfficialEndpoint(raw.Endpoint())}
}

// isOfficialEndpoint returns true for the 3 endpoints where go-ovh's "/v1"
// and "/v2" root-aliasing is known to be valid.
func isOfficialEndpoint(endpoint string) bool {
	switch endpoint {
	case ovh.OvhEU, ovh.OvhCA, ovh.OvhUS:
		return true
	}
	return false
}

// fixPath undoes go-ovh's "/v1" root-aliasing when it doesn't apply to the
// active endpoint. Dropping the "/v1" segment lets the path append under the
// endpoint's own "/1.0" base (the only prefix these endpoints actually
// serve) instead of being routed to the bare domain root.
func (c *APIClient) fixPath(path string) string {
	if !c.rewriteV1Paths {
		return path
	}
	if rest, ok := strings.CutPrefix(path, "/v1/"); ok {
		return "/" + rest
	}
	return path
}

func (c *APIClient) Get(path string, resType interface{}) error {
	return c.Client.Get(c.fixPath(path), resType)
}

func (c *APIClient) Post(path string, reqBody, resType interface{}) error {
	return c.Client.Post(c.fixPath(path), reqBody, resType)
}

func (c *APIClient) Put(path string, reqBody, resType interface{}) error {
	return c.Client.Put(c.fixPath(path), reqBody, resType)
}

func (c *APIClient) Delete(path string, resType interface{}) error {
	return c.Client.Delete(c.fixPath(path), resType)
}

func (c *APIClient) NewRequest(method, path string, reqBody interface{}, needAuth bool) (*http.Request, error) {
	return c.Client.NewRequest(method, c.fixPath(path), reqBody, needAuth)
}

// OVH API client
var Client *APIClient

func InitClient() {
	InitClientWithProfile(nil, "")
}

func InitClientWithProfile(cfg *ini.File, profileOverride string) {
	var err error
	var headers map[string]string
	var rawClient *ovh.Client

	// Init API client
	if runtime.GOARCH == "wasm" && runtime.GOOS == "js" {
		// In WASM mode, we use an unauthenticated client
		rawClient = &ovh.Client{
			Client: &http.Client{},
		}
		rawClient.UserAgent = os.Getenv("OVH_USER_AGENT")
		rawClient.SetEndpoint(os.Getenv("OVH_ENDPOINT"))
	} else {
		profileName := config.GetActiveProfileName(cfg, profileOverride)
		if profileName != "" && !config.IsDefaultProfile(profileName) {
			// Profile mode: read credentials from the profile section
			var endpoint, appKey, appSecret, consumerKey string
			endpoint, appKey, appSecret, consumerKey, err = config.GetProfileCredentials(cfg, profileName)
			if err == nil {
				rawClient, err = ovh.NewClient(endpoint, appKey, appSecret, consumerKey)
				headers = config.GetProfileCustomHeaders(cfg, profileName)
			}
		} else {
			// Legacy mode: let go-ovh read from env/config files
			rawClient, err = ovh.NewDefaultClient()
			if err == nil && cfg != nil {
				if endpoint, _ := config.GetConfigValue(cfg, "default", "endpoint"); endpoint != "" {
					headers = config.GetCustomHeaders(cfg, endpoint)
				}
			}
		}
		if rawClient != nil {
			rawClient.UserAgent = "ovh-cli/" + version.Version
		}
	}
	if err != nil {
		log.Printf(`OVHcloud API client not initialized, please run "ovhcloud login" to authenticate (%s)`, err)
	} else if rawClient != nil {
		// Chain transports: customHeaders (injects user-configured headers) → schemasVersion
		// (adds X-Schemas-Version for /v2/ paths) → debug logging (logs request/response)
		// → default transport (sends over the wire).
		rawClient.Client.Transport = newCustomHeadersTransport(
			newSchemasVersionTransport(NewTransport("OVH", http.DefaultTransport)), headers)

		Client = NewAPIClient(rawClient)
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
