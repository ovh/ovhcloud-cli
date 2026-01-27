// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovh/ovhcloud-cli/internal/config"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
)

// ViewMode represents the current view mode
type ViewMode int

const (
	TableView ViewMode = iota
	DetailView
	LoadingView
	ErrorView
)

// ProductType represents a product category
type ProductType int

const (
	ProductInstances ProductType = iota
	ProductKubernetes
	ProductDatabases
	ProductStorage
	ProductNetworks
	ProductProjects
)

// Model represents the TUI application state
type Model struct {
	width           int
	height          int
	mode            ViewMode
	currentProduct  ProductType
	navIdx          int // Index in navigation bar
	table           table.Model
	detailData      map[string]interface{}
	currentData     []map[string]interface{}
	errorMsg        string
	cloudProject    string
	currentItemName string // Name of the currently viewed item
}

// Navigation items for the top bar
type NavItem struct {
	Label   string
	Icon    string
	Product ProductType
	Path    string
}

// Styles
var (
	// Header / Logo
	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7B68EE"))

	// Navigation bar
	navBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1)

	navItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(0, 2)

	navItemSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF7F")).
				Bold(true).
				Padding(0, 2).
				Background(lipgloss.Color("#2a2a2a"))

	// Content area
	contentBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(1, 2)

	// Title for current product
	productTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#7B68EE")).
				Padding(0, 2)

	// Detail view boxes
	boxTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7B68EE"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Width(18)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	statusRunningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF7F"))

	statusStoppedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6B6B"))

	// Footer
	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Padding(0, 1)

	// Error and loading
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Padding(1, 2)

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7B68EE")).
			Padding(1, 2)
)

// Messages for async operations
type projectsLoadedMsg struct {
	projects []map[string]interface{}
	err      error
}

type instancesLoadedMsg struct {
	instances []map[string]interface{}
	err       error
}

type dataLoadedMsg struct {
	data []map[string]interface{}
	err  error
}

// Navigation items
func getNavItems() []NavItem {
	return []NavItem{
		{Label: "Instances", Icon: "💻", Product: ProductInstances, Path: "/instances"},
		{Label: "Kubernetes", Icon: "☸️", Product: ProductKubernetes, Path: "/kubernetes"},
		{Label: "Databases", Icon: "🗄️", Product: ProductDatabases, Path: "/databases"},
		{Label: "Storage", Icon: "💾", Product: ProductStorage, Path: "/storage/s3"},
		{Label: "Networks", Icon: "🌐", Product: ProductNetworks, Path: "/networks/private"},
		{Label: "Projects", Icon: "📦", Product: ProductProjects, Path: "/projects"},
	}
}

// StartBrowser is the entry point for the browser TUI
func StartBrowser(cmd *cobra.Command, args []string) {
	httpLib.InitClient()

	cloudProject, _ := getDefaultCloudProject()

	initialModel := Model{
		mode:           LoadingView,
		currentProduct: ProductInstances,
		navIdx:         0,
		cloudProject:   cloudProject,
	}

	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Load the default product (instances)
	return m.fetchDataForPath("/instances")
}

// Update handles all messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case projectsLoadedMsg:
		return m.handleProjectsLoaded(msg)

	case instancesLoadedMsg:
		return m.handleInstancesLoaded(msg)

	case dataLoadedMsg:
		return m.handleDataLoaded(msg)
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	var content strings.Builder

	// Calculate available width
	width := m.width
	if width < 80 {
		width = 80
	}

	// Header with logo
	content.WriteString(m.renderHeader())
	content.WriteString("\n")

	// Navigation bar
	content.WriteString(m.renderNavBar(width))
	content.WriteString("\n\n")

	// Content box with current product
	content.WriteString(m.renderContentBox(width))
	content.WriteString("\n\n")

	// Footer
	content.WriteString(m.renderFooter())

	return content.String()
}

func (m Model) renderHeader() string {
	logo := logoStyle.Render("☁ OVHcloud Manager")
	return logo
}

func (m Model) renderNavBar(width int) string {
	navItems := getNavItems()
	var items []string

	for i, nav := range navItems {
		var style lipgloss.Style
		if i == m.navIdx {
			style = navItemSelectedStyle
		} else {
			style = navItemStyle
		}
		items = append(items, style.Render(fmt.Sprintf("%s %s", nav.Icon, nav.Label)))
	}

	navContent := lipgloss.JoinHorizontal(lipgloss.Top, items...)
	return navBarStyle.Width(width - 2).Render(navContent)
}

func (m Model) renderContentBox(width int) string {
	navItems := getNavItems()
	currentNav := navItems[m.navIdx]

	// Product title - show item name in detail view
	var titleText string
	if m.mode == DetailView && m.currentItemName != "" {
		titleText = fmt.Sprintf(" %s %s > %s ", currentNav.Icon, currentNav.Label, m.currentItemName)
	} else {
		titleText = fmt.Sprintf(" %s %s ", currentNav.Icon, currentNav.Label)
	}
	title := productTitleStyle.Render(titleText)

	// Content based on mode
	var contentStr string
	switch m.mode {
	case LoadingView:
		contentStr = loadingStyle.Render("⏳ Loading data...")
	case ErrorView:
		contentStr = errorStyle.Render("❌ Error: " + m.errorMsg)
	case TableView:
		contentStr = m.renderTable()
	case DetailView:
		contentStr = m.renderDetailView(width - 6)
	}

	// Combine title and content
	fullContent := title + "\n\n" + contentStr

	return contentBoxStyle.Width(width - 4).Render(fullContent)
}

func (m Model) renderTable() string {
	if m.table.Rows() == nil || len(m.table.Rows()) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("No data available")
	}
	return m.table.View()
}

func (m Model) renderDetailView(width int) string {
	if m.detailData == nil {
		return "No data"
	}

	// Determine what type of resource we're viewing
	switch m.currentProduct {
	case ProductInstances:
		return m.renderInstanceDetail(width)
	case ProductKubernetes:
		return m.renderKubernetesDetail(width)
	case ProductProjects:
		return m.renderProjectDetail(width)
	default:
		return m.renderGenericDetail(width)
	}
}

func (m Model) renderInstanceDetail(width int) string {
	var content strings.Builder

	// Get values safely
	_ = getStringValue(m.detailData, "name", "Unknown") // name is shown in title bar
	status := getStringValue(m.detailData, "status", "Unknown")
	id := getStringValue(m.detailData, "id", "N/A")
	region := getStringValue(m.detailData, "region", "N/A")
	flavorName := getStringValue(m.detailData, "flavorId", "N/A")
	imageName := getStringValue(m.detailData, "imageId", "N/A")
	created := getStringValue(m.detailData, "created", "N/A")

	// Get IP addresses
	ipv4 := "N/A"
	ipv6 := "N/A"
	if addresses, ok := m.detailData["ipAddresses"].([]interface{}); ok {
		for _, addr := range addresses {
			if addrMap, ok := addr.(map[string]interface{}); ok {
				ip := getStringValue(addrMap, "ip", "")
				version := int(getFloatValue(addrMap, "version", 0))
				if version == 4 && ipv4 == "N/A" {
					ipv4 = ip
				} else if version == 6 && ipv6 == "N/A" {
					ipv6 = ip
				}
			}
		}
	}

	// Status indicator
	statusIcon := "🟢"
	statusStyle := statusRunningStyle
	if strings.ToLower(status) != "active" && strings.ToLower(status) != "running" {
		statusIcon = "🔴"
		statusStyle = statusStoppedStyle
	}

	// Build the detail view with boxes
	boxWidth := (width - 6) / 2
	if boxWidth < 35 {
		boxWidth = 35
	}

	// Left column - Information box
	infoContent := strings.Builder{}
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Status"), statusStyle.Render(statusIcon+" "+status)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("ID"), valueStyle.Render(truncate(id, 30))))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Region"), valueStyle.Render(region)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Flavor"), valueStyle.Render(truncate(flavorName, 25))))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Image"), valueStyle.Render(truncate(imageName, 25))))
	infoContent.WriteString(fmt.Sprintf("%s %s", labelStyle.Render("Created"), valueStyle.Render(truncate(created, 25))))

	infoBox := renderBox("Informations", infoContent.String(), boxWidth)

	// Right column - Network box
	networkContent := strings.Builder{}
	networkContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("IPv4"), valueStyle.Render(ipv4)))
	networkContent.WriteString(fmt.Sprintf("%s %s", labelStyle.Render("IPv6"), valueStyle.Render(truncate(ipv6, 25))))

	networkBox := renderBox("Réseau", networkContent.String(), boxWidth)

	// Actions box (top)
	actionsContent := "[Reboot] [Rescue Mode] [Stop] [Console] [Reinstall]"
	actionsBox := renderBox("Actions rapides", actionsContent, width-4)

	// Combine everything
	content.WriteString(actionsBox)
	content.WriteString("\n\n")

	// Side by side boxes
	leftRight := lipgloss.JoinHorizontal(lipgloss.Top, infoBox, "  ", networkBox)
	content.WriteString(leftRight)

	return content.String()
}

func (m Model) renderKubernetesDetail(width int) string {
	var content strings.Builder

	clusterName := getStringValue(m.detailData, "name", "Unknown")
	status := getStringValue(m.detailData, "status", "Unknown")
	id := getStringValue(m.detailData, "id", "N/A")
	region := getStringValue(m.detailData, "region", "N/A")
	version := getStringValue(m.detailData, "version", "N/A")
	nodesCount := getFloatValue(m.detailData, "nodesCount", 0)
	updatePolicy := getStringValue(m.detailData, "updatePolicy", "N/A")

	statusIcon := "🟢"
	statusStyle := statusRunningStyle
	if strings.ToLower(status) != "ready" && strings.ToLower(status) != "running" {
		statusIcon = "🟡"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	}

	boxWidth := (width - 6) / 2
	if boxWidth < 35 {
		boxWidth = 35
	}

	// Cluster info
	infoContent := strings.Builder{}
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Status"), statusStyle.Render(statusIcon+" "+status)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("ID"), valueStyle.Render(truncate(id, 30))))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Region"), valueStyle.Render(region)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Version"), valueStyle.Render(version)))
	infoContent.WriteString(fmt.Sprintf("%s %s", labelStyle.Render("Nodes"), valueStyle.Render(fmt.Sprintf("%.0f", nodesCount))))

	infoBox := renderBox("Cluster "+clusterName, infoContent.String(), boxWidth)

	// Configuration
	configContent := strings.Builder{}
	configContent.WriteString(fmt.Sprintf("%s %s", labelStyle.Render("Update Policy"), valueStyle.Render(updatePolicy)))

	configBox := renderBox("Configuration", configContent.String(), boxWidth)

	// Actions
	actionsContent := "[kubectl config] [Scale] [Upgrade] [Reset Kubeconfig]"
	actionsBox := renderBox("Actions", actionsContent, width-4)

	content.WriteString(actionsBox)
	content.WriteString("\n\n")
	leftRight := lipgloss.JoinHorizontal(lipgloss.Top, infoBox, "  ", configBox)
	content.WriteString(leftRight)

	return content.String()
}

func (m Model) renderProjectDetail(width int) string {
	var content strings.Builder

	name := getStringValue(m.detailData, "description", "Unknown Project")
	projectID := getStringValue(m.detailData, "project_id", "N/A")
	status := getStringValue(m.detailData, "status", "N/A")
	createdAt := getStringValue(m.detailData, "creationDate", "N/A")

	statusIcon := "🟢"
	statusStyle := statusRunningStyle
	if strings.ToLower(status) != "ok" {
		statusIcon = "🟡"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	}

	boxWidth := (width - 6) / 2
	if boxWidth < 35 {
		boxWidth = 35
	}

	// Project info
	infoContent := strings.Builder{}
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Status"), statusStyle.Render(statusIcon+" "+status)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Project ID"), valueStyle.Render(truncate(projectID, 30))))
	infoContent.WriteString(fmt.Sprintf("%s %s", labelStyle.Render("Created"), valueStyle.Render(truncate(createdAt, 25))))

	infoBox := renderBox("Project: "+name, infoContent.String(), boxWidth)

	// Actions
	actionsContent := "[Select as Default] [View Resources] [Settings]"
	actionsBox := renderBox("Actions", actionsContent, boxWidth)

	leftRight := lipgloss.JoinHorizontal(lipgloss.Top, infoBox, "  ", actionsBox)
	content.WriteString(leftRight)

	return content.String()
}

func (m Model) renderGenericDetail(width int) string {
	var content strings.Builder

	boxWidth := width - 4

	// Sort keys for consistent display
	keys := make([]string, 0, len(m.detailData))
	for k := range m.detailData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	detailContent := strings.Builder{}
	for _, key := range keys {
		value := m.detailData[key]
		valueStr := fmt.Sprintf("%v", value)
		if len(valueStr) > 50 {
			valueStr = valueStr[:47] + "..."
		}
		detailContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render(key), valueStyle.Render(valueStr)))
	}

	detailBox := renderBox("Details", detailContent.String(), boxWidth)
	content.WriteString(detailBox)

	return content.String()
}

func renderBox(title string, content string, width int) string {
	titleRendered := boxTitleStyle.Render("─ " + title + " ")
	titleLen := lipgloss.Width(titleRendered)

	// Build top border with title
	topBorder := "┌" + titleRendered
	remainingWidth := width - titleLen - 2
	if remainingWidth > 0 {
		topBorder += strings.Repeat("─", remainingWidth) + "┐"
	} else {
		topBorder += "┐"
	}

	// Content lines with side borders
	lines := strings.Split(content, "\n")
	var contentLines []string
	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		padding := width - 4 - lineWidth
		if padding < 0 {
			padding = 0
		}
		contentLines = append(contentLines, "│ "+line+strings.Repeat(" ", padding)+" │")
	}

	// Bottom border
	bottomBorder := "└" + strings.Repeat("─", width-2) + "┘"

	return topBorder + "\n" + strings.Join(contentLines, "\n") + "\n" + bottomBorder
}

func (m Model) renderFooter() string {
	var help string
	switch m.mode {
	case TableView:
		help = "←→: Switch Product • ↑↓: Navigate Table • Enter: View Details • q: Quit"
	case DetailView:
		help = "←→: Switch Product • Esc: Back to List • q: Quit"
	default:
		help = "←→: Switch Product • Enter: Select • q: Quit"
	}
	return footerStyle.Render(help)
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "left":
		// Navigate to previous product
		if m.navIdx > 0 {
			m.navIdx--
			return m.loadCurrentProduct()
		}
		return m, nil

	case "right":
		// Navigate to next product
		navItems := getNavItems()
		if m.navIdx < len(navItems)-1 {
			m.navIdx++
			return m.loadCurrentProduct()
		}
		return m, nil

	case "esc":
		// Go back to table view from detail view
		if m.mode == DetailView {
			m.mode = TableView
		}
		return m, nil

	case "enter":
		// In table view, show details
		if m.mode == TableView {
			selectedRow := m.table.Cursor()
			if selectedRow >= 0 && selectedRow < len(m.currentData) {
				m.detailData = m.currentData[selectedRow]
				m.currentItemName = getStringValue(m.detailData, "name", "Item")
				m.mode = DetailView
			}
		}
		return m, nil

	case "up", "down", "j", "k":
		// Table navigation
		if m.mode == TableView {
			var cmd tea.Cmd
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	return m, nil
}

func (m Model) loadCurrentProduct() (Model, tea.Cmd) {
	navItems := getNavItems()
	currentNav := navItems[m.navIdx]
	m.currentProduct = currentNav.Product
	m.mode = LoadingView
	m.detailData = nil
	m.currentData = nil
	return m, m.fetchDataForPath(currentNav.Path)
}

// Helper functions
func getStringValue(data map[string]interface{}, key string, defaultVal string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", val)
	}
	return defaultVal
}

func getFloatValue(data map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := data[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return defaultVal
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func getDefaultCloudProject() (string, error) {
	projectID, err := config.GetConfigValue(flags.CliConfig, "", "default_cloud_project")
	if err != nil || projectID == "" {
		return "", err
	}
	return projectID, nil
}
