// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	_ "embed"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
)

// These two commands answer the questions somebody has to answer before they
// can write the `storage` block of a reinstall: which partition schemes the
// chosen template allows, and whether the machine has a RAID controller at all.
//
// A third would have sized a RAID configuration. install/hardwareRaidSize is in
// the embedded schema, badged "Stable production version", and does not exist:
// it answers 404 "Got an invalid (or empty) URL" — a routing error, not a
// business one — on every server tried, in v1 and v2, with and without its
// parameters, while install/hardwareRaidProfile resolves on those same servers
// and answers 403 with a message. A command for it would fail every time it was
// run, so there is none.
//
// The block itself already works — `--from-file` carries it through untouched,
// and the sample shipped with `--init-file` shows a full layout. What was
// missing is every way of finding out what to put in it, which meant writing a
// layout and learning at install time whether the server would take it. On a
// reinstall, learning at install time means the disks are already wiped.

//go:embed templates/install_status.tmpl
var installStatusTemplate string

var (
	// InstallTemplate is the OS template these answers are relative to.
	InstallTemplate string
)

// ListBaremetalPartitionSchemes lists the partition schemes a template allows
// on this server.
func ListBaremetalPartitionSchemes(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v1/dedicated/server/%s/install/compatibleTemplatePartitionSchemes?templateName=%s",
		url.PathEscape(args[0]), url.QueryEscape(InstallTemplate))

	var schemes []string
	if err := httpLib.Client.Get(path, &schemes); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to fetch the partition schemes of %s for %s: %s", args[0], InstallTemplate, err)
		return
	}

	rows := make([]map[string]any, 0, len(schemes))
	for _, scheme := range schemes {
		rows = append(rows, map[string]any{"name": scheme, "template": InstallTemplate})
	}

	display.RenderTable(rows, []string{"name", "template"}, &flags.OutputFormatConfig)
}

// GetBaremetalRaidProfile shows the RAID controllers of a server and the disks
// on them.
//
// Most servers have none. Measured on a real account, every one of the 14
// checked answered "Hardware RAID is not supported by this server" — which is
// exactly the answer worth having before writing a hardwareRaid block the API
// will refuse. So that refusal is reported as an answer rather than as a
// failure: it tells the operator what to do next, which is to configure
// software RAID in the partitioning layout instead.
func GetBaremetalRaidProfile(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v1/dedicated/server/%s/install/hardwareRaidProfile", url.PathEscape(args[0]))

	var profile struct {
		Controllers []struct {
			Model string `json:"model"`
			Type  string `json:"type"`
			Disks []struct {
				Capacity  any      `json:"capacity"`
				DiskGroup int      `json:"diskGroupId"`
				Names     []string `json:"names"`
				Speed     any      `json:"speed"`
				Type      any      `json:"type"`
			} `json:"disks"`
		} `json:"controllers"`
	}

	if err := httpLib.Client.Get(path, &profile); err != nil {
		if isUnsupportedHardwareRaid(err) {
			display.OutputInfo(&flags.OutputFormatConfig,
				map[string]any{"server": args[0], "hardwareRaid": false},
				"%s has no hardware RAID controller.\nUse software RAID instead: set `raidLevel` on the partitions of the storage layout.",
				args[0])
			return
		}
		display.OutputError(&flags.OutputFormatConfig,
			"failed to fetch the RAID profile of %s: %s", args[0], err)
		return
	}

	var rows []map[string]any
	for _, controller := range profile.Controllers {
		for _, disk := range controller.Disks {
			rows = append(rows, map[string]any{
				"controller":  controller.Model,
				"type":        controller.Type,
				"diskGroupId": disk.DiskGroup,
				"disks":       len(disk.Names),
				"capacity":    disk.Capacity,
				"diskType":    disk.Type,
			})
		}
	}

	display.RenderTable(rows,
		[]string{"controller", "type", "diskGroupId", "disks", "capacity", "diskType"},
		&flags.OutputFormatConfig)
}

// isUnsupportedHardwareRaid recognises the API's way of saying a server has no
// RAID controller.
//
// It is matched on the message because that is all the API gives: the answer
// arrives as an error rather than as an empty profile, and the status code it
// carries is the same one a genuine failure would use. Matching text is
// fragile, so it only ever downgrades an error into a plainer sentence — a
// message that stops matching costs the friendlier wording, never the answer.
func isUnsupportedHardwareRaid(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "hardware raid is not supported")
}

// InstallationProgress is what install/status answers while a machine is being
// installed: how long it has been going, and one entry per step.
//
// The steps carry no name of their own — `comment` is the whole description —
// so "which step is this" is answered by position, and "which one is running"
// by the single entry whose status is `doing`.
type InstallationProgress struct {
	ElapsedTime int `json:"elapsedTime"`
	Progress    []struct {
		Comment string `json:"comment"`
		Error   string `json:"error"`
		Status  string `json:"status"`
	} `json:"progress"`
}

// current returns the step being worked on, its position, and whether there is
// one at all. A finished-but-not-yet-cleared installation has every step done.
func (p InstallationProgress) current() (comment string, position int, ok bool) {
	for i, step := range p.Progress {
		if step.Status == "doing" {
			return step.Comment, i + 1, true
		}
	}
	return "", 0, false
}

// failed returns the first step that reported an error.
func (p InstallationProgress) failed() (comment, reason string, ok bool) {
	for _, step := range p.Progress {
		if step.Status == "error" {
			return step.Comment, step.Error, true
		}
	}
	return "", "", false
}

// isNotBeingInstalled recognises the API saying nothing is running.
//
// The route answers 404 outside an installation — verified on three servers —
// with "Server is not being installed or reinstalled at the moment". That is
// an answer, not a breakdown, and it is also the reason this endpoint cannot
// carry the verdict of a --wait: it says the same thing before an install has
// started and after it has finished. The task remains the authority on
// whether the work is done; this one only says where it has got to.
func isNotBeingInstalled(err error) bool {
	return err != nil &&
		strings.Contains(strings.ToLower(err.Error()), "not being installed or reinstalled")
}

// fetchInstallProgress reads install/status, separating "nothing is running"
// from "the call failed".
func fetchInstallProgress(server string) (InstallationProgress, bool, error) {
	var progress InstallationProgress

	path := fmt.Sprintf("/v1/dedicated/server/%s/install/status", url.PathEscape(server))
	if err := httpLib.Client.Get(path, &progress); err != nil {
		if isNotBeingInstalled(err) {
			return progress, false, nil
		}
		return progress, false, err
	}

	return progress, true, nil
}

// ShowBaremetalInstallStatus reports where a running installation has got to.
func ShowBaremetalInstallStatus(_ *cobra.Command, args []string) {
	server := args[0]

	progress, installing, err := fetchInstallProgress(server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to fetch the installation status of %s: %s", server, err)
		return
	}

	if !installing {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"server": server, "installing": false},
			"%s is not being installed or reinstalled at the moment.", server)
		return
	}

	steps := make([]map[string]any, 0, len(progress.Progress))
	for i, step := range progress.Progress {
		steps = append(steps, map[string]any{
			"position": i + 1,
			"status":   step.Status,
			"comment":  step.Comment,
			"error":    step.Error,
		})
	}

	// elapsedTime is attributed to the API rather than presented as the age of
	// the installation, because it is not one. Measured across a real
	// reinstall: -1935 during the hardware reboot, +42 four minutes later,
	// then -87 again after the final reboot. It counts up at one per second in
	// each stretch and does not reset between steps — step 5 at 42s became
	// step 10 at 72s — so it is a continuous counter whose origin is rebased
	// every time the machine restarts, which is what its clock does before NTP
	// catches up. Reporting it as "running for 1m47s" when the install had
	// been going for six minutes would state something measured to be false.
	elapsed := ""
	if progress.ElapsedTime >= 0 {
		elapsed = " · API reports " + formatElapsed(progress.ElapsedTime) + " elapsed"
	}

	summary := fmt.Sprintf("%d step(s)%s", len(steps), elapsed)
	if comment, position, ok := progress.current(); ok {
		summary = fmt.Sprintf("step %d of %d — %s%s", position, len(steps), comment, elapsed)
	}

	display.OutputObject(map[string]any{
		"installing":  true,
		"elapsedTime": progress.ElapsedTime,
		"summary":     summary,
		"steps":       steps,
	}, server, installStatusTemplate, &flags.OutputFormatConfig)
}

// unknownElapsed is what is printed instead of a duration the API did not
// give a usable value for.
const unknownElapsed = "an unknown time"

// formatElapsed prints seconds the way somebody watching a reinstall reads
// them. The API counts in seconds, and a reinstall runs for tens of minutes,
// so "1834" is the one form nobody wants.
//
// Negative values are not turned into durations. elapsedTime is not always one:
// measured on a real reinstall of an ADVANCE-1, it answered -1935 while the
// machine had been installing for under two minutes, and it then counted up at
// one per second — so the field is a correct counter with an origin set some
// thirty minutes ahead. Its differences are true and its absolute value is not,
// which is why --wait times itself instead of asking.
func formatElapsed(seconds int) string {
	if seconds < 0 {
		return unknownElapsed
	}
	d := time.Duration(seconds) * time.Second
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	return fmt.Sprintf("%dm%02ds", int(d.Minutes()), int(d.Seconds())%60)
}

// installProgressNote is the extra line --wait prints while it follows the
// reinstall task, so that a twenty-minute wait says something other than the
// same sentence over and over.
//
// It answers an empty string whenever it cannot say anything useful: this is
// decoration on top of the task poll, and a progress read that fails must
// never turn a running installation into a failed command.
func installProgressNote(server string, since time.Time) string {
	progress, installing, err := fetchInstallProgress(server)
	if err != nil || !installing {
		return ""
	}

	// The wait knows when it sent the request, so it says how long it has been
	// waiting rather than repeating a field that answered -1935 seconds on the
	// installation this was measured against.
	elapsed := formatElapsed(int(time.Since(since).Seconds()))

	if comment, position, ok := progress.current(); ok {
		return fmt.Sprintf("step %d/%d: %s (%s elapsed)",
			position, len(progress.Progress), comment, elapsed)
	}

	if comment, reason, ok := progress.failed(); ok {
		return fmt.Sprintf("step %q reported an error: %s", comment, reason)
	}

	return fmt.Sprintf("%d step(s), %s elapsed", len(progress.Progress), elapsed)
}
