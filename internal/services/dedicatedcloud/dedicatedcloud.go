// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package dedicatedcloud

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
)

var (
	dedicatedcloudColumnsToDisplay = []string{"serviceName", "regionLocation", "version", "state", "description"}

	//go:embed templates/dedicatedcloud.tmpl
	dedicatedcloudTemplate string

	//go:embed templates/datacenter.tmpl
	datacenterTemplate string
)

// toFloat64 converts any numeric type to float64
func toFloat64(v any) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	if i, ok := v.(int); ok {
		return float64(i)
	}
	if n, ok := v.(json.Number); ok {
		if f, err := n.Float64(); err == nil {
			return f
		}
	}
	return 0
}

// toInt converts any numeric type to int
func toInt(v any) int {
	if i, ok := v.(int); ok {
		return i
	}
	if f, ok := v.(float64); ok {
		return int(f)
	}
	if n, ok := v.(json.Number); ok {
		if i, err := n.Int64(); err == nil {
			return int(i)
		}
	}
	return 0
}

// getLocationMap fetches the location mapping from the API
func getLocationMap() (map[string]string, error) {
	// Fetch location list (array of location IDs)
	var locationIDs []string
	if err := httpLib.Client.Get("/dedicatedCloud/location", &locationIDs); err != nil {
		return nil, fmt.Errorf("failed to fetch location list: %w", err)
	}

	// Build location map by fetching each location detail
	locationMap := make(map[string]string)
	for _, locationID := range locationIDs {
		var locationInfo map[string]any
		locationPath := fmt.Sprintf("/dedicatedCloud/location/%s", url.PathEscape(locationID))
		if err := httpLib.Client.Get(locationPath, &locationInfo); err != nil {
			// Skip if we can't fetch this location
			continue
		}
		if regionLocation, ok := locationInfo["regionLocation"].(string); ok {
			locationMap[locationID] = regionLocation
		}
	}

	return locationMap, nil
}

func ListDedicatedCloud(_ *cobra.Command, _ []string) {
	// Fetch dedicatedcloud list
	body, err := httpLib.FetchExpandedArray("/dedicatedCloud", "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch results: %s", err)
		return
	}

	// Fetch location mapping
	locationMap, err := getLocationMap()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Enrich each object with version and regionLocation
	for _, obj := range body {
		// Format version
		if versionObj, ok := obj["version"].(map[string]any); ok {
			versionStr, _ := versionObj["major"].(string)
			if minor, ok := versionObj["minor"].(string); ok && minor != "" {
				versionStr += "." + minor
			}
			if build, ok := versionObj["build"].(string); ok && build != "" {
				versionStr += "." + build
			}
			obj["version"] = versionStr
		}

		// Add regionLocation from location mapping
		if location, ok := obj["location"].(string); ok {
			if regionLocation, ok := locationMap[location]; ok {
				obj["regionLocation"] = regionLocation
			} else {
				obj["regionLocation"] = location
			}
		}
	}

	// Filter results
	body, err = filters.FilterLines(body, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	// Display table
	display.RenderTable(body, dedicatedcloudColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetDedicatedCloud(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/dedicatedCloud/%s", url.PathEscape(args[0]))

	// Fetch dedicatedcloud
	var object map[string]any
	if err := httpLib.Client.Get(endpoint, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching %s: %s", endpoint, err)
		return
	}

	// Fetch location mapping and add regionLocation
	locationMap, err := getLocationMap()
	if err == nil {
		if location, ok := object["location"].(string); ok {
			if regionLocation, ok := locationMap[location]; ok {
				object["regionLocation"] = regionLocation
			} else {
				object["regionLocation"] = location
			}
		}
	}

	// Fetch datacenters list
	datacentersEndpoint := fmt.Sprintf("/dedicatedCloud/%s/datacenter", url.PathEscape(args[0]))
	datacenters, err := httpLib.FetchExpandedArray(datacentersEndpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching datacenters for %s: %s", args[0], err)
		return
	}
	object["datacenters"] = datacenters

	display.OutputObject(object, args[0], dedicatedcloudTemplate, &flags.OutputFormatConfig)
}

func ListDatacenter(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/dedicatedCloud/%s/datacenter", url.PathEscape(args[0]))
	datacenters, err := httpLib.FetchExpandedArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch results: %s", err)
		return
	}

	// Enrich each datacenter with totals
	for i := range datacenters {
		datacenterId := ""
		if idRaw, ok := datacenters[i]["datacenterId"]; ok && idRaw != nil {
			datacenterId = fmt.Sprint(idRaw)
		}

		if datacenterId == "" {
			continue
		}

		// Fetch hosts for this datacenter
		hostsEndpoint := fmt.Sprintf("/dedicatedCloud/%s/datacenter/%s/host", url.PathEscape(args[0]), url.PathEscape(datacenterId))
		hosts, err := httpLib.FetchExpandedArray(hostsEndpoint, "")
		if err != nil {
			// If error, continue with empty totals
			hosts = []map[string]any{}
		}

		// Fetch filers for this datacenter
		localFilersEndpoint := fmt.Sprintf("/dedicatedCloud/%s/datacenter/%s/filer", url.PathEscape(args[0]), url.PathEscape(datacenterId))
		localFilers, err := httpLib.FetchExpandedArray(localFilersEndpoint, "")
		if err != nil {
			localFilers = []map[string]any{}
		}

		globalFilersEndpoint := fmt.Sprintf("/dedicatedCloud/%s/filer", url.PathEscape(args[0]))
		globalFilers, err := httpLib.FetchExpandedArray(globalFilersEndpoint, "")
		if err != nil {
			globalFilers = []map[string]any{}
		}
		allFilers := append(localFilers, globalFilers...)

		// Calculate totals
		totalCores := 0
		totalRAM := 0.0
		totalVMs := 0
		totalDiskSpace := 0.0

		// Sum from hosts
		for _, host := range hosts {
			// Sum cores
			if cpuNumRaw, ok := host["cpuNum"]; ok && cpuNumRaw != nil {
				totalCores += int(toFloat64(cpuNumRaw))
			}

			// Sum RAM
			if ram, ok := host["ram"].(map[string]any); ok {
				if ramValueRaw, ok := ram["value"]; ok && ramValueRaw != nil {
					totalRAM += toFloat64(ramValueRaw)
				}
			}

			// Sum VMs
			if vmTotalRaw, ok := host["vmTotal"]; ok && vmTotalRaw != nil {
				totalVMs += toInt(vmTotalRaw)
			}
		}

		// Sum disk space from filers
		for _, filer := range allFilers {
			if size, ok := filer["size"].(map[string]any); ok {
				if sizeValueRaw, ok := size["value"]; ok && sizeValueRaw != nil {
					sizeValue := toFloat64(sizeValueRaw)
					// Convert to GB if needed
					if sizeUnit, ok := size["unit"].(string); ok {
						if strings.ToUpper(sizeUnit) == "TB" {
							sizeValue *= 1000 // Convert TB to GB
						} else if strings.ToUpper(sizeUnit) == "MB" {
							sizeValue /= 1000 // Convert MB to GB
						}
					}
					totalDiskSpace += sizeValue
				}
			}
		}

		// Add totals to datacenter
		datacenters[i]["totalCores"] = totalCores
		datacenters[i]["totalRAM"] = fmt.Sprintf("%.0f GB", totalRAM)
		datacenters[i]["totalVMs"] = totalVMs
		datacenters[i]["totalDiskSpace"] = fmt.Sprintf("%.0f GB", totalDiskSpace)
	}

	// Filter results
	datacenters, err = filters.FilterLines(datacenters, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	// Display table
	display.RenderTable(datacenters, []string{"datacenterId", "name", "version", "commercialName", "totalCores", "totalRAM", "totalVMs", "totalDiskSpace"}, &flags.OutputFormatConfig)
}

func GetDatacenter(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/dedicatedCloud/%s/datacenter/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))

	// Fetch datacenter
	var object map[string]any
	if err := httpLib.Client.Get(path, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching %s: %s", path, err)
		return
	}

	// Fetch hosts list
	hostsEndpoint := fmt.Sprintf("/dedicatedCloud/%s/datacenter/%s/host", url.PathEscape(args[0]), url.PathEscape(args[1]))
	hosts, err := httpLib.FetchExpandedArray(hostsEndpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching hosts for %s: %s", args[1], err)
		return
	}

	// Enrich hosts with formatted data and group by cluster
	hostsByCluster := make(map[string][]map[string]any)
	for i := range hosts {
		// Format Core Number with GHz in parentheses
		coreNumber := ""
		if cpuNumRaw, ok := hosts[i]["cpuNum"]; ok && cpuNumRaw != nil {
			cpuNumValue := toFloat64(cpuNumRaw)
			if cpuNumValue > 0 {
				coreNumber = fmt.Sprintf("%.0f", cpuNumValue)
				if cpu, ok := hosts[i]["cpu"].(map[string]any); ok {
					if cpuValueRaw, ok := cpu["value"]; ok && cpuValueRaw != nil {
						cpuValue := toFloat64(cpuValueRaw)
						if cpuValue > 0 {
							// Use format: cores (freqGHz)
							coreNumber += fmt.Sprintf(" (%.0fGHz)", cpuValue)
						}
					}
				}
			}
		}
		hosts[i]["coreNumber"] = coreNumber

		// Add VM count
		if vmTotalRaw, ok := hosts[i]["vmTotal"]; ok && vmTotalRaw != nil {
			hosts[i]["vmCount"] = toInt(vmTotalRaw)
		} else {
			hosts[i]["vmCount"] = 0
		}

		// Add maintenance status emoji
		if inMaintenance, ok := hosts[i]["inMaintenance"].(bool); ok && inMaintenance {
			hosts[i]["maintenanceStatus"] = "🔧"
		} else {
			hosts[i]["maintenanceStatus"] = "✅"
		}

		// Add connection state indicator
		connectionStateIndicator := "🔴"
		if connectionState, ok := hosts[i]["connectionState"].(string); ok && connectionState == "connected" {
			connectionStateIndicator = "🟢"
		}
		hosts[i]["connectionStateIndicator"] = connectionStateIndicator

		// Group by cluster
		clusterName := "Unknown"
		if cn, ok := hosts[i]["clusterName"].(string); ok && cn != "" {
			clusterName = cn
		}
		hostsByCluster[clusterName] = append(hostsByCluster[clusterName], hosts[i])
	}
	object["hostsByCluster"] = hostsByCluster
	object["hosts"] = hosts // Keep original for backward compatibility

	// Fetch clusters information
	clustersEndpoint := fmt.Sprintf("/dedicatedCloud/%s/datacenter/%s/cluster", url.PathEscape(args[0]), url.PathEscape(args[1]))
	clustersList, err := httpLib.FetchExpandedArray(clustersEndpoint, "")
	if err != nil {
		// If error, just continue with empty list
		clustersList = []map[string]any{}
	}

	// Build a map of clusterName to clusterId from hosts
	clusterNameToId := make(map[string]int)
	for _, host := range hosts {
		if clusterName, ok := host["clusterName"].(string); ok && clusterName != "" {
			if clusterIdRaw, ok := host["clusterId"]; ok && clusterIdRaw != nil {
				clusterId := toInt(clusterIdRaw)
				if clusterId > 0 {
					clusterNameToId[clusterName] = clusterId
				}
			}
		}
	}

	// Fetch details for each cluster that has hosts
	clustersWithDetails := make([]map[string]any, 0)
	for clusterName := range hostsByCluster {
		clusterId, found := clusterNameToId[clusterName]
		if !found {
			// Try to find clusterId from clustersList by name
			for _, cluster := range clustersList {
				if name, ok := cluster["name"].(string); ok && name == clusterName {
					if idRaw, ok := cluster["clusterId"]; ok && idRaw != nil {
						clusterId = toInt(idRaw)
						break
					}
				}
			}
		}
		if clusterId > 0 {
			clusterDetailEndpoint := fmt.Sprintf("/dedicatedCloud/%s/datacenter/%s/cluster/%d", url.PathEscape(args[0]), url.PathEscape(args[1]), clusterId)
			var clusterDetail map[string]any
			if err := httpLib.Client.Get(clusterDetailEndpoint, &clusterDetail); err == nil {
				// Format drsStatus
				if drsStatus, ok := clusterDetail["drsStatus"].(string); ok {
					if drsStatus == "enabled" {
						clusterDetail["drsStatusFormatted"] = "🟢 enabled"
					} else {
						clusterDetail["drsStatusFormatted"] = "🔴 " + drsStatus
					}
				}
				// Format haStatus
				if haStatus, ok := clusterDetail["haStatus"].(string); ok {
					if haStatus == "enabled" {
						clusterDetail["haStatusFormatted"] = "🟢 enabled"
					} else {
						clusterDetail["haStatusFormatted"] = "🔴 " + haStatus
					}
				}
				// Remove unwanted fields
				delete(clusterDetail, "vmwareClusterId")
				delete(clusterDetail, "autoscale")
				clustersWithDetails = append(clustersWithDetails, clusterDetail)
			}
		}
	}
	object["clusters"] = clustersWithDetails

	// Fetch local filers (datacenter level)
	localFilersEndpoint := fmt.Sprintf("/dedicatedCloud/%s/datacenter/%s/filer", url.PathEscape(args[0]), url.PathEscape(args[1]))
	localFilers, err := httpLib.FetchExpandedArray(localFilersEndpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching local filers for %s: %s", args[1], err)
		return
	}

	// Fetch global filers (service level)
	globalFilersEndpoint := fmt.Sprintf("/dedicatedCloud/%s/filer", url.PathEscape(args[0]))
	globalFilers, err := httpLib.FetchExpandedArray(globalFilersEndpoint, "")
	if err != nil {
		// If error, just continue with empty list
		globalFilers = []map[string]any{}
	}

	// Combine both lists
	allFilers := make([]map[string]any, 0, len(localFilers)+len(globalFilers))
	allFilers = append(allFilers, localFilers...)
	allFilers = append(allFilers, globalFilers...)

	// Enrich filers with formatted data
	for i := range allFilers {
		// Set visibility based on source
		if i < len(localFilers) {
			allFilers[i]["visibility"] = "Local"
		} else {
			allFilers[i]["visibility"] = "Global"
		}
		// Format size
		sizeStr := ""
		if size, ok := allFilers[i]["size"].(map[string]any); ok {
			if sizeValueRaw, ok := size["value"]; ok && sizeValueRaw != nil {
				sizeValue := toFloat64(sizeValueRaw)
				if sizeValue > 0 {
					sizeStr = fmt.Sprintf("%.0f", sizeValue)
					if sizeUnit, ok := size["unit"].(string); ok {
						sizeStr += " " + sizeUnit
					}
				}
			}
		}
		allFilers[i]["sizeFormatted"] = sizeStr

		// Format spaceFree
		spaceFreeStr := ""
		if spaceFreeRaw, ok := allFilers[i]["spaceFree"]; ok && spaceFreeRaw != nil {
			spaceFreeValue := toFloat64(spaceFreeRaw)
			if spaceFreeValue > 0 {
				spaceFreeStr = fmt.Sprintf("%.0f GB", spaceFreeValue)
			}
		}
		allFilers[i]["spaceFreeFormatted"] = spaceFreeStr

		// Extract cluster name from master (first part of domain)
		clusterName := ""
		if master, ok := allFilers[i]["master"].(string); ok && master != "" {
			// Extract first part before first dot
			parts := strings.Split(master, ".")
			if len(parts) > 0 {
				clusterName = parts[0]
			}
		}
		allFilers[i]["clusterName"] = clusterName

		// Add VM count
		if vmTotalRaw, ok := allFilers[i]["vmTotal"]; ok && vmTotalRaw != nil {
			allFilers[i]["vmCount"] = toInt(vmTotalRaw)
		} else {
			allFilers[i]["vmCount"] = 0
		}

		// Add connection state indicator
		connectionStateIndicator := "🔴"
		if connectionState, ok := allFilers[i]["connectionState"].(string); ok && connectionState == "online" {
			connectionStateIndicator = "🟢"
		}
		allFilers[i]["connectionStateIndicator"] = connectionStateIndicator
	}
	object["filers"] = allFilers

	// Calculate totals
	totalCores := 0
	totalRAM := 0.0
	totalVMs := 0
	totalDiskSpace := 0.0

	// Sum from hosts
	for _, host := range hosts {
		// Sum cores
		if cpuNumRaw, ok := host["cpuNum"]; ok && cpuNumRaw != nil {
			totalCores += int(toFloat64(cpuNumRaw))
		}

		// Sum RAM
		if ram, ok := host["ram"].(map[string]any); ok {
			if ramValueRaw, ok := ram["value"]; ok && ramValueRaw != nil {
				totalRAM += toFloat64(ramValueRaw)
			}
		}

		// Sum VMs
		if vmTotalRaw, ok := host["vmTotal"]; ok && vmTotalRaw != nil {
			totalVMs += toInt(vmTotalRaw)
		}
	}

	// Sum disk space from filers
	for _, filer := range allFilers {
		if size, ok := filer["size"].(map[string]any); ok {
			if sizeValueRaw, ok := size["value"]; ok && sizeValueRaw != nil {
				sizeValue := toFloat64(sizeValueRaw)
				// Convert to GB if needed
				if sizeUnit, ok := size["unit"].(string); ok {
					if strings.ToUpper(sizeUnit) == "TB" {
						sizeValue *= 1000 // Convert TB to GB
					} else if strings.ToUpper(sizeUnit) == "MB" {
						sizeValue /= 1000 // Convert MB to GB
					}
				}
				totalDiskSpace += sizeValue
			}
		}
	}

	// Format totals
	object["totalCores"] = totalCores
	object["totalRAM"] = fmt.Sprintf("%.0f GB", totalRAM)
	object["totalVMs"] = totalVMs
	object["totalDiskSpace"] = fmt.Sprintf("%.0f GB", totalDiskSpace)

	display.OutputObject(object, args[1], datacenterTemplate, &flags.OutputFormatConfig)
}
