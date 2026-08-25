// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package baremetal

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type reinstallWizardStep int

const (
	reinstallStepOs reinstallWizardStep = iota
	reinstallStepInputs
	reinstallStepEnum
	reinstallStepSSHKey
	reinstallStepKeyValue
	reinstallStepKeyValueEdit
	reinstallStepDiskGroup
	reinstallStepDiskCount
	reinstallStepPartitionMode
	reinstallStepScheme
	reinstallStepLayout
	reinstallStepLayoutEdit
	reinstallStepConfirm
	reinstallStepSavePath
)

var (
	wizardTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	wizardDescStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	wizardSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	wizardDimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	wizardSubtleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	wizardHintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	wizardErrorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	wizardLoadingStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	wizardLabelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(26)
	wizardValueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	wizardInputStyle    = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00FF7F")).
				Padding(0, 1).Width(60)
	wizardNumberInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00FF7F")).
				Padding(0, 1).Width(10)
	wizardTextAreaStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00FF7F")).
				Padding(0, 1).Width(76).Height(8)
	wizardKVBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#666666")).
				Padding(0, 1).Width(28)
	wizardKVBoxFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00FF7F")).
				Padding(0, 1).Width(28)
	wizardConfirmInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#FF6B6B")).
				Foreground(lipgloss.Color("#FF6B6B")).
				Padding(0, 1).Width(10)
	// wizardConfirmBoxStyle boxes the whole reinstallation summary (OS,
	// customizations, storage) on the confirm screen.
	wizardConfirmBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00FF7F")).
				Padding(1, 2)
)

// maxKeyValuePairs caps how many entries a "keyValue" customization input can
// carry, keeping the list (and the request payload) manageable.
const maxKeyValuePairs = 10

// Size of a list screen until the terminal reports its own.
const (
	defaultPickerWidth  = 60
	defaultPickerHeight = 15
)

// osListTitle is osList's title, rendered by hand rather than through the
// list's own (so the column header can be inserted between the two, above
// the items, when the details table is showing).
const osListTitle = "Select the operating system to install:"

type osNamesLoadedMsg struct {
	nameItems      []pickerItem
	tableItems     []pickerItem
	tableHeader    string
	tableSeparator string
	infoByName     map[string]map[string]any
	err            error
}

type osDetailsLoadedMsg struct {
	details *osDetails
	sshKeys []string
	err     error
}

type storageLoadedMsg struct {
	schemes  []osPartitionScheme
	hardware *hardwareSpecifications
	err      error
}

type sshKeyLoadedMsg struct {
	content string
	err     error
}

// pickerItem is one choice of a "pick one from a list" screen: label is what
// typing filters on and, absent a row, what is shown; row, when set,
// replaces label on-screen with a wider, pre-aligned table line (label
// stays the actual value, e.g. an OS name, so selecting the item is
// unaffected); lines are extra details only shown while the item is
// selected; index is its position in the data the list was built from,
// which filtering makes the list's own index unusable for.
type pickerItem struct {
	label string
	row   string
	lines []string
	index int
}

func (i pickerItem) FilterValue() string { return i.label }

type pickerItemDelegate struct{}

func (d pickerItemDelegate) Height() int                             { return 1 }
func (d pickerItemDelegate) Spacing() int                            { return 0 }
func (d pickerItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d pickerItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(pickerItem)
	if !ok {
		return
	}

	text := item.label
	if item.row != "" {
		text = item.row
	}

	if index != m.Index() {
		// wizardSelectedStyle below adds 1 column of padding on each side, on
		// top of its own "▶ " prefix: match that with a 3rd leading space
		// here, so the text doesn't shift right by one column when selected.
		fmt.Fprint(w, wizardDimStyle.Render("   "+text))
		return
	}

	fmt.Fprint(w, wizardSelectedStyle.Render("▶ "+text))
	for _, line := range item.lines {
		fmt.Fprint(w, "\n"+wizardSubtleStyle.Render("    "+line))
	}
}

// pickerListItems wraps the choices for the list, which only takes an interface.
func pickerListItems(items []pickerItem) []list.Item {
	listItems := make([]list.Item, 0, len(items))
	for _, item := range items {
		listItems = append(listItems, item)
	}

	return listItems
}

// labelPickerItems builds the choices of a picker whose labels are all there is
// to show, such as the OS names or the values of an enum.
func labelPickerItems(labels []string) []pickerItem {
	items := make([]pickerItem, 0, len(labels))
	for i, label := range labels {
		items = append(items, pickerItem{label: label, index: i})
	}

	return items
}

// newPickerList builds a filterable list screen, all of them being set up the
// same way.
func (m *reinstallWizardModel) newPickerList(title string) list.Model {
	picker := list.New(nil, pickerItemDelegate{}, m.listWidth, m.listHeight)
	picker.Title = title
	picker.SetShowStatusBar(false)
	picker.SetFilteringEnabled(true)
	picker.Styles.Title = wizardTitleStyle
	picker.Styles.HelpStyle = wizardHintStyle
	// The wizard, not the list, decides when to leave.
	picker.KeyMap.Quit.SetEnabled(false)

	return picker
}

// fillPicker gives a list its choices, starting from a clean filter so a screen
// is never reopened on the text typed the previous time.
func fillPicker(picker *list.Model, items []pickerItem, selected int) {
	picker.ResetFilter()
	picker.SetItems(pickerListItems(items))
	picker.Select(selected)
}

type reinstallWizardModel struct {
	serviceName string

	step           reinstallWizardStep
	isLoading      bool
	loadingMessage string
	errorMsg       string
	err            error

	// Size given to every filterable list of the wizard.
	listWidth  int
	listHeight int

	// Screen 1: operating system
	osList list.Model
	osName string
	// osNameItems is the plain, name-only choice for every compatible OS,
	// shown in technical mode; osTableItems is the default view, one aligned
	// table row per OS that templateInfos has details for (deprecated or
	// otherwise obscure templates are usually the ones left out).
	// osTechnicalMode picks which of the two osList currently shows.
	osNameItems      []pickerItem
	osTableItems     []pickerItem
	osTableHeader    string
	osTableSeparator string
	osTechnicalMode  bool
	// osInfoByName holds each OS's templateInfos fields (description,
	// category, family, subfamily, endOfInstall), keyed by name, for the
	// end-of-install warning and the confirm screen's OS details, when the
	// chosen OS has any (technical-mode-only OSes don't).
	osInfoByName map[string]map[string]any

	// Screen 2: customizations
	details *osDetails
	formIdx int
	// answers holds a live value for every input of the form, while
	// customizations only keeps what is actually sent, once the form is left.
	answers        map[string]any
	customizations map[string]any
	enumList       list.Model
	sshKeyNames    []string
	sshKeyList     list.Model
	kvPairs        []keyValuePair
	kvIdx          int
	kvEditIdx      int
	kvEditKey      string
	kvEditValue    string
	kvEditFieldIdx int

	// Screen 3: storage
	hardware     *hardwareSpecifications
	schemes      []osPartitionScheme
	diskGroupIdx int
	diskCount    int
	partModeList list.Model
	schemeList   list.Model
	// schemeDetailViewport shows the currently highlighted scheme's partitions,
	// scrollable independently of schemeList: unlike a bubbles/list item (whose
	// declared Height() must be one fixed value for every item, list-wide), a
	// viewport genuinely clips and scrolls content taller than it, so a scheme
	// with many partitions can never silently overflow the terminal.
	schemeDetailViewport viewport.Model
	schemeName           string
	schemeStepOrigin     reinstallWizardStep
	layout               []layoutPartition
	layoutIdx            int
	editIdx              int
	editPartition        layoutPartition
	editFieldIdx         int
	editSizeText         string
	confirmOrigin        reinstallWizardStep

	// Screen 4: confirmation
	saveEnabled bool
	savePath    string
	savedPath   string
	launch      bool
	// confirmText is what the user typed to confirm the reinstallation:
	// only "yes" (case-insensitive) actually launches it.
	confirmText string
}

func newReinstallWizardModel(serviceName string) *reinstallWizardModel {
	model := &reinstallWizardModel{
		serviceName:    serviceName,
		step:           reinstallStepOs,
		isLoading:      true,
		loadingMessage: "Loading compatible operating systems…",
		customizations: map[string]any{},
		listWidth:      defaultPickerWidth,
		listHeight:     defaultPickerHeight,
	}

	// Every list is built up front, so a screen opened later already has the
	// size the terminal reported.
	model.osList = model.newPickerList(osListTitle)
	model.osList.SetShowTitle(false)
	model.enumList = model.newPickerList("")
	model.sshKeyList = model.newPickerList("")
	model.partModeList = model.newPickerList("How should the disks be partitioned?")
	model.schemeList = model.newPickerList("Select a partitioning scheme:")
	model.schemeDetailViewport = viewport.New(model.listWidth, model.listHeight)

	fillPicker(&model.partModeList, labelPickerItems(partitionModeChoices), 0)

	return model
}

// pickers lists every filterable screen of the wizard, so they all follow the
// size of the terminal.
func (m *reinstallWizardModel) pickers() []*list.Model {
	return []*list.Model{
		&m.osList, &m.enumList, &m.sshKeyList,
		&m.partModeList, &m.schemeList,
	}
}

// activePicker is the list the current step is showing, if it is one.
func (m *reinstallWizardModel) activePicker() *list.Model {
	switch m.step {
	case reinstallStepOs:
		return &m.osList
	case reinstallStepEnum:
		return &m.enumList
	case reinstallStepSSHKey:
		return &m.sshKeyList
	case reinstallStepPartitionMode:
		return &m.partModeList
	case reinstallStepScheme:
		return &m.schemeList
	}

	return nil
}

// filteringPicker tells whether a filter is currently being typed, which both
// Esc and the step handlers must leave to the list itself.
func (m *reinstallWizardModel) filteringPicker() bool {
	picker := m.activePicker()

	return picker != nil && picker.FilterState() == list.Filtering
}

func (m *reinstallWizardModel) Init() tea.Cmd {
	return m.fetchOsNamesCmd()
}

func (m *reinstallWizardModel) fetchOsNamesCmd() tea.Cmd {
	return func() tea.Msg {
		nameItems, tableItems, tableHeader, tableSeparator, infoByName, err := fetchCompatibleOsPickerData(m.serviceName)
		return osNamesLoadedMsg{
			nameItems: nameItems, tableItems: tableItems, tableHeader: tableHeader,
			tableSeparator: tableSeparator, infoByName: infoByName, err: err,
		}
	}
}

// fetchCompatibleOsPickerData returns the OS picker's two views: nameItems
// is a plain, name-only choice per OS compatible with the given server
// (technical mode), tableItems an aligned table row per one of them that
// templateInfos has details for, tableHeader/tableSeparator its matching
// column header and horizontal rule below it (the default view), and
// infoByName that same templateInfos data, keyed by name, for the
// end-of-install warning and the confirm screen's OS details.
func fetchCompatibleOsPickerData(serviceName string) (
	nameItems []pickerItem, tableItems []pickerItem, tableHeader string, tableSeparator string,
	infoByName map[string]map[string]any, err error,
) {
	names, err := fetchCompatibleOsNames(serviceName)
	if err != nil {
		return nil, nil, "", "", nil, err
	}

	seenNames := make(map[string]bool, len(names))
	for i, name := range names {
		seenNames[name] = true
		nameItems = append(nameItems, pickerItem{label: name, index: i})
	}

	infos, err := fetchOsTemplatesInfo(func(templateName string) bool { return seenNames[templateName] })
	if err != nil {
		return nil, nil, "", "", nil, err
	}

	tableItems, tableHeader, tableSeparator = osTemplateTableRows(infos)

	infoByName = make(map[string]map[string]any, len(infos))
	for _, info := range infos {
		if name, _ := info["name"].(string); name != "" {
			infoByName[name] = info
		}
	}

	return nameItems, tableItems, tableHeader, tableSeparator, infoByName, nil
}

// osTemplateColumns are the templateOsInfo fields shown as table columns,
// beside the OS name, in the order they appear.
var osPickerTableColumns = []struct {
	key    string
	header string
}{
	{"description", "Description"},
	{"category", "Category"},
	{"family", "Family"},
	{"subfamily", "Subfamily"},
	{"endOfInstall", "End of install"},
}

// osTemplateTableRows turns fetchOsTemplatesInfo's rows into picker items
// whose row is a single line with every detail field aligned into
// fixed-width columns, separated by a vertical bar (label stays the plain
// OS name, the actual value selecting the item answers), plus the column
// header line and a matching horizontal rule below it, aligned the same way.
func osTemplateTableRows(infos []map[string]any) (items []pickerItem, header string, separator string) {
	if len(infos) == 0 {
		return nil, "", ""
	}

	nameWidth := len("Name")
	widths := make(map[string]int, len(osPickerTableColumns))
	for _, column := range osPickerTableColumns {
		widths[column.key] = len(column.header)
	}
	for _, info := range infos {
		if name, _ := info["name"].(string); len(name) > nameWidth {
			nameWidth = len(name)
		}
		for _, column := range osPickerTableColumns {
			if value, _ := info[column.key].(string); len(value) > widths[column.key] {
				widths[column.key] = len(value)
			}
		}
	}

	columnWidths := make([]int, 0, len(osPickerTableColumns)+1)
	columnWidths = append(columnWidths, nameWidth)
	for _, column := range osPickerTableColumns {
		columnWidths = append(columnWidths, widths[column.key])
	}

	headerCells := make([]string, 0, len(columnWidths))
	headerCells = append(headerCells, padRight("Name", nameWidth))
	for _, column := range osPickerTableColumns {
		headerCells = append(headerCells, padRight(column.header, widths[column.key]))
	}
	header = strings.Join(headerCells, " │ ")

	separatorCells := make([]string, len(columnWidths))
	for i, width := range columnWidths {
		separatorCells[i] = strings.Repeat("─", width)
	}
	separator = strings.Join(separatorCells, "─┼─")

	items = make([]pickerItem, 0, len(infos))
	for i, info := range infos {
		name, _ := info["name"].(string)
		cells := make([]string, 0, len(columnWidths))
		cells = append(cells, padRight(name, nameWidth))
		for _, column := range osPickerTableColumns {
			value, _ := info[column.key].(string)
			cells = append(cells, padRight(value, widths[column.key]))
		}
		items = append(items, pickerItem{label: name, row: strings.Join(cells, " │ "), index: i})
	}

	return items, header, separator
}

// padRight right-pads s with spaces up to width, so table columns line up;
// s itself is returned unchanged if it is already at least that wide.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}

	return s + strings.Repeat(" ", width-len(s))
}

func (m *reinstallWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.listWidth, m.listHeight = msg.Width-4, msg.Height-8
		for _, picker := range m.pickers() {
			picker.SetSize(m.listWidth, m.listHeight)
		}
		m.resizeSchemeViews()
		return m, nil

	case osNamesLoadedMsg:
		m.isLoading = false
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.osNameItems = msg.nameItems
		m.osTableItems = msg.tableItems
		m.osTableHeader = msg.tableHeader
		m.osTableSeparator = msg.tableSeparator
		m.osInfoByName = msg.infoByName
		m.fillOsList()
		return m, nil

	case osDetailsLoadedMsg:
		m.isLoading = false
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.details = msg.details
		m.sshKeyNames = msg.sshKeys
		m.answers = seedCustomizations(msg.details.Inputs)
		return m.openInputsForm()

	case sshKeyLoadedMsg:
		m.isLoading = false
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			return m, nil
		}
		// Picking a listed key answers the question: move on to the next one.
		m.setFocusedAnswer(msg.content)
		return m.focusNextInput()

	case storageLoadedMsg:
		m.isLoading = false
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.schemes = msg.schemes
		m.hardware = msg.hardware
		m.initDiskGroupSelection()
		m.initSchemeSelection()
		m.step = reinstallStepDiskGroup
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	if picker := m.activePicker(); picker != nil {
		var cmd tea.Cmd
		*picker, cmd = picker.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *reinstallWizardModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" {
		m.launch = false
		return m, tea.Quit
	}

	if m.isLoading {
		return m, nil
	}

	// Esc leaves the wizard, except on the steps where it goes back to the
	// list they were opened from, and while a filter is being typed.
	if key == "esc" &&
		m.step != reinstallStepLayoutEdit &&
		m.step != reinstallStepKeyValueEdit &&
		m.step != reinstallStepSavePath &&
		!m.filteringPicker() {
		m.launch = false
		return m, m.leaveWizard()
	}

	switch m.step {
	case reinstallStepOs:
		return m.handleOsKeys(msg)
	case reinstallStepInputs:
		return m.handleInputsKeys(msg)
	case reinstallStepEnum:
		return m.handleEnumKeys(msg)
	case reinstallStepSSHKey:
		return m.handleSSHKeyKeys(msg)
	case reinstallStepKeyValue:
		return m.handleKeyValueKeys(key)
	case reinstallStepKeyValueEdit:
		return m.handleKeyValueEditKeys(msg)
	case reinstallStepDiskGroup:
		return m.handleDiskGroupKeys(key)
	case reinstallStepDiskCount:
		return m.handleDiskCountKeys(key)
	case reinstallStepPartitionMode:
		return m.handlePartitionModeKeys(msg)
	case reinstallStepScheme:
		return m.handleSchemeKeys(msg)
	case reinstallStepLayout:
		return m.handleLayoutKeys(key)
	case reinstallStepLayoutEdit:
		return m.handleLayoutEditKeys(msg)
	case reinstallStepConfirm:
		return m.handleConfirmKeys(msg)
	case reinstallStepSavePath:
		return m.handleSavePathKeys(msg)
	}

	return m, nil
}

// leaveWizard writes the parameters file when the user asked for it, whether or
// not the reinstallation itself is launched.
func (m *reinstallWizardModel) leaveWizard() tea.Cmd {
	if m.saveEnabled && m.savedPath == "" {
		if err := m.writeParametersFile(); err != nil {
			m.errorMsg = err.Error()
			m.launch = false
			return nil
		}
	}

	return tea.Quit
}

func (m *reinstallWizardModel) handleOsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.filteringPicker() {
		switch msg.String() {
		case "enter":
			selected, ok := selectedPickerItem(&m.osList)
			if !ok {
				return m, nil
			}
			m.osName = selected.label
			m.isLoading = true
			m.loadingMessage = fmt.Sprintf("Loading installation parameters of %s…", m.osName)
			return m, m.fetchOsDetailsCmd(m.osName)
		case "t":
			m.osTechnicalMode = !m.osTechnicalMode
			m.fillOsList()
			return m, nil
		}
	}

	return m.updateActivePicker(msg)
}

// osShowingTable reports whether osList is currently showing the details
// table (the default view) rather than the plain name list — technical mode
// always shows the name list, and so does the default view if not one
// compatible OS has details to show.
func (m *reinstallWizardModel) osShowingTable() bool {
	return !m.osTechnicalMode && len(m.osTableItems) > 0
}

// visibleOsItems returns the OS choices osList should currently show.
func (m *reinstallWizardModel) visibleOsItems() []pickerItem {
	if m.osShowingTable() {
		return m.osTableItems
	}

	return m.osNameItems
}

// fillOsList (re)fills osList from osNameItems/osTableItems, following the
// current technical mode.
func (m *reinstallWizardModel) fillOsList() {
	fillPicker(&m.osList, m.visibleOsItems(), 0)
}

// osTechnicalModeHint describes what "t" currently does: leaving technical
// mode for the details table also drops any OS templateInfos has no
// details for, which the hint calls out so their disappearance isn't a
// surprise.
func (m *reinstallWizardModel) osTechnicalModeHint() string {
	if m.osTechnicalMode {
		return "t: Close the technical view"
	}

	hidden := len(m.osNameItems) - len(m.osTableItems)
	if hidden <= 0 {
		return "t: Close the technical view"
	}

	return "t: Show the technical view"
}

// selectedOsEndOfInstallWarning warns as soon as the currently highlighted
// OS is close to its end of install, so it's known before it is even picked.
func (m *reinstallWizardModel) selectedOsEndOfInstallWarning() string {
	selected, ok := selectedPickerItem(&m.osList)
	if !ok {
		return ""
	}

	return m.osEndOfInstallWarningFor(selected.label, true)
}

// osEndOfInstallWarningFor warns when the given OS is close to its end of
// install, empty if osName has no known date (e.g. not fetched yet) or is
// comfortably far from it. includeDate is passed through to
// osEndOfInstallWarning.
func (m *reinstallWizardModel) osEndOfInstallWarningFor(osName string, includeDate bool) string {
	endOfInstall, _ := m.osInfoByName[osName]["endOfInstall"].(string)

	return osEndOfInstallWarning(endOfInstall, time.Now(), includeDate)
}

// updateActivePicker hands the key over to the list of the current step, which
// owns navigation as well as the filter text being typed.
func (m *reinstallWizardModel) updateActivePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	picker := m.activePicker()
	if picker == nil {
		return m, nil
	}

	var cmd tea.Cmd
	*picker, cmd = picker.Update(msg)

	return m, cmd
}

func selectedPickerItem(picker *list.Model) (pickerItem, bool) {
	item, ok := picker.SelectedItem().(pickerItem)

	return item, ok
}

func (m *reinstallWizardModel) View() string {
	if m.err != nil {
		return ""
	}

	var content strings.Builder
	content.WriteString(wizardTitleStyle.Render("OS reinstallation on server "+m.serviceName) + "\n")
	content.WriteString(wizardSubtleStyle.Render(m.stepLabel()) + "\n\n")

	if m.isLoading {
		content.WriteString(wizardLoadingStyle.Render("⏳ " + m.loadingMessage))
		return content.String()
	}

	switch m.step {
	case reinstallStepOs:
		content.WriteString(wizardTitleStyle.Render(osListTitle) + "\n\n")
		if m.osShowingTable() {
			content.WriteString(wizardSubtleStyle.Bold(true).Render("   "+m.osTableHeader) + "\n")
			content.WriteString(wizardSubtleStyle.Render("   "+m.osTableSeparator) + "\n")
		}
		// osList's title is hidden (rendered by hand as osListTitle above),
		// but the list still reserves a line for it (blank, but padded to
		// the list's width) unless a filter is actively being typed there —
		// drop that line so nothing sits between the header separator and
		// the first entry.
		listView := m.osList.View()
		if nl := strings.IndexByte(listView, '\n'); nl >= 0 && strings.TrimSpace(listView[:nl]) == "" {
			listView = listView[nl+1:]
		}
		content.WriteString(listView + "\n")
		if warning := m.selectedOsEndOfInstallWarning(); warning != "" {
			content.WriteString(wizardErrorStyle.Render("⚠  "+warning+"  ⚠") + "\n")
		}
		content.WriteString(wizardHintStyle.Render(m.osTechnicalModeHint()))
	case reinstallStepInputs:
		content.WriteString(m.renderInputsStep())
	case reinstallStepEnum:
		content.WriteString(m.renderEnumStep())
	case reinstallStepSSHKey:
		content.WriteString(m.renderSSHKeyStep())
	case reinstallStepKeyValue:
		content.WriteString(m.renderKeyValueStep())
	case reinstallStepKeyValueEdit:
		content.WriteString(m.renderKeyValueEditStep())
	case reinstallStepDiskGroup:
		content.WriteString(m.renderDiskGroupStep())
	case reinstallStepDiskCount:
		content.WriteString(m.renderDiskCountStep())
	case reinstallStepPartitionMode:
		content.WriteString(m.renderPartitionModeStep())
	case reinstallStepScheme:
		content.WriteString(m.renderSchemeStep())
	case reinstallStepLayout:
		content.WriteString(m.renderLayoutStep())
	case reinstallStepLayoutEdit:
		content.WriteString(m.renderLayoutEditStep())
	case reinstallStepConfirm:
		content.WriteString(m.renderConfirmStep())
	case reinstallStepSavePath:
		content.WriteString(m.renderSavePathStep())
	}

	return content.String()
}

func (m *reinstallWizardModel) stepLabel() string {
	switch m.step {
	case reinstallStepOs:
		return "Step 1/4 — Operating system"
	case reinstallStepInputs, reinstallStepEnum, reinstallStepSSHKey,
		reinstallStepKeyValue, reinstallStepKeyValueEdit:
		return "Step 2/4 — Installation parameters"
	case reinstallStepDiskGroup, reinstallStepDiskCount, reinstallStepPartitionMode,
		reinstallStepScheme, reinstallStepLayout, reinstallStepLayoutEdit:
		return "Step 3/4 — Storage"
	default:
		return "Step 4/4 — Summary"
	}
}

func (m *reinstallWizardModel) renderError() string {
	if m.errorMsg == "" {
		return ""
	}

	return wizardErrorStyle.Render("Error: "+m.errorMsg) + "\n\n"
}

func (m *reinstallWizardModel) result() reinstallWizardResult {
	jbodOnly := false
	if group := m.selectedDiskGroup(); group != nil {
		jbodOnly = group.hasHardwareRaid()
	}

	return reinstallWizardResult{
		osName:         m.osName,
		customizations: m.customizations,
		diskGroupID:    m.selectedDiskGroupID(),
		diskCount:      m.diskCount,
		schemeName:     m.schemeName,
		layout:         m.layout,
		jbodOnly:       jbodOnly,
	}
}

// runReinstallWizard walks the user through the installation parameters and
// returns the request body it built. launch tells whether the reinstallation was
// confirmed, savedPath is set when the parameters were written to a file: both
// are independent from each other.
func runReinstallWizard(serviceName string) (body map[string]any, launch bool, savedPath string, err error) {
	model := newReinstallWizardModel(serviceName)

	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		return nil, false, "", fmt.Errorf("failed to run wizard: %w", err)
	}

	if model.err != nil {
		return nil, false, "", model.err
	}

	return buildReinstallPayload(model.result()), model.launch, model.savedPath, nil
}
