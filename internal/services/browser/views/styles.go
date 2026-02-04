// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package views

import "github.com/charmbracelet/lipgloss"

// Color palette - OVHcloud brand colors and common UI colors
var (
	// Brand colors
	ColorPrimary   = lipgloss.Color("#7B68EE") // OVHcloud purple
	ColorSecondary = lipgloss.Color("#00FF7F") // Green accent
	ColorDanger    = lipgloss.Color("#FF6B6B") // Red for errors/warnings
	ColorWarning   = lipgloss.Color("#FFD700") // Yellow for warnings
	ColorMuted     = lipgloss.Color("#888888") // Gray for secondary text
	ColorDimmed    = lipgloss.Color("#666666") // Darker gray
	ColorWhite     = lipgloss.Color("#FFFFFF")
	ColorBlack     = lipgloss.Color("#000000")

	// Background colors
	ColorBgDark    = lipgloss.Color("#1a1a1a")
	ColorBgMedium  = lipgloss.Color("#2a2a2a")
	ColorBgBorder  = lipgloss.Color("#444444")
	ColorBgBorder2 = lipgloss.Color("#240")
)

// Shared styles for consistent UI rendering
var (
	// Header / Logo
	StyleLogo = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	// Navigation bar
	StyleNavBar = lipgloss.NewStyle().
			Background(ColorBgDark).
			Padding(0, 1)

	StyleNavItem = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 2)

	StyleNavItemSelected = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				Bold(true).
				Padding(0, 2).
				Background(ColorBgMedium)

	// Content area
	StyleContentBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorBgBorder).
			Padding(1, 2)

	// Title for current product
	StyleProductTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWhite).
				Background(ColorPrimary).
				Padding(0, 2)

	// Detail view boxes
	StyleBoxTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	StyleLabel = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Width(18)

	StyleValue = lipgloss.NewStyle().
			Foreground(ColorWhite)

	StyleStatusRunning = lipgloss.NewStyle().
				Foreground(ColorSecondary)

	StyleStatusStopped = lipgloss.NewStyle().
				Foreground(ColorDanger)

	StyleStatusWarning = lipgloss.NewStyle().
				Foreground(ColorWarning)

	// Footer
	StyleFooter = lipgloss.NewStyle().
			Foreground(ColorDimmed).
			Padding(0, 1)

	// Error and loading
	StyleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Padding(1, 2)

	StyleLoading = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Padding(1, 2)

	// Buttons and actions
	StyleButtonSelected = lipgloss.NewStyle().
				Background(ColorPrimary).
				Foreground(ColorWhite).
				Bold(true).
				Padding(0, 1)

	StyleButton = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 1)

	StyleButtonDanger = lipgloss.NewStyle().
				Foreground(ColorDanger).
				Padding(0, 1)

	StyleButtonSuccess = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				Padding(0, 1)

	// Input fields
	StyleInput = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	StyleInputLabel = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Help text
	StyleHelp = lipgloss.NewStyle().
			Foreground(ColorDimmed)

	// Headers and titles
	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	StyleSubheader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite)

	// Table styles
	StyleTableHeader = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(ColorBgBorder2).
				BorderBottom(true).
				Bold(true)

	StyleTableSelected = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("57")).
				Bold(false)

	// Notification styles
	StyleNotificationSuccess = lipgloss.NewStyle().
					Foreground(ColorSecondary).
					Bold(true)

	StyleNotificationError = lipgloss.NewStyle().
				Foreground(ColorDanger).
				Bold(true)

	// Additional status styles
	StyleStatusReady = lipgloss.NewStyle().
				Foreground(ColorSecondary)

	StyleStatusError = lipgloss.NewStyle().
				Foreground(ColorDanger)

	// Filter style
	StyleFilter = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Background(ColorBgMedium)

	// Subtle text
	StyleSubtle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Highlight style
	StyleHighlight = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	// Info style
	StyleInfo = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)
)

// RenderBox creates a bordered box with a title.
func RenderBox(title, content string, width int) string {
	titleStr := StyleBoxTitle.Render("▸ " + title)
	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(ColorBgBorder).
		Padding(0, 1).
		Width(width)
	return titleStr + "\n" + boxStyle.Render(content)
}

// RenderKeyValue renders a label-value pair.
func RenderKeyValue(label, value string) string {
	return StyleLabel.Render(label+":") + " " + StyleValue.Render(value)
}

// RenderStatus renders a status with appropriate coloring.
func RenderStatus(status string) string {
	statusLower := lipgloss.NewStyle().Foreground(ColorWhite)
	switch status {
	case "ACTIVE", "RUNNING", "READY", "HEALTHY":
		statusLower = StyleStatusRunning
	case "SHUTOFF", "STOPPED", "ERROR", "FAILED", "UNHEALTHY":
		statusLower = StyleStatusStopped
	case "BUILD", "BUILDING", "PENDING", "INSTALLING", "UPDATING":
		statusLower = StyleStatusWarning
	}
	return statusLower.Render(status)
}
