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
		// A plan code no public catalogue lists has no invoiceName: "" comes
		// back and the caller keeps the plan code. Inventing would be worse.
		{"27risele01-v1", "", ""},
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

// countUncatalogued serves the line that explains the rows no public catalogue
// lists. It counts the `catalogue` field buildCatalogRows sets, empty for none;
// a row that does not carry the field did not come from the table.
func TestCountUncatalogued(t *testing.T) {
	rows := []map[string]any{
		{"name": "ADVANCE-3 2024", "catalogue": "eco"},
		{"name": "ADVANCE-1 2026", "catalogue": "baremetalServers"},
		{"name": "27risele01-v1", "catalogue": ""},
		{"name": "26scaleamd09-v2", "catalogue": ""},
		{"name": "-"},
	}
	if got := countUncatalogued(rows); got != 2 {
		t.Errorf("countUncatalogued = %d, want 2", got)
	}
}
