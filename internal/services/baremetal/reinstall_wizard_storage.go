// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package baremetal

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type layoutField int

const (
	layoutFieldFileSystem layoutField = iota
	layoutFieldMountPoint
	layoutFieldSize
	layoutFieldRaidLevel
	layoutFieldLvName
	layoutFieldZpName
)

var layoutFieldLabels = map[layoutField]string{
	layoutFieldFileSystem: "File system",
	layoutFieldMountPoint: "Mount point",
	layoutFieldSize:       "Size in MiB (0 = rest)",
	layoutFieldRaidLevel:  "RAID level",
	layoutFieldLvName:     "LVM volume name",
	layoutFieldZpName:     "ZFS pool name",
}

func (m *reinstallWizardModel) loadStorage() tea.Cmd {
	m.isLoading = true
	m.loadingMessage = "Loading storage configuration…"

	osName, serviceName := m.osName, m.serviceName

	return func() tea.Msg {
		schemes, err := fetchPartitionSchemes(osName)
		if err != nil {
			return storageLoadedMsg{err: err}
		}

		hardware, err := fetchHardwareSpecifications(serviceName)
		if err != nil {
			return storageLoadedMsg{err: err}
		}

		return storageLoadedMsg{schemes: schemes, hardware: hardware}
	}
}

// partitionModeChoices are the two ways the disks can be partitioned, in the
// order the screen offers them.
var partitionModeChoices = []string{
	"Use one of the default partitioning schemes",
	"Define a custom partitioning layout",
}

func (m *reinstallWizardModel) initDiskGroupSelection() {
	m.diskGroupIdx = 0
	if m.hardware != nil {
		m.diskGroupIdx = max(slices.IndexFunc(m.hardware.DiskGroups, func(group diskGroup) bool {
			return group.DiskGroupID == m.hardware.DefaultDiskGroupID
		}), 0)
	}

	m.diskCount = m.maxDiskCount()
}

// initSchemeSelection fills the scheme name list: just names, since the
// selected scheme's own partitions are shown separately, in the scrollable
// schemeDetailViewport (see resizeSchemeViews / updateSchemeDetailContent).
func (m *reinstallWizardModel) initSchemeSelection() {
	items := make([]pickerItem, 0, len(m.schemes))
	for i, scheme := range m.schemes {
		items = append(items, pickerItem{
			label: fmt.Sprintf("%s (%d partitions)", scheme.Name, len(scheme.Partitions)),
			index: i,
		})
	}

	fillPicker(&m.schemeList, items, 0)
	m.resizeSchemeViews()
	m.updateSchemeDetailContent()
}

// resizeSchemeViews splits the space budgeted for a list screen between the
// scheme name list and the detail viewport underneath it.
func (m *reinstallWizardModel) resizeSchemeViews() {
	listHeight := schemeListHeight(len(m.schemes))
	m.schemeList.SetSize(m.listWidth, listHeight)

	m.schemeDetailViewport.Width = m.listWidth
	m.schemeDetailViewport.Height = max(m.listHeight-listHeight-2, 3)
}

// updateSchemeDetailContent fills the detail viewport with the highlighted
// scheme's partitions, and scrolls it back to the top: switching schemes
// should not carry over a scroll position that no longer means anything.
func (m *reinstallWizardModel) updateSchemeDetailContent() {
	selected, ok := selectedPickerItem(&m.schemeList)
	if !ok || selected.index >= len(m.schemes) {
		m.schemeDetailViewport.SetContent("")
		return
	}

	scheme := m.schemes[selected.index]
	width := schemePartitionsBoxWidth(scheme.Partitions)

	var lines []string
	for _, partition := range scheme.Partitions {
		lines = append(lines, describeSchemePartition(partition, width, m.rootMountpoint()))
	}

	m.schemeDetailViewport.SetContent(strings.Join(lines, "\n\n"))
	m.schemeDetailViewport.GotoTop()
}

// rootMountpoint is the OS's mandatory root mount point, used to tag which
// partition is the system root; empty when the OS's details aren't known yet.
func (m *reinstallWizardModel) rootMountpoint() string {
	if m.details == nil {
		return ""
	}

	return m.details.RootMountpoint
}

func (m *reinstallWizardModel) selectedDiskGroup() *diskGroup {
	if m.hardware == nil || m.diskGroupIdx >= len(m.hardware.DiskGroups) {
		return nil
	}

	return &m.hardware.DiskGroups[m.diskGroupIdx]
}

func (m *reinstallWizardModel) selectedDiskGroupID() int {
	if group := m.selectedDiskGroup(); group != nil {
		return group.DiskGroupID
	}

	return 0
}

func (m *reinstallWizardModel) maxDiskCount() int {
	count := 1
	if group := m.selectedDiskGroup(); group != nil && group.NumberOfDisks > 0 {
		count = group.NumberOfDisks
	}

	if m.details != nil && m.details.SoftRaidOnlyMirroring && count > 2 {
		count = 2
	}

	return count
}

func (m *reinstallWizardModel) allowedRaidLevels() []int {
	if m.details != nil && m.details.SoftRaidOnlyMirroring {
		return []int{0, 1}
	}

	return raidLevels
}

func (m *reinstallWizardModel) handleDiskGroupKeys(key string) (tea.Model, tea.Cmd) {
	groups := m.diskGroups()

	switch key {
	case "up", "k":
		if m.diskGroupIdx > 0 {
			m.diskGroupIdx--
			m.diskCount = m.maxDiskCount()
		}
	case "down", "j":
		if m.diskGroupIdx < len(groups)-1 {
			m.diskGroupIdx++
			m.diskCount = m.maxDiskCount()
		}
	case "left":
		m.backToLastInput()
	case "enter":
		m.errorMsg = ""
		if len(m.otherDiskGroups()) == 0 {
			m.step = reinstallStepDiskCount
			return m, nil
		}
		m.initEraseOthersSelection()
		m.step = reinstallStepEraseOthers
	}

	return m, nil
}

// diskGroups returns the server's disk groups, or nil if hardware
// specifications haven't loaded (or reported none).
func (m *reinstallWizardModel) diskGroups() []diskGroup {
	if m.hardware == nil {
		return nil
	}

	return m.hardware.DiskGroups
}

// otherDiskGroups returns every disk group except the one the OS is
// installed on (diskGroupIdx) — the only ones the reinstallStepEraseOthers
// step, and the API's own "erase" attribute, apply to.
func (m *reinstallWizardModel) otherDiskGroups() []diskGroup {
	groups := m.diskGroups()

	others := make([]diskGroup, 0, len(groups))
	for i, group := range groups {
		if i != m.diskGroupIdx {
			others = append(others, group)
		}
	}

	return others
}

// initEraseOthersSelection prepares reinstallStepEraseOthers: every other
// disk group defaults to erased (the API's own default), previous choices
// are kept across a trip back to reinstallStepDiskGroup unless the group
// they were for is no longer "other" (e.g. it just became the install
// target).
func (m *reinstallWizardModel) initEraseOthersSelection() {
	others := m.otherDiskGroups()

	if m.eraseOtherGroups == nil {
		m.eraseOtherGroups = make(map[int]bool, len(others))
	}

	stillOther := make(map[int]bool, len(others))
	for _, group := range others {
		stillOther[group.DiskGroupID] = true
		if _, ok := m.eraseOtherGroups[group.DiskGroupID]; !ok {
			m.eraseOtherGroups[group.DiskGroupID] = true
		}
	}
	for id := range m.eraseOtherGroups {
		if !stillOther[id] {
			delete(m.eraseOtherGroups, id)
		}
	}

	m.eraseOthersIdx = min(m.eraseOthersIdx, max(len(others)-1, 0))
}

func (m *reinstallWizardModel) handleEraseOthersKeys(key string) (tea.Model, tea.Cmd) {
	others := m.otherDiskGroups()

	switch key {
	case "up", "k":
		if m.eraseOthersIdx > 0 {
			m.eraseOthersIdx--
		}
	case "down", "j":
		if m.eraseOthersIdx < len(others)-1 {
			m.eraseOthersIdx++
		}
	case " ":
		if m.eraseOthersIdx < len(others) {
			id := others[m.eraseOthersIdx].DiskGroupID
			m.eraseOtherGroups[id] = !m.eraseOtherGroups[id]
		}
	case "left":
		m.errorMsg = ""
		m.step = reinstallStepDiskGroup
	case "enter":
		m.errorMsg = ""
		m.step = reinstallStepDiskCount
	}

	return m, nil
}

func (m *reinstallWizardModel) backToLastInput() {
	m.errorMsg = ""
	if m.details == nil || len(m.details.Inputs) == 0 {
		m.step = reinstallStepOs
		return
	}

	m.formIdx = len(m.details.Inputs) - 1
	m.step = reinstallStepInputs
}

func (m *reinstallWizardModel) handleDiskCountKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.diskCount < m.maxDiskCount() {
			m.diskCount++
		}
	case "down", "j":
		if m.diskCount > 1 {
			m.diskCount--
		}
	case "left":
		m.errorMsg = ""
		m.step = reinstallStepDiskGroup
		if len(m.otherDiskGroups()) > 0 {
			m.step = reinstallStepEraseOthers
		}
	case "enter":
		m.errorMsg = ""
		if m.details != nil && m.details.NoPartitioning {
			m.goToSchemeSelection(reinstallStepDiskCount)
			return m, nil
		}
		m.partModeList.Title = partitionModeQuestion(m.diskCount)
		m.step = reinstallStepPartitionMode
	}

	return m, nil
}

// goToSchemeSelection shows the scheme review/selection screen. With a single
// scheme there is nothing to choose, but its name and partitions are still
// shown to the customer before moving on — only up/down navigation is moot.
func (m *reinstallWizardModel) goToSchemeSelection(previous reinstallWizardStep) {
	m.layout = nil

	// Nothing to choose from: zero schemes leaves the API's own default in
	// place, and with exactly one there is no decision to make either — go
	// straight to the summary instead of a picker with a single, forced entry.
	if len(m.schemes) <= 1 {
		m.schemeName = ""
		if len(m.schemes) == 1 {
			m.schemeName = m.schemes[0].Name
		}
		m.confirmOrigin = previous
		m.confirmText = ""
		m.step = reinstallStepConfirm
		return
	}

	m.schemeList.ResetFilter()
	m.schemeList.Select(max(slices.IndexFunc(m.schemes, func(scheme osPartitionScheme) bool {
		return scheme.Name == m.schemeName
	}), 0))
	m.updateSchemeDetailContent()
	m.schemeStepOrigin = previous
	m.step = reinstallStepScheme
}

func (m *reinstallWizardModel) handlePartitionModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.filteringPicker() {
		switch msg.String() {
		case "left":
			m.errorMsg = ""
			m.step = reinstallStepDiskCount
			return m, nil
		case "enter":
			selected, ok := selectedPickerItem(&m.partModeList)
			if !ok {
				return m, nil
			}
			m.errorMsg = ""
			if selected.index == 0 {
				m.goToSchemeSelection(reinstallStepPartitionMode)
				return m, nil
			}

			m.schemeName = ""
			if len(m.layout) == 0 {
				m.layout = layoutFromScheme(defaultPartitionScheme(m.schemes))
				m.normalizeLayout()
			}
			m.layoutIdx = 0
			m.step = reinstallStepLayout
			return m, nil
		}
	}

	return m.updateActivePicker(msg)
}

// normalizeLayout drops what the current disk selection cannot express.
func (m *reinstallWizardModel) normalizeLayout() {
	allowed := m.allowedRaidLevels()

	for i := range m.layout {
		partition := &m.layout[i]
		if partition.FileSystem == "swap" || m.diskCount <= 1 {
			partition.RaidLevel = nil
			continue
		}
		if partition.RaidLevel != nil && !slices.Contains(allowed, *partition.RaidLevel) {
			partition.RaidLevel = nil
		}
	}
}

func (m *reinstallWizardModel) handleSchemeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.filteringPicker() {
		switch msg.String() {
		case "left":
			m.errorMsg = ""
			m.step = m.schemeStepOrigin
			return m, nil
		case "enter":
			selected, ok := selectedPickerItem(&m.schemeList)
			if !ok {
				return m, nil
			}
			m.errorMsg = ""
			m.schemeName = m.schemes[selected.index].Name
			m.layout = nil
			m.confirmOrigin = reinstallStepScheme
			m.confirmText = ""
			m.step = reinstallStepConfirm
			return m, nil
		case "pgup", "b":
			m.schemeDetailViewport.PageUp()
			return m, nil
		case "pgdown", "f":
			m.schemeDetailViewport.PageDown()
			return m, nil
		}
	}

	model, cmd := m.updateActivePicker(msg)
	m.updateSchemeDetailContent()

	return model, cmd
}

func (m *reinstallWizardModel) handleLayoutKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.layoutIdx > 0 {
			m.layoutIdx--
		}
	case "down", "j":
		if m.layoutIdx < len(m.layout)-1 {
			m.layoutIdx++
		}
	case "K":
		if m.layoutIdx > 0 {
			m.layout[m.layoutIdx], m.layout[m.layoutIdx-1] = m.layout[m.layoutIdx-1], m.layout[m.layoutIdx]
			m.layoutIdx--
		}
	case "J":
		if m.layoutIdx < len(m.layout)-1 {
			m.layout[m.layoutIdx], m.layout[m.layoutIdx+1] = m.layout[m.layoutIdx+1], m.layout[m.layoutIdx]
			m.layoutIdx++
		}
	case "a":
		m.startPartitionEdition(len(m.layout))
	case "e":
		if len(m.layout) > 0 {
			m.startPartitionEdition(m.layoutIdx)
		}
	case "d":
		if len(m.layout) > 0 {
			m.layout = slices.Delete(m.layout, m.layoutIdx, m.layoutIdx+1)
			m.layoutIdx = min(m.layoutIdx, max(len(m.layout)-1, 0))
			m.errorMsg = ""
		}
	case "left":
		m.errorMsg = ""
		m.step = reinstallStepPartitionMode
	case "enter":
		rootMountpoint := ""
		if m.details != nil {
			rootMountpoint = m.details.RootMountpoint
		}
		if err := validateLayout(m.layout, rootMountpoint); err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}
		m.errorMsg = ""
		m.schemeName = ""
		m.confirmOrigin = reinstallStepLayout
		m.confirmText = ""
		m.step = reinstallStepConfirm
	}

	return m, nil
}

func (m *reinstallWizardModel) startPartitionEdition(index int) {
	m.errorMsg = ""
	m.editIdx = index
	m.editFieldIdx = 0

	if index < len(m.layout) {
		m.editPartition = m.layout[index]
	} else {
		m.editPartition = layoutPartition{}
		if m.details != nil && len(m.details.Filesystems) > 0 {
			m.editPartition.FileSystem = m.details.Filesystems[0]
		}
	}

	m.editSizeText = strconv.Itoa(m.editPartition.Size)
	m.step = reinstallStepLayoutEdit
}

func (m *reinstallWizardModel) editFields() []layoutField {
	fields := []layoutField{layoutFieldFileSystem, layoutFieldMountPoint, layoutFieldSize}

	if m.diskCount > 1 && m.editPartition.FileSystem != "swap" {
		fields = append(fields, layoutFieldRaidLevel)
	}
	if m.details != nil && m.details.LvmReady &&
		m.editPartition.FileSystem != "swap" && m.editPartition.FileSystem != "zfs" {
		fields = append(fields, layoutFieldLvName)
	}
	if m.editPartition.FileSystem == "zfs" {
		fields = append(fields, layoutFieldZpName)
	}

	return fields
}

func (m *reinstallWizardModel) cycleFileSystem(delta int) {
	if m.details == nil || len(m.details.Filesystems) == 0 {
		return
	}

	index := slices.Index(m.details.Filesystems, m.editPartition.FileSystem)
	index = (index + delta + len(m.details.Filesystems)) % len(m.details.Filesystems)
	m.editPartition.FileSystem = m.details.Filesystems[index]

	if m.editPartition.FileSystem != "zfs" {
		m.editPartition.ZpName = ""
	} else {
		m.editPartition.LvName = ""
	}
	if m.editPartition.FileSystem == "swap" {
		m.editPartition.RaidLevel = nil
		m.editPartition.LvName = ""
	}
	m.editFieldIdx = min(m.editFieldIdx, len(m.editFields())-1)
}

func (m *reinstallWizardModel) cycleRaidLevel(delta int) {
	levels := m.allowedRaidLevels()

	current := 0
	if m.editPartition.RaidLevel != nil {
		current = slices.Index(levels, *m.editPartition.RaidLevel) + 1
	}

	current = (current + delta + len(levels) + 1) % (len(levels) + 1)
	if current == 0 {
		m.editPartition.RaidLevel = nil
		return
	}

	level := levels[current-1]
	m.editPartition.RaidLevel = &level
}

func (m *reinstallWizardModel) handleLayoutEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	fields := m.editFields()
	field := fields[min(m.editFieldIdx, len(fields)-1)]
	key := msg.String()

	switch key {
	case "esc":
		m.errorMsg = ""
		m.step = reinstallStepLayout
		return m, nil
	case "tab", "down":
		m.editFieldIdx = (m.editFieldIdx + 1) % len(fields)
		return m, nil
	case "shift+tab", "up":
		m.editFieldIdx = (m.editFieldIdx + len(fields) - 1) % len(fields)
		return m, nil
	case "enter":
		if err := m.savePartitionEdition(); err != nil {
			m.errorMsg = err.Error()
			return m, nil
		}
		m.errorMsg = ""
		m.step = reinstallStepLayout
		return m, nil
	}

	switch field {
	case layoutFieldFileSystem:
		switch key {
		case "left", "h":
			m.cycleFileSystem(-1)
		case "right", "l", " ":
			m.cycleFileSystem(1)
		}
		return m, nil

	case layoutFieldRaidLevel:
		switch key {
		case "left", "h":
			m.cycleRaidLevel(-1)
		case "right", "l", " ":
			m.cycleRaidLevel(1)
		}
		return m, nil
	}

	switch key {
	case "backspace":
		switch field {
		case layoutFieldMountPoint:
			m.editPartition.MountPoint = trimLastRune(m.editPartition.MountPoint)
		case layoutFieldSize:
			m.editSizeText = trimLastRune(m.editSizeText)
		case layoutFieldLvName:
			m.editPartition.LvName = trimLastRune(m.editPartition.LvName)
		case layoutFieldZpName:
			m.editPartition.ZpName = trimLastRune(m.editPartition.ZpName)
		}
	default:
		if len(msg.Runes) == 0 {
			return m, nil
		}
		value := string(msg.Runes)
		switch field {
		case layoutFieldMountPoint:
			m.editPartition.MountPoint += value
		case layoutFieldSize:
			if strings.IndexFunc(value, func(r rune) bool { return r < '0' || r > '9' }) < 0 {
				m.editSizeText = normalizeDigits(m.editSizeText + value)
			}
		case layoutFieldLvName:
			m.editPartition.LvName += value
		case layoutFieldZpName:
			m.editPartition.ZpName += value
		}
	}

	return m, nil
}

func (m *reinstallWizardModel) savePartitionEdition() error {
	partition := m.editPartition

	if partition.FileSystem == "" {
		return fmt.Errorf("a file system must be selected")
	}

	partition.MountPoint = strings.TrimSpace(partition.MountPoint)
	if partition.MountPoint == "" {
		return fmt.Errorf("a mount point is required")
	}

	size, err := strconv.Atoi(strings.TrimSpace(m.editSizeText))
	if err != nil {
		return fmt.Errorf("invalid size, expected a number of MiB")
	}
	partition.Size = size

	if partition.FileSystem != "zfs" {
		partition.ZpName = ""
	}
	if m.diskCount <= 1 || partition.FileSystem == "swap" {
		partition.RaidLevel = nil
	}
	if partition.FileSystem == "swap" || partition.FileSystem == "zfs" {
		partition.LvName = ""
	}

	if m.editIdx < len(m.layout) {
		m.layout[m.editIdx] = partition
	} else {
		m.layout = append(m.layout, partition)
		m.layoutIdx = len(m.layout) - 1
	}

	return nil
}

func (m *reinstallWizardModel) renderDiskGroupStep() string {
	var content strings.Builder
	content.WriteString(wizardTitleStyle.Render("Select the disk group the OS will be installed on:") + "\n\n")

	groups := m.diskGroups()
	if len(groups) == 0 {
		content.WriteString(wizardDimStyle.Render("  No disk group reported for this server.") + "\n\n")
		content.WriteString(m.renderError())
		content.WriteString(wizardHintStyle.Render("Enter: Continue • ←: Back • Esc: Cancel"))
		return content.String()
	}

	width := diskGroupsBoxWidth(groups)
	for i, group := range groups {
		box := formatPartitionBox(diskGroupAttrs(group, diskGroupAvailableDisksLabel), width, diskGroupTags(group))
		content.WriteString(renderPartitionEntry(box, i == m.diskGroupIdx))
	}
	content.WriteString("\n")

	if group := m.selectedDiskGroup(); group != nil {
		if warning := hardwareRaidWarning(*group); warning != "" {
			content.WriteString(wizardErrorStyle.Render("⚠  "+warning+"  ⚠") + "\n\n")
		}
	}

	content.WriteString(m.renderError())
	content.WriteString(wizardHintStyle.Render("↑↓ Navigate • Enter: Continue • ←: Back • Esc: Cancel"))

	return content.String()
}

func (m *reinstallWizardModel) renderEraseOthersStep() string {
	var content strings.Builder
	content.WriteString(wizardTitleStyle.Render("Data on the other disk groups:") + "\n")
	content.WriteString(wizardDescStyle.Render(
		"Every disk group is erased during reinstallation by default — pick which of these should keep their data instead.") + "\n\n")

	others := m.otherDiskGroups()
	width := otherDiskGroupsBoxWidth(others)
	for i, group := range others {
		erase := m.eraseOtherGroups[group.DiskGroupID]
		attrs := append(diskGroupAttrs(group, diskGroupDisksLabel), partitionAttr{"Erase data", eraseChoiceLabel(erase)})
		box := formatPartitionBox(attrs, width, diskGroupTags(group))
		content.WriteString(renderPartitionEntry(box, i == m.eraseOthersIdx))
	}
	content.WriteString("\n")

	content.WriteString(m.renderError())
	content.WriteString(wizardHintStyle.Render(
		"↑↓ Navigate • Space: Toggle erase/keep • Enter: Continue • ←: Back • Esc: Cancel"))

	return content.String()
}

func (m *reinstallWizardModel) renderDiskCountStep() string {
	group := m.selectedDiskGroup()

	title := "Number of disks to use:"
	if group != nil {
		title = fmt.Sprintf("Number of disks to use on disk group %d:", group.DiskGroupID)
	}

	var content strings.Builder
	content.WriteString(wizardTitleStyle.Render(title) + "\n\n")

	if group != nil {
		attrs := diskGroupAttrs(*group, diskGroupAvailableDisksLabel)
		box := formatPartitionBox(attrs, partitionRowWidth(attrs), diskGroupTags(*group))
		for _, line := range strings.Split(box, "\n") {
			content.WriteString(wizardDimStyle.Render("  "+line) + "\n")
		}
		content.WriteString("\n")
	}

	content.WriteString(wizardNumberInputStyle.Render(strconv.Itoa(m.diskCount)) + "\n")
	content.WriteString(wizardSubtleStyle.Render(fmt.Sprintf("  Between 1 and %d", m.maxDiskCount())) + "\n")

	if m.details != nil && m.details.SoftRaidOnlyMirroring {
		content.WriteString(wizardDescStyle.Render("  This OS only supports RAID mirroring, at most 2 disks can be used.") + "\n")
	}
	content.WriteString("\n")

	content.WriteString(m.renderError())
	content.WriteString(wizardHintStyle.Render("↑↓ Change • Enter: Continue • ←: Back • Esc: Cancel"))

	return content.String()
}

func (m *reinstallWizardModel) renderPartitionModeStep() string {
	var content strings.Builder
	content.WriteString(m.partModeList.View() + "\n\n")

	content.WriteString(m.renderError())
	content.WriteString(wizardHintStyle.Render("Enter: Continue • ←: Back • Esc: Cancel"))

	return content.String()
}

func (m *reinstallWizardModel) renderSchemeStep() string {
	var content strings.Builder
	content.WriteString(m.schemeList.View() + "\n")
	content.WriteString(wizardDescStyle.Render("  Partitions of the highlighted scheme:") + "\n")
	content.WriteString(m.schemeDetailViewport.View() + "\n")
	if !m.schemeDetailViewport.AtTop() || !m.schemeDetailViewport.AtBottom() {
		content.WriteString(wizardSubtleStyle.Render(fmt.Sprintf("  %.0f%% scrolled",
			m.schemeDetailViewport.ScrollPercent()*100)) + "\n")
	}
	content.WriteString("\n")

	content.WriteString(m.renderError())
	content.WriteString(wizardHintStyle.Render(
		"↑↓ Navigate • PgUp/PgDn: Scroll details • Enter: Continue • ←: Back • Esc: Cancel"))

	return content.String()
}

func (m *reinstallWizardModel) renderLayoutStep() string {
	var content strings.Builder
	content.WriteString(wizardTitleStyle.Render("Custom partitioning layout:") + "\n")
	content.WriteString(wizardDescStyle.Render("Partitions are created in the order of this list.") + "\n\n")

	if len(m.layout) == 0 {
		content.WriteString(wizardDimStyle.Render("  No partition defined yet.") + "\n")
	}

	width := layoutPartitionsBoxWidth(m.layout)
	for i, partition := range m.layout {
		content.WriteString(renderPartitionEntry(describeLayoutPartition(partition, width, m.rootMountpoint()), i == m.layoutIdx))
	}
	content.WriteString("\n")

	content.WriteString(m.renderError())
	content.WriteString(wizardHintStyle.Render(
		"↑↓ Navigate • K/J: Move • a: Add • e: Edit • d: Delete • Enter: Continue • ←: Back • Esc: Cancel"))

	return content.String()
}

func (m *reinstallWizardModel) renderLayoutEditStep() string {
	var content strings.Builder

	title := "Edit partition"
	if m.editIdx >= len(m.layout) {
		title = "New partition"
	}
	content.WriteString(wizardTitleStyle.Render(title+":") + "\n\n")

	fields := m.editFields()
	for i, field := range fields {
		var value string
		switch field {
		case layoutFieldFileSystem:
			value = m.editPartition.FileSystem
		case layoutFieldMountPoint:
			value = m.editPartition.MountPoint
		case layoutFieldSize:
			value = m.editSizeText
		case layoutFieldRaidLevel:
			value = "none"
			if m.editPartition.RaidLevel != nil {
				value = strconv.Itoa(*m.editPartition.RaidLevel)
			}
		case layoutFieldLvName:
			value = m.editPartition.LvName
			if value == "" {
				value = "(no LVM)"
			}
		case layoutFieldZpName:
			value = m.editPartition.ZpName
			if value == "" {
				value = "(auto)"
			}
		}

		line := wizardLabelStyle.Render("  "+layoutFieldLabels[field]) + " "
		if i == m.editFieldIdx {
			line += wizardSelectedStyle.Render("▶ " + value)
		} else {
			line += wizardDimStyle.Render("  " + value)
		}
		content.WriteString(line + "\n")
	}
	content.WriteString("\n")

	content.WriteString(m.renderError())
	content.WriteString(wizardHintStyle.Render(
		"↑↓/Tab: Field • ←→: Change choice • Type: Edit value • Enter: Save • Esc: Back to the list"))

	return content.String()
}

func partitionModeQuestion(diskCount int) string {
	if diskCount == 1 {
		return "How should the disk be partitioned?"
	}

	return fmt.Sprintf("How should the %d disks be partitioned?", diskCount)
}

// diskGroupDisksLabel is the "Disks" row's label in a disk group's box.
// "Available Disks" is more accurate while picking the install target and
// its disk count (before other, non-erased disk groups can even be
// factored in), but plain "Disks" fits every other context (e.g. the
// reinstallStepEraseOthers step, which isn't about availability at all).
const diskGroupDisksLabel = "Disks"
const diskGroupAvailableDisksLabel = "Available Disks"

// diskGroupAttrs lists a disk group's attributes for display as its own box,
// the same way a partition's are.
func diskGroupAttrs(group diskGroup, disksLabel string) []partitionAttr {
	attrs := []partitionAttr{
		{disksLabel, strconv.Itoa(group.NumberOfDisks)},
	}
	if size := formatQuantity(group.DiskSize); size != "" {
		attrs = append(attrs, partitionAttr{"Disk size", size})
	}
	if group.DiskType != "" {
		attrs = append(attrs, partitionAttr{"Disk type", group.DiskType})
	}
	if group.Description != "" {
		attrs = append(attrs, partitionAttr{"Description", group.Description})
	}
	if group.RaidController != "" {
		attrs = append(attrs, partitionAttr{"RAID controller", group.RaidController})
	}

	return attrs
}

// diskGroupLabel names a disk group for display in prose (e.g. a warning),
// as opposed to diskGroupAttrs' own box of individual attribute rows: its
// id, then how many disks it has, what type/size they are, and how they're
// assembled (a hardware RAID controller's name, or "JBOD" otherwise) —
// e.g. "Group 1 (2 X Disk SSD 480 GB, JBOD)".
func diskGroupLabel(group diskGroup) string {
	disks := fmt.Sprintf("%d X Disk", group.NumberOfDisks)
	if group.DiskType != "" {
		disks += " " + strings.ToUpper(group.DiskType)
	}
	if size := formatQuantity(group.DiskSize); size != "" {
		disks += " " + size
	}

	raid := "JBOD"
	if group.hasHardwareRaid() {
		raid = group.RaidController
	}

	return fmt.Sprintf("Group %d (%s, %s)", group.DiskGroupID, disks, raid)
}

// erasedDiskGroupLabels names every disk group that will actually be erased:
// the one the OS is installed on, always, plus any other one not set to
// keep its data — the ones the confirm screen's erasure warning must list.
func (m *reinstallWizardModel) erasedDiskGroupLabels() []string {
	var labels []string
	if group := m.selectedDiskGroup(); group != nil {
		labels = append(labels, diskGroupLabel(*group))
	}
	for _, group := range m.otherDiskGroups() {
		if m.eraseOtherGroups[group.DiskGroupID] {
			labels = append(labels, diskGroupLabel(group))
		}
	}

	return labels
}

// diskGroupTags names the badge shown above a disk group's box: "HARDWARE
// RAID" when the OVHcloud CLI can't configure its controller (it always
// installs such a group in JBOD mode instead).
func diskGroupTags(group diskGroup) []string {
	if group.hasHardwareRaid() {
		return []string{"HARDWARE RAID"}
	}

	return nil
}

// diskGroupsBoxWidth gives the shared box width for a list of disk groups
// shown together, so every box is exactly as wide as the widest one needs.
func diskGroupsBoxWidth(groups []diskGroup) int {
	attrSets := make([][]partitionAttr, len(groups))
	for i, group := range groups {
		attrSets[i] = diskGroupAttrs(group, diskGroupAvailableDisksLabel)
	}

	return partitionBoxWidth(attrSets)
}

// eraseChoiceLabel describes what an "erase" choice does to a disk group's
// existing data, for display next to the yes/no toggle.
func eraseChoiceLabel(erase bool) string {
	if erase {
		return "Yes (data erased)"
	}

	return "No (data kept)"
}

// otherDiskGroupsBoxWidth is diskGroupsBoxWidth, but also accounting for the
// "Erase data" row reinstallStepEraseOthers adds to every box.
func otherDiskGroupsBoxWidth(groups []diskGroup) int {
	attrSets := make([][]partitionAttr, len(groups))
	for i, group := range groups {
		attrSets[i] = append(diskGroupAttrs(group, diskGroupDisksLabel), partitionAttr{"Erase data", eraseChoiceLabel(true)})
	}

	return partitionBoxWidth(attrSets)
}

// renderPartitionEntry indents every line of a partition box the same way the
// rest of this list style marks a selection: an arrow (only on the box's
// first line) or two blank spaces, never breaking the box's left alignment.
func renderPartitionEntry(box string, selected bool) string {
	marker, style := "  ", wizardDimStyle
	if selected {
		marker, style = "▶ ", wizardSelectedStyle
	}

	var b strings.Builder
	for i, line := range strings.Split(box, "\n") {
		prefix := "  "
		if i == 0 {
			prefix = marker
		}
		b.WriteString(style.Render(prefix+line) + "\n")
	}

	return b.String()
}
