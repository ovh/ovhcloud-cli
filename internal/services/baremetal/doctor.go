// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// The things that quietly break a dedicated server are spread across five
// routes and none of them is where somebody would look. A machine left on the
// rescue system runs no service and says nothing; monitoring switched off means
// OVHcloud will not call when a disk dies; a renewal switched off means the
// server goes back at expiry. Each is one field, in a different object.
//
// `doctor` reads them together and answers the only question worth asking:
// is anything wrong with this server, or with any of them.
//
// Every command named in a Fix was checked against the command tree of this
// binary before being written down. The first draft sent operators to
// `set-boot`, `list-boots` and `list-planned-interventions`, none of which
// exist — the same defect an earlier lot of this audit shipped in its own
// error message, and the reason this note is here.

var (
	// DoctorExpiryDays is how far ahead an expiry counts as imminent.
	DoctorExpiryDays int

	// DoctorStrict makes the command exit non-zero when it finds something.
	DoctorStrict bool
)

// severity orders the findings, worst first.
type severity int

const (
	critical severity = iota
	warning
	note
)

func (s severity) String() string {
	switch s {
	case critical:
		return "critical"
	case warning:
		return "warning"
	default:
		return "note"
	}
}

// finding is one thing that is wrong, on one server.
type finding struct {
	Server   string
	Severity severity
	Check    string
	Detail   string
	Fix      string
}

// diagnosis is everything read about one server, plus what it says.
type diagnosis struct {
	Server   string
	Findings []finding
	Err      error
}

// Doctor reports what is wrong with a server, or with every server.
func Doctor(_ *cobra.Command, args []string) {
	servers := args
	if len(servers) == 0 {
		var all []string
		if err := httpLib.Client.Get("/v1/dedicated/server", &all); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to list your servers: %s", err)
			return
		}
		servers = all
	}

	if len(servers) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"servers": 0},
			"You have no dedicated server.")
		return
	}

	diagnoses, err := diagnoseAll(servers)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var (
		findings []finding
		failed   []string
	)
	for _, d := range diagnoses {
		if d.Err != nil {
			failed = append(failed, fmt.Sprintf("%s (%s)", d.Server, d.Err))
			continue
		}
		findings = append(findings, d.Findings...)
	}

	// A server that could not be read is not a healthy server. Counting it as
	// checked would be the one mistake this command cannot afford: it exists to
	// be trusted when it says nothing is wrong.
	if len(failed) > 0 {
		display.OutputError(&flags.OutputFormatConfig,
			"could not check %d of %d servers, so this is not a clean bill of health:\n   %s",
			len(failed), len(servers), strings.Join(failed, "\n   "))
		return
	}

	if len(findings) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"servers": len(servers), "findings": []any{}},
			"✅ Nothing to report on %d server(s).", len(servers))
		return
	}

	sortFindings(findings)

	rows := make([]map[string]any, 0, len(findings))
	for _, f := range findings {
		rows = append(rows, map[string]any{
			"server":   f.Server,
			"severity": f.Severity.String(),
			"check":    f.Check,
			"detail":   f.Detail,
			"fix":      f.Fix,
		})
	}

	filtered, err := filtersLib.FilterLines(rows, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(filtered,
		[]string{"server", "severity", "check", "detail", "fix"},
		&flags.OutputFormatConfig)

	// The default exit code stays 0 because a command that ran and answered did
	// not fail, and every other command of this CLI reads a non-zero code that
	// way. --strict is for the pipeline that wants the answer as a gate, and it
	// has to be asked for rather than discovered.
	//
	// It changes the exit code and nothing else. Calling OutputError here would
	// print a second document after the table, and under -o json only the last
	// one survives -- the pipeline that asked for --strict would receive the
	// error message instead of the findings it came for.
	// Reached only when there is something to report: the no-finding case
	// returned above. A len(findings) > 0 guard here would be a condition that
	// cannot be false, which is the dead code this repository keeps finding.
	if DoctorStrict {
		display.ExitFunc(1)
	}
}

// diagnoseAll checks every server, ten at a time.
func diagnoseAll(servers []string) ([]diagnosis, error) {
	const parallelChecks = 10

	var (
		sem        = semaphore.NewWeighted(parallelChecks)
		diagnoses  = make([]diagnosis, len(servers))
		group, ctx = errgroup.WithContext(context.Background())
	)

	for i, server := range servers {
		if err := sem.Acquire(ctx, 1); err != nil {
			return nil, fmt.Errorf("failed to check your servers: %w", err)
		}

		group.Go(func() error {
			defer sem.Release(1)
			diagnoses[i] = diagnose(server)
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return diagnoses, nil
}

// diagnose reads one server and reports what is wrong with it.
func diagnose(server string) diagnosis {
	escaped := url.PathEscape(server)

	var detail map[string]any
	if err := httpLib.Client.Get(fmt.Sprintf("/v1/dedicated/server/%s", escaped), &detail); err != nil {
		return diagnosis{Server: server, Err: err}
	}

	found := checkState(server, detail)
	found = append(found, checkMonitoring(server, detail)...)
	found = append(found, checkIntervention(server, detail)...)

	var boot map[string]any
	if bootID := bootIdentifier(detail); bootID != 0 {
		if err := httpLib.Client.Get(
			fmt.Sprintf("/v1/dedicated/server/%s/boot/%d", escaped, bootID), &boot); err == nil {
			found = append(found, checkBoot(server, boot)...)
		}
	}

	if readings := readServiceInfos(escaped, renewalReadings); len(readings) > 0 {
		found = append(found, checkRenewal(server, readings)...)
	}

	var running []int64
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/dedicated/server/%s/task?status=doing", escaped), &running); err == nil {
		found = append(found, checkTasks(server, running)...)
	}

	var planned []int64
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/dedicated/server/%s/plannedIntervention", escaped), &planned); err == nil {
		found = append(found, checkPlannedIntervention(server, planned)...)
	}

	return diagnosis{Server: server, Findings: found}
}

// checkBoot reports a server that will not come up on its own disk.
//
// The obvious test — bootType == "rescue" — misses half of them. On the account
// measured, six servers of thirty-five were booted on a rescue system: three
// carried bootType "rescue", and three carried bootType "internal" on a boot
// entry whose kernel is "rescue-customer" and whose description ends with
// "[REMOVAL ON 2025-06-23]". The type is the API's, and it is wrong for that
// entry; the kernel is not. So both are read.
func checkBoot(server string, boot map[string]any) []finding {
	bootType := stringValue(boot, "bootType")
	kernel := stringValue(boot, "kernel")
	description := stringValue(boot, "description")

	switch {
	case bootType == "harddisk":
		return nil

	case bootType == "power":
		return []finding{{
			Server: server, Severity: critical, Check: "boot",
			Detail: fmt.Sprintf("next boot is %q, so the server will not come back up", description),
			Fix:    fmt.Sprintf("ovhcloud baremetal boot list %s, then boot set", server),
		}}

	case isRescueBoot(bootType, kernel):
		found := []finding{{
			Server: server, Severity: warning, Check: "boot",
			Detail: fmt.Sprintf("booted on the rescue system (%s), so it is not running its own OS", kernel),
			Fix:    fmt.Sprintf("ovhcloud baremetal boot list %s, then boot set + reboot", server),
		}}

		// The rescue image itself can be past its removal date. Three servers
		// of the fleet measured sat on one announced for removal fourteen
		// months earlier.
		if strings.Contains(description, "REMOVAL ON") {
			found = append(found, finding{
				Server: server, Severity: warning, Check: "boot-image",
				Detail: fmt.Sprintf("its rescue image is retired: %q", description),
				Fix:    fmt.Sprintf("ovhcloud baremetal boot list %s", server),
			})
		}

		return found

	default:
		return []finding{{
			Server: server, Severity: note, Check: "boot",
			Detail: fmt.Sprintf("boots on %q rather than its disk (%s)", bootType, description),
			Fix:    fmt.Sprintf("ovhcloud baremetal boot list %s", server),
		}}
	}
}

// isRescueBoot answers whether a boot entry is a rescue system, by its type or
// by its kernel. See checkBoot for why one of the two is not enough.
func isRescueBoot(bootType, kernel string) bool {
	return bootType == "rescue" || strings.HasPrefix(kernel, "rescue")
}

// checkState reports a server the API does not consider healthy.
func checkState(server string, detail map[string]any) []finding {
	var found []finding

	if state := stringValue(detail, "state"); state != "" && state != "ok" {
		found = append(found, finding{
			Server: server, Severity: critical, Check: "state",
			Detail: fmt.Sprintf("the API reports this server as %q", state),
			Fix:    "ovhcloud support-tickets create",
		})
	}

	if power := stringValue(detail, "powerState"); power != "" && power != "poweron" {
		found = append(found, finding{
			Server: server, Severity: critical, Check: "power",
			Detail: fmt.Sprintf("the server is %s", power),
			Fix:    fmt.Sprintf("ovhcloud baremetal reboot %s", server),
		})
	}

	return found
}

// checkMonitoring reports a server nobody will be called about.
func checkMonitoring(server string, detail map[string]any) []finding {
	if monitoring, set := detail["monitoring"].(bool); set && !monitoring {
		return []finding{{
			Server: server, Severity: warning, Check: "monitoring",
			Detail: "monitoring is off, so OVHcloud will not raise an alert if this server stops answering",
			Fix:    fmt.Sprintf("ovhcloud baremetal edit %s --monitoring", server),
		}}
	}

	return nil
}

// checkIntervention reports a server whose hardware will not be touched.
func checkIntervention(server string, detail map[string]any) []finding {
	if refused, set := detail["noIntervention"].(bool); set && refused {
		return []finding{{
			Server: server, Severity: warning, Check: "intervention",
			Detail: "hardware intervention is refused, so a failed disk will not be replaced without asking first",
			Fix:    fmt.Sprintf("ovhcloud baremetal edit %s --no-intervention=false", server),
		}}
	}

	return nil
}

// renewalReadings is how many times the renewal state is read.
//
// Once is not enough, and that is measured rather than defensive:
// `renew.automatic` on this route comes back differently between consecutive
// reads of the same object. Twenty sequential reads of five servers gave
// 10/10, 9/11, 8/12, 6/14 and 5/15 splits between true and false, with the
// `domain` field proving each answer belonged to the server asked about; three
// other servers answered the same value twenty times out of twenty, and
// `monitoring` on the server object was stable ten out of ten. So it is that
// one field, on that one route, and a health check built on a single read of
// it reports a coin toss.
//
// Five rather than three: on a field behaving like a fair coin, three readings
// agree by chance once in four, and a health check wrong one time in four is
// not one anybody should act on. Five brings that to about one in sixteen, and
// what is reported when they do agree says how it was established rather than
// claiming certainty this route cannot give.
const renewalReadings = 5

// readServiceInfos reads the billing object several times.
func readServiceInfos(escaped string, times int) []map[string]any {
	readings := make([]map[string]any, 0, times)

	for range times {
		var infos map[string]any
		if err := httpLib.Client.Get(
			fmt.Sprintf("/v1/dedicated/server/%s/serviceInfos", escaped), &infos); err != nil {
			return readings
		}
		readings = append(readings, infos)
	}

	return readings
}

// checkRenewal reports a server on its way out — or the fact that the API will
// not say whether it is.
func checkRenewal(server string, readings []map[string]any) []finding {
	var found []finding

	infos := readings[len(readings)-1]
	renew, _ := infos["renew"].(map[string]any)
	expiration := stringValue(infos, "expiration")

	switch automatic, agreed := agreedRenewal(readings); {
	case !agreed:
		// Reporting "renewal is off" from one of these readings would be a coin
		// toss, and reporting nothing would hide that nobody can tell. What is
		// certain is that the answer is not trustworthy, and that is the
		// finding.
		found = append(found, finding{
			Server: server, Severity: warning, Check: "renewal",
			Detail: fmt.Sprintf("the API gave different answers about automatic renewal across %d reads, so whether this server renews on %s cannot be established from here",
				len(readings), expiration),
			Fix: fmt.Sprintf("ovhcloud baremetal service-info get %s (and check the Manager)", server),
		})

	case !automatic:
		found = append(found, finding{
			Server: server, Severity: warning, Check: "renewal",
			Detail: fmt.Sprintf("automatic renewal read as off %d times out of %d, and this server expires on %s — this field is unreliable on this route, so confirm before acting",
				len(readings), len(readings), expiration),
			Fix: fmt.Sprintf("ovhcloud baremetal service-info edit %s --renew-automatic", server),
		})
	}

	if deleteAtExpiration, set := renew["deleteAtExpiration"].(bool); set && deleteAtExpiration {
		found = append(found, finding{
			Server: server, Severity: critical, Check: "renewal",
			Detail: fmt.Sprintf("this server is set to be deleted when it expires on %s", expiration),
			Fix:    fmt.Sprintf("ovhcloud baremetal service-info edit %s --renew-delete-at-expiration=false", server),
		})
	}

	if days, ok := daysUntil(expiration); ok && days <= int64(DoctorExpiryDays) {
		severity := note
		if automatic, agreed := agreedRenewal(readings); agreed && !automatic {
			severity = warning
		}
		found = append(found, finding{
			Server: server, Severity: severity, Check: "expiry",
			Detail: fmt.Sprintf("expires in %d day(s), on %s", days, expiration),
			Fix:    fmt.Sprintf("ovhcloud baremetal service-info get %s", server),
		})
	}

	return found
}

// agreedRenewal answers what the readings say about automatic renewal, and
// whether they agree at all. See renewalReadings for why the question has to be
// asked that way.
func agreedRenewal(readings []map[string]any) (bool, bool) {
	var (
		value bool
		known bool
	)

	for _, infos := range readings {
		renew, _ := infos["renew"].(map[string]any)
		automatic, set := renew["automatic"].(bool)
		if !set {
			return false, false
		}

		if !known {
			value, known = automatic, true
			continue
		}

		if automatic != value {
			return false, false
		}
	}

	return value, known
}

// checkTasks reports work still running on the server, because most of the
// other checks read a state that is about to change.
func checkTasks(server string, running []int64) []finding {
	if len(running) == 0 {
		return nil
	}

	return []finding{{
		Server: server, Severity: note, Check: "tasks",
		Detail: fmt.Sprintf("%d task(s) still running, so this server is being changed right now", len(running)),
		Fix:    fmt.Sprintf("ovhcloud baremetal list-tasks %s", server),
	}}
}

// checkPlannedIntervention reports maintenance already scheduled.
func checkPlannedIntervention(server string, planned []int64) []finding {
	if len(planned) == 0 {
		return nil
	}

	return []finding{{
		Server: server, Severity: warning, Check: "planned-intervention",
		Detail: fmt.Sprintf("%d intervention(s) planned on this server", len(planned)),
		Fix:    fmt.Sprintf("ovhcloud baremetal list-interventions %s", server),
	}}
}

// sortFindings puts the worst first, then groups by server so one machine's
// problems are read together.
func sortFindings(findings []finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Severity != findings[j].Severity {
			return findings[i].Severity < findings[j].Severity
		}
		if findings[i].Server != findings[j].Server {
			return findings[i].Server < findings[j].Server
		}
		return findings[i].Check < findings[j].Check
	})
}

// daysUntil answers how many days are left before a date, and whether the date
// could be read at all.
//
// An unreadable date reports "no" rather than zero: zero would mean "expires
// today" and raise an alarm about a field nobody managed to parse.
func daysUntil(date string) (int64, bool) {
	if date == "" {
		return 0, false
	}

	for _, layout := range []string{"2006-01-02", time.RFC3339} {
		if parsed, err := time.Parse(layout, date); err == nil {
			return int64(time.Until(parsed).Hours() / 24), true
		}
	}

	return 0, false
}

func stringValue(object map[string]any, key string) string {
	value, _ := object[key].(string)
	return value
}

// bootIdentifier reads the active boot entry whatever shape the decoder left
// it in. go-ovh decodes with UseNumber, so this arrives as a json.Number and
// never as a float64 — a type switch that only knew float64 turned a vRack
// branch into dead code once in this CLI.
func bootIdentifier(detail map[string]any) int64 {
	switch value := detail["bootId"].(type) {
	case json.Number:
		parsed, err := value.Int64()
		if err != nil {
			return 0
		}
		return parsed
	case float64:
		return int64(value)
	case int64:
		return value
	case int:
		return int64(value)
	}

	return 0
}
