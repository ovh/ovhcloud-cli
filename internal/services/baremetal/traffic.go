// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
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

// TrafficPeriods and TrafficTypes come from the API enums, quoted here so that
// a wrong value is refused by the CLI with the list rather than by the API with
// a 400.
var (
	trafficPeriods = []string{"hourly", "daily", "weekly", "monthly", "yearly"}
	trafficTypes   = []string{
		"traffic:download", "traffic:upload",
		"packets:download", "packets:upload",
		"errors:download", "errors:upload",
	}
)

// mrtgPoint is one sample of a graph.
type mrtgPoint struct {
	Timestamp int64 `json:"timestamp"`
	Value     struct {
		Unit  string  `json:"unit"`
		Value float64 `json:"value"`
	} `json:"value"`
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

	peak, total := points[0].Value.Value, 0.0
	for _, p := range points {
		if p.Value.Value > peak {
			peak = p.Value.Value
		}
		total += p.Value.Value
	}

	unit := points[0].Value.Unit
	average := total / float64(len(points))
	latest := points[len(points)-1].Value.Value

	return map[string]any{
		"points": len(points),
		"unit":   unit,

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

	if !slicesContain(trafficPeriods, TrafficPeriod) {
		display.OutputError(&flags.OutputFormatConfig,
			"unknown period %q; use one of %s", TrafficPeriod, strings.Join(trafficPeriods, ", "))
		return
	}
	for _, requested := range TrafficTypes {
		if !slicesContain(trafficTypes, requested) {
			display.OutputError(&flags.OutputFormatConfig,
				"unknown type %q; use one of %s", requested, strings.Join(trafficTypes, ", "))
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

	rows := make([]map[string]any, 0, len(macs)*len(TrafficTypes))
	for _, mac := range macs {
		for _, requested := range TrafficTypes {
			path := fmt.Sprintf("/v1/dedicated/server/%s/networkInterfaceController/%s/mrtg?period=%s&type=%s",
				url.PathEscape(server), url.PathEscape(mac),
				url.QueryEscape(TrafficPeriod), url.QueryEscape(requested))

			var points []mrtgPoint
			if err := httpLib.Client.Get(path, &points); err != nil {
				display.OutputError(&flags.OutputFormatConfig,
					"failed to read %s on %s: %s", requested, mac, err)
				return
			}

			row := summarise(points)
			row["nic"] = mac
			row["type"] = requested
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
