// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"fmt"
	"net/url"
)

// API v2 path segments of the managed database and managed analytics services.
const (
	managedDatabaseCategory  = "managedDatabase"
	managedAnalyticsCategory = "managedAnalytics"
)

// databaseAnalyticsEndpoint returns the API v2 endpoint of the given category,
// managedDatabase or managedAnalytics, followed by the given path elements.
func databaseAnalyticsEndpoint(projectID, category string, elems ...string) string {
	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/%s", projectID, category)
	for _, elem := range elems {
		endpoint += "/" + url.PathEscape(elem)
	}
	return endpoint
}
