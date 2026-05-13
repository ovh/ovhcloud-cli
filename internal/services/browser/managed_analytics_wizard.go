// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
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

// fetchAnalyticsCapabilities fetches capabilities and availability for managed analytics services.
// It uses the same general availability endpoint as the databases wizard, then filters
// the engines list to only those with category "analysis".
func (m Model) fetchAnalyticsCapabilities() tea.Cmd {
	return func() tea.Msg {
		capEndpoint := fmt.Sprintf("/v1/cloud/project/%s/database/capabilities", m.cloudProject)
		availEndpoint := fmt.Sprintf("/v1/cloud/project/%s/database/availability", m.cloudProject)

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

		// Only keep analytics engines (category == "analysis")
		engines, _ := caps["engines"].([]interface{})
		var engMaps []map[string]interface{}
		for _, e := range engines {
			if em, ok := e.(map[string]interface{}); ok {
				if strings.ToLower(getStringValue(em, "category", "")) == "analysis" {
					engMaps = append(engMaps, em)
				}
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

// fetchAnalyticsEngineAvail fetches availability data for a specific analytics engine.
// The general /database/availability endpoint does not return analytics engines,
// so we must call /database/{engine}/availability per engine after selection.
func (m Model) fetchAnalyticsEngineAvail(engineName string) tea.Cmd {
	return func() tea.Msg {
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/database/%s/availability",
			m.cloudProject, url.PathEscape(engineName))
		var items []map[string]interface{}
		if err := httpLib.Client.Get(endpoint, &items); err != nil {
			return analyticsEngineAvailLoadedMsg{engine: engineName, err: err}
		}
		// Ensure each item carries the engine name so downstream filters work.
		for i := range items {
			if getStringValue(items[i], "engine", "") == "" {
				items[i]["engine"] = engineName
			}
		}
		return analyticsEngineAvailLoadedMsg{engine: engineName, availItems: items}
	}
}

// createAnalyticsFromWizard calls the OVHcloud API to create the managed analytics service.
func (m Model) createAnalyticsFromWizard() tea.Cmd {
	return func() tea.Msg {
		engine := m.wizard.dbEngine
		if engine == "" {
			return analyticsCreatedMsg{err: fmt.Errorf("no engine selected")}
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
			return analyticsCreatedMsg{err: fmt.Errorf("failed to create analytics service: %w", err)}
		}
		name := m.wizard.dbName
		if name == "" {
			name = getStringValue(result, "id", "analytics")
		}
		return analyticsCreatedMsg{name: name}
	}
}

// ─── Render functions ─────────────────────────────────────────────────────────

func (m Model) renderAnalyticsWizardNameStep(_ int) string {
	var sb strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	sb.WriteString(title.Render("Service name:") + "\n\n")
	sb.WriteString(desc.Render("Choose a unique name to identify your managed analytics service.") + "\n\n")
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

func (m Model) renderAnalyticsWizardEngineStep(_ int) string {
	var sb strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	sel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	sub := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	sb.WriteString(title.Render("Select an analytics engine:") + "\n\n")
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

func (m Model) renderAnalyticsWizardVersionStep(_ int) string {
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

func (m Model) renderAnalyticsWizardRegionStep(_ int) string {
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

func (m Model) renderAnalyticsWizardPlanStep(_ int) string {
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

func (m Model) renderAnalyticsWizardFlavorStep(_ int) string {
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

func (m Model) renderAnalyticsWizardNodesStep(_ int) string {
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

func (m Model) renderAnalyticsWizardStorageStep(_ int) string {
	var sb strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	sb.WriteString(title.Render("Storage size (GB):") + "\n\n")
	sb.WriteString(desc.Render(fmt.Sprintf(
		"Flavor: %s  •  Plan: %s",
		m.wizard.dbFlavor, strings.ToUpper(m.wizard.dbPlan))) + "\n\n")
	// Storage is fixed for analytics — show the pre-selected value
	val := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F"))
	sb.WriteString(info.Render("Storage is fixed for this configuration:") + " " +
		val.Render(fmt.Sprintf("%d GB", m.wizard.dbDiskSize)) + "\n\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Enter: Continue • ← Back • Esc: Cancel"))
	return sb.String()
}

func (m Model) renderAnalyticsWizardNetworkStep(_ int) string {
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

func (m Model) renderAnalyticsWizardConfirmStep(_ int) string {
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
		sb.WriteString(loadingStyle.Render("⏳ Creating analytics service..."))
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

func (m Model) handleAnalyticsWizardNameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		m.wizard.step = AnalyticsWizardStepEngine
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Loading capabilities..."
		return m, m.fetchAnalyticsCapabilities()
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

func (m Model) handleAnalyticsWizardEngineKeys(key string) (tea.Model, tea.Cmd) {
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
		m.wizard.dbEngineCategory = getStringValue(e, "category", "analysis")
		m.wizard.dbVersionIdx = 0
		m.wizard.dbVersion = ""
		m.wizard.errorMsg = ""
		m.wizard.step = AnalyticsWizardStepVersion
	case "left":
		m.wizard.step = AnalyticsWizardStepName
	}
	return m, nil
}

func (m Model) handleAnalyticsWizardVersionKeys(key string) (tea.Model, tea.Cmd) {
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
		m.wizard.step = AnalyticsWizardStepRegion
	case "left":
		m.wizard.step = AnalyticsWizardStepEngine
	}
	return m, nil
}

func (m Model) handleAnalyticsWizardRegionKeys(key string) (tea.Model, tea.Cmd) {
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
		m.wizard.step = AnalyticsWizardStepPlan
	case "left":
		m.wizard.step = AnalyticsWizardStepVersion
	}
	return m, nil
}

func (m Model) handleAnalyticsWizardPlanKeys(key string) (tea.Model, tea.Cmd) {
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
		m.wizard.step = AnalyticsWizardStepFlavor
	case "left":
		m.wizard.step = AnalyticsWizardStepRegion
	}
	return m, nil
}

func (m Model) handleAnalyticsWizardFlavorKeys(key string) (tea.Model, tea.Cmd) {
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
		// Storage is fixed for analytics — read the max from specifications directly.
		// specifications.storage.maximum.value is the actual required disk size (e.g. 30 GB).
		// We do NOT use min or step since deprecated fields can be stale/wrong.
		diskSize := 0
		if avail := m.dbActiveAvail(); avail != nil {
			if specs, ok := avail["specifications"].(map[string]interface{}); ok {
				if storage, ok := specs["storage"].(map[string]interface{}); ok {
					if maxS, ok := storage["maximum"].(map[string]interface{}); ok {
						if v, ok := toFloat64(maxS["value"]); ok && v > 0 {
							diskSize = int(v)
						}
					}
					// If maximum is absent or zero, fall back to minimum
					if diskSize == 0 {
						if minS, ok := storage["minimum"].(map[string]interface{}); ok {
							if v, ok := toFloat64(minS["value"]); ok && v > 0 {
								diskSize = int(v)
							}
						}
					}
				}
			}
		}
		m.wizard.dbDiskSize = diskSize
		m.wizard.dbStorageInput = strconv.Itoa(diskSize)
		m.wizard.errorMsg = ""
		m.wizard.step = AnalyticsWizardStepNodes
	case "left":
		m.wizard.step = AnalyticsWizardStepPlan
	}
	return m, nil
}

func (m Model) handleAnalyticsWizardNodesKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		m.wizard.step = AnalyticsWizardStepStorage
	case "backspace":
		if len(m.wizard.dbNodesInput) > 0 {
			m.wizard.dbNodesInput = m.wizard.dbNodesInput[:len(m.wizard.dbNodesInput)-1]
		}
	case "left":
		m.wizard.step = AnalyticsWizardStepFlavor
	default:
		if len(msg.Runes) == 1 && msg.Runes[0] >= '0' && msg.Runes[0] <= '9' {
			m.wizard.dbNodesInput += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handleAnalyticsWizardStorageKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Storage is fixed for analytics — just navigate forward or back
	switch msg.String() {
	case "left":
		m.wizard.step = AnalyticsWizardStepNodes
	default:
		m.wizard.step = AnalyticsWizardStepNetwork
	}
	return m, nil
}

func (m Model) handleAnalyticsWizardNetworkKeys(key string) (tea.Model, tea.Cmd) {
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
		m.wizard.step = AnalyticsWizardStepConfirm
	case "left":
		m.wizard.step = AnalyticsWizardStepStorage
	}
	return m, nil
}

func (m Model) handleAnalyticsWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		m.wizard.dbConfirmIdx = 0
	case "right", "l":
		m.wizard.dbConfirmIdx = 1
	case "enter":
		if m.wizard.dbConfirmIdx == 1 {
			m.wizard.step = AnalyticsWizardStepNetwork
			return m, nil
		}
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Creating analytics service..."
		return m, m.createAnalyticsFromWizard()
	}
	return m, nil
}
