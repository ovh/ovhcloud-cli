// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// Four ranges as the API described them on 20 October 2026, trimmed to the
// fields the command reads. Between them they carry every case the range view
// has to describe: on-die ECC (ADVANCE-2), ECC (RISE-1), no ECC (KS-GAME),
// NVMe with and without a PCIe generation, two processors (SYS-5), and a range
// with no private bandwidth. ADVANCE-2 also comes in a 128 GB configuration
// that is deliverable sooner than its default one, which is the trap below.
const (
	specsCatalog        = `{"locale": {"currencyCode": "EUR"}, "plans": [{"planCode": "24adv02-v3", "invoiceName": "ADVANCE-2 | AMD EPYC 4344P", "product": "24adv02", "addonFamilies": [{"name": "storage", "default": "softraid-2x960nvme-pcie-gen4-24adv02-v3"}, {"name": "bandwidth", "default": "bandwidth-3000-adv-gen3"}, {"name": "vrack", "default": "vrack-bandwidth-25000-adv-gen3"}, {"name": "memory", "default": "ram-64g-on-die-ecc-5200-24adv02-v3"}], "pricings": [{"mode": "default", "price": 12499000000, "interval": 0, "intervalUnit": "none", "capacities": ["installation"], "promotions": []}, {"mode": "default", "price": 12499000000, "interval": 1, "intervalUnit": "month", "capacities": ["renew"], "promotions": []}]}, {"planCode": "24sys053", "invoiceName": "SYS-5 | Intel Xeon Silver 4214R", "product": "24sys05", "addonFamilies": [{"name": "memory", "default": "ram-96g-ecc-2400-24sys"}, {"name": "storage", "default": "softraid-2x960nvme-24sys"}, {"name": "bandwidth", "default": "bandwidth-1000-24sys"}, {"name": "vrack", "default": "vrack-bandwidth-500-24sys"}], "pricings": [{"mode": "default", "price": 9999000000, "interval": 0, "intervalUnit": "none", "capacities": ["installation"], "promotions": []}, {"mode": "default", "price": 9999000000, "interval": 1, "intervalUnit": "month", "capacities": ["renew"], "promotions": []}]}, {"planCode": "24skgame012", "invoiceName": "KS-GAME | Intel Core i7 7700K", "product": "24skgame01", "addonFamilies": [{"name": "storage", "default": "softraid-2x450nvme-24skgame"}, {"name": "bandwidth", "default": "bandwidth-300-24skgame"}, {"name": "memory", "default": "ram-32g-noecc-2133-24skgame01"}], "pricings": [{"mode": "default", "price": 2899000000, "interval": 0, "intervalUnit": "none", "capacities": ["installation"], "promotions": []}, {"mode": "default", "price": 2899000000, "interval": 1, "intervalUnit": "month", "capacities": ["renew"], "promotions": []}]}, {"planCode": "24rise01-v2", "invoiceName": "RISE-1 | Intel Xeon-E 2386G", "product": "24rise01", "addonFamilies": [{"name": "storage", "default": "softraid-2x512nvme-24rise01-v2"}, {"name": "bandwidth", "default": "bandwidth-1000-unguaranteed-rise-gen2"}, {"name": "memory", "default": "ram-32g-ecc-3200-24rise01-v2"}, {"name": "vrack", "default": "vrack-bandwidth-1000-24rise01-v2"}], "pricings": [{"mode": "default", "price": 5699000000, "interval": 0, "intervalUnit": "none", "capacities": ["installation"], "promotions": []}, {"mode": "default", "price": 5699000000, "interval": 1, "intervalUnit": "month", "capacities": ["renew"], "promotions": []}]}], "products": [{"name": "24adv02", "blobs": {"technical": {"server": {"cpu": {"boost": 5.3, "brand": "AMD", "cores": 8, "model": "EPYC 4344P", "score": 35000, "number": 1, "threads": 16, "frequency": 3.8}, "frame": {"size": "0.5U", "model": "OVH", "dualPowerSupply": false}, "range": "advance", "services": {"sla": 99.95, "antiddos": "standard", "includedBackup": 500}}}}}, {"name": "24rise01", "blobs": {"technical": {"server": {"cpu": {"boost": 4.7, "brand": "Intel", "cores": 6, "model": "Xeon-E 2386G", "score": 13976, "number": 1, "threads": 12, "frequency": 3.5}, "frame": {"size": "0.5U", "model": "OVH", "dualPowerSupply": false}, "range": "rise", "services": {"sla": 99.95, "antiddos": "standard"}}}}}, {"name": "24skgame01", "blobs": {"technical": {"server": {"cpu": {"boost": 4.5, "brand": "Intel", "cores": 4, "model": "Core i7-7700K", "score": 9688, "number": 1, "threads": 8, "frequency": 4.2}, "frame": {"size": "1U", "model": "OVH", "dualPowerSupply": false}, "range": "kimsufi", "services": {"sla": 99.9, "antiddos": "standard"}}}}}, {"name": "24sys05", "blobs": {"technical": {"server": {"cpu": {"boost": 3.5, "brand": "Intel", "cores": 12, "model": "Xeon Silver 4214R", "score": 20000, "number": 2, "threads": 24, "frequency": 2.2}, "frame": {"size": "1U", "model": "OVH", "dualPowerSupply": false}, "range": "soyoustart", "services": {"sla": 99.9, "antiddos": "standard", "includedBackup": 500}}}}}, {"name": "bandwidth-1000", "blobs": {"technical": {"bandwidth": {"burst": 0, "level": 1000, "limit": 0, "guaranteed": true}}}}, {"name": "bandwidth-1000-unguaranteed", "blobs": {"technical": {"bandwidth": {"burst": 0, "level": 1000, "limit": 0, "guaranteed": false}}}}, {"name": "bandwidth-300", "blobs": {"technical": {"bandwidth": {"burst": 0, "level": 300, "limit": 0, "guaranteed": false}}}}, {"name": "bandwidth-3000", "blobs": {"technical": {"bandwidth": {"burst": 0, "level": 3000, "limit": 0, "guaranteed": true}}}}, {"name": "ram-128g-on-die-ecc-3600", "blobs": {"technical": {"memory": {"ecc": false, "size": 128, "ramType": "DDR5", "frequency": 3600}}}}, {"name": "ram-32g-ecc-3200", "blobs": {"technical": {"memory": {"ecc": true, "size": 32, "ramType": "DDR4", "frequency": 3200}}}}, {"name": "ram-32g-noecc-2133", "blobs": {"technical": {"memory": {"ecc": false, "size": 32, "ramType": "DDR4", "frequency": 2133}}}}, {"name": "ram-64g-on-die-ecc-5200", "blobs": {"technical": {"memory": {"ecc": false, "size": 64, "ramType": "DDR5", "frequency": 5200}}}}, {"name": "ram-96g-ecc-2400", "blobs": {"technical": {"memory": {"ecc": true, "size": 96, "ramType": "DDR4", "frequency": 2400}}}}, {"name": "softraid-2x450nvme", "blobs": {"technical": {"storage": {"raid": "Soft RAID", "disks": [{"specs": "2.5in 7mm", "usage": "data", "number": 2, "capacity": 450, "interface": "NVMe", "technology": "SSD"}], "hotSwap": false, "raidDetails": {"type": "Soft RAID"}}}}}, {"name": "softraid-2x512nvme", "blobs": {"technical": {"storage": {"raid": "Soft RAID", "disks": [{"specs": "M.2 80", "usage": "data", "number": 2, "capacity": 512, "interface": "NVMe", "technology": "SSD"}], "hotSwap": false, "raidDetails": {"type": "Soft RAID"}}}}}, {"name": "softraid-2x960nvme", "blobs": {"technical": {"storage": {"raid": "Soft RAID", "disks": [{"specs": "2.5in 7mm", "usage": "data", "number": 2, "capacity": 960, "interface": "NVMe", "technology": "SSD"}], "hotSwap": false, "raidDetails": {"type": "Soft RAID"}}}}}, {"name": "softraid-2x960nvme-pcie-gen4", "blobs": {"technical": {"storage": {"raid": "Soft RAID", "disks": [{"specs": "M.2 80", "usage": "data", "number": 2, "capacity": 960, "interface": "NVMe", "technology": "SSD", "physicalInterface": "pcie-gen4"}], "hotSwap": false, "raidDetails": {"type": "Soft RAID"}}}}}, {"name": "vrack-bandwidth-1000", "blobs": {"technical": {"vrack": {"burst": 0, "level": 1000, "limit": 0, "guaranteed": true}}}}, {"name": "vrack-bandwidth-25000", "blobs": {"technical": {"vrack": {"burst": 0, "level": 25000, "limit": 0, "guaranteed": false}}}}, {"name": "vrack-bandwidth-500", "blobs": {"technical": {"vrack": {"burst": 0, "level": 500, "limit": 0, "guaranteed": true}}}}], "addons": [{"planCode": "bandwidth-1000-24sys", "product": "bandwidth-1000"}, {"planCode": "bandwidth-1000-unguaranteed-rise-gen2", "product": "bandwidth-1000-unguaranteed"}, {"planCode": "bandwidth-300-24skgame", "product": "bandwidth-300"}, {"planCode": "bandwidth-3000-adv-gen3", "product": "bandwidth-3000"}, {"planCode": "ram-32g-ecc-3200-24rise01-v2", "product": "ram-32g-ecc-3200"}, {"planCode": "ram-32g-noecc-2133-24skgame01", "product": "ram-32g-noecc-2133"}, {"planCode": "ram-64g-on-die-ecc-5200-24adv02-v3", "product": "ram-64g-on-die-ecc-5200"}, {"planCode": "ram-96g-ecc-2400-24sys", "product": "ram-96g-ecc-2400"}, {"planCode": "softraid-2x450nvme-24skgame", "product": "softraid-2x450nvme"}, {"planCode": "softraid-2x512nvme-24rise01-v2", "product": "softraid-2x512nvme"}, {"planCode": "softraid-2x960nvme-24sys", "product": "softraid-2x960nvme"}, {"planCode": "softraid-2x960nvme-pcie-gen4-24adv02-v3", "product": "softraid-2x960nvme-pcie-gen4"}, {"planCode": "vrack-bandwidth-1000-24rise01-v2", "product": "vrack-bandwidth-1000"}, {"planCode": "vrack-bandwidth-25000-adv-gen3", "product": "vrack-bandwidth-25000"}, {"planCode": "vrack-bandwidth-500-24sys", "product": "vrack-bandwidth-500"}]}`
	specsAvailabilities = `[{"fqn": "24adv02-v3.ram-128g-on-die-ecc-3600.softraid-2x960nvme-pcie-gen4", "planCode": "24adv02-v3", "server": "24adv02", "memory": "ram-128g-on-die-ecc-3600", "storage": "softraid-2x960nvme-pcie-gen4", "datacenters": [{"availability": "1H-high", "datacenter": "bhs"}, {"availability": "1H-low", "datacenter": "fra"}, {"availability": "unavailable", "datacenter": "gra"}, {"availability": "1H-low", "datacenter": "lon"}, {"availability": "unavailable", "datacenter": "rbx"}, {"availability": "1H-low", "datacenter": "sbg"}, {"availability": "72H", "datacenter": "waw"}]}, {"fqn": "24adv02-v3.ram-64g-on-die-ecc-5200.softraid-2x960nvme-pcie-gen4", "planCode": "24adv02-v3", "server": "24adv02", "memory": "ram-64g-on-die-ecc-5200", "storage": "softraid-2x960nvme-pcie-gen4", "datacenters": [{"availability": "72H", "datacenter": "bhs"}, {"availability": "72H", "datacenter": "fra"}, {"availability": "unavailable", "datacenter": "gra"}, {"availability": "72H", "datacenter": "lon"}, {"availability": "unavailable", "datacenter": "rbx"}, {"availability": "unavailable", "datacenter": "sbg"}, {"availability": "1H-low", "datacenter": "waw"}]}, {"fqn": "24rise01-v2.ram-32g-ecc-3200.softraid-2x512nvme", "planCode": "24rise01-v2", "server": "24rise01", "memory": "ram-32g-ecc-3200", "storage": "softraid-2x512nvme", "datacenters": [{"availability": "1H-high", "datacenter": "bhs"}, {"availability": "1H-high", "datacenter": "fra"}, {"availability": "1H-low", "datacenter": "gra"}, {"availability": "unavailable", "datacenter": "lon"}, {"availability": "unavailable", "datacenter": "rbx"}, {"availability": "unavailable", "datacenter": "sbg"}, {"availability": "unavailable", "datacenter": "waw"}]}, {"fqn": "24skgame012.ram-32g-noecc-2133.softraid-2x450nvme", "planCode": "24skgame012", "server": "24skgame01", "memory": "ram-32g-noecc-2133", "storage": "softraid-2x450nvme", "datacenters": [{"availability": "unavailable", "datacenter": "bhs"}, {"availability": "unavailable", "datacenter": "fra"}, {"availability": "unavailable", "datacenter": "gra"}, {"availability": "unavailable", "datacenter": "lon"}, {"availability": "unavailable", "datacenter": "sbg"}, {"availability": "unavailable", "datacenter": "waw"}]}, {"fqn": "24sys053.ram-96g-ecc-2400.softraid-2x960nvme", "planCode": "24sys053", "server": "24sys05", "memory": "ram-96g-ecc-2400", "storage": "softraid-2x960nvme", "datacenters": [{"availability": "1H-low", "datacenter": "bhs"}, {"availability": "unknown", "datacenter": "fra"}, {"availability": "1H-low", "datacenter": "gra"}, {"availability": "unknown", "datacenter": "lon"}, {"availability": "unknown", "datacenter": "rbx"}, {"availability": "unknown", "datacenter": "sbg"}, {"availability": "unknown", "datacenter": "waw"}]}]`
)

func registerSpecsCatalog(t *td.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/me",
		httpmock.NewStringResponder(200, `{"ovhSubsidiary": "FR"}`))
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/order/catalog/public/eco?ovhSubsidiary=FR",
		httpmock.NewStringResponder(200, specsCatalog))
	httpmock.RegisterResponder("GET", `=~^https://eu\.api\.ovh\.com/v1/dedicated/server/datacenter/availabilities`,
		httpmock.NewStringResponder(200, specsAvailabilities))
}

// A range is recognised by its processor before its plan code: the commercial
// page leads with it, and two ranges at the same price differ by nothing else.
// A dual-socket machine counts the cores of both processors, because that is
// what the operator gets.
func (ms *MockSuite) TestBaremetalCatalogShowsTheProcessor(assert, require *td.T) {
	registerSpecsCatalog(assert)

	out, err := cmd.Execute("baremetal", "catalog")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("EPYC 4344P 8c"))
	assert.Cmp(out, td.Contains("2x Xeon Silver 4214R 24c"), "both processors' cores are counted")
}

// "64 GB" says how much memory, not which. The type comes from the catalogue,
// because the option code does not carry it; the ECC wording comes from the
// code, because the catalogue's own `ecc` flag is false on six of the seven
// on-die modules, and on-die ECC is not the ECC a server buyer means.
func (ms *MockSuite) TestBaremetalCatalogNamesTheMemoryType(assert, require *td.T) {
	registerSpecsCatalog(assert)

	out, err := cmd.Execute("baremetal", "catalog", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains(`"ram": "64 GB DDR5 on-die ECC"`), "on-die ECC is named as such")
	assert.Cmp(out, td.Contains(`"ram": "32 GB DDR4 ECC"`), "ECC memory says so")
	assert.Cmp(out, td.Contains(`"ram": "32 GB DDR4"`), "memory without ECC claims none")
}

// The PCIe generation of an NVMe drive is in the option code, and it is what
// separates two ranges with the same capacity. A drive whose code carries no
// generation is shown without one rather than with a guess.
func (ms *MockSuite) TestBaremetalCatalogNamesThePCIeGeneration(assert, require *td.T) {
	registerSpecsCatalog(assert)

	out, err := cmd.Execute("baremetal", "catalog", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains(`"disks": "2x960 GB NVMe Gen4"`))
	assert.Cmp(out, td.Contains(`"disks": "2x512 GB NVMe"`), "no generation is invented")
}

// The public and private bandwidth of the default configuration, in Gbps,
// public first. A range with no vRack option has no private bandwidth to show.
func (ms *MockSuite) TestBaremetalCatalogShowsTheIncludedBandwidth(assert, require *td.T) {
	registerSpecsCatalog(assert)

	out, err := cmd.Execute("baremetal", "catalog", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains(`"bandwidth": "3 / 25"`), "ADVANCE-2: 3 Gbps public, 25 Gbps private")
	assert.Cmp(out, td.Contains(`"bandwidth": "0.3 / -"`), "KS-GAME has no private bandwidth")
	assert.Cmp(out, td.Contains(`"publicBandwidthMbps": 3000`), "the raw figure stays for scripts")
}

// One row per range shows the range as it is sold: its default configuration.
// It used to show whichever configuration could be delivered soonest, at the
// range's base price — ADVANCE-2 with 128 GB of memory at the price of 64 GB,
// because every configuration of a plan carries the plan's price here.
func (ms *MockSuite) TestBaremetalCatalogShowsTheDefaultConfigurationPerRange(assert, require *td.T) {
	registerSpecsCatalog(assert)

	out, err := cmd.Execute("baremetal", "catalog")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("64 GB DDR5 on-die ECC"), "the default configuration is the one shown")
	assert.Cmp(out, td.Not(td.Contains("128 GB")), "not the one that happens to ship sooner")
}

// Seven facts per range do not fit 132 columns: measured on the 115 ranges the
// API listed on 20 October 2026, the range view is 147 characters across. 150
// is the limit it is held to; the plan code and the amount due at order, which
// would take it past 190, live in the narrowed view and in -o json.
func (ms *MockSuite) TestBaremetalCatalogRangeViewFits150Columns(assert, require *td.T) {
	registerSpecsCatalog(assert)

	out, err := cmd.Execute("baremetal", "catalog")

	require.CmpNoError(err)
	for _, line := range strings.Split(out, "\n") {
		assert.Cmp(len([]rune(line)) <= 150, true,
			"a %d-character line does not fit: %s", len([]rune(line)), line)
	}
}

// The cache is keyed by its version. A price table cached by the previous
// build has no processor, no memory type and no bandwidth in it, and read as
// fresh it would empty three columns without a word.
func (ms *MockSuite) TestBaremetalCatalogDoesNotReadTheCacheOfTheEarlierShape(assert, require *td.T) {
	registerSpecsCatalog(assert)
	dir := filepath.Join(os.Getenv("XDG_CACHE_HOME"), "ovhcloud", "catalog")
	require.CmpNoError(os.MkdirAll(dir, 0o700))
	require.CmpNoError(os.WriteFile(filepath.Join(dir, "eco-fr-v3.json"),
		[]byte(`{"currency": "EUR", "plans": {"24adv02-v3": {"default": {"recurring": 74.99, "interval": 1}}}, "names": {}}`), 0o600))

	out, err := cmd.Execute("baremetal", "catalog")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("EPYC 4344P 8c"), "the specs come from a fresh download")
}
