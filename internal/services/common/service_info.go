// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package common

import "github.com/spf13/cobra"

// ServiceInfoRenewFlag describes one renewal flag: what it is called on the
// command line, and which field of the API object it sets.
type ServiceInfoRenewFlag struct {
	Name  string
	Field string
	Usage string

	// Period is the one flag that carries a number rather than a yes or no.
	Period bool
}

// ServiceInfoRenewFlags is the single description of the renewal flags shared
// by every `service-info edit` command.
//
// It is exported rather than kept private because the command layer registers
// the flags and this layer reads them back: the flag name is the only thing
// tying the two halves together, so they must not each hold their own copy of
// it. One table, read twice — a rename in it changes both sides at once, and a
// rename anywhere else does not compile.
var ServiceInfoRenewFlags = []ServiceInfoRenewFlag{
	{Name: "renew-automatic", Field: "automatic", Usage: "Renew the service automatically"},
	{Name: "renew-delete-at-expiration", Field: "deleteAtExpiration", Usage: "Delete the service when it expires"},
	{Name: "renew-forced", Field: "forced", Usage: "Force the renewal"},
	{Name: "renew-manual-payment", Field: "manualPayment", Usage: "Pay the renewal manually"},
	{Name: "renew-period", Field: "period", Usage: "Renewal period, in months", Period: true},
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

	for _, flag := range ServiceInfoRenewFlags {
		if !cmd.Flags().Changed(flag.Name) {
			continue
		}

		if flag.Period {
			period, err := cmd.Flags().GetInt(flag.Name)
			if err != nil {
				continue
			}
			renew[flag.Field] = period
			continue
		}

		value, err := cmd.Flags().GetBool(flag.Name)
		if err != nil {
			continue
		}
		renew[flag.Field] = value
	}

	if len(renew) == 0 {
		return map[string]any{}
	}

	return map[string]any{"renew": renew}
}
