// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package ip

import (
	"encoding/json"
	"regexp"
	"sort"
	"testing"

	"github.com/ovh/ovhcloud-cli/internal/assets"
)

// The eight products are a list in Go, and the comment above it is right that
// the API has no index of them: there is no enum anywhere in ip.json, and no
// GET /ip/{ip}/license to read. Checked, because the audit had this down as a
// value retyped instead of read, and it is not.
//
// What is true is weaker and still worth a guard: the eight are eight paths of
// the embedded schema, so the Go list can drift away from the schema without a
// sound. This test is the sound. It reads the paths rather than a hardcoded
// count, so a product added upstream fails here rather than going unnoticed.
func TestLicenseProductsMatchTheEmbeddedSchema(t *testing.T) {
	var schema struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(assets.IpOpenapiSchema, &schema); err != nil {
		t.Fatalf("could not read the embedded ip schema: %s", err)
	}

	leaf := regexp.MustCompile(`^/ip/\{ip\}/license/([^/]+)$`)
	var fromSchema []string
	for path := range schema.Paths {
		if m := leaf.FindStringSubmatch(path); m != nil {
			fromSchema = append(fromSchema, m[1])
		}
	}

	if len(fromSchema) == 0 {
		t.Fatal("no /ip/{ip}/license/* path in the embedded schema: this test is measuring nothing")
	}

	sorted := append([]string(nil), licenseProducts...)
	sort.Strings(sorted)
	sort.Strings(fromSchema)

	if len(sorted) != len(fromSchema) {
		t.Fatalf("the Go list has %d products, the schema has %d: %v vs %v",
			len(sorted), len(fromSchema), sorted, fromSchema)
	}
	for i := range sorted {
		if sorted[i] != fromSchema[i] {
			t.Errorf("product %d: Go says %q, the schema says %q", i, sorted[i], fromSchema[i])
		}
	}
}
