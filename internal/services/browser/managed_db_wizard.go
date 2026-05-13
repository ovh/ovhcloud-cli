// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// ─── API / fetch ──────────────────────────────────────────────────────────────

// fetchDBCapabilities fetches capabilities and availability for managed databases.
func (m Model) fetchDBCapabilities() tea.Cmd {
	return func() tea.Msg {
		capEndpoint := fmt.Sprintf("/v1/cloud/project/%s/database/capabilities", m.cloudProject)
		availEndpoint := fmt.Sprintf("/v1/cloud/project/%s/database/availability?action=create", m.cloudProject)

		var caps map[string]interface{}
		if err := httpLib.Client.Get(capEndpoint, &caps); err != nil {
			return dbCapabilitiesLoadedMsg{err: err}
		}

		var availItems []map[string]interface{}
		_ = httpLib.Client.Get(availEndpoint, &availItems)

		// Fallback regions from capabilities top-level list
		var capsRegions []string
		if raws, ok := caps["regions"].([]interface{}); ok {
			for _, r := range raws {
				if s, ok := r.(string); ok {
					capsRegions = append(capsRegions, s)
				}
			}
			sort.Strings(capsRegions)
		}

		engines, _ := caps["engines"].([]interface{})
		var engMaps []map[string]interface{}
		for _, e := range engines {
			if em, ok := e.(map[string]interface{}); ok {
				engMaps = append(engMaps, em)
			}
		}
		sort.Slice(engMaps, func(i, j int) bool {
			return getStringValue(engMaps[i], "name", "") < getStringValue(engMaps[j], "name", "")
		})

		flavors, _ := caps["flavors"].([]interface{})
		var flavMaps []map[string]interface{}
		for _, f := range flavors {
			if fm, ok := f.(map[string]interface{}); ok {
				flavMaps = append(flavMaps, fm)
			}
		}
		sort.Slice(flavMaps, func(i, j int) bool {
			oi, _ := toFloat64(flavMaps[i]["order"])
			oj, _ := toFloat64(flavMaps[j]["order"])
			if oi != oj {
				return oi < oj
			}
			return getStringValue(flavMaps[i], "name", "") < getStringValue(flavMaps[j], "name", "")
		})

		plans, _ := caps["plans"].([]interface{})
		var planMaps []map[string]interface{}
		for _, p := range plans {
			if pm, ok := p.(map[string]interface{}); ok {
				planMaps = append(planMaps, pm)
			}
		}
		sort.Slice(planMaps, func(i, j int) bool {
			oi, _ := toFloat64(planMaps[i]["order"])
			oj, _ := toFloat64(planMaps[j]["order"])
			return oi < oj
		})

		return dbCapabilitiesLoadedMsg{
			engines:     engMaps,
			flavors:     flavMaps,
			plans:       planMaps,
			availItems:  availItems,
			capsRegions: capsRegions,
		}
	}
}

// createManagedDBFromWizard calls the OVHcloud API to create the managed DB service.
func (m Model) createManagedDBFromWizard() tea.Cmd {
	return func() tea.Msg {
		engine := m.wizard.dbEngine
		if engine == "" {
			return dbCreatedMsg{err: fmt.Errorf("no engine selected")}
		}
		body := map[string]interface{}{
			"description": m.wizard.dbName,
			"version":     m.wizard.dbVersion,
			"plan":        m.wizard.dbPlan,
			"nodesPattern": map[string]interface{}{
				"flavor": m.wizard.dbFlavor,
				"region": m.wizard.dbRegion,
				"number": m.wizard.dbNodes,
			},
		}
		if m.wizard.dbNetworkIdx == 1 && m.wizard.dbNetworkId != "" {
			body["networkId"] = m.wizard.dbNetworkId
		}
		if m.wizard.dbDiskSize > 0 {
			body["disk"] = map[string]interface{}{"size": m.wizard.dbDiskSize}
		}

		endpoint := fmt.Sprintf("/v1/cloud/project/%s/database/%s",
			m.cloudProject, url.PathEscape(engine))
		var result map[string]interface{}
		if err := httpLib.Client.Post(endpoint, body, &result); err != nil {
			return dbCreatedMsg{err: fmt.Errorf("failed to create database: %w", err)}
		}
		name := m.wizard.dbName
		if name == "" {
			name = getStringValue(result, "id", "database")
		}
		return dbCreatedMsg{dbName: name}
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// getAvailFlavor returns the flavor name from an availability item,
// checking the deprecated top-level field first, then specifications.flavor (new format).
func getAvailFlavor(a map[string]interface{}) string {
	if f := getStringValue(a, "flavor", ""); f != "" {
		return f
	}
	if specs, ok := a["specifications"].(map[string]interface{}); ok {
		return getStringValue(specs, "flavor", "")
	}
	return ""
}

// toFloat64 converts a JSON value to float64, handling both float64 and json.Number
// (go-ovh uses UseNumber() so all numbers come as json.Number).
func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	case int64:
		return float64(n), true
	case int:
		return float64(n), true
	}
	return 0, false
}

// dbFilteredVersions returns versions available for the selected engine.
func (m Model) dbFilteredVersions() []string {
	engLower := strings.ToLower(m.wizard.dbEngine)
	for _, e := range m.wizard.dbEngines {
		if strings.ToLower(getStringValue(e, "name", "")) == engLower {
			if versions, ok := e["versions"].([]interface{}); ok {
				var vs []string
				for _, v := range versions {
					if s, ok := v.(string); ok {
						vs = append(vs, s)
					}
				}
				return vs
			}
		}
	}
	return nil
}

// dbFilteredRegions returns unique regions from availability matching engine+version.
// Falls back to capabilities regions if the availability endpoint returned nothing.
func (m Model) dbFilteredRegions() []string {
	seen := map[string]bool{}
	var regions []string
	engLower := strings.ToLower(m.wizard.dbEngine)
	for _, a := range m.wizard.dbAvailItems {
		if strings.ToLower(getStringValue(a, "engine", "")) != engLower {
			continue
		}
		if getStringValue(a, "version", "") != m.wizard.dbVersion {
			continue
		}
		r := getStringValue(a, "region", "")
		if r != "" && !seen[r] {
			seen[r] = true
			regions = append(regions, r)
		}
	}
	if len(regions) == 0 {
		// Availability endpoint returned nothing — use capabilities regions as fallback
		return m.wizard.dbCapsRegions
	}
	sort.Strings(regions)
	return regions
}

// dbFilteredPlans returns unique plans from availability matching engine+version+region.
// Falls back to capabilities plans when availability returns nothing.
func (m Model) dbFilteredPlans() []string {
	seen := map[string]bool{}
	var plans []string
	engLower := strings.ToLower(m.wizard.dbEngine)
	for _, a := range m.wizard.dbAvailItems {
		if strings.ToLower(getStringValue(a, "engine", "")) != engLower {
			continue
		}
		if getStringValue(a, "version", "") != m.wizard.dbVersion {
			continue
		}
		if !strings.EqualFold(getStringValue(a, "region", ""), m.wizard.dbRegion) {
			continue
		}
		p := getStringValue(a, "plan", "")
		if p != "" && !seen[p] {
			seen[p] = true
			plans = append(plans, p)
		}
	}
	if len(plans) == 0 {
		// Fallback: use capabilities plan list
		plans = append(plans, m.wizard.dbCapPlans...)
	}
	// Sort by predefined order
	order := map[string]int{"discovery": 0, "production": 1, "advanced": 2}
	sort.Slice(plans, func(i, j int) bool {
		oi, iok := order[plans[i]]
		oj, jok := order[plans[j]]
		if iok && jok {
			return oi < oj
		}
		return plans[i] < plans[j]
	})
	return plans
}

// dbFilteredFlavors returns unique flavors from availability matching engine+version+region+plan.
// Falls back to capabilities flavors filtered by engine category prefix when availability is empty.
func (m Model) dbFilteredFlavors() []string {
	seen := map[string]bool{}
	var flavors []string
	engLower := strings.ToLower(m.wizard.dbEngine)
	for _, a := range m.wizard.dbAvailItems {
		if strings.ToLower(getStringValue(a, "engine", "")) != engLower {
			continue
		}
		if getStringValue(a, "version", "") != m.wizard.dbVersion {
			continue
		}
		if !strings.EqualFold(getStringValue(a, "region", ""), m.wizard.dbRegion) {
			continue
		}
		if !strings.EqualFold(getStringValue(a, "plan", ""), m.wizard.dbPlan) {
			continue
		}
		f := getAvailFlavor(a)
		if f != "" && !seen[f] {
			seen[f] = true
			flavors = append(flavors, f)
		}
	}
	if len(flavors) == 0 {
		// Fallback: availability returned nothing, show all capabilities flavors.
		for _, f := range m.wizard.dbFlavors {
			name := getStringValue(f, "name", "")
			if name != "" {
				flavors = append(flavors, name)
			}
		}
		return flavors
	}
	// Sort by order in dbFlavors list
	orderMap := map[string]int{}
	for i, f := range m.wizard.dbFlavors {
		orderMap[getStringValue(f, "name", "")] = i
	}
	sort.Slice(flavors, func(i, j int) bool {
		oi := orderMap[flavors[i]]
		oj := orderMap[flavors[j]]
		if oi != oj {
			return oi < oj
		}
		return flavors[i] < flavors[j]
	})
	return flavors
}

// dbActiveAvail returns the availability entry matching current engine+version+region+plan+flavor.
// It tries progressively relaxed matches to handle analytics per-engine items
// where some top-level deprecated fields may be absent or formatted differently.
func (m Model) dbActiveAvail() map[string]interface{} {
	engLower := strings.ToLower(m.wizard.dbEngine)

	// Pass 1: strict — all 5 fields
	for _, a := range m.wizard.dbAvailItems {
		if strings.ToLower(getStringValue(a, "engine", "")) != engLower {
			continue
		}
		if getStringValue(a, "version", "") != m.wizard.dbVersion {
			continue
		}
		if !strings.EqualFold(getStringValue(a, "region", ""), m.wizard.dbRegion) {
			continue
		}
		if !strings.EqualFold(getStringValue(a, "plan", ""), m.wizard.dbPlan) {
			continue
		}
		if !strings.EqualFold(getAvailFlavor(a), m.wizard.dbFlavor) {
			continue
		}
		return a
	}

	// Pass 2: engine + region + plan + flavor, relaxing version.
	// If the item has empty region or plan (per-engine endpoints may omit them),
	// treat that field as a wildcard match.
	for _, a := range m.wizard.dbAvailItems {
		if strings.ToLower(getStringValue(a, "engine", "")) != engLower {
			continue
		}
		if r := getStringValue(a, "region", ""); r != "" && !strings.EqualFold(r, m.wizard.dbRegion) {
			continue
		}
		if p := getStringValue(a, "plan", ""); p != "" && !strings.EqualFold(p, m.wizard.dbPlan) {
			continue
		}
		if !strings.EqualFold(getAvailFlavor(a), m.wizard.dbFlavor) {
			continue
		}
		return a
	}

	// Pass 3: engine + flavor only (absolute last resort)
	for _, a := range m.wizard.dbAvailItems {
		if strings.ToLower(getStringValue(a, "engine", "")) != engLower {
			continue
		}
		if !strings.EqualFold(getAvailFlavor(a), m.wizard.dbFlavor) {
			continue
		}
		return a
	}

	return nil
}

// dbFlavorInfo returns the capabilities.flavor entry for the specified flavor name.
func (m Model) dbFlavorInfo(flavorName string) map[string]interface{} {
	for _, f := range m.wizard.dbFlavors {
		if strings.EqualFold(getStringValue(f, "name", ""), flavorName) {
			return f
		}
	}
	return nil
}

// dbNodesConstraints returns (min, max) nodes allowed for the selected plan+flavor from availability.
func (m Model) dbNodesConstraints() (int, int) {
	avail := m.dbActiveAvail()
	if avail == nil {
		// Fallback based on plan
		switch m.wizard.dbPlan {
		case "discovery":
			return 1, 1
		case "production":
			return 2, 2
		case "advanced":
			return 3, 3
		default:
			return 1, 3
		}
	}
	min := 0
	max := 0
	// Prefer specifications.nodes (new format) over deprecated top-level fields
	if specs, ok := avail["specifications"].(map[string]interface{}); ok {
		if nodes, ok := specs["nodes"].(map[string]interface{}); ok {
			if v, ok := toFloat64(nodes["minimum"]); ok && v > 0 {
				min = int(v)
			}
			if v, ok := toFloat64(nodes["maximum"]); ok && v > 0 {
				max = int(v)
			}
		}
	}
	// Fall back to deprecated fields
	if min == 0 {
		if v, ok := toFloat64(avail["minNodeNumber"]); ok && v > 0 {
			min = int(v)
		}
	}
	if max == 0 {
		if v, ok := toFloat64(avail["maxNodeNumber"]); ok && v > 0 {
			max = int(v)
		}
	}
	if min < 1 {
		min = 1
	}
	if max < min {
		max = min
	}
	return min, max
}

// dbStorageConstraints returns (min, max, step) disk sizes in GB from availability.
func (m Model) dbStorageConstraints() (int, int, int) {
	avail := m.dbActiveAvail()
	min, max, step := 10, 1000, 10
	if avail != nil {
		// Prefer specifications.storage (new format) over deprecated top-level fields
		if specs, ok := avail["specifications"].(map[string]interface{}); ok {
			if storage, ok := specs["storage"].(map[string]interface{}); ok {
				if minS, ok := storage["minimum"].(map[string]interface{}); ok {
					if v, ok := toFloat64(minS["value"]); ok && v > 0 {
						min = int(v)
					}
				}
				if maxS, ok := storage["maximum"].(map[string]interface{}); ok {
					if v, ok := toFloat64(maxS["value"]); ok && v > 0 {
						max = int(v)
					}
				}
				if stepS, ok := storage["step"].(map[string]interface{}); ok {
					if v, ok := toFloat64(stepS["value"]); ok && v > 0 {
						step = int(v)
					}
				}
			}
		}
		// Fall back to deprecated fields
		if v, ok := toFloat64(avail["minDiskSize"]); ok && v > 0 && min == 10 {
			min = int(v)
		}
		if v, ok := toFloat64(avail["maxDiskSize"]); ok && v > 0 && max == 1000 {
			max = int(v)
		}
		if v, ok := toFloat64(avail["stepDiskSize"]); ok && v > 0 && step == 10 {
			step = int(v)
		}
	}
	return min, max, step
}

// dbEngineDescription returns the description for a given engine name.
func dbEngineDescription(name string) string {
	switch strings.ToLower(name) {
	case "postgresql":
		return "Relational and object-relational database"
	case "mysql":
		return "Relational database management system"
	case "mongodb":
		return "Document-oriented database system"
	case "valkey":
		return "In-memory key-value data store"
	case "redis":
		return "In-memory key-value data store"
	case "kafka":
		return "Distributed event streaming platform"
	case "cassandra":
		return "Wide-column distributed database"
	case "opensearch":
		return "Distributed search and analytics engine"
	case "grafana":
		return "Analytics and monitoring platform"
	case "m3db":
		return "Distributed time series database"
	}
	return ""
}

// dbPlanDescription returns the description for a given plan name.
func dbPlanDescription(plan string) string {
	switch plan {
	case "discovery":
		return "1 node • Manual and automatic backups • Private networks"
	case "production":
		return "2 nodes • Manual and automatic backups • Private networks"
	case "advanced":
		return "3 nodes • Manual and automatic backups • Private networks"
	}
	return ""
}

// ─── Render functions ─────────────────────────────────────────────────────────

func (m Model) renderDBWizardNameStep(_ int) string {
	var sb strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	sb.WriteString(title.Render("Service name:") + "\n\n")
	sb.WriteString(desc.Render("Choose a unique name to identify your managed database service.") + "\n\n")
	if m.wizard.errorMsg != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}
	input := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(0, 1).Width(40)
	sb.WriteString(input.Render(m.wizard.dbNameInput+"▌") + "\n\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Type the name • Enter: Continue • Esc: Cancel"))
	return sb.String()
}

func (m Model) renderDBWizardEngineStep(_ int) string {
	var sb strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	sel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	sub := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	sb.WriteString(title.Render("Select a database engine:") + "\n\n")
	if m.wizard.isLoading {
		sb.WriteString(loadingStyle.Render("⏳ Loading capabilities..."))
		return sb.String()
	}
	if m.wizard.errorMsg != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}
	for i, e := range m.wizard.dbEngines {
		name := getStringValue(e, "name", "")
		engineDesc := getStringValue(e, "description", dbEngineDescription(name))
		if i == m.wizard.dbEngineIdx {
			sb.WriteString(sel.Render("▶ "+strings.ToUpper(name)) + "\n")
			sb.WriteString(sub.Render("    "+engineDesc) + "\n\n")
		} else {
			sb.WriteString(dim.Render("  "+strings.ToUpper(name)) + "\n")
			sb.WriteString(sub.Render("    "+engineDesc) + "\n\n")
		}
	}
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ Navigate • Enter: Select • ← Back • Esc: Cancel"))
	return sb.String()
}

func (m Model) renderDBWizardVersionStep(_ int) string {
	versions := m.dbFilteredVersions()
	var sb strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	sel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	sb.WriteString(title.Render("Select a version:") + "\n\n")
	sb.WriteString(desc.Render(fmt.Sprintf("Engine: %s", strings.ToUpper(m.wizard.dbEngine))) + "\n\n")
	if m.wizard.errorMsg != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}
	if len(versions) == 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).
			Render("No versions available for this engine.") + "\n")
	} else {
		for i, v := range versions {
			if i == m.wizard.dbVersionIdx {
				sb.WriteString(sel.Render("▶ "+v) + "\n")
			} else {
				sb.WriteString(dim.Render("  "+v) + "\n")
			}
		}
	}
	sb.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ Navigate • Enter: Select • ← Back • Esc: Cancel"))
	return sb.String()
}

func (m Model) renderDBWizardRegionStep(_ int) string {
	regions := m.dbFilteredRegions()
	var sb strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	sel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	sb.WriteString(title.Render("Select a datacenter:") + "\n\n")
	sb.WriteString(desc.Render(fmt.Sprintf("Engine: %s %s", strings.ToUpper(m.wizard.dbEngine), m.wizard.dbVersion)) + "\n\n")
	if m.wizard.errorMsg != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}
	if len(regions) == 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).
			Render("No regions available for this configuration.") + "\n")
	} else {
		maxVisible := 14
		startIdx := 0
		if m.wizard.dbRegionIdx >= maxVisible {
			startIdx = m.wizard.dbRegionIdx - maxVisible + 1
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(regions) {
			endIdx = len(regions)
		}
		for i := startIdx; i < endIdx; i++ {
			r := regions[i]
			if i == m.wizard.dbRegionIdx {
				sb.WriteString(sel.Render("▶ "+r) + "\n")
			} else {
				sb.WriteString(dim.Render("  "+r) + "\n")
			}
		}
		if len(regions) > maxVisible {
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
				Render(fmt.Sprintf("\n  %d / %d regions", m.wizard.dbRegionIdx+1, len(regions))))
		}
	}
	sb.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ Navigate • Enter: Select • ← Back • Esc: Cancel"))
	return sb.String()
}

func (m Model) renderDBWizardPlanStep(_ int) string {
	plans := m.dbFilteredPlans()
	var sb strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	sel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	sb.WriteString(title.Render("Select a plan:") + "\n\n")
	sb.WriteString(desc.Render(fmt.Sprintf("Engine: %s %s  •  Region: %s",
		strings.ToUpper(m.wizard.dbEngine), m.wizard.dbVersion, m.wizard.dbRegion)) + "\n\n")
	if m.wizard.errorMsg != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}
	if len(plans) == 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).
			Render("No plans available for this configuration.") + "\n")
	} else {
		for i, p := range plans {
			planDesc := dbPlanDescription(p)
			if i == m.wizard.dbPlanIdx {
				sb.WriteString(sel.Render("▶ "+strings.ToUpper(p)) + "\n")
				sb.WriteString(info.Render("    "+planDesc) + "\n\n")
			} else {
				sb.WriteString(dim.Render("  "+strings.ToUpper(p)) + "\n")
				sb.WriteString(info.Render("    "+planDesc) + "\n\n")
			}
		}
	}
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ Navigate • Enter: Select • ← Back • Esc: Cancel"))
	return sb.String()
}

func (m Model) renderDBWizardFlavorStep(_ int) string {
	flavors := m.dbFilteredFlavors()
	var sb strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	sel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	sb.WriteString(title.Render("Select an instance type:") + "\n\n")
	sb.WriteString(desc.Render(fmt.Sprintf("Engine: %s %s  •  Region: %s  •  Plan: %s",
		strings.ToUpper(m.wizard.dbEngine), m.wizard.dbVersion,
		m.wizard.dbRegion, strings.ToUpper(m.wizard.dbPlan))) + "\n\n")
	if m.wizard.errorMsg != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}
	if len(flavors) == 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).
			Render("No instance types available for this configuration.") + "\n")
	} else {
		for i, f := range flavors {
			fi := m.dbFlavorInfo(f)
			detail := ""
			if fi != nil {
				core, _ := toFloat64(fi["core"])
				mem, _ := toFloat64(fi["memory"])
				if core > 0 || mem > 0 {
					detail = fmt.Sprintf("  %d vCores  %d GB RAM", int(core), int(mem))
				}
			}
			if i == m.wizard.dbFlavorIdx {
				sb.WriteString(sel.Render("▶ "+f) + "\n")
				if detail != "" {
					sb.WriteString(info.Render("   "+detail) + "\n\n")
				} else {
					sb.WriteString("\n")
				}
			} else {
				sb.WriteString(dim.Render("  "+f) + "\n")
				if detail != "" {
					sb.WriteString(info.Render("   "+detail) + "\n\n")
				} else {
					sb.WriteString("\n")
				}
			}
		}
	}
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ Navigate • Enter: Select • ← Back • Esc: Cancel"))
	return sb.String()
}

func (m Model) renderDBWizardNodesStep(_ int) string {
	var sb strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	minN, maxN := m.dbNodesConstraints()
	sb.WriteString(title.Render("Number of cluster nodes:") + "\n\n")
	sb.WriteString(desc.Render(fmt.Sprintf(
		"Engine: %s %s  •  Plan: %s  •  Flavor: %s",
		strings.ToUpper(m.wizard.dbEngine), m.wizard.dbVersion,
		strings.ToUpper(m.wizard.dbPlan), m.wizard.dbFlavor)) + "\n\n")
	if m.wizard.errorMsg != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}
	input := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(0, 1).Width(14)
	sb.WriteString(input.Render(m.wizard.dbNodesInput+"▌") + "\n\n")
	sb.WriteString(info.Render(fmt.Sprintf("Allowed range: %d – %d node(s)", minN, maxN)) + "\n\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Type a number • Enter: Continue • ← Back • Esc: Cancel"))
	return sb.String()
}

func (m Model) renderDBWizardStorageStep(_ int) string {
	var sb strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	sb.WriteString(title.Render("Storage size (GB):") + "\n\n")
	sb.WriteString(desc.Render(fmt.Sprintf(
		"Flavor: %s  •  Plan: %s",
		m.wizard.dbFlavor, strings.ToUpper(m.wizard.dbPlan))) + "\n\n")
	if m.wizard.errorMsg != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}
	if m.dbActiveAvail() == nil {
		sb.WriteString(info.Render("No storage constraints available — default storage will be used.") + "\n\n")
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
			Render("Enter: Continue with default • ← Back • Esc: Cancel"))
		return sb.String()
	}
	minS, maxS, stepS := m.dbStorageConstraints()
	input := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(0, 1).Width(14)
	sb.WriteString(input.Render(m.wizard.dbStorageInput+"▌") + "\n\n")
	sb.WriteString(info.Render(fmt.Sprintf("Range: %d – %d GB  (step: %d GB)", minS, maxS, stepS)) + "\n\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Type a number • Enter: Continue • ← Back • Esc: Cancel"))
	return sb.String()
}

func (m Model) renderDBWizardNetworkStep(_ int) string {
	var sb strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	sel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	sb.WriteString(title.Render("Network type:") + "\n\n")
	if m.wizard.errorMsg != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}
	opts := []struct{ label, detail string }{
		{"Public network (Internet)", "Accessible from the internet"},
		{"Private network (vRack)", "Requires a configured private network"},
	}
	for i, opt := range opts {
		if i == m.wizard.dbNetworkIdx {
			sb.WriteString(sel.Render("▶ "+opt.label) + "\n")
			sb.WriteString(info.Render("    "+opt.detail) + "\n\n")
		} else {
			sb.WriteString(dim.Render("  "+opt.label) + "\n")
			sb.WriteString(info.Render("    "+opt.detail) + "\n\n")
		}
	}
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ Navigate • Enter: Select • ← Back • Esc: Cancel"))
	return sb.String()
}

func (m Model) renderDBWizardConfirmStep(_ int) string {
	var sb strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	lbl := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(20)
	val := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	sb.WriteString(title.Render("Confirm creation:") + "\n\n")
	sb.WriteString(lbl.Render("  Name:") + val.Render(m.wizard.dbName) + "\n")
	sb.WriteString(lbl.Render("  Engine:") + val.Render(strings.ToUpper(m.wizard.dbEngine)+" "+m.wizard.dbVersion) + "\n")
	sb.WriteString(lbl.Render("  Region:") + val.Render(m.wizard.dbRegion) + "\n")
	sb.WriteString(lbl.Render("  Plan:") + val.Render(strings.ToUpper(m.wizard.dbPlan)) + "\n")
	sb.WriteString(lbl.Render("  Flavor:") + val.Render(m.wizard.dbFlavor) + "\n")
	sb.WriteString(lbl.Render("  Nodes:") + val.Render(strconv.Itoa(m.wizard.dbNodes)) + "\n")
	if m.wizard.dbDiskSize > 0 {
		sb.WriteString(lbl.Render("  Storage:") + val.Render(strconv.Itoa(m.wizard.dbDiskSize)+" GB") + "\n")
	}
	networkLabel := "Public"
	if m.wizard.dbNetworkIdx == 1 {
		networkLabel = "Private (vRack)"
	}
	sb.WriteString(lbl.Render("  Network:") + val.Render(networkLabel) + "\n\n")
	if m.wizard.isLoading {
		sb.WriteString(loadingStyle.Render("⏳ Creating database service..."))
		return sb.String()
	}
	if m.wizard.errorMsg != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}
	btnCreate := lipgloss.NewStyle().
		Background(lipgloss.Color("#00FF7F")).Foreground(lipgloss.Color("#000000")).
		Bold(true).Padding(0, 2).Render(" Create ")
	btnCancel := lipgloss.NewStyle().
		Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).
		Padding(0, 2).Render(" Cancel ")
	if m.wizard.dbConfirmIdx == 1 {
		btnCreate = lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).
			Padding(0, 2).Render(" Create ")
		btnCancel = lipgloss.NewStyle().
			Background(lipgloss.Color("#FF6B6B")).Foreground(lipgloss.Color("#000000")).
			Bold(true).Padding(0, 2).Render(" Cancel ")
	}
	sb.WriteString(btnCreate + "  " + btnCancel + "\n\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("←→ Select • Enter: Confirm • Esc: Cancel"))
	return sb.String()
}

// ─── Key handlers ─────────────────────────────────────────────────────────────

func (m Model) handleDBWizardNameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "enter":
		name := strings.TrimSpace(m.wizard.dbNameInput)
		if name == "" {
			m.wizard.errorMsg = "Name cannot be empty"
			return m, nil
		}
		m.wizard.dbName = name
		m.wizard.errorMsg = ""
		m.wizard.step = DBWizardStepEngine
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Loading capabilities..."
		return m, m.fetchDBCapabilities()
	case "backspace":
		if len(m.wizard.dbNameInput) > 0 {
			m.wizard.dbNameInput = m.wizard.dbNameInput[:len(m.wizard.dbNameInput)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.wizard.dbNameInput += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handleDBWizardEngineKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.dbEngineIdx > 0 {
			m.wizard.dbEngineIdx--
		}
	case "down", "j":
		if m.wizard.dbEngineIdx < len(m.wizard.dbEngines)-1 {
			m.wizard.dbEngineIdx++
		}
	case "enter":
		if len(m.wizard.dbEngines) == 0 {
			return m, nil
		}
		e := m.wizard.dbEngines[m.wizard.dbEngineIdx]
		m.wizard.dbEngine = getStringValue(e, "name", "")
		m.wizard.dbEngineCategory = getStringValue(e, "category", "operational")
		m.wizard.dbVersionIdx = 0
		m.wizard.dbVersion = ""
		m.wizard.errorMsg = ""
		m.wizard.step = DBWizardStepVersion
	case "left":
		m.wizard.step = DBWizardStepName
	}
	return m, nil
}

func (m Model) handleDBWizardVersionKeys(key string) (tea.Model, tea.Cmd) {
	versions := m.dbFilteredVersions()
	switch key {
	case "up", "k":
		if m.wizard.dbVersionIdx > 0 {
			m.wizard.dbVersionIdx--
		}
	case "down", "j":
		if m.wizard.dbVersionIdx < len(versions)-1 {
			m.wizard.dbVersionIdx++
		}
	case "enter":
		if len(versions) == 0 {
			return m, nil
		}
		m.wizard.dbVersion = versions[m.wizard.dbVersionIdx]
		m.wizard.dbRegionIdx = 0
		m.wizard.dbRegion = ""
		m.wizard.errorMsg = ""
		m.wizard.step = DBWizardStepRegion
	case "left":
		m.wizard.step = DBWizardStepEngine
	}
	return m, nil
}

func (m Model) handleDBWizardRegionKeys(key string) (tea.Model, tea.Cmd) {
	regions := m.dbFilteredRegions()
	switch key {
	case "up", "k":
		if m.wizard.dbRegionIdx > 0 {
			m.wizard.dbRegionIdx--
		}
	case "down", "j":
		if m.wizard.dbRegionIdx < len(regions)-1 {
			m.wizard.dbRegionIdx++
		}
	case "enter":
		if len(regions) == 0 {
			return m, nil
		}
		m.wizard.dbRegion = regions[m.wizard.dbRegionIdx]
		m.wizard.dbPlanIdx = 0
		m.wizard.dbPlan = ""
		m.wizard.errorMsg = ""
		m.wizard.step = DBWizardStepPlan
	case "left":
		m.wizard.step = DBWizardStepVersion
	}
	return m, nil
}

func (m Model) handleDBWizardPlanKeys(key string) (tea.Model, tea.Cmd) {
	plans := m.dbFilteredPlans()
	switch key {
	case "up", "k":
		if m.wizard.dbPlanIdx > 0 {
			m.wizard.dbPlanIdx--
		}
	case "down", "j":
		if m.wizard.dbPlanIdx < len(plans)-1 {
			m.wizard.dbPlanIdx++
		}
	case "enter":
		if len(plans) == 0 {
			return m, nil
		}
		m.wizard.dbPlan = plans[m.wizard.dbPlanIdx]
		m.wizard.dbFlavorIdx = 0
		m.wizard.dbFlavor = ""
		m.wizard.errorMsg = ""
		m.wizard.step = DBWizardStepFlavor
	case "left":
		m.wizard.step = DBWizardStepRegion
	}
	return m, nil
}

func (m Model) handleDBWizardFlavorKeys(key string) (tea.Model, tea.Cmd) {
	flavors := m.dbFilteredFlavors()
	switch key {
	case "up", "k":
		if m.wizard.dbFlavorIdx > 0 {
			m.wizard.dbFlavorIdx--
		}
	case "down", "j":
		if m.wizard.dbFlavorIdx < len(flavors)-1 {
			m.wizard.dbFlavorIdx++
		}
	case "enter":
		if len(flavors) == 0 {
			return m, nil
		}
		m.wizard.dbFlavor = flavors[m.wizard.dbFlavorIdx]
		// Pre-fill nodes based on constraints
		minN, _ := m.dbNodesConstraints()
		m.wizard.dbNodes = minN
		m.wizard.dbNodesInput = strconv.Itoa(minN)
		// Pre-fill storage only when availability data provides valid constraints
		if avail := m.dbActiveAvail(); avail != nil {
			minS, _, _ := m.dbStorageConstraints()
			m.wizard.dbDiskSize = minS
			m.wizard.dbStorageInput = strconv.Itoa(minS)
		} else {
			m.wizard.dbDiskSize = 0
			m.wizard.dbStorageInput = ""
		}
		m.wizard.errorMsg = ""
		m.wizard.step = DBWizardStepNodes
	case "left":
		m.wizard.step = DBWizardStepPlan
	}
	return m, nil
}

func (m Model) handleDBWizardNodesKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	minN, maxN := m.dbNodesConstraints()
	switch key {
	case "enter":
		n, err := strconv.Atoi(strings.TrimSpace(m.wizard.dbNodesInput))
		if err != nil || n < minN || n > maxN {
			m.wizard.errorMsg = fmt.Sprintf("Please enter a number between %d and %d", minN, maxN)
			return m, nil
		}
		m.wizard.dbNodes = n
		m.wizard.errorMsg = ""
		m.wizard.step = DBWizardStepStorage
	case "backspace":
		if len(m.wizard.dbNodesInput) > 0 {
			m.wizard.dbNodesInput = m.wizard.dbNodesInput[:len(m.wizard.dbNodesInput)-1]
		}
	case "left":
		m.wizard.step = DBWizardStepFlavor
	default:
		if len(msg.Runes) == 1 && msg.Runes[0] >= '0' && msg.Runes[0] <= '9' {
			m.wizard.dbNodesInput += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handleDBWizardStorageKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	// If no availability data, skip this step entirely
	if m.dbActiveAvail() == nil {
		m.wizard.dbDiskSize = 0
		switch key {
		case "left":
			m.wizard.step = DBWizardStepNodes
		default:
			m.wizard.step = DBWizardStepNetwork
		}
		return m, nil
	}
	minS, maxS, _ := m.dbStorageConstraints()
	switch key {
	case "enter":
		storageStr := strings.TrimSpace(m.wizard.dbStorageInput)
		if storageStr == "" {
			// Empty = use default (omit from request)
			m.wizard.dbDiskSize = 0
			m.wizard.errorMsg = ""
			m.wizard.step = DBWizardStepNetwork
			return m, nil
		}
		s, err := strconv.Atoi(storageStr)
		if err != nil || s < minS || s > maxS {
			m.wizard.errorMsg = fmt.Sprintf("Please enter a value between %d and %d GB", minS, maxS)
			return m, nil
		}
		m.wizard.dbDiskSize = s
		m.wizard.errorMsg = ""
		m.wizard.step = DBWizardStepNetwork
	case "backspace":
		if len(m.wizard.dbStorageInput) > 0 {
			m.wizard.dbStorageInput = m.wizard.dbStorageInput[:len(m.wizard.dbStorageInput)-1]
		}
	case "left":
		m.wizard.step = DBWizardStepNodes
	default:
		if len(msg.Runes) == 1 && msg.Runes[0] >= '0' && msg.Runes[0] <= '9' {
			m.wizard.dbStorageInput += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handleDBWizardNetworkKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.dbNetworkIdx > 0 {
			m.wizard.dbNetworkIdx--
		}
	case "down", "j":
		if m.wizard.dbNetworkIdx < 1 {
			m.wizard.dbNetworkIdx++
		}
	case "enter":
		m.wizard.errorMsg = ""
		m.wizard.step = DBWizardStepConfirm
	case "left":
		m.wizard.step = DBWizardStepStorage
	}
	return m, nil
}

func (m Model) handleDBWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		m.wizard.dbConfirmIdx = 0
	case "right", "l":
		m.wizard.dbConfirmIdx = 1
	case "enter":
		if m.wizard.dbConfirmIdx == 1 {
			m.wizard.step = DBWizardStepNetwork
			return m, nil
		}
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Creating database service..."
		return m, m.createManagedDBFromWizard()
	}
	return m, nil
}
