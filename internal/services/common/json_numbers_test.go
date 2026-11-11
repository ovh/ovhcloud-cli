// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// An int64 above 2^53 does not fit in a float64 mantissa. Decoding into an
// interface used to round it, and the rounded value was what reached the API.
func TestDecodeJSONObject_KeepsLargeIntegersExact(t *testing.T) {
	type spec struct {
		QuotaBytes int64 `json:"quotaBytes"`
	}

	raw, err := json.Marshal(spec{QuotaBytes: 9007199254740993})
	td.Require(t).CmpNoError(err)

	var body map[string]any
	td.Require(t).CmpNoError(decodeJSONObject(raw, &body))

	out, err := json.Marshal(body)
	td.Require(t).CmpNoError(err)
	td.Cmp(t, string(out), `{"quotaBytes":9007199254740993}`)
}

// The same guarantee must hold for parameters read from a file.
func TestDecodeJSONObjectFrom_KeepsLargeIntegersExact(t *testing.T) {
	var body map[string]any
	td.Require(t).CmpNoError(decodeJSONObjectFrom(
		strings.NewReader(`{"quotaBytes":9007199254740993}`), &body))

	out, err := json.Marshal(body)
	td.Require(t).CmpNoError(err)
	td.Cmp(t, string(out), `{"quotaBytes":9007199254740993}`)
}

// Decimals must survive too, and a whole number must not grow a decimal part.
func TestDecodeJSONObject_KeepsNumberShapes(t *testing.T) {
	var body map[string]any
	td.Require(t).CmpNoError(decodeJSONObject(
		[]byte(`{"price":12.99,"count":3,"ratio":0.5}`), &body))

	out, err := json.Marshal(body)
	td.Require(t).CmpNoError(err)
	td.Cmp(t, string(out), `{"count":3,"price":12.99,"ratio":0.5}`)
}
