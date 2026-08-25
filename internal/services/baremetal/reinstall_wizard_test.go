// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/go-ovh/ovh"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// withMockedAPI points the shared client at httpmock for the duration of a test.
func withMockedAPI(t *testing.T) {
	t.Helper()
	httpmock.Activate(t)

	origClient := httpLib.Client
	client, err := ovh.NewClient("ovh-eu", "app_key", "app_secret", "consumer_key")
	td.Require(t).CmpNoError(err)
	httpLib.Client = client

	t.Cleanup(func() {
		httpLib.Client = origClient
	})

	// go-ovh computes its clock delta before the first signed call.
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
		httpmock.NewStringResponder(200, "0"))
}

func TestFetchCompatibleOsNames(t *testing.T) {
	withMockedAPI(t)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/srv/install/compatibleTemplates",
		httpmock.NewStringResponder(200, `{"ovh": ["debian12_64", "ubuntu2404-server_64"]}`))

	names, err := fetchCompatibleOsNames("srv")

	td.Require(t).CmpNoError(err)
	td.Cmp(t, names, []string{"debian12_64", "ubuntu2404-server_64"})
}

// The response only carries an "ovh" key: an empty one is not a failure.
func TestFetchCompatibleOsNames_EmptyList(t *testing.T) {
	withMockedAPI(t)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/srv/install/compatibleTemplates",
		httpmock.NewStringResponder(200, `{"ovh": []}`))

	names, err := fetchCompatibleOsNames("srv")

	td.Require(t).CmpNoError(err)
	td.Cmp(t, names, td.Empty())
}

func TestFetchCompatibleOsNames_APIError(t *testing.T) {
	withMockedAPI(t)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/srv/install/compatibleTemplates",
		httpmock.NewStringResponder(404, `{"message": "not found"}`))

	_, err := fetchCompatibleOsNames("srv")

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains("failed to fetch compatible OSes"))
}

func TestBuildReinstallPayload_WithSchemeName(t *testing.T) {
	body := buildReinstallPayload(reinstallWizardResult{
		osName:         "debian12_64",
		customizations: map[string]any{"hostname": "my-server"},
		diskGroupID:    1,
		diskCount:      2,
		schemeName:     "default",
	})

	td.Cmp(t, body, map[string]any{
		"operatingSystem": "debian12_64",
		"customizations":  map[string]any{"hostname": "my-server"},
		"storage": []any{
			map[string]any{
				"diskGroupId": 1,
				"partitioning": map[string]any{
					"disks":      2,
					"schemeName": "default",
				},
			},
		},
	})
}

// A disk group with a hardware RAID controller isn't configurable by the
// wizard, so the payload forces JBOD mode instead of leaving the controller's
// own default in place.
func TestBuildReinstallPayload_JbodOnly(t *testing.T) {
	body := buildReinstallPayload(reinstallWizardResult{
		osName:      "debian12_64",
		diskGroupID: 1,
		diskCount:   2,
		schemeName:  "default",
		jbodOnly:    true,
	})

	td.Cmp(t, body, map[string]any{
		"operatingSystem": "debian12_64",
		"storage": []any{
			map[string]any{
				"diskGroupId":  1,
				"partitioning": map[string]any{"disks": 2, "schemeName": "default"},
				"hardwareRaid": []any{map[string]any{"raidLevel": nil}},
			},
		},
	})
}

// A custom layout and a scheme name are mutually exclusive: the layout wins
// because it is the only one the user built by hand.
func TestBuildReinstallPayload_WithCustomLayout(t *testing.T) {
	raid := 1
	body := buildReinstallPayload(reinstallWizardResult{
		osName:      "debian12_64",
		diskGroupID: 0,
		diskCount:   2,
		schemeName:  "default",
		layout: []layoutPartition{
			{FileSystem: "ext4", MountPoint: "/", Size: 20480, RaidLevel: &raid, LvName: "root"},
			{FileSystem: "swap", MountPoint: "swap", Size: 2048},
			{FileSystem: "zfs", MountPoint: "/data", Size: 0, ZpName: "poule"},
		},
	})

	td.Cmp(t, body, map[string]any{
		"operatingSystem": "debian12_64",
		"storage": []any{
			map[string]any{
				"diskGroupId": 0,
				"partitioning": map[string]any{
					"disks": 2,
					"layout": []any{
						map[string]any{
							"fileSystem": "ext4",
							"mountPoint": "/",
							"size":       20480,
							"raidLevel":  1,
							"extras":     map[string]any{"lv": map[string]any{"name": "root"}},
						},
						map[string]any{
							"fileSystem": "swap",
							"mountPoint": "swap",
							"size":       2048,
						},
						map[string]any{
							"fileSystem": "zfs",
							"mountPoint": "/data",
							"size":       0,
							"extras":     map[string]any{"zp": map[string]any{"name": "poule"}},
						},
					},
				},
			},
		},
	})
}

// Answers left at their default value are not recorded, so an untouched
// installation must not send an empty customizations object.
func TestBuildReinstallPayload_WithoutCustomizations(t *testing.T) {
	body := buildReinstallPayload(reinstallWizardResult{
		osName:         "debian12_64",
		customizations: map[string]any{},
		diskCount:      1,
	})

	td.Cmp(t, body, td.Not(td.ContainsKey("customizations")))
}

func TestValidateLayout(t *testing.T) {
	raid := 1

	testCases := []struct {
		name       string
		partitions []layoutPartition
		root       string
		expected   any
	}{
		{
			name: "valid layout",
			partitions: []layoutPartition{
				{FileSystem: "ext4", MountPoint: "/", Size: 0},
				{FileSystem: "swap", MountPoint: "swap", Size: 2048},
			},
			root:     "/",
			expected: nil,
		},
		{
			name: "two swap partitions",
			partitions: []layoutPartition{
				{FileSystem: "ext4", MountPoint: "/", Size: 0},
				{FileSystem: "swap", MountPoint: "swap", Size: 2048},
				{FileSystem: "swap", MountPoint: "swap", Size: 1024},
			},
			root:     "/",
			expected: td.Contains("only one swap partition"),
		},
		{
			name: "swap file system without the swap mount point",
			partitions: []layoutPartition{
				{FileSystem: "ext4", MountPoint: "/", Size: 0},
				{FileSystem: "swap", MountPoint: "/swap", Size: 2048},
			},
			root:     "/",
			expected: td.Contains("swap file system and the swap mount point"),
		},
		{
			name: "swap mount point without the swap file system",
			partitions: []layoutPartition{
				{FileSystem: "ext4", MountPoint: "/", Size: 0},
				{FileSystem: "ext4", MountPoint: "swap", Size: 2048},
			},
			root:     "/",
			expected: td.Contains("swap file system and the swap mount point"),
		},
		{
			name: "swap partition with a RAID level",
			partitions: []layoutPartition{
				{FileSystem: "ext4", MountPoint: "/", Size: 0},
				{FileSystem: "swap", MountPoint: "swap", Size: 2048, RaidLevel: &raid},
			},
			root:     "/",
			expected: td.Contains("swap partition cannot have a RAID level"),
		},
		{
			name: "several partitions filling the remaining space",
			partitions: []layoutPartition{
				{FileSystem: "ext4", MountPoint: "/", Size: 0},
				{FileSystem: "ext4", MountPoint: "/data", Size: 0},
			},
			root:     "/",
			expected: td.Contains("only one partition can have a size of 0"),
		},
		{
			name: "duplicated mount point",
			partitions: []layoutPartition{
				{FileSystem: "ext4", MountPoint: "/", Size: 1024},
				{FileSystem: "ext4", MountPoint: "/", Size: 2048},
			},
			root:     "/",
			expected: td.Contains(`mount point "/" is used by several partitions`),
		},
		{
			name: "missing root mount point",
			partitions: []layoutPartition{
				{FileSystem: "ext4", MountPoint: "/data", Size: 0},
			},
			root:     "/",
			expected: td.Contains("a partition mounted on / is required"),
		},
		{
			name:       "empty layout",
			partitions: nil,
			root:       "/",
			expected:   td.Contains("at least one partition is required"),
		},
		{
			name: "partition without a file system",
			partitions: []layoutPartition{
				{MountPoint: "/", Size: 0},
			},
			root:     "/",
			expected: td.Contains("has no file system"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateLayout(testCase.partitions, testCase.root)

			if testCase.expected == nil {
				td.CmpNoError(t, err)
				return
			}

			td.Require(t).CmpError(err)
			td.Cmp(t, err.Error(), testCase.expected)
		})
	}
}

func TestDefaultPartitionScheme(t *testing.T) {
	schemes := []osPartitionScheme{
		{Name: "custom", Priority: 1},
		{Name: "default", Priority: 5},
	}

	td.Cmp(t, defaultPartitionScheme(schemes).Name, "default")
	td.Cmp(t, defaultPartitionScheme(nil), td.Nil())
}

func TestLayoutFromScheme(t *testing.T) {
	scheme := osPartitionScheme{
		Name: "default",
		Partitions: []schemePartition{
			{Filesystem: "ext4", Mountpoint: "/", Order: 2, Raid: "1", Size: 0, Type: "lv", VolumeName: "root"},
			{Filesystem: "ext4", Mountpoint: "/boot", Order: 1, Size: 1024, Type: "primary"},
		},
	}

	raid := 1
	td.Cmp(t, layoutFromScheme(&scheme), []layoutPartition{
		{FileSystem: "ext4", MountPoint: "/", Size: 0, RaidLevel: &raid, LvName: "root"},
		{FileSystem: "ext4", MountPoint: "/boot", Size: 1024},
	}, "partitions are seeded in the order the scheme declares")
}

func TestSortSchemesByPriority(t *testing.T) {
	schemes := []osPartitionScheme{
		{Name: "low", Priority: 1},
		{Name: "high", Priority: 10},
		{Name: "mid", Priority: 5},
	}

	sortSchemesByPriority(schemes)

	td.Cmp(t, []string{schemes[0].Name, schemes[1].Name, schemes[2].Name}, []string{"high", "mid", "low"},
		"schemes are listed from highest to lowest priority, so the default selection is the highest one")
}

func TestValidateInputValue(t *testing.T) {
	cases := []struct {
		inputType string
		value     string
		wantErr   bool
	}{
		{"date", "2024-01-31", false},
		{"date", "not-a-date", true},
		{"date", "31-01-2024", true},
		{"time", "23:59", false},
		{"time", "23:59:59", false},
		{"time", "25:99", true},
		{"email", "user@example.com", false},
		{"email", "not-an-email", true},
		{"hexstring", "0a1B2c", false},
		{"hexstring", "not-hex!", true},
		{"hexstring", "", true},
		{"ip", "192.0.2.1", false},
		{"ip", "::1", false},
		{"ip", "not-an-ip", true},
		{"url", "https://example.com/path", false},
		{"url", "not-a-url", true},
		{"uuid", "123e4567-e89b-12d3-a456-426614174000", false},
		{"uuid", "not-a-uuid", true},
		{"string", "anything goes", false},
		{"text", "anything goes", false},
		{"number", "anything goes", false},
		{"hostname", "anything goes", false},
	}

	for _, c := range cases {
		err := validateInputValue(c.inputType, c.value)
		if c.wantErr {
			td.CmpError(t, err, fmt.Sprintf("%s %q should be rejected", c.inputType, c.value))
		} else {
			td.CmpNoError(t, err, fmt.Sprintf("%s %q should be accepted", c.inputType, c.value))
		}
	}
}

func TestKvPairsFromAnswer(t *testing.T) {
	td.Cmp(t, kvPairsFromAnswer(map[string]string{"b": "2", "a": "1"}), []keyValuePair{
		{Key: "a", Value: "1"},
		{Key: "b", Value: "2"},
	}, "pairs are sorted by key for a stable display order")

	td.Cmp(t, kvPairsFromAnswer(nil), td.Empty())
	td.Cmp(t, kvPairsFromAnswer("not a map"), td.Empty())
}

func TestKvMapFromPairs(t *testing.T) {
	td.Cmp(t, kvMapFromPairs([]keyValuePair{
		{Key: "a", Value: "1"},
		{Key: "", Value: "dropped, no key"},
		{Key: "b", Value: "2"},
	}), map[string]string{"a": "1", "b": "2"})

	td.Cmp(t, kvMapFromPairs(nil), td.Nil(), "no pairs means no answer, not an empty map")
	td.Cmp(t, kvMapFromPairs([]keyValuePair{{Key: "  "}}), td.Nil(), "a blank-only key is dropped too")
}

func TestSeedCustomizations(t *testing.T) {
	answers := seedCustomizations([]osInput{
		{Name: "hostname", Type: "string", Default: "srv"},
		{Name: "secure", Type: "boolean", Default: "yes"},
		{Name: "language", Type: "enum", Default: "fr", Enum: []string{"en", "fr"}},
		{Name: "region", Type: "enum", Default: "unknown", Enum: []string{"en", "fr"}},
		{Name: "labels", Type: "keyValue"},
		{Name: "key", Type: "sshPubKey"},
	})

	td.Cmp(t, answers, map[string]any{
		"hostname": "srv",
		"secure":   "true",
		"language": "fr",
		"region":   "en",
		"labels":   map[string]string{},
		"key":      "",
	}, "every input gets a live answer, a default outside of the enum falls back to its first value")
}

func TestFinalizeCustomizations(t *testing.T) {
	inputs := []osInput{
		{Name: "hostname", Type: "string", Default: "srv"},
		{Name: "comment", Type: "text"},
		{Name: "secure", Type: "boolean", Default: "yes"},
		{Name: "debug", Type: "boolean", Default: "false"},
		{Name: "language", Type: "enum", Default: "fr", Enum: []string{"en", "fr"}},
		{Name: "labels", Type: "keyValue"},
		{Name: "tags", Type: "keyValue"},
		{Name: "unknown", Type: "string", Default: "kept out"},
	}

	customizations := finalizeCustomizations(inputs, map[string]any{
		"hostname": "my-server",
		"comment":  "",
		"secure":   "true",
		"debug":    "true",
		"language": "fr",
		"labels":   map[string]string{},
		"tags":     map[string]string{"a": "1"},
	})

	td.Cmp(t, customizations, map[string]any{
		"hostname": "my-server",
		"debug":    true,
		"tags":     map[string]string{"a": "1"},
	}, "answers left at their default value, empty pair lists and unanswered inputs are dropped")
}

func TestValidateAnswer(t *testing.T) {
	td.CmpNoError(t, validateAnswer(osInput{Name: "language", Type: "enum", Mandatory: true}, ""),
		"a choice always has a value, there is nothing to check")

	err := validateAnswer(osInput{Name: "hostname", Type: "string", Mandatory: true}, "  ")
	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), "hostname is mandatory")

	err = validateAnswer(osInput{Name: "contact", Type: "email"}, "not-an-email")
	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains("contact: expected a valid email address"))

	td.CmpNoError(t, validateAnswer(osInput{Name: "contact", Type: "email"}, ""),
		"an optional answer left empty is not checked further")

	err = validateAnswer(osInput{Name: "labels", Type: "keyValue", Mandatory: true}, map[string]string{})
	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), "labels is mandatory")
}

func TestFirstInvalidAnswer(t *testing.T) {
	inputs := []osInput{
		{Name: "hostname", Type: "string"},
		{Name: "contact", Type: "email"},
		{Name: "comment", Type: "text", Mandatory: true},
	}

	index, err := firstInvalidAnswer(inputs, map[string]any{
		"hostname": "srv",
		"contact":  "user@example.com",
		"comment":  "hello",
	})
	td.CmpNoError(t, err)
	td.Cmp(t, index, -1)

	index, err = firstInvalidAnswer(inputs, map[string]any{
		"hostname": "srv",
		"contact":  "not-an-email",
		"comment":  "",
	})
	td.Require(t).CmpError(err)
	td.Cmp(t, index, 1, "the focus goes to the first question to fix, not to the last one")
	td.Cmp(t, err.Error(), td.Contains("contact"))
}

func TestCompactSummary(t *testing.T) {
	td.Cmp(t, compactSummary("string", ""), "", "an unanswered question has no summary")
	td.Cmp(t, compactSummary("string", "srv"), "srv")
	td.Cmp(t, compactSummary("sshPubKey", "ssh-ed25519 AAAA"), "*****ssh-ed25519 AAAA")
	td.Cmp(t, compactSummary("keyValue", map[string]string{"b": "2", "a": "1"}), "a=1, b=2")
	td.Cmp(t, compactSummary("text", "l1\nl2"), "l1 ↵ l2", "a multiline answer is kept on one line")
}

func TestVisibleTail(t *testing.T) {
	td.Cmp(t, visibleTail("short", 10), "short")
	td.Cmp(t, visibleTail("abcdefghij", 5), "…ghij", "the end of the value is what matters while typing")
}

func TestMaskSecret(t *testing.T) {
	td.Cmp(t, maskSecret("short", 25), "*****short")
	td.Cmp(t, maskSecret("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMcQ9RiJoiZ2iX7wf49SoNSILtxpk5UDbXaZpqExample", 25),
		"*****oNSILtxpk5UDbXaZpqExample")
}

func TestFormatKeyValueSummary(t *testing.T) {
	td.Cmp(t, formatKeyValueSummary(map[string]string{"b": "2", "a": "1"}), "a=1, b=2")
	td.Cmp(t, formatKeyValueSummary(nil), "")
}

func TestTextSummary(t *testing.T) {
	td.Cmp(t, textSummary("single line"), "single line")
	td.Cmp(t, textSummary("l1\nl2\nl3"), "l1\nl2\nl3",
		"5 lines or fewer are shown in full, nothing is actually hidden")
	td.Cmp(t, textSummary("l1\nl2\nl3\nl4\nl5"), "l1\nl2\nl3\nl4\nl5")
	td.Cmp(t, textSummary("l1\nl2\nl3\nl4\nl5\nl6\nl7"), "l1\nl2\nl3\n...\nl6\nl7",
		"more than 5 lines: first 3, an ellipsis line, then the last 2")

	long := strings.Repeat("a", 60)
	td.Cmp(t, textSummary(long), strings.Repeat("a", 49)+"…",
		"a line wider than 50 characters is truncated to fit")
}

func TestLayoutPartitionAttrs(t *testing.T) {
	raid := 1
	partition := layoutPartition{
		MountPoint: "/", FileSystem: "ext4", Size: 0, RaidLevel: &raid, LvName: "root", ZpName: "tank",
	}

	td.Cmp(t, layoutPartitionAttrs(partition), []partitionAttr{
		{Label: "Mount point", Value: "/"},
		{Label: "File system", Value: "ext4"},
		{Label: "Size", Value: "remaining space"},
		{Label: "RAID level", Value: "1"},
		{Label: "LVM volume", Value: "root"},
		{Label: "ZFS pool", Value: "tank"},
	}, "a size of 0 reads as remaining space, and optional attributes are only listed when set")

	td.Cmp(t, layoutPartitionAttrs(layoutPartition{MountPoint: "/boot", FileSystem: "ext4", Size: 1024}),
		[]partitionAttr{
			{Label: "Mount point", Value: "/boot"},
			{Label: "File system", Value: "ext4"},
			{Label: "Size", Value: "1024 MiB"},
		}, "RAID/LVM/ZFS rows are dropped when the partition doesn't use them")
}

func TestFormatPartitionBox(t *testing.T) {
	attrs := []partitionAttr{
		{Label: "Mount point", Value: "/"},
		{Label: "Size", Value: "remaining space"},
	}
	box := formatPartitionBox(attrs, partitionRowWidth(attrs), nil)

	td.Cmp(t, box, ""+
		"┌───────────────────────────────┐\n"+
		"│ Mount point : /               │\n"+
		"│ Size        : remaining space │\n"+
		"└───────────────────────────────┘")

	for i, line := range strings.Split(box, "\n") {
		td.Cmp(t, len([]rune(line)), len([]rune(strings.Split(box, "\n")[0])),
			fmt.Sprintf("line %d must be as wide as the border, or the box misaligns", i))
	}

	td.Cmp(t, formatPartitionBox(nil, 10, nil), "")

	// A row too wide for the given width is truncated, not left overflowing.
	narrow := formatPartitionBox(attrs, 10, nil)
	for _, line := range strings.Split(narrow, "\n") {
		td.Cmp(t, len([]rune(line)), 14, "every line stays exactly rowWidth+4 wide")
	}

	tagged := formatPartitionBox(attrs, partitionRowWidth(attrs), []string{"SYSTEM", "LVM"})
	taggedLines := strings.Split(tagged, "\n")
	width := partitionRowWidth(attrs)
	td.Cmp(t, taggedLines[0], "┌"+strings.Repeat("─", width+2)+"┐",
		"the box's top border, not a tags line, comes first")
	td.Cmp(t, taggedLines[1], fmt.Sprintf("│ %*s │", width, "[SYSTEM] [LVM]"),
		"tags are the box's own first row, inside the border")
}

func TestPartitionTags(t *testing.T) {
	td.Cmp(t, partitionTags(layoutPartition{MountPoint: "/"}, "/"), []string{"SYSTEM"})
	td.Cmp(t, partitionTags(layoutPartition{MountPoint: "/", LvName: "root"}, "/"), []string{"SYSTEM", "LVM"})
	td.Cmp(t, partitionTags(layoutPartition{MountPoint: "/boot"}, "/"), td.Empty(),
		"a partition that isn't the root and has no LVM volume gets no tag")
	td.Cmp(t, partitionTags(layoutPartition{MountPoint: "/", LvName: "root"}, ""), []string{"LVM"},
		"an unknown root mount point (empty) never tags a partition as the system root")
}

func TestPartitionBoxWidth(t *testing.T) {
	narrow := []partitionAttr{{Label: "Mount point", Value: "/"}}
	wide := []partitionAttr{{Label: "Mount point", Value: strings.Repeat("x", 80)}}

	td.Cmp(t, partitionBoxWidth([][]partitionAttr{narrow}), partitionRowWidth(narrow),
		"a single, short partition doesn't get padded up to the cap")
	td.Cmp(t, partitionBoxWidth([][]partitionAttr{narrow, wide}), maxPartitionBoxWidth-4,
		"the widest partition in the set drives the shared width, capped at maxPartitionBoxWidth")
}

func TestSchemePartitionToLayout(t *testing.T) {
	raid := 1
	td.Cmp(t,
		schemePartitionToLayout(schemePartition{
			Filesystem: "ext4", Mountpoint: "/", Raid: "1", Size: 0, Type: "lv", VolumeName: "root",
		}),
		layoutPartition{FileSystem: "ext4", MountPoint: "/", Size: 0, RaidLevel: &raid, LvName: "root"},
	)
}

func TestDiskGroupHasHardwareRaid(t *testing.T) {
	td.Cmp(t, diskGroup{RaidController: "HBA330"}.hasHardwareRaid(), true)
	td.Cmp(t, diskGroup{}.hasHardwareRaid(), false)
}

func TestHardwareRaidWarning(t *testing.T) {
	td.Cmp(t, hardwareRaidWarning(diskGroup{}), "", "no controller, no warning")

	warning := hardwareRaidWarning(diskGroup{DiskGroupID: 1, RaidController: "HBA330"})
	td.Cmp(t, warning, td.Contains("Disk group 1"))
	td.Cmp(t, warning, td.Contains("HBA330"))
	td.Cmp(t, warning, td.Contains("JBOD"))
}

func TestOsEndOfInstallWarning(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	td.Cmp(t, osEndOfInstallWarning("", now, true), "", "no date, no warning")
	td.Cmp(t, osEndOfInstallWarning("not-a-date", now, true), "", "unparseable date, no warning")
	td.Cmp(t, osEndOfInstallWarning("2029-03-06", now, true), "", "comfortably far away, no warning")

	warning := osEndOfInstallWarning("2026-06-01", now, true)
	td.Cmp(t, warning, td.Contains("2026-06-01"), "less than 6 months away warns, with the date")

	td.Cmp(t, osEndOfInstallWarning("2025-12-31", now, true), td.Contains("2025-12-31"), "a past date still warns")

	warning = osEndOfInstallWarning("2026-06-01", now, false)
	td.Cmp(t, warning, td.Not(td.Contains("2026-06-01")), "includeDate false leaves the date out")
	td.Cmp(t, warning, td.Not(td.Empty()), "but still warns")
}

func TestSchemeListHeight(t *testing.T) {
	td.Cmp(t, schemeListHeight(0), 2, "at least one row's worth of height even with no schemes yet")
	td.Cmp(t, schemeListHeight(2), 3)
	td.Cmp(t, schemeListHeight(20), 6, "capped, so a long scheme list still leaves room for the detail viewport")
}
