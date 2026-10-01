// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
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

	common.RenderFilteredTable(rows, []string{"name", "template"})
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
				// capacity and speed are complexType.UnitAndValue in the
				// schema, not scalars. Left as `any` they reached the table as
				// the decoded map and rendered as "map[unit:GB value:1000]",
				// truncated to the cell width — on every server that actually
				// has a controller, which is none of the fourteen measured,
				// which is why nobody saw it.
				Capacity  unitAndValue `json:"capacity"`
				DiskGroup int          `json:"diskGroupId"`
				Names     []string     `json:"names"`
				Speed     unitAndValue `json:"speed"`
				Type      any          `json:"type"`
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
				"capacity":    disk.Capacity.String(),
				"speed":       disk.Speed.String(),
				"diskType":    disk.Type,
			})
		}
	}

	common.RenderFilteredTable(rows,
		[]string{"controller", "type", "diskGroupId", "disks", "capacity", "speed", "diskType"})
}

// unitAndValue renders a quantity as one string, whichever of the two shapes
// the API sends.
//
// The schema types capacity as complexType.UnitAndValue_long and speed as
// UnitAndValue_string, so both are objects. Left as `any`, they reached the
// table as the decoded map and rendered as "map[unit:GB value:1000]", truncated
// to the cell width.
//
// Both forms are accepted because the object form cannot be confirmed here:
// hardwareRaidProfile answers 403 on all 35 servers of the account, so nobody
// has seen a populated response. Decoding into a struct alone would turn an
// ugly cell into a broken command if the API sends the scalar the previous
// fixture assumed; accepting either costs eight lines and cannot be wrong.
type unitAndValue struct {
	Unit  string
	Value string
}

func (u *unitAndValue) UnmarshalJSON(data []byte) error {
	var scalar string
	if err := json.Unmarshal(data, &scalar); err == nil {
		u.Value, u.Unit = scalar, ""
		return nil
	}

	var object struct {
		Unit  string      `json:"unit"`
		Value json.Number `json:"value"`
	}
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}
	u.Unit, u.Value = object.Unit, object.Value.String()

	return nil
}

// String is empty rather than "0 " when nothing came back, so an absent field
// renders as an empty cell instead of a number the API never gave.
func (u unitAndValue) String() string {
	switch {
	case u.Value == "":
		return ""
	case u.Unit == "":
		return u.Value
	}

	return fmt.Sprintf("%s %s", u.Value, u.Unit)
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
