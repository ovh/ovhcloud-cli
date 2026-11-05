// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// API v2 path segments of the managed database and managed analytics services.
const (
	managedDatabaseCategory  = "managedDatabase"
	managedAnalyticsCategory = "managedAnalytics"
)

type (
	// databaseAnalyticsTargetSpec is the targetSpec of a managed database or
	// analytics service creation.
	databaseAnalyticsTargetSpec struct {
		Backup struct {
			Regions []string `json:"regions,omitempty"`
			Time    string   `json:"time,omitempty"`
		} `json:"backup,omitzero"`
		DeletionProtected     bool   `json:"deletionProtected,omitempty"`
		Description           string `json:"description,omitempty"`
		Flavor                string `json:"flavor,omitempty"`
		ForkSourceBackupID    string `json:"forkSourceBackupId,omitempty"`
		ForkSourcePointInTime string `json:"forkSourcePointInTime,omitempty"`
		ForkSourceServiceID   string `json:"forkSourceServiceId,omitempty"`
		MaintenanceWindow     struct {
			StartTimeUtc string `json:"startTimeUtc,omitempty"`
		} `json:"maintenanceWindow,omitzero"`
		Network struct {
			AuthorizedIps *databaseAnalyticsAuthorizedIps `json:"authorizedIps,omitempty"`
			NetworkID     string                          `json:"networkId,omitempty"`
			SubnetID      string                          `json:"subnetId,omitempty"`
		} `json:"network,omitzero"`
		NodesCount        int    `json:"nodesCount,omitempty"`
		Plan              string `json:"plan,omitempty"`
		PrometheusEnabled bool   `json:"prometheusEnabled,omitempty"`
		Region            string `json:"region,omitempty"`
		Storage           struct {
			Size *databaseAnalyticsUnitAndValue `json:"size,omitempty"`
		} `json:"storage,omitzero"`
		Version string `json:"version,omitempty"`
	}

	databaseAnalyticsAuthorizedIps struct {
		Rules []databaseAnalyticsAuthorizedIpsRule `json:"rules"`
	}

	databaseAnalyticsAuthorizedIpsRule struct {
		Cidr string `json:"cidr"`
	}

	databaseAnalyticsUnitAndValue struct {
		Unit  string `json:"unit"`
		Value int    `json:"value"`
	}
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

// getDatabaseAnalyticsService returns the engine of a service and the service,
// with its engine added. A v2 service does not carry its engine and the route
// of an engine answers 404 for a service of another engine, so every engine is
// tried.
func getDatabaseAnalyticsService(projectID, category, serviceID string, engines []string) (string, map[string]any, error) {
	services := make([]map[string]any, len(engines))
	errs := make([]error, len(engines))

	var wg sync.WaitGroup
	for i, engine := range engines {
		wg.Go(func() {
			errs[i] = httpLib.Client.Get(databaseAnalyticsEndpoint(projectID, category, engine, serviceID), &services[i])
		})
	}
	wg.Wait()

	for i, engine := range engines {
		if errs[i] == nil {
			services[i]["engine"] = engine
			return engine, services[i], nil
		}
	}

	// An error other than a 404 (e.g. an engine the IAM policy denies) may hide the service
	for _, err := range errs {
		if apiErr, ok := errors.AsType[*ovh.APIError](err); !ok || apiErr.Code != http.StatusNotFound {
			return "", nil, err
		}
	}

	return "", nil, fmt.Errorf("service %s not found", serviceID)
}

// listDatabaseAnalyticsServices displays the services of a category with their
// engine, listed engine by engine since a v2 service does not carry its engine.
func listDatabaseAnalyticsServices(projectID, category string, engines, columnsToDisplay []string) {
	lists := make([][]any, len(engines))
	errs := make([]error, len(engines))

	var wg sync.WaitGroup
	for i, engine := range engines {
		wg.Go(func() {
			lists[i], errs[i] = httpLib.FetchArray(databaseAnalyticsEndpoint(projectID, category, engine), "")
		})
	}
	wg.Wait()

	var services []map[string]any
	for i, engine := range engines {
		if errs[i] != nil {
			if !flags.IgnoreErrors {
				display.OutputError(&flags.OutputFormatConfig, "failed to fetch %s services: %s", engine, errs[i])
				return
			}
			if flags.Debug {
				log.Printf("failed to fetch %s services: %s", engine, errs[i])
			}
			continue
		}

		for _, item := range lists[i] {
			service := item.(map[string]any)
			service["engine"] = engine
			services = append(services, service)
		}
	}

	services, err := filtersLib.FilterLines(services, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(services, columnsToDisplay, &flags.OutputFormatConfig)
}

// newAuthorizedIps returns the authorized IPs of the given CIDR blocks, or nil
// when there are none.
func newAuthorizedIps(cidrs []string) *databaseAnalyticsAuthorizedIps {
	if len(cidrs) == 0 {
		return nil
	}

	ips := &databaseAnalyticsAuthorizedIps{}
	for _, cidr := range cidrs {
		ips.Rules = append(ips.Rules, databaseAnalyticsAuthorizedIpsRule{Cidr: cidr})
	}
	return ips
}

// setNodes sets the flavor, region and number of nodes from a nodes list
// (flavor1:region1,flavor2:region2...). API v2 takes a single flavor and
// region for all the nodes of a service.
func (s *databaseAnalyticsTargetSpec) setNodes(nodes []string) error {
	for _, node := range nodes {
		parts := strings.Split(node, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid node format: %s (expected format: flavor1:region1,flavor2:region2...)", node)
		}
		if s.NodesCount > 0 && (parts[0] != s.Flavor || parts[1] != s.Region) {
			return fmt.Errorf("invalid nodes list: all nodes must have the same flavor and region")
		}
		s.Flavor, s.Region = parts[0], parts[1]
		s.NodesCount++
	}
	return nil
}

// resetDatabaseAnalyticsUserCredentials resets the password of the user at
// endpoint and returns the user with its new password. The keys created by
// "ovhcloud login" grant no PATCH, so the reset uses resetPassword rather than a
// PATCH of credentialsLastUpdatedAt.
func resetDatabaseAnalyticsUserCredentials(endpoint string) (map[string]any, error) {
	var user map[string]any
	if err := httpLib.Client.Post(endpoint+"/resetPassword", nil, &user); err != nil {
		return nil, err
	}
	return user, nil
}

// getDatabaseAnalyticsBackup displays a backup of the given service. API v2
// serves backups for a whole category, so the service is checked here.
func getDatabaseAnalyticsBackup(projectID, category, serviceID, backupID, template string) {
	var backup map[string]any
	if err := httpLib.Client.Get(databaseAnalyticsEndpoint(projectID, category, "backup", backupID), &backup); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch backup: %s", err)
		return
	}

	if backup["serviceId"] != serviceID {
		display.OutputError(&flags.OutputFormatConfig, "backup %s not found in service %s", backupID, serviceID)
		return
	}

	display.OutputObject(backup, backupID, template, &flags.OutputFormatConfig)
}
