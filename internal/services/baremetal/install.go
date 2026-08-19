// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"fmt"
	"net/url"
	"strings"

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
