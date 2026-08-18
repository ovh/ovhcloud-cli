// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"bytes"
	"encoding/json"
	"io"
)

// decodeJSONObject decodes a JSON object into a map without going through
// float64.
//
// encoding/json turns every number into a float64 when the destination is an
// interface, and a float64 carries 53 bits of mantissa: an int64 above 2^53
// comes back changed. A quota of 9007199254740993 bytes became
// 9007199254740992 on its way to the request body, silently.
//
// UseNumber keeps the literal as written, which is also what the API client
// does when it decodes responses — so both sides of the merge now hold the
// same representation.
func decodeJSONObject(data []byte, target *map[string]any) error {
	return decodeJSONObjectFrom(bytes.NewReader(data), target)
}

// decodeJSONObjectFrom is decodeJSONObject reading from a stream, for the
// parameters given in a file.
func decodeJSONObjectFrom(reader io.Reader, target *map[string]any) error {
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()

	return decoder.Decode(target)
}
