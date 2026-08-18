// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/cache"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

// Flags of `power`.
var (
	PowerWait    bool
	PowerTimeout time.Duration
	PowerBootID  int
)

// There is no power endpoint. Not /power, not /shutdown, not /poweroff — the
// 126 paths of the dedicated server API carry none of them, which is how this
// capability came to be written off as missing.
//
// It is not missing, it is spelled differently: `power` is a first-class member
// of dedicated.server.BootTypeEnum, every server carries an entry of that type
// ("Power-off server", bootId 95083 in Europe and 95644 in Canada on the 35
// servers this was measured against), and bootId is writable. Setting that boot
// and rebooting halts the machine; `powerState` says so.
//
// Measured end to end on 18 August 2026 against a real server: off at t+143s,
// back on at t+207s.
const (
	powerBootType = "power"
	diskBootType  = "harddisk"

	// The bootId a server was on before it was powered off, so that powering it
	// back on returns it to where it was rather than to a guess. It is kept a
	// week: longer than any power-off worth doing, short enough that a stale
	// entry cannot claim to describe a server reinstalled since.
	powerCacheNamespace = "power"
	powerCacheTTL       = 7 * 24 * time.Hour
)

// bootEntry is one boot configuration of a server.
type bootEntry struct {
	BootID      int    `json:"bootId"`
	BootType    string `json:"bootType"`
	Kernel      string `json:"kernel"`
	Description string `json:"description"`
}

// GetBaremetalPowerStatus answers the question the other two commands act on.
func GetBaremetalPowerStatus(_ *cobra.Command, args []string) {
	server := args[0]

	state, err := readPowerState(server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	boot, err := readBootEntry(server, state.BootID)
	if err != nil {
		// The boot identifier is known; only its description is missing, and
		// that is not worth failing a read-only command over.
		boot = bootEntry{BootID: state.BootID, Description: "(unreadable)"}
	}

	details := map[string]any{
		"server":          server,
		"powerState":      state.PowerState,
		"state":           state.State,
		"bootId":          state.BootID,
		"bootType":        boot.BootType,
		"bootDescription": boot.Description,
	}

	message := fmt.Sprintf("%s is %s, set to boot on %q (bootId %d)",
		server, state.PowerState, boot.Description, state.BootID)

	// A server left on the power-off boot entry powers itself down again at the
	// next reboot, whoever asks for it and from wherever. That is worth saying
	// out loud rather than leaving in a bootId nobody reads.
	if boot.BootType == powerBootType {
		message += "\n⚠️  This server is set to power off: any reboot, including one from the manager, will shut it down again."
	}

	display.OutputInfo(&flags.OutputFormatConfig, details, "%s", message)
}

// PowerOffBaremetal halts a server by putting it on its power-off boot entry
// and rebooting into it.
func PowerOffBaremetal(_ *cobra.Command, args []string) {
	server := args[0]

	state, err := readPowerState(server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	target, err := findBootOfType(server, powerBootType)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"%s has no power-off boot entry, so it cannot be shut down this way: %s", server, err)
		return
	}

	if state.PowerState == "poweroff" {
		display.OutputError(&flags.OutputFormatConfig, "%s is already off", server)
		return
	}

	setBoot := fmt.Sprintf("/v1/dedicated/server/%s", url.PathEscape(server))
	reboot := fmt.Sprintf("/v1/dedicated/server/%s/reboot", url.PathEscape(server))

	if !common.ConfirmAction(common.Disruptive, server, fmt.Sprintf(
		"Powering off %s interrupts everything running on it, and it stays off until it is powered on again.", server)) {
		display.OutputError(&flags.OutputFormatConfig, "power off of %s cancelled", server)
		return
	}

	// Two calls, and the first one outlives the second: it rewrites the boot
	// configuration. A preview that showed the reboot alone would hide the
	// change that makes every later reboot shut the machine down.
	if common.ReportDryRun(
		common.Call{Method: "PUT", Endpoint: setBoot},
		common.Call{Method: "POST", Endpoint: reboot},
	) {
		return
	}

	// Remembered before anything is changed, because after the PUT the previous
	// value is gone from the API and only `power on` knows it mattered.
	rememberBoot(server, state.BootID)

	if err := httpLib.Client.Put(setBoot, map[string]any{"bootId": target.BootID}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to set the power-off boot entry on %s: %s", server, err)
		return
	}
	if err := httpLib.Client.Post(reboot, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"the boot entry of %s is now %q but the reboot failed, so it is still running: %s\nRun `ovhcloud baremetal power on %s` to put it back.",
			server, target.Description, err, server)
		return
	}

	details := map[string]any{
		"server": server, "bootId": target.BootID, "previousBootId": state.BootID,
	}

	if !PowerWait {
		display.OutputInfo(&flags.OutputFormatConfig, details,
			"⚡️ Power off of %s launched. It takes about two minutes; watch it with:\n  ovhcloud baremetal power status %s",
			server, server)
		return
	}

	if err := waitForPower(server, "poweroff"); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, details, "✅ %s is off.", server)
}

// PowerOnBaremetal starts a server again, and puts back the boot it was on.
func PowerOnBaremetal(_ *cobra.Command, args []string) {
	server := args[0]

	state, err := readPowerState(server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	restore, source, err := bootToRestore(server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Powering on is not the interesting half. Leaving the machine on its
	// power-off entry is: it would shut down again at the next reboot, and
	// nothing on screen would have said why. So the boot is put back even when
	// the server is already running.
	setBoot := fmt.Sprintf("/v1/dedicated/server/%s", url.PathEscape(server))
	reboot := fmt.Sprintf("/v1/dedicated/server/%s/reboot", url.PathEscape(server))

	calls := []common.Call{{Method: "PUT", Endpoint: setBoot}}
	if state.PowerState != "poweron" {
		calls = append(calls, common.Call{Method: "POST", Endpoint: reboot})
	}
	if common.ReportDryRun(calls...) {
		return
	}

	if err := httpLib.Client.Put(setBoot, map[string]any{"bootId": restore}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to set boot %d on %s: %s", restore, server, err)
		return
	}
	forgetBoot(server)

	if state.PowerState == "poweron" {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"server": server, "bootId": restore, "bootSource": source},
			"%s was already running; its boot is now %d (%s), so a reboot will no longer shut it down.",
			server, restore, source)
		return
	}

	if err := httpLib.Client.Post(reboot, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to power on %s: %s", server, err)
		return
	}

	details := map[string]any{"server": server, "bootId": restore, "bootSource": source}

	if !PowerWait {
		display.OutputInfo(&flags.OutputFormatConfig, details,
			"⚡️ Power on of %s launched, booting on %d (%s). It takes about three minutes.",
			server, restore, source)
		return
	}

	if err := waitForPower(server, "poweron"); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, details,
		"✅ %s is on, booting on %d (%s).", server, restore, source)
}

// bootToRestore decides what to boot on, and says where that came from.
//
// Three sources, in order of how well they answer "where was this server
// before": what the operator named, what this machine remembered, and — when
// neither exists — the disk. The last one is a fallback, not an answer, which
// is why the source travels with the value all the way to the message: a
// `power on` run from a different machine than the `power off` cannot know the
// server was sitting in rescue, and must not pretend it does.
func bootToRestore(server string) (int, string, error) {
	if PowerBootID != 0 {
		return PowerBootID, "given with --boot", nil
	}

	if remembered, found := recallBoot(server); found {
		return remembered, "the boot it was on before it was powered off", nil
	}

	disk, err := findBootOfType(server, diskBootType)
	if err != nil {
		return 0, "", fmt.Errorf(
			"this machine has no record of the boot %s was on, and the server has no disk boot entry to fall back on (%s); name one with --boot, or list them with `ovhcloud baremetal boot list %s`",
			server, err, server)
	}
	return disk.BootID, "the disk, because this machine has no record of the previous boot", nil
}

// waitForPower polls the server until it reports the state asked for.
//
// It polls powerState and not the reboot task, because the task is done long
// before the machine is: measured on 18 August 2026, `hardReboot` reported done
// after 60 seconds while the server only went dark at 143. A wait on the task
// would return with the machine still running and call it a success.
func waitForPower(server, expected string) error {
	deadline := time.Now().Add(PowerTimeout)
	for {
		state, err := readPowerState(server)
		if err != nil {
			return err
		}
		if state.PowerState == expected {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf(
				"%s still reports %q after %s; the request was accepted, so check again with `ovhcloud baremetal power status %s`",
				server, state.PowerState, PowerTimeout, server)
		}
		time.Sleep(10 * time.Second)
	}
}

type powerState struct {
	PowerState string `json:"powerState"`
	State      string `json:"state"`
	BootID     int    `json:"bootId"`
}

func readPowerState(server string) (powerState, error) {
	var state powerState
	endpoint := fmt.Sprintf("/v1/dedicated/server/%s", url.PathEscape(server))
	if err := httpLib.Client.Get(endpoint, &state); err != nil {
		return powerState{}, fmt.Errorf("failed to read the state of %s: %s", server, err)
	}
	return state, nil
}

func readBootEntry(server string, bootID int) (bootEntry, error) {
	var entry bootEntry
	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/boot/%d", url.PathEscape(server), bootID)
	if err := httpLib.Client.Get(endpoint, &entry); err != nil {
		return bootEntry{}, err
	}
	return entry, nil
}

// findBootOfType asks the API which entry does the job rather than carrying a
// table of identifiers. They are not the same everywhere — the power-off entry
// is 95083 in Europe and 95644 in Canada — and a list copied into Go stops
// being true silently.
func findBootOfType(server, bootType string) (bootEntry, error) {
	var ids []int
	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/boot?bootType=%s",
		url.PathEscape(server), url.QueryEscape(bootType))
	if err := httpLib.Client.Get(endpoint, &ids); err != nil {
		return bootEntry{}, err
	}
	if len(ids) == 0 {
		return bootEntry{}, fmt.Errorf("this server has no boot entry of type %q", bootType)
	}
	return readBootEntry(server, ids[0])
}

// rememberBoot, recallBoot and forgetBoot keep the boot a server was on across
// the two invocations that power it off and back on.
//
// This is a convenience and never a dependency: every caller works without it,
// and the fallback is stated in the output rather than hidden. A cache that
// silently decided where a server boots would be worse than no cache.
func rememberBoot(server string, bootID int) {
	data, err := json.Marshal(map[string]int{"bootId": bootID})
	if err != nil {
		return
	}
	cache.Write(powerCacheNamespace, powerCacheKey(server), data, powerCacheTTL)
}

func recallBoot(server string) (int, bool) {
	data, found := cache.Read(powerCacheNamespace, powerCacheKey(server), powerCacheTTL)
	if !found {
		return 0, false
	}
	var stored struct {
		BootID int `json:"bootId"`
	}
	if err := json.Unmarshal(data, &stored); err != nil || stored.BootID == 0 {
		return 0, false
	}
	return stored.BootID, true
}

func forgetBoot(server string) {
	cache.Remove(powerCacheNamespace, powerCacheKey(server))
}

// powerCacheKey escapes the server name because it names a file: a service name
// is not chosen by this CLI and must not be able to choose a path.
func powerCacheKey(server string) string {
	return url.PathEscape(server) + ".json"
}
