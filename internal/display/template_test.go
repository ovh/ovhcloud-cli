// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package display

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

func TestFormatBytes_SIUnits(t *testing.T) {
	tests := []struct {
		bytes    float64
		decimals int
		expected string
	}{
		{0, 2, "0"},
		{500, 0, "500 B"},
		{1500, 2, "1.50 KB"},
		{1048576, 1, "1.0 MB"},
		{1234567890, 2, "1.23 GB"},
		{1e15, 0, "1 PB"},
	}

	for _, tt := range tests {
		got := formatByteSize(tt.bytes, tt.decimals, 1000)
		td.Cmp(t, got, tt.expected)
	}
}

func TestFormatBytes_IECUnits(t *testing.T) {
	tests := []struct {
		bytes    float64
		decimals int
		expected string
	}{
		{0, 2, "0"},
		{1024, 0, "1 KiB"},
		{1048576, 2, "1.00 MiB"},
		{1073741824, 1, "1.0 GiB"},
		{1099511627776, 2, "1.00 TiB"},
	}

	for _, tt := range tests {
		got := formatByteSize(tt.bytes, tt.decimals, 1024)
		td.Cmp(t, got, tt.expected)
	}
}

func TestFormatBytes_InvalidFormat(t *testing.T) {
	// Should fallback to SI units (1000)
	got := formatByteSize(1000, 1, 999)
	td.Cmp(t, got, "1.0 KB")
}

func TestFormatBytes_LargeBytes(t *testing.T) {
	// Test for bytes larger than the largest unit
	got := formatByteSize(1e25, 2, 1000)
	td.Cmp(t, got, "10.00 YB")
}
