// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

func point(timestamp int64, unit string, value float64) mrtgPoint {
	var p mrtgPoint
	p.Timestamp = timestamp
	p.Value.Unit = unit
	p.Value.Value = value
	return p
}

// Every unit this API returns is a rate — bps, pps and eps, all three seen on a
// real server — so one scale serves them all.
func TestReadableRateScalesEveryUnit(t *testing.T) {
	td.Cmp(t, readableRate(33978092.769, "bps"), "33.98 Mbps")
	td.Cmp(t, readableRate(233842.415, "bps"), "233.84 kbps")
	td.Cmp(t, readableRate(944.676, "bps"), "944.68 bps")
	td.Cmp(t, readableRate(2867.649, "pps"), "2.87 kpps")
	td.Cmp(t, readableRate(0, "eps"), "0 eps")
	td.Cmp(t, readableRate(2.5e9, "bps"), "2.50 Gbps")
}

// An average of 0.16518435754189942 bps is the true number and the wrong
// answer. The readable column exists so the table stops printing it.
func TestReadableRateDoesNotPrintSeventeenDecimals(t *testing.T) {
	td.Cmp(t, readableRate(0.16518435754189942, "bps"), "0.17 bps")
}

func TestSummariseReportsPeakAverageAndLatest(t *testing.T) {
	summary := summarise([]mrtgPoint{
		point(1, "bps", 100),
		point(2, "bps", 300),
		point(3, "bps", 200),
	})

	td.Cmp(t, summary["peak"], 300.0)
	td.Cmp(t, summary["average"], 200.0)
	td.Cmp(t, summary["latest"], 200.0, "the last sample, not the largest")
	td.Cmp(t, summary["points"], 3)
	td.Cmp(t, summary["unit"], "bps")
}

// The series travels with the summary rather than being thrown away: a day is
// 173 samples, which no table can show and every script wants. Dropping it
// would make -o json poorer than the API it wraps.
func TestSummariseKeepsTheSeries(t *testing.T) {
	summary := summarise([]mrtgPoint{point(1, "bps", 100), point(2, "bps", 300)})

	series, ok := summary["series"].([]mrtgPoint)
	td.Require(t).Cmp(ok, true)
	td.Cmp(t, len(series), 2)
	td.Cmp(t, series[1].Timestamp, int64(2))
}

// A controller with no samples is not a crash, and not a zero either: saying
// "0 bps" for a graph that returned nothing would state a measurement that was
// never made.
func TestSummariseSaysNothingRatherThanZero(t *testing.T) {
	summary := summarise(nil)

	td.Cmp(t, summary["points"], 0)
	td.Cmp(t, summary["peak"], nil)
	td.Cmp(t, summary["average"], nil)
}

// The accepted periods and types are the API's, read from the embedded schema.
// They were retyped in Go until a cross-review pointed it out: the values
// matched at the time, which is exactly how such a list rots unnoticed.
func TestTrafficPeriodsAndTypesComeFromTheSchema(t *testing.T) {
	assert := td.Assert(t)

	periods, err := trafficPeriods()
	assert.CmpNoError(err)
	assert.Cmp(periods, td.Bag("hourly", "daily", "weekly", "monthly", "yearly"))

	types, err := trafficTypes()
	assert.CmpNoError(err)
	assert.Cmp(types, td.SuperBagOf("traffic:download", "packets:upload", "errors:download"))
}

// A value the schema does not carry is refused before any request, and the
// refusal lists what is accepted.
func TestAnUnknownPeriodIsRefusedWithTheAcceptedList(t *testing.T) {
	assert := td.Assert(t)

	err := checkAgainstSchema("period", "fortnightly", trafficPeriods)

	assert.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("fortnightly"))
	assert.Cmp(err.Error(), td.Contains("monthly"), "the refusal says what is accepted")

	assert.CmpNoError(checkAgainstSchema("period", "daily", trafficPeriods))
}
