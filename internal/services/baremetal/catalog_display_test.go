// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import "testing"

// Every noraid code used to read "on request", which was true of none of them.
// The codes below are the ones the availabilities used on 20 October 2026.
func TestHumaniseStorageReadsNoRaidLayouts(t *testing.T) {
	cases := map[string]string{
		// KS-B and KS-C: one 120 GB SSD, no RAID.
		"noraid-1x120ssd": "1x120 GB SSD",
		// RISE-8 and RISE-9 in their default configuration: no disk at all,
		// the disks being options.
		"noraid-0": "no disk",
		// The catalogue's own name for that default.
		"softraid-0disk-24rise": "no disk",
		// Unchanged layouts, as a control.
		"softraid-2x960nvme-pcie-gen4": "2x960 GB NVMe Gen4",
	}
	for code, want := range cases {
		if got := humaniseStorage(code); got != want {
			t.Errorf("humaniseStorage(%q) = %q, want %q", code, got, want)
		}
	}
}

// A Scale or a High Grade boots from a pair of disks of its own, given apart
// from its data disks. Leaving them out called a machine with two NVMe drives
// "no disk". Beside data disks they are left out of the column, for its width,
// and kept in -o json. The codes are the availabilities' own.
func TestDescribeStorageCountsTheSystemDisks(t *testing.T) {
	cases := []struct{ data, system, want string }{
		{"noraid-0", "softraid-2x1920nvme-pcie-gen5-system", "2x1.92 TB NVMe Gen5 system"},
		{"softraid-2x3840nvme", "softraid-2x960nvme-system", "2x3.84 TB NVMe"},
		{"softraid-2x960nvme", "", "2x960 GB NVMe"},
		{"noraid-0", "", "no disk"},
	}
	for _, c := range cases {
		if got := describeStorage(c.data, c.system); got != c.want {
			t.Errorf("describeStorage(%q, %q) = %q, want %q", c.data, c.system, got, c.want)
		}
	}
}

// "—" and "?" read as missing data in a table. They were answers.
func TestShortDelayNamesWhatIsNotADelay(t *testing.T) {
	cases := map[string]string{
		"unavailable": "no stock",
		"unknown":     "unknown",
		"":            "unknown",
	}
	for code, want := range cases {
		if got := shortDelay(code); got != want {
			t.Errorf("shortDelay(%q) = %q, want %q", code, got, want)
		}
	}
}

// 1H-low is deliverable within the hour, with little stock left. The sort key
// ranks it as 3 so that it follows 1H-high, and the table used to print that
// key: "3 h (low)".
func TestDelayShowsTheStatedDelayNotTheSortKey(t *testing.T) {
	cases := []struct{ code, short, long string }{
		{"1H-low", "< 1 h (low)", "within the hour (low stock)"},
		{"1H-high", "< 1 h", "within the hour"},
		{"72H", "3 d", "3d"},
		{"240H", "10 d", "10d"},
	}
	for _, c := range cases {
		if got := shortDelay(c.code); got != c.short {
			t.Errorf("shortDelay(%q) = %q, want %q", c.code, got, c.short)
		}
		if got := humaniseDelay(c.code); got != c.long {
			t.Errorf("humaniseDelay(%q) = %q, want %q", c.code, got, c.long)
		}
	}
	if deliveryHours("1H-low") <= deliveryHours("1H-high") {
		t.Errorf("1H-low must still sort after 1H-high")
	}
}

// A memory code that spells its generation out is all there is for a plan no
// catalogue describes; one that does not is not given a guess.
func TestDescribeMemoryReadsTheGenerationInTheCode(t *testing.T) {
	cases := []struct{ code, ramType, want string }{
		{"ram-32g-ddr3-noecc-1333", "", "32 GB DDR3"},
		{"ram-64g-ddr4-ecc-2666", "", "64 GB DDR4 ECC"},
		{"ram-512g-ecc-5600", "", "512 GB ECC"},
		// The catalogue's type, when there is one, is the authority.
		{"ram-64g-on-die-ecc-5200", "DDR5", "64 GB DDR5 on-die ECC"},
	}
	for _, c := range cases {
		if got := describeMemory(c.code, c.ramType); got != c.want {
			t.Errorf("describeMemory(%q, %q) = %q, want %q", c.code, c.ramType, got, c.want)
		}
	}
}
