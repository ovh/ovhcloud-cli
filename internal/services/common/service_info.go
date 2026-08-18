// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package common

import "github.com/spf13/cobra"

// The renewal flags shared by every `service-info edit` command, paired with
// the field each one sets in the API object.
//
// The registration and the payload builder live side by side on purpose: the
// flag name is the only thing that ties them together, so a rename that
// touches one and not the other would silently stop sending a setting rather
// than fail to compile.
var serviceInfoRenewFlags = []struct {
	name  string
	field string
	usage string
}{
	{"renew-automatic", "automatic", "Renew the service automatically"},
	{"renew-delete-at-expiration", "deleteAtExpiration", "Delete the service when it expires"},
	{"renew-forced", "forced", "Force the renewal"},
	{"renew-manual-payment", "manualPayment", "Pay the renewal manually"},
	{"renew-period", "period", "Renewal period, in months"},
}

// AddServiceInfoRenewFlags registers the renewal flags on a `service-info
// edit` command.
func AddServiceInfoRenewFlags(cmd *cobra.Command) {
	for _, flag := range serviceInfoRenewFlags {
		if flag.field == "period" {
			cmd.Flags().Int(flag.name, 0, flag.usage)
			continue
		}
		cmd.Flags().Bool(flag.name, false, flag.usage)
	}
}

// ServiceInfoRenewPayload returns the renewal settings the operator actually
// asked to change, and nothing else.
//
// The distinction matters more than it looks. These settings are booleans
// bound to a struct with no `omitempty`, so building the payload from that
// struct sends every one of them on every call — and a merge that lets the
// command line win then turns `--renew-period 12` into "set the period to 12
// AND switch automatic renewal off". The service kept renewing itself for
// years; one unrelated edit stopped it, and nothing said so.
//
// Reading `Changed` rather than the values also keeps `--renew-automatic=false`
// working: pflag records a flag as changed whatever value it was given, so an
// explicit false is sent while an absent flag stays absent.
func ServiceInfoRenewPayload(cmd *cobra.Command) map[string]any {
	renew := map[string]any{}

	for _, flag := range serviceInfoRenewFlags {
		if !cmd.Flags().Changed(flag.name) {
			continue
		}

		if flag.field == "period" {
			period, err := cmd.Flags().GetInt(flag.name)
			if err != nil {
				continue
			}
			renew[flag.field] = period
			continue
		}

		value, err := cmd.Flags().GetBool(flag.name)
		if err != nil {
			continue
		}
		renew[flag.field] = value
	}

	if len(renew) == 0 {
		return map[string]any{}
	}

	return map[string]any{"renew": renew}
}
