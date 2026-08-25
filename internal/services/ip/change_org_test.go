// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package ip

import (
	"strings"
	"testing"
)

// Ce que la garde annonce EST le sujet de cette commande : elle ne casse rien,
// elle donne l adresse. Le texte precedent disait « registers … to the
// organisation » et « published to the regional registry » -- vrai, et muet sur
// la seule consequence que pese la personne qui tape.
func TestChangeOrgWarningSaysTheAddressLeaves(t *testing.T) {
	got := changeOrgWarning("192.0.2.0/24", "RIPE_66451")

	for _, attendu := range []string{
		"192.0.2.0/24",        // de quelle adresse on parle
		"RIPE_66451",          // et vers qui elle part
		"leaves this account", // la consequence
		"agree",               // et que le retour ne depend plus de nous
	} {
		if !strings.Contains(got, attendu) {
			t.Errorf("l avertissement ne contient pas %q :\n  %s", attendu, got)
		}
	}
}
