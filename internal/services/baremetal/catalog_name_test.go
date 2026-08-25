package baremetal

import "testing"

// Le catalogue affichait « 24adv03-v3 » la ou une facture dit « ADVANCE-3 ».
// Le retour de revue etait exactement celui-la : « on affiche les plan code au
// lieu d afficher les noms commerciaux plus lisibles ».
//
// Chaque cas ci-dessous vient d une mesure du 25/08 sur
// /v1/order/catalog/public/eco?ovhSubsidiary=FR (99 plans).
func TestCommercialName(t *testing.T) {
	cases := []struct {
		planCode    string
		invoiceName string
		want        string
	}{
		// Le cas cite dans la revue.
		{"24adv03-v3", "ADVANCE-3 | AMD EPYC 4464P", "ADVANCE-3 2024"},
		// Le suffixe de region ne change ni le nom ni la generation.
		{"24adv03-v3-syd", "ADVANCE-3 | AMD EPYC 4464P", "ADVANCE-3 2024"},
		// Deux generations d une meme gamme : c est l annee qui les separe, et
		// c est pour cela qu elle est affichee.
		{"24rise01-v1", "RISE-1 | Intel Xeon E-2386G", "RISE-1 2024"},
		{"25risel01-v1", "RISE-L | AMD EPYC 4585PX", "RISE-L 2025"},
		// La casse n est pas retouchee : un Title-case naif ecrirait « Rise-Xl »
		// et « Rise-Gpu-1 », donc abimerait les sigles de la gamme.
		{"25risexl01-v1", "RISE-XL | AMD EPYC 4585PX", "RISE-XL 2025"},
		{"26risegpu01-v1", "RISE-GPU-1 | Intel Xeon", "RISE-GPU-1 2026"},
		// Un nom sans barre verticale : un seul plan du catalogue est dans ce
		// cas, et il ne doit pas perdre son nom pour autant.
		{"24sk102", "KS-1", "KS-1 2024"},
		// 145 des 244 plans annonces par /availabilities ne sont pas au
		// catalogue public : Scale, High Grade, HCI, SAP. Sans nom, on rend ""
		// et l appelant garde le plan code. Inventer serait pire.
		{"23scaleamd01-v3", "", ""},
		// Un plan code sans prefixe d annee garde le nom, sans annee inventee.
		{"legacy-plan", "SOMETHING | CPU", "SOMETHING"},
	}
	for _, c := range cases {
		if got := commercialName(c.planCode, c.invoiceName); got != c.want {
			t.Errorf("commercialName(%q, %q) = %q, attendu %q",
				c.planCode, c.invoiceName, got, c.want)
		}
	}
}

// countUnnamed sert la ligne qui explique les tirets. Elle compte ce que la
// table montre, pas ce que le catalogue contient.
func TestCountUnnamed(t *testing.T) {
	rows := []map[string]any{
		{"name": "ADVANCE-3 2024"},
		{"name": "-"},
		{"name": ""},
		{},
	}
	if got := countUnnamed(rows); got != 3 {
		t.Errorf("countUnnamed = %d, attendu 3", got)
	}
}
