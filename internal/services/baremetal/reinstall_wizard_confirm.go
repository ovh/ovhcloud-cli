// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package baremetal

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *reinstallWizardModel) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left":
		m.errorMsg = ""
		m.step = m.confirmOrigin

	case "ctrl+s":
		m.errorMsg = ""
		if m.saveEnabled {
			m.saveEnabled = false
			m.savePath = ""
			return m, nil
		}
		m.saveEnabled = true
		if m.savePath == "" {
			m.savePath = fmt.Sprintf("%s-reinstall.json", m.serviceName)
		}
		m.step = reinstallStepSavePath

	case "enter":
		if !strings.EqualFold(strings.TrimSpace(m.confirmText), "yes") {
			m.errorMsg = `Type "yes" to confirm the reinstallation`
			return m, nil
		}
		m.errorMsg = ""
		m.launch = true
		return m, m.leaveWizard()

	case "backspace":
		m.confirmText = trimLastRune(m.confirmText)

	case "ctrl+u":
		m.confirmText = ""

	default:
		if len(msg.Runes) > 0 {
			m.errorMsg = ""
			m.confirmText += string(msg.Runes)
		}
	}

	return m, nil
}

func (m *reinstallWizardModel) handleSavePathKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.errorMsg = ""
		m.saveEnabled = false
		m.savePath = ""
		m.step = reinstallStepConfirm
	case "enter":
		if strings.TrimSpace(m.savePath) == "" {
			m.errorMsg = "Destination path cannot be empty"
			return m, nil
		}
		m.savePath = strings.TrimSpace(m.savePath)
		m.errorMsg = ""
		m.step = reinstallStepConfirm
	case "backspace":
		m.savePath = trimLastRune(m.savePath)
	case "ctrl+u":
		m.savePath = ""
	default:
		if len(msg.Runes) > 0 {
			m.savePath += string(msg.Runes)
		}
	}

	return m, nil
}

func (m *reinstallWizardModel) writeParametersFile() error {
	body, err := json.MarshalIndent(buildReinstallPayload(m.result()), "", "    ")
	if err != nil {
		return fmt.Errorf("failed to format installation parameters: %w", err)
	}

	if err := os.WriteFile(m.savePath, body, 0o644); err != nil {
		return fmt.Errorf("failed to write installation parameters: %w", err)
	}

	m.savedPath = m.savePath

	return nil
}

// customizationSummary renders a customization answer for the confirm
// screen: masked for a SSH key, compacted for a key/value list, verbatim
// otherwise.
func (m *reinstallWizardModel) customizationSummary(name string) string {
	return answerSummary(m.inputTypeFor(name), m.customizations[name])
}

// renderConfirmOsSection is the confirm screen's "Operating system"
// section: the chosen OS, its templateInfos details when available, and
// the end-of-install warning.
func (m *reinstallWizardModel) renderConfirmOsSection() string {
	var b strings.Builder
	b.WriteString(wizardLabelStyle.Render("  Operating system:") + wizardValueStyle.Render(m.osName) + "\n")
	if info, ok := m.osInfoByName[m.osName]; ok {
		for _, column := range osPickerTableColumns {
			value, _ := info[column.key].(string)
			if value == "" {
				continue
			}
			b.WriteString(wizardLabelStyle.Render("    "+column.header+":") + wizardSubtleStyle.Render(value) + "\n")
		}
	}
	if warning := m.osEndOfInstallWarningFor(m.osName, false); warning != "" {
		b.WriteString(wizardErrorStyle.Render("    ⚠  "+warning+"  ⚠") + "\n")
	}

	return strings.TrimSuffix(b.String(), "\n")
}

// renderConfirmCustomizationsSection is the confirm screen's
// "Customizations" section: every answer, or "default values" when none
// differ from the OS's own defaults.
func (m *reinstallWizardModel) renderConfirmCustomizationsSection() string {
	var b strings.Builder

	if len(m.customizations) == 0 {
		b.WriteString(wizardLabelStyle.Render("  Customizations:") + wizardValueStyle.Render("default values") + "\n")
		return strings.TrimSuffix(b.String(), "\n")
	}

	b.WriteString(wizardLabelStyle.Render("  Customizations:") + "\n")
	names := slices.Sorted(maps.Keys(m.customizations))
	for _, name := range names {
		if m.inputTypeFor(name) == "text" {
			b.WriteString(wizardLabelStyle.Render("    "+name) + "\n")
			for _, line := range strings.Split(m.customizationSummary(name), "\n") {
				b.WriteString(wizardValueStyle.Render("      "+line) + "\n")
			}
			continue
		}
		b.WriteString(wizardLabelStyle.Render("    "+name) +
			wizardValueStyle.Render(m.customizationSummary(name)) + "\n")
	}

	return strings.TrimSuffix(b.String(), "\n")
}

// renderConfirmStorageSection is the confirm screen's "Storage" section:
// the disk group, disk count, and the partitioning that will be applied.
func (m *reinstallWizardModel) renderConfirmStorageSection() string {
	var b strings.Builder
	b.WriteString(wizardLabelStyle.Render("  Storage:") + "\n")

	if group := m.selectedDiskGroup(); group != nil {
		b.WriteString(wizardLabelStyle.Render("    Disk group id:") +
			wizardValueStyle.Render(fmt.Sprintf("%d", group.DiskGroupID)) + "\n")
	}
	b.WriteString(wizardLabelStyle.Render("    Disks used:") +
		wizardValueStyle.Render(fmt.Sprintf("%d", m.diskCount)) + "\n")

	for _, group := range m.otherDiskGroups() {
		status := "erased"
		style := wizardErrorStyle
		if !m.eraseOtherGroups[group.DiskGroupID] {
			status, style = "data kept", wizardValueStyle
		}
		b.WriteString(wizardLabelStyle.Render(fmt.Sprintf("    Disk group %d:", group.DiskGroupID)) +
			style.Render(status) + "\n")
	}

	switch {
	case len(m.layout) > 0:
		b.WriteString(wizardLabelStyle.Render("    Partitioning:") + wizardValueStyle.Render("custom layout") + "\n")
		width := layoutPartitionsBoxWidth(m.layout)
		for _, partition := range m.layout {
			for _, line := range strings.Split(describeLayoutPartition(partition, width, m.rootMountpoint()), "\n") {
				b.WriteString(wizardSubtleStyle.Render("      "+line) + "\n")
			}
		}
	case m.schemeName != "":
		b.WriteString(wizardLabelStyle.Render("    Partitioning scheme:") + wizardValueStyle.Render(m.schemeName) + "\n")
		if scheme := findSchemeByName(m.schemes, m.schemeName); scheme != nil {
			width := schemePartitionsBoxWidth(scheme.Partitions)
			for _, partition := range scheme.Partitions {
				for _, line := range strings.Split(describeSchemePartition(partition, width, m.rootMountpoint()), "\n") {
					b.WriteString(wizardSubtleStyle.Render("      "+line) + "\n")
				}
			}
		}
	default:
		b.WriteString(wizardLabelStyle.Render("    Partitioning:") + wizardValueStyle.Render("default") + "\n")
	}

	return strings.TrimSuffix(b.String(), "\n")
}

func (m *reinstallWizardModel) renderConfirmStep() string {
	var content strings.Builder
	content.WriteString(wizardTitleStyle.Render("Confirm the reinstallation of "+m.serviceName+":") + "\n\n")

	sections := lipgloss.JoinVertical(lipgloss.Left,
		m.renderConfirmOsSection(), "",
		m.renderConfirmCustomizationsSection(), "",
		m.renderConfirmStorageSection(),
	)
	content.WriteString(wizardConfirmBoxStyle.Render(sections) + "\n\n")

	saveLabel := "○ Also save these parameters to a file"
	if m.saveEnabled {
		saveLabel = fmt.Sprintf("● Also save these parameters to %s", m.savePath)
	}
	content.WriteString(wizardSelectedStyle.Render(saveLabel) + "\n\n")

	if group := m.selectedDiskGroup(); group != nil {
		if warning := hardwareRaidWarning(*group); warning != "" {
			content.WriteString(wizardErrorStyle.Render("⚠  "+warning+"  ⚠") + "\n\n")
		}

		if labels := m.erasedDiskGroupLabels(); len(labels) > 1 {
			content.WriteString(wizardErrorStyle.Render("⚠  All data on the following disk groups will be erased ⚠") + "\n")
			for _, label := range labels {
				content.WriteString(wizardErrorStyle.Render("- "+label) + "\n")
			}
		} else {
			content.WriteString(wizardErrorStyle.Render(fmt.Sprintf(
				"⚠  All data on %s of server %s will be erased  ⚠", labels[0], m.serviceName)) + "\n")
		}
		content.WriteString(wizardErrorStyle.Render("Are you sure you want to continue?") + "\n\n")
	}

	content.WriteString(lipgloss.JoinHorizontal(lipgloss.Center,
		wizardDescStyle.Render("  Type \"yes\" to confirm:  "),
		wizardConfirmInputStyle.Render(m.confirmText+"▌")) + "\n\n")

	content.WriteString(m.renderError())
	content.WriteString(wizardHintStyle.Render(
		"Type \"yes\" then Enter: Reinstall • Ctrl+S: Toggle saving to a file • ←: Back • Esc: Quit without reinstalling"))

	return content.String()
}

func (m *reinstallWizardModel) renderSavePathStep() string {
	var content strings.Builder
	content.WriteString(wizardTitleStyle.Render("Save the installation parameters to:") + "\n")
	content.WriteString(wizardDescStyle.Render("The file can be replayed with: ovhcloud baremetal reinstall "+
		m.serviceName+" --from-file <path>") + "\n\n")

	content.WriteString(wizardInputStyle.Render(m.savePath+"▌") + "\n\n")

	content.WriteString(m.renderError())
	content.WriteString(wizardHintStyle.Render("Type a path • Enter: Confirm • Esc: Do not save"))

	return content.String()
}
