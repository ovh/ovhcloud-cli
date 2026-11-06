// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

// addServiceInfoRenewFlags registers the renewal flags on a `service-info edit`
// command.
//
// It sits here, beside the other shared flag helpers, rather than in the
// service package: declaring cobra flags is the command layer's job, and
// internal/services/common was the only service package doing it. The flag
// names still come from common.ServiceInfoRenewFlags, which is also what the
// payload builder reads — one table, so the two halves cannot drift.
func addServiceInfoRenewFlags(cmd *cobra.Command) {
	for _, flag := range common.ServiceInfoRenewFlags {
		if flag.Period {
			cmd.Flags().Int(flag.Name, 0, flag.Usage)
			continue
		}
		cmd.Flags().Bool(flag.Name, false, flag.Usage)
	}
}
