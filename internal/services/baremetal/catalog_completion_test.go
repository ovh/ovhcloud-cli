// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import "testing"

// The values offered by --datacenter and --country are read from the embedded
// schema. When that read fails, shell completion goes silent and nothing else
// notices, since the command itself still runs.
func TestCatalogCompletionsReadTheSchema(t *testing.T) {
	for flag, values := range map[string]func() ([]string, error){
		"datacenter": availabilityDatacenters,
		"country":    catalogCountries,
	} {
		got, err := values()
		if err != nil {
			t.Errorf("--%s: %s", flag, err)
			continue
		}
		if len(got) == 0 {
			t.Errorf("--%s: the schema enumerates no value", flag)
		}
	}
}
