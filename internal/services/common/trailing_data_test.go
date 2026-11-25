// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"strings"
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// json.Unmarshal rejected anything after the first value. A Decoder stops at
// the first value and ignores the rest, so a second object — or a typo that
// splits the payload in two — would be dropped and the request sent with only
// half of what the user wrote.
func TestDecodeJSONObject_RejectsTrailingData(t *testing.T) {
	for _, input := range []string{
		`{"quotaBytes":1}{"quotaBytes":2}`,
		`{"quotaBytes":1} garbage`,
		`{"quotaBytes":1}[]`,
	} {
		var body map[string]any
		err := decodeJSONObject([]byte(input), &body)
		td.CmpError(t, err, "input %q must be rejected, not truncated", input)
	}
}

func TestDecodeJSONObjectFrom_RejectsTrailingData(t *testing.T) {
	var body map[string]any
	err := decodeJSONObjectFrom(strings.NewReader(`{"a":1}{"b":2}`), &body)
	td.CmpError(t, err, "a file holding two objects must be rejected")
}

// Trailing whitespace and newlines are not trailing data.
func TestDecodeJSONObject_AcceptsTrailingWhitespace(t *testing.T) {
	var body map[string]any
	td.CmpNoError(t, decodeJSONObject([]byte("{\"a\":1}\n\n  \t\n"), &body))
	td.Cmp(t, len(body), 1)
}
