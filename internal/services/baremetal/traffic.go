// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
	"github.com/spf13/cobra"
)

// Traffic graphs come from mrtg, and the CLI reached none of it. Three routes
// carry it and two are dead ends: /mrtg and /vrack/{vrack}/mrtg were deprecated
// on 2017-10-23 with deletion announced for 2018-04-23, both pointing at
// /networkInterfaceController. Only the per-controller one is live, so it is
// the only one wired here.
//
// It wants a MAC address, and nothing in this CLI could tell you one: there is
// no command that lists the controllers of a server, and `vni list` shows the
// virtual interfaces, whose own MAC list only appears under -o json. Asking an
// operator for a MAC they have no way to look up would be a command that
// cannot be run, so the server is resolved to its controllers here.

var (
	// TrafficPeriod is the window the graph covers.
	TrafficPeriod string

	// TrafficTypes are the series to read. Both directions of traffic are
	// read by default because "how much is this server doing" is a question
	// about both, and reading one of them silently answers half of it.
	TrafficTypes []string

	// TrafficNIC restricts the reading to one controller.
	TrafficNIC string
)

// defaultTrafficTypes is what `traffic` reads when --type is not given: both
// directions of ordinary traffic, which is the question the command is opened
// for. It lives here rather than in the flag declaration because a slice flag
// cannot keep a default across two commands in one process — see the note at
// its registration.
var defaultTrafficTypes = []string{"traffic:download", "traffic:upload"}

// TrafficPeriods and TrafficTypes come from the API enums, quoted here so that
// a wrong value is refused by the CLI with the list rather than by the API with
// a 400.
var (
	// Read from the embedded schema rather than retyped here. The values happen
	// to match today, but a list copied into Go goes stale in silence: this
	// repository already shipped a game-protocol list that refused a protocol
	// the API had accepted for months.
	trafficPeriods = sync.OnceValues(func() ([]string, error) {
		return openapi.GetComponentEnum(assets.BaremetalOpenapiSchema, "dedicated.server.MrtgPeriodEnum")
	})
	trafficTypes = sync.OnceValues(func() ([]string, error) {
		return openapi.GetComponentEnum(assets.BaremetalOpenapiSchema, "dedicated.server.MrtgTypeEnum")
	})
)

// mrtgPoint is one sample of a graph.
//
// Value is a pointer because the API declares it nullable, and means it: swept
// over ten servers of this account, 348 of 5311 yearly samples came back null —
// 90 of 273 on one machine, and its very first sample was one of them.
//
// A non-pointer struct turned every one of those into a real zero. The cost was
// two wrong numbers presented as facts: that machine's average read 14039 bps
// where the truth is 20944, understated by 33% by samples that do not exist,
// and because the unit was taken from the first sample the whole row printed
// "20.94 k" with no unit at all.
type mrtgPoint struct {
	Timestamp int64      `json:"timestamp"`
	Value     *mrtgValue `json:"value"`
}

type mrtgValue struct {
	Unit  string  `json:"unit"`
	Value float64 `json:"value"`
}

// controllersOf lists the network controllers of a server, so that nobody has
// to know a MAC address to read a graph.
func controllersOf(server string) ([]string, error) {
	var macs []string

	path := fmt.Sprintf("/v1/dedicated/server/%s/networkInterfaceController", url.PathEscape(server))
	if err := httpLib.Client.Get(path, &macs); err != nil {
		return nil, fmt.Errorf("failed to list the network controllers of %s: %w", server, err)
	}

	sort.Strings(macs)

	return macs, nil
}

// summarise reduces a series to the three numbers a table can hold.
//
// The series itself is not dropped: it travels in the same object, so `-o json`
// hands a script all 173 points while the table shows what a person reads. A
// day of samples is not something a terminal can usefully print, and dropping
// it would make the JSON output poorer than the API it wraps.
func summarise(points []mrtgPoint) map[string]any {
	if len(points) == 0 {
		return map[string]any{"points": 0}
	}

	// An absent sample is not a sample worth zero. It is counted and reported,
	// and it takes part in nothing: not the peak, not the total, not the divisor
	// of the average, and not the unit — which comes from the first sample that
	// has one rather than from the first sample.
	var (
		peak, total float64
		measured    int
		unit        string
		latest      float64
	)
	for _, p := range points {
		if p.Value == nil {
			continue
		}
		if measured == 0 || p.Value.Value > peak {
			peak = p.Value.Value
		}
		if unit == "" {
			unit = p.Value.Unit
		}
		total += p.Value.Value
		latest = p.Value.Value
		measured++
	}

	if measured == 0 {
		// Every sample of the window is absent. Reporting three zeros here would
		// say the interface carried no traffic, which is a different statement
		// from "nothing was recorded".
		return map[string]any{
			"points":   len(points),
			"measured": 0,
			"missing":  len(points),
		}
	}

	average := total / float64(measured)

	return map[string]any{
		"points": len(points),
		"unit":   unit,

		// measured and missing are how many of those points carried a value. A
		// window a third of which was never recorded is a caveat on every figure
		// on the line, so it travels with them.
		"measured": measured,
		"missing":  len(points) - measured,

		// The raw figures keep the names, so `-o json` and `-o peak` hand a
		// script a number rather than something it has to parse back.
		"peak":    peak,
		"average": average,
		"latest":  latest,

		// The table shows these instead. A day of downloads peaking at
		// 33978092.769 bps is a true answer to a question nobody asked in
		// those units, and an average printed as 0.16518435754189942 is worse.
		"peakReadable":    readableRate(peak, unit),
		"averageReadable": readableRate(average, unit),
		"latestReadable":  readableRate(latest, unit),

		"series": points,
	}
}

// readableRate scales a rate the way somebody reads it. Every unit this API
// returns is a rate — bps for traffic, pps for packets, eps for errors —
// measured on a real server, so one scale serves all three.
func readableRate(value float64, unit string) string {
	switch {
	case value >= 1e9:
		return fmt.Sprintf("%.2f G%s", value/1e9, unit)
	case value >= 1e6:
		return fmt.Sprintf("%.2f M%s", value/1e6, unit)
	case value >= 1e3:
		return fmt.Sprintf("%.2f k%s", value/1e3, unit)
	case value == 0:
		return "0 " + unit
	default:
		return fmt.Sprintf("%.2f %s", value, unit)
	}
}

// ShowBaremetalTraffic reads the traffic graphs of a server's controllers.
func ShowBaremetalTraffic(_ *cobra.Command, args []string) {
	server := args[0]

	if err := checkAgainstSchema("period", TrafficPeriod, trafficPeriods); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	// An empty --type means the default, and it is resolved here rather than
	// carried by the flag. cobra can hold a default for a string slice, but
	// PostExecute puts a used slice flag back to nil instead of to its default
	// — DefValue is "[]" for a slice, so there is nothing else it could use —
	// and this is the only slice flag in the CLI that had a non-empty default.
	// In a process that runs more than one command, the second `baremetal
	// traffic` looped over nothing and printed an empty table with exit 0.
	//
	// Resolving it here, rather than fixing the reset in root.go, keeps the
	// change out of the one file every branch of this series touches.
	requested := TrafficTypes
	if len(requested) == 0 {
		requested = defaultTrafficTypes
	}

	for _, series := range requested {
		if err := checkAgainstSchema("type", series, trafficTypes); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}
	}

	macs := []string{TrafficNIC}
	if TrafficNIC == "" {
		var err error
		if macs, err = controllersOf(server); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}
		if len(macs) == 0 {
			display.OutputInfo(&flags.OutputFormatConfig,
				map[string]any{"server": server, "controllers": 0},
				"%s has no network controller to read a graph from.", server)
			return
		}
	}

	rows := make([]map[string]any, 0, len(macs)*len(requested))
	for _, mac := range macs {
		for _, series := range requested {
			path := fmt.Sprintf("/v1/dedicated/server/%s/networkInterfaceController/%s/mrtg?period=%s&type=%s",
				url.PathEscape(server), url.PathEscape(mac),
				url.QueryEscape(TrafficPeriod), url.QueryEscape(series))

			var points []mrtgPoint
			if err := httpLib.Client.Get(path, &points); err != nil {
				display.OutputError(&flags.OutputFormatConfig,
					"failed to read %s on %s: %s", series, mac, err)
				return
			}

			row := summarise(points)
			row["nic"] = mac
			row["type"] = series
			row["period"] = TrafficPeriod
			rows = append(rows, row)
		}
	}

	display.RenderTable(rows,
		[]string{"nic", "type", "period", "peakReadable peak", "averageReadable average",
			"latestReadable latest", "points"},
		&flags.OutputFormatConfig)
}

func slicesContain(haystack []string, needle string) bool {
	for _, candidate := range haystack {
		if candidate == needle {
			return true
		}
	}
	return false
}

// checkAgainstSchema refuses a value the API does not declare, and says which
// ones it does. The accepted list comes from the embedded schema: a list
// retyped in Go goes stale without a word, which is how a game protocol the
// API had accepted for months came to be refused locally.
func checkAgainstSchema(flag, value string, read func() ([]string, error)) error {
	accepted, err := read()
	if err != nil {
		return fmt.Errorf("failed to read the values accepted for --%s: %w", flag, err)
	}

	if slicesContain(accepted, value) {
		return nil
	}

	return fmt.Errorf("--%s does not accept %q; accepted values are: %s",
		flag, value, strings.Join(accepted, ", "))
}
