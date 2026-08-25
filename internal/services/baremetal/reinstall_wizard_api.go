// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"encoding/json"
	"fmt"
	"maps"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// osInput is a question the API asks for the selected OS.
type osInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Default     string   `json:"default"`
	Mandatory   bool     `json:"mandatory"`
	Enum        []string `json:"enum"`
}

type osDetails struct {
	Filesystems           []string  `json:"filesystems"`
	Inputs                []osInput `json:"inputs"`
	LvmReady              bool      `json:"lvmReady"`
	NoPartitioning        bool      `json:"noPartitioning"`
	RootMountpoint        string    `json:"rootMountpoint"`
	SoftRaidOnlyMirroring bool      `json:"softRaidOnlyMirroring"`
}

type schemePartition struct {
	Filesystem string `json:"filesystem"`
	Mountpoint string `json:"mountpoint"`
	Order      int    `json:"order"`
	Raid       any    `json:"raid"`
	Size       any    `json:"size"`
	Type       string `json:"type"`
	VolumeName string `json:"volumeName"`
}

type osPartitionScheme struct {
	Name       string
	Priority   int
	Partitions []schemePartition
}

type diskGroup struct {
	DiskGroupID    int    `json:"diskGroupId"`
	NumberOfDisks  int    `json:"numberOfDisks"`
	DiskSize       any    `json:"diskSize"`
	DiskType       string `json:"diskType"`
	Description    string `json:"description"`
	RaidController string `json:"raidController"`
}

// hasHardwareRaid reports whether a disk group is managed by a hardware RAID
// controller. The wizard doesn't support configuring it, so such a group is
// always installed in JBOD mode (storage[].hardwareRaid[].raidLevel: null).
func (g diskGroup) hasHardwareRaid() bool {
	return g.RaidController != ""
}

type hardwareSpecifications struct {
	DefaultDiskGroupID int         `json:"defaultDiskGroupId"`
	DiskGroups         []diskGroup `json:"diskGroups"`
}

// layoutPartition is one entry of the custom partitioning layout being edited.
type layoutPartition struct {
	FileSystem string
	MountPoint string
	RaidLevel  *int
	Size       int
	LvName     string
	ZpName     string
}

// raidLevels are the values accepted by the API for a software RAID.
var raidLevels = []int{0, 1, 5, 6, 7, 10}

func fetchOsDetails(osName string) (*osDetails, error) {
	path := fmt.Sprintf("/v1/dedicated/installationTemplate/%s", url.PathEscape(osName))

	var details osDetails
	if err := httpLib.Client.Get(path, &details); err != nil {
		return nil, fmt.Errorf("failed to fetch details of OS %s: %w", osName, err)
	}

	return &details, nil
}

func fetchPartitionSchemes(osName string) ([]osPartitionScheme, error) {
	basePath := fmt.Sprintf("/v1/dedicated/installationTemplate/%s/partitionScheme", url.PathEscape(osName))

	var names []string
	if err := httpLib.Client.Get(basePath, &names); err != nil {
		return nil, fmt.Errorf("failed to fetch partitioning schemes: %w", err)
	}

	schemes := make([]osPartitionScheme, 0, len(names))
	for _, name := range names {
		schemePath := fmt.Sprintf("%s/%s", basePath, url.PathEscape(name))

		var scheme struct {
			Name     string `json:"name"`
			Priority int    `json:"priority"`
		}
		if err := httpLib.Client.Get(schemePath, &scheme); err != nil {
			return nil, fmt.Errorf("failed to fetch partitioning scheme %s: %w", name, err)
		}
		if scheme.Name == "" {
			scheme.Name = name
		}

		var mountpoints []string
		if err := httpLib.Client.Get(schemePath+"/partition", &mountpoints); err != nil {
			return nil, fmt.Errorf("failed to fetch partitions of scheme %s: %w", name, err)
		}

		partitions := make([]schemePartition, 0, len(mountpoints))
		for _, mountpoint := range mountpoints {
			var partition schemePartition
			partitionPath := fmt.Sprintf("%s/partition/%s", schemePath, url.PathEscape(mountpoint))
			if err := httpLib.Client.Get(partitionPath, &partition); err != nil {
				return nil, fmt.Errorf("failed to fetch partition %s of scheme %s: %w", mountpoint, name, err)
			}
			if partition.Mountpoint == "" {
				partition.Mountpoint = mountpoint
			}
			partitions = append(partitions, partition)
		}

		sortPartitionsByOrder(partitions)

		schemes = append(schemes, osPartitionScheme{
			Name:       scheme.Name,
			Priority:   scheme.Priority,
			Partitions: partitions,
		})
	}

	sortSchemesByPriority(schemes)

	return schemes, nil
}

func fetchHardwareSpecifications(serviceName string) (*hardwareSpecifications, error) {
	path := fmt.Sprintf("/v1/dedicated/server/%s/specifications/hardware", url.PathEscape(serviceName))

	var hardware hardwareSpecifications
	if err := httpLib.Client.Get(path, &hardware); err != nil {
		return nil, fmt.Errorf("failed to fetch hardware specifications: %w", err)
	}

	return &hardware, nil
}

func fetchSSHKeyNames() ([]string, error) {
	var names []string
	if err := httpLib.Client.Get("/v1/me/sshKey", &names); err != nil {
		return nil, fmt.Errorf("failed to fetch SSH keys: %w", err)
	}

	return names, nil
}

func fetchSSHKeyContent(keyName string) (string, error) {
	path := fmt.Sprintf("/v1/me/sshKey/%s", url.PathEscape(keyName))

	var key struct {
		KeyName string `json:"keyName"`
		Key     string `json:"key"`
	}
	if err := httpLib.Client.Get(path, &key); err != nil {
		return "", fmt.Errorf("failed to fetch SSH key %s: %w", keyName, err)
	}

	return key.Key, nil
}

func sortPartitionsByOrder(partitions []schemePartition) {
	for i := 1; i < len(partitions); i++ {
		for j := i; j > 0 && partitions[j].Order < partitions[j-1].Order; j-- {
			partitions[j], partitions[j-1] = partitions[j-1], partitions[j]
		}
	}
}

// sortSchemesByPriority orders schemes from highest to lowest priority, so
// the scheme screen lists them that way and defaults its selection to the
// first (highest-priority) entry.
func sortSchemesByPriority(schemes []osPartitionScheme) {
	for i := 1; i < len(schemes); i++ {
		for j := i; j > 0 && schemes[j].Priority > schemes[j-1].Priority; j-- {
			schemes[j], schemes[j-1] = schemes[j-1], schemes[j]
		}
	}
}

// findSchemeByName looks up a scheme by name, so its partitions can be shown
// (e.g. on the confirm screen) alongside the name the user picked.
func findSchemeByName(schemes []osPartitionScheme, name string) *osPartitionScheme {
	for i := range schemes {
		if schemes[i].Name == name {
			return &schemes[i]
		}
	}

	return nil
}

// defaultPartitionScheme returns the scheme with the highest priority, which is
// the one the API would apply if no scheme was given.
func defaultPartitionScheme(schemes []osPartitionScheme) *osPartitionScheme {
	var best *osPartitionScheme
	for i := range schemes {
		if best == nil || schemes[i].Priority > best.Priority {
			best = &schemes[i]
		}
	}

	return best
}

// schemePartitionToLayout converts a read-only scheme partition into the same
// shape a custom layout partition uses, so both share one attribute/display
// mapping instead of two.
func schemePartitionToLayout(partition schemePartition) layoutPartition {
	converted := layoutPartition{
		FileSystem: partition.Filesystem,
		MountPoint: partition.Mountpoint,
		Size:       intValue(partition.Size),
	}
	if raid, ok := optionalIntValue(partition.Raid); ok {
		converted.RaidLevel = &raid
	}
	if partition.Type == "lv" {
		converted.LvName = partition.VolumeName
	}

	return converted
}

// layoutFromScheme seeds an editable layout from the partitions of a scheme.
func layoutFromScheme(scheme *osPartitionScheme) []layoutPartition {
	if scheme == nil {
		return nil
	}

	layout := make([]layoutPartition, 0, len(scheme.Partitions))
	for _, partition := range scheme.Partitions {
		layout = append(layout, schemePartitionToLayout(partition))
	}

	return layout
}

// partitionAttr is one labeled attribute of a partition, shown as its own
// "Label : value" row instead of packed into a single dense line.
type partitionAttr struct {
	Label string
	Value string
}

func layoutPartitionAttrs(partition layoutPartition) []partitionAttr {
	sizeValue := fmt.Sprintf("%d MiB", partition.Size)
	if partition.Size == 0 {
		sizeValue = "remaining space"
	}

	attrs := []partitionAttr{
		{"Mount point", partition.MountPoint},
		{"File system", partition.FileSystem},
		{"Size", sizeValue},
	}
	if partition.RaidLevel != nil {
		attrs = append(attrs, partitionAttr{"RAID level", strconv.Itoa(*partition.RaidLevel)})
	}
	if partition.LvName != "" {
		attrs = append(attrs, partitionAttr{"LVM volume", partition.LvName})
	}
	if partition.ZpName != "" {
		attrs = append(attrs, partitionAttr{"ZFS pool", partition.ZpName})
	}

	return attrs
}

// maxPartitionBoxWidth is the largest a partition box is allowed to be,
// border to border.
const maxPartitionBoxWidth = 50

// partitionRowWidth measures how wide a single "Label : value" row of a
// partition box needs to be, given the label column shared by its own
// attributes.
func partitionRowWidth(attrs []partitionAttr) int {
	labelWidth := 0
	for _, attr := range attrs {
		labelWidth = max(labelWidth, len(attr.Label))
	}

	width := 0
	for _, attr := range attrs {
		width = max(width, len(fmt.Sprintf("%-*s : %s", labelWidth, attr.Label, attr.Value)))
	}

	return width
}

// partitionBoxWidth is the row width every box of a list of partitions should
// share: wide enough for the widest of them, capped so no box exceeds
// maxPartitionBoxWidth.
func partitionBoxWidth(attrSets [][]partitionAttr) int {
	width := 0
	for _, attrs := range attrSets {
		width = max(width, partitionRowWidth(attrs))
	}

	return min(width, maxPartitionBoxWidth-4)
}

// layoutPartitionsBoxWidth and schemePartitionsBoxWidth give the shared box
// width for a list of partitions shown together, so every box in that list is
// exactly as wide as the widest one needs to be.
func layoutPartitionsBoxWidth(partitions []layoutPartition) int {
	attrSets := make([][]partitionAttr, len(partitions))
	for i, partition := range partitions {
		attrSets[i] = layoutPartitionAttrs(partition)
	}

	return partitionBoxWidth(attrSets)
}

func schemePartitionsBoxWidth(partitions []schemePartition) int {
	attrSets := make([][]partitionAttr, len(partitions))
	for i, partition := range partitions {
		attrSets[i] = layoutPartitionAttrs(schemePartitionToLayout(partition))
	}

	return partitionBoxWidth(attrSets)
}

// schemeListHeight is how many rows the scheme picker's name list itself
// needs: enough for its items (capped), leaving the rest of the screen to the
// scrollable detail viewport below it.
func schemeListHeight(schemeCount int) int {
	return min(max(schemeCount, 1)+1, 6)
}

// hardwareRaidWarning explains why a disk group's hardware RAID controller is
// being ignored, or "" when the group has none.
func hardwareRaidWarning(group diskGroup) string {
	if !group.hasHardwareRaid() {
		return ""
	}

	return fmt.Sprintf(
		"Disk group %d has a hardware RAID controller (%s), which the OVHcloud CLI does not support configuring — it will be installed in JBOD mode instead",
		group.DiskGroupID, group.RaidController)
}

// osEndOfInstallDateLayout is the format of the endOfInstall field returned
// by installationTemplate/templateInfos: a plain calendar date, no time.
const osEndOfInstallDateLayout = "2006-01-02"

// osEndOfInstallWarning warns once an OS's end of install date (endOfInstall,
// formatted as osEndOfInstallDateLayout) is less than 6 months away from now,
// so it isn't picked for a reinstallation that may need repeating later.
// Empty, unparseable, or comfortably far away dates warn about nothing.
// includeDate adds the date to the message; leave it out when the date is
// already shown right next to the warning.
func osEndOfInstallWarning(endOfInstall string, now time.Time, includeDate bool) string {
	if endOfInstall == "" {
		return ""
	}

	date, err := time.Parse(osEndOfInstallDateLayout, endOfInstall)
	if err != nil {
		return ""
	}

	if date.After(now.AddDate(0, 6, 0)) {
		return ""
	}

	if !includeDate {
		return "This OS will soon no longer be available for reinstallation"
	}

	return fmt.Sprintf(
		"This OS will soon no longer be available for reinstallation (end of install: %s)",
		date.Format(osEndOfInstallDateLayout))
}

// partitionTags names the badges shown above a partition's box: "SYSTEM"
// when its mount point is the OS's rootMountpoint, "LVM" when it has a
// logical volume name.
func partitionTags(partition layoutPartition, rootMountpoint string) []string {
	var tags []string
	if rootMountpoint != "" && partition.MountPoint == rootMountpoint {
		tags = append(tags, "SYSTEM")
	}
	if partition.LvName != "" {
		tags = append(tags, "LVM")
	}

	return tags
}

// formatPartitionBox draws a partition's attributes as a labeled ASCII box of
// the given row width (rows longer than that are truncated, shorter ones
// padded), its tags shown right-aligned as the box's own first row when there
// are any, e.g.:
//
//	┌──────────────────────────┐
//	│           [SYSTEM] [LVM] │
//	│ Mount point : /           │
//	│ File system : ext4        │
//	│ Size        : remaining space │
//	└──────────────────────────┘
func formatPartitionBox(attrs []partitionAttr, rowWidth int, tags []string) string {
	if len(attrs) == 0 {
		return ""
	}

	labelWidth := 0
	for _, attr := range attrs {
		labelWidth = max(labelWidth, len(attr.Label))
	}

	var rows []string
	if len(tags) > 0 {
		var tagRow strings.Builder
		for _, tag := range tags {
			fmt.Fprintf(&tagRow, "[%s] ", tag)
		}
		text := truncateLine(strings.TrimRight(tagRow.String(), " "), rowWidth)
		// Pre-padded to rowWidth here, right-aligned: the render loop below
		// pads every row left-aligned, which is a no-op once already full width.
		rows = append(rows, fmt.Sprintf("%*s", rowWidth, text))
	}
	for _, attr := range attrs {
		row := fmt.Sprintf("%-*s : %s", labelWidth, attr.Label, attr.Value)
		rows = append(rows, truncateLine(row, rowWidth))
	}

	border := strings.Repeat("─", rowWidth+2)

	var b strings.Builder
	fmt.Fprintf(&b, "┌%s┐\n", border)
	for _, row := range rows {
		fmt.Fprintf(&b, "│ %-*s │\n", rowWidth, row)
	}
	fmt.Fprintf(&b, "└%s┘", border)

	return b.String()
}

// describeLayoutPartition and describeSchemePartition render one partition as
// a labeled box of the given shared width (see layoutPartitionsBoxWidth /
// schemePartitionsBoxWidth), tagged against the OS's rootMountpoint, sharing
// the same attribute mapping via schemePartitionToLayout.
func describeLayoutPartition(partition layoutPartition, rowWidth int, rootMountpoint string) string {
	return formatPartitionBox(layoutPartitionAttrs(partition), rowWidth, partitionTags(partition, rootMountpoint))
}

func describeSchemePartition(partition schemePartition, rowWidth int, rootMountpoint string) string {
	converted := schemePartitionToLayout(partition)
	return formatPartitionBox(layoutPartitionAttrs(converted), rowWidth, partitionTags(converted, rootMountpoint))
}

// reinstallWizardResult holds everything the wizard collected, in the order the
// screens collected it.
type reinstallWizardResult struct {
	osName         string
	customizations map[string]any
	diskGroupID    int
	diskCount      int
	schemeName     string
	layout         []layoutPartition
	// jbodOnly is set when the selected disk group has a hardware RAID
	// controller: the wizard doesn't support configuring it, so the disks are
	// always installed in JBOD mode instead.
	jbodOnly bool
}

func buildReinstallPayload(result reinstallWizardResult) map[string]any {
	partitioning := map[string]any{
		"disks": result.diskCount,
	}

	if len(result.layout) > 0 {
		layout := make([]any, 0, len(result.layout))
		for _, partition := range result.layout {
			entry := map[string]any{
				"fileSystem": partition.FileSystem,
				"mountPoint": partition.MountPoint,
				"size":       partition.Size,
			}
			if partition.RaidLevel != nil {
				entry["raidLevel"] = *partition.RaidLevel
			}

			extras := map[string]any{}
			if partition.LvName != "" {
				extras["lv"] = map[string]any{"name": partition.LvName}
			}
			if partition.ZpName != "" {
				extras["zp"] = map[string]any{"name": partition.ZpName}
			}
			if len(extras) > 0 {
				entry["extras"] = extras
			}

			layout = append(layout, entry)
		}
		partitioning["layout"] = layout
	} else if result.schemeName != "" {
		partitioning["schemeName"] = result.schemeName
	}

	storageEntry := map[string]any{
		"diskGroupId":  result.diskGroupID,
		"partitioning": partitioning,
	}
	if result.jbodOnly {
		// A null raidLevel tells the API to configure this disk group's
		// hardware RAID controller in JBOD mode instead of an actual array,
		// since the wizard has no UI to configure the controller itself.
		storageEntry["hardwareRaid"] = []any{map[string]any{"raidLevel": nil}}
	}

	body := map[string]any{
		"operatingSystem": result.osName,
		"storage":         []any{storageEntry},
	}

	if len(result.customizations) > 0 {
		body["customizations"] = result.customizations
	}

	return body
}

// validateLayout enforces the rules the API expects from a custom layout, so
// that the wizard does not send a request that is known to be refused.
func validateLayout(partitions []layoutPartition, rootMountpoint string) error {
	if len(partitions) == 0 {
		return fmt.Errorf("at least one partition is required")
	}

	var (
		swapCount     int
		fillCount     int
		mountpoints   = map[string]bool{}
		hasRootMount  bool
		trimmedRoot   = strings.TrimSpace(rootMountpoint)
		duplicateName string
	)

	for _, partition := range partitions {
		if partition.FileSystem == "" {
			return fmt.Errorf("partition %q has no file system", partition.MountPoint)
		}
		if partition.MountPoint == "" {
			return fmt.Errorf("a partition has no mount point")
		}

		if partition.FileSystem == "swap" || partition.MountPoint == "swap" {
			if partition.FileSystem != "swap" || partition.MountPoint != "swap" {
				return fmt.Errorf("a swap partition must use both the swap file system and the swap mount point")
			}
			if partition.RaidLevel != nil {
				return fmt.Errorf("a swap partition cannot have a RAID level")
			}
			swapCount++
		}

		if partition.Size == 0 {
			fillCount++
		}

		if mountpoints[partition.MountPoint] {
			duplicateName = partition.MountPoint
		}
		mountpoints[partition.MountPoint] = true

		if partition.MountPoint == trimmedRoot {
			hasRootMount = true
		}
	}

	if swapCount > 1 {
		return fmt.Errorf("only one swap partition is allowed")
	}
	if fillCount > 1 {
		return fmt.Errorf("only one partition can have a size of 0 (fill remaining space)")
	}
	if duplicateName != "" {
		return fmt.Errorf("mount point %q is used by several partitions", duplicateName)
	}
	if trimmedRoot != "" && !hasRootMount {
		return fmt.Errorf("a partition mounted on %s is required", trimmedRoot)
	}

	return nil
}

func intValue(value any) int {
	result, _ := optionalIntValue(value)
	return result
}

func optionalIntValue(value any) (int, bool) {
	switch typed := value.(type) {
	case nil:
		return 0, false
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil {
			return 0, false
		}
		return int(parsed), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err != nil {
			return 0, false
		}
		return parsed, true
	case map[string]any:
		return optionalIntValue(typed["value"])
	}

	return 0, false
}

// formatQuantity renders a value the API may return either as a plain number or
// as a {unit, value} object.
func formatQuantity(value any) string {
	if quantity, ok := value.(map[string]any); ok {
		amount, found := optionalIntValue(quantity["value"])
		if !found {
			return ""
		}
		if unit, ok := quantity["unit"].(string); ok {
			return fmt.Sprintf("%d %s", amount, unit)
		}
		return strconv.Itoa(amount)
	}

	if amount, ok := optionalIntValue(value); ok {
		return strconv.Itoa(amount)
	}

	return ""
}

// keyValuePair is one entry being edited for a "keyValue" customization input.
type keyValuePair struct {
	Key   string
	Value string
}

// kvPairsFromAnswer turns a stored "keyValue" answer back into an editable,
// stably ordered list of pairs.
func kvPairsFromAnswer(answer any) []keyValuePair {
	values, ok := answer.(map[string]string)
	if !ok {
		return nil
	}

	pairs := make([]keyValuePair, 0, len(values))
	for _, key := range slices.Sorted(maps.Keys(values)) {
		pairs = append(pairs, keyValuePair{Key: key, Value: values[key]})
	}

	return pairs
}

// kvMapFromPairs builds the map to store as the customization answer, dropping
// pairs left with an empty key. A nil result means "no answer" (untouched).
func kvMapFromPairs(pairs []keyValuePair) map[string]string {
	values := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		key := strings.TrimSpace(pair.Key)
		if key == "" {
			continue
		}
		values[key] = pair.Value
	}

	if len(values) == 0 {
		return nil
	}

	return values
}

// seedCustomizations gives every input of the form a live answer as soon as it
// is displayed, so typing (or cycling a choice) updates it in place instead of
// only when the question is left.
func seedCustomizations(inputs []osInput) map[string]any {
	answers := make(map[string]any, len(inputs))

	for _, input := range inputs {
		switch input.Type {
		case "keyValue":
			answers[input.Name] = map[string]string{}
		case "boolean":
			answers[input.Name] = strconv.FormatBool(isTrueValue(input.Default))
		case "enum":
			value := input.Default
			if len(input.Enum) > 0 && !slices.Contains(input.Enum, value) {
				value = input.Enum[0]
			}
			answers[input.Name] = value
		default:
			answers[input.Name] = input.Default
		}
	}

	return answers
}

// finalizeCustomizations turns the live answers of the form into what the
// request carries: booleans become real booleans, and an answer still equal to
// the value the API proposed is dropped, so an untouched installation sends
// nothing for it.
func finalizeCustomizations(inputs []osInput, answers map[string]any) map[string]any {
	customizations := make(map[string]any, len(answers))

	for _, input := range inputs {
		answer, ok := answers[input.Name]
		if !ok {
			continue
		}

		switch input.Type {
		case "keyValue":
			values, _ := answer.(map[string]string)
			if len(values) == 0 {
				continue
			}
			customizations[input.Name] = values
		case "boolean":
			value := isTrueValue(fmt.Sprintf("%v", answer))
			if value == isTrueValue(input.Default) {
				continue
			}
			customizations[input.Name] = value
		default:
			value := fmt.Sprintf("%v", answer)
			if value == input.Default {
				continue
			}
			customizations[input.Name] = value
		}
	}

	return customizations
}

// validateAnswer checks one form answer against what its input declares: a
// mandatory question must be answered, and a typed value must have the shape
// its type expects.
func validateAnswer(input osInput, answer any) error {
	switch input.Type {
	case "enum", "boolean":
		return nil
	case "keyValue":
		if values, _ := answer.(map[string]string); input.Mandatory && len(values) == 0 {
			return fmt.Errorf("%s is mandatory", input.Name)
		}
		return nil
	}

	value := strings.TrimSpace(fmt.Sprintf("%v", answer))
	if value == "" {
		if input.Mandatory {
			return fmt.Errorf("%s is mandatory", input.Name)
		}
		return nil
	}

	if err := validateInputValue(input.Type, value); err != nil {
		return fmt.Errorf("%s: %w", input.Name, err)
	}

	return nil
}

// firstInvalidAnswer reports the first question the form cannot accept, so the
// focus can be moved there along with the error instead of only telling the
// user that something is wrong.
func firstInvalidAnswer(inputs []osInput, answers map[string]any) (int, error) {
	for i, input := range inputs {
		if err := validateAnswer(input, answers[input.Name]); err != nil {
			return i, err
		}
	}

	return -1, nil
}

func isTrueValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes":
		return true
	}

	return false
}

// answerSummary renders a customization answer for a summary line: masked for
// a SSH key, compacted for a key/value list, shortened for a multiline text.
func answerSummary(inputType string, value any) string {
	switch inputType {
	case "sshPubKey":
		return maskSecret(fmt.Sprintf("%v", value), 25)
	case "keyValue":
		if values, ok := value.(map[string]string); ok {
			return formatKeyValueSummary(values)
		}
	case "text":
		return textSummary(fmt.Sprintf("%v", value))
	}

	return fmt.Sprintf("%v", value)
}

// compactSummary is answerSummary kept on a single line, for the form rows
// which give one line to each unfocused question.
func compactSummary(inputType string, value any) string {
	if text, ok := value.(string); ok && text == "" {
		return ""
	}

	summary := answerSummary(inputType, value)
	if inputType == "text" {
		summary = truncateLine(strings.ReplaceAll(summary, "\n", " ↵ "), textSummaryLineWidth)
	}

	return summary
}

// visibleTail keeps the end of a long value in sight inside a fixed-width
// input box, which is the part that matters while typing.
func visibleTail(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}

	return "…" + string(runes[len(runes)-width+1:])
}

// maskSecret hides a sensitive value on the confirm screen: 5 asterisks
// followed by only the last `visible` characters, so the customer can
// recognize which key they picked without the full value being shown.
func maskSecret(value string, visible int) string {
	if len(value) <= visible {
		return "*****" + value
	}

	return "*****" + value[len(value)-visible:]
}

// formatKeyValueSummary renders a "keyValue" answer as a compact, stably
// ordered "k1=v1, k2=v2" string instead of Go's default map formatting.
func formatKeyValueSummary(values map[string]string) string {
	if len(values) == 0 {
		return ""
	}

	parts := make([]string, 0, len(values))
	for _, key := range slices.Sorted(maps.Keys(values)) {
		parts = append(parts, fmt.Sprintf("%s=%s", key, values[key]))
	}

	return strings.Join(parts, ", ")
}

const textSummaryLineWidth = 50

// textSummary renders a multiline "text" answer for the confirm screen: the
// first 3 lines, then "..." and the last 2 lines when there's more than
// that, each line capped to textSummaryLineWidth so it doesn't break the
// screen's layout.
func textSummary(value string) string {
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		lines[i] = truncateLine(line, textSummaryLineWidth)
	}

	const head, tail = 3, 2
	if len(lines) <= head+tail {
		return strings.Join(lines, "\n")
	}

	return strings.Join(lines[:head], "\n") + "\n...\n" + strings.Join(lines[len(lines)-tail:], "\n")
}

func truncateLine(line string, width int) string {
	runes := []rune(line)
	if len(runes) <= width {
		return line
	}

	return string(runes[:width-1]) + "…"
}

var hexstringPattern = regexp.MustCompile(`^[0-9a-fA-F]+$`)

// validateInputValue rejects a value that doesn't match the shape the API
// expects for the given input type, so the wizard catches the mistake right
// away instead of the user finding out from an API error. Input types with no
// specific shape to check (string, text, number, hostname, keyValue, ...)
// are left alone here.
func validateInputValue(inputType, value string) error {
	switch inputType {
	case "date":
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return fmt.Errorf("expected a date in the YYYY-MM-DD format")
		}
	case "time":
		if _, err := time.Parse("15:04:05", value); err != nil {
			if _, err := time.Parse("15:04", value); err != nil {
				return fmt.Errorf("expected a time in the HH:MM or HH:MM:SS format")
			}
		}
	case "email":
		if _, err := mail.ParseAddress(value); err != nil {
			return fmt.Errorf("expected a valid email address")
		}
	case "hexstring":
		if !hexstringPattern.MatchString(value) {
			return fmt.Errorf("expected an hexadecimal string")
		}
	case "ip":
		if net.ParseIP(value) == nil {
			return fmt.Errorf("expected a valid IP address")
		}
	case "url":
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("expected a valid URL")
		}
	case "uuid":
		if _, err := uuid.Parse(value); err != nil {
			return fmt.Errorf("expected a valid UUID")
		}
	}

	return nil
}
