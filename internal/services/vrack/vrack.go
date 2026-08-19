// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package vrack

import (
	_ "embed"
	"fmt"
	"net/url"

	httpLib "github.com/ovh/ovhcloud-cli/internal/http"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	vrackColumnsToDisplay = []string{"serviceName", "name", "description"}

	//go:embed templates/vrack.tmpl
	vrackTemplate string

	VrackSpec struct {
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
	}
)

func ListVrack(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/vrack", "", vrackColumnsToDisplay, flags.GenericFilters)
}

// GetVrack shows a vRack and what is inside it.
//
// It used to show the name, the description and the IAM URN, which meant a
// vRack holding two servers and an OVHcloud Connect printed exactly like an
// empty one. The contents are the reason anybody looks at a vRack, so they are
// read here and folded into the same object -- which also puts them in the
// -o json output, where a script can act on them.
func GetVrack(_ *cobra.Command, args []string) {
	vrack := args[0]

	var object map[string]any
	if err := httpLib.Client.Get(fmt.Sprintf("/v1/vrack/%s", url.PathEscape(vrack)), &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch vRack %s: %s", vrack, err)
		return
	}

	// The contents are additional information, not the object: a vRack that
	// cannot be listed is still worth printing. So a failure here costs the
	// section, never the command.
	servers, err := attachedInterfaces(vrack)
	if err != nil {
		display.OutputWarning(&flags.OutputFormatConfig, "%s", err)
	}
	others, unreadable := otherContents(vrack)

	object["servers"] = serverRows(servers)
	object["otherContents"] = others
	object["summary"] = summarise(len(servers), others, unreadable)
	object["anyServerNamed"] = anyNamed(servers)

	display.OutputObject(object, vrack, vrackTemplate, &flags.OutputFormatConfig)
}

// serverRows turns the interfaces into what the template prints.
func serverRows(interfaces []serverInterface) []map[string]any {
	rows := make([]map[string]any, 0, len(interfaces))
	for _, itf := range interfaces {
		name := ""
		if itf.DisplayName != "" && itf.DisplayName != itf.Server {
			name = itf.DisplayName
		}
		rows = append(rows, map[string]any{
			"name":          name,
			"server":        itf.Server,
			"interface":     itf.Name,
			"interfaceUuid": itf.UUID,
		})
	}
	return rows
}

// anyNamed reports whether the Name column has anything to say.
//
// Written after seeing the real output: a vRack whose servers are all unnamed
// printed the hostname twice per row, once under Name and once under Server.
// A column repeating its neighbour costs width -- and on a terminal narrow
// enough it is what pushes the row onto two lines, which is the thing the
// table was chosen to avoid. So the column appears when at least one server
// has a name of its own, and not otherwise.
func anyNamed(interfaces []serverInterface) bool {
	for _, itf := range interfaces {
		if itf.DisplayName != "" && itf.DisplayName != itf.Server {
			return true
		}
	}
	return false
}

func EditVrack(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/vrack/{serviceName}",
		fmt.Sprintf("/v1/vrack/%s", url.PathEscape(args[0])),
		VrackSpec,
		assets.VrackOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}
