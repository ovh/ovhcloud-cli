// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/backupservices"
	"github.com/spf13/cobra"
)

func init() {
	backupCmd := &cobra.Command{
		Use:   "backup-services",
		Short: "Retrieve information and manage your Veeam Backup services",
		Long: "Retrieve information and manage your Veeam Backup services.\n\n" +
			"A backup tenant holds storage vaults and a Veeam Service Provider Console tenant; " +
			"the console tenant is what drives the agents installed on your machines. Both levels " +
			"are resolved when the account has only one of them, so --tenant and --vspc are only " +
			"needed when there is a choice to make.\n\n" +
			"The agents themselves are managed from the machine they protect: see " +
			"`ovhcloud baremetal backup-agent`.",
	}

	// Both levels of the hierarchy are UUIDs, and both are resolved when there
	// is only one. These flags exist for the accounts where there is not.
	backupCmd.PersistentFlags().StringVar(&backupservices.Tenant, "tenant", "",
		"Backup tenant to work on (default: the only one on the account)")
	backupCmd.PersistentFlags().StringVar(&backupservices.Vspc, "vspc", "",
		"VSPC tenant to work on (default: the only one in the backup tenant)")

	// Tenants
	backupTenantCmd := &cobra.Command{
		Use:   "tenant",
		Short: "Show the backup tenants of the account",
	}
	backupCmd.AddCommand(backupTenantCmd)

	backupTenantCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your backup tenants",
		Run:     backupservices.ListTenants,
	}))

	backupTenantCmd.AddCommand(&cobra.Command{
		Use:   "get [<tenant_id>]",
		Short: "Show one backup tenant",
		Args:  cobra.MaximumNArgs(1),
		Run:   backupservices.ShowTenant,
	})

	// Vaults
	backupVaultCmd := &cobra.Command{
		Use:   "vault",
		Short: "Show and rename the storage vaults of a backup tenant",
	}
	backupCmd.AddCommand(backupVaultCmd)

	backupVaultCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List the storage vaults",
		Run:     backupservices.ListVaults,
	}))

	backupVaultCmd.AddCommand(&cobra.Command{
		Use:   "get <vault_id>",
		Short: "Show one storage vault",
		Args:  cobra.ExactArgs(1),
		Run:   backupservices.ShowVault,
	})

	backupVaultCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "buckets <vault_id>",
		Short: "List the buckets behind a storage vault",
		Args:  cobra.ExactArgs(1),
		Run:   backupservices.ListBuckets,
	}))

	backupVaultEditCmd := &cobra.Command{
		Use:   "edit <vault_id>",
		Short: "Rename a storage vault",
		Args:  cobra.ExactArgs(1),
		Run:   backupservices.EditVault,
	}
	backupVaultEditCmd.Flags().StringVar(&backupservices.EditSpec.Name, "name", "", "New name of the vault")
	addConfirmationFlags(backupVaultEditCmd, "Print the call that would be made without making it")
	backupVaultCmd.AddCommand(backupVaultEditCmd)

	// VSPC tenants
	backupVspcCmd := &cobra.Command{
		Use:   "vspc",
		Short: "Show and rename the Veeam Service Provider Console tenants",
	}
	backupCmd.AddCommand(backupVspcCmd)

	backupVspcCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List the VSPC tenants",
		Run:     backupservices.ListVspc,
	}))

	backupVspcCmd.AddCommand(&cobra.Command{
		Use:   "get [<vspc_id>]",
		Short: "Show one VSPC tenant",
		Args:  cobra.MaximumNArgs(1),
		Run:   backupservices.ShowVspc,
	})

	backupVspcEditCmd := &cobra.Command{
		Use:   "edit <vspc_id>",
		Short: "Rename a VSPC tenant",
		Args:  cobra.ExactArgs(1),
		Run:   backupservices.EditVspc,
	}
	backupVspcEditCmd.Flags().StringVar(&backupservices.EditSpec.Name, "name", "", "New name of the VSPC tenant")
	addConfirmationFlags(backupVspcEditCmd, "Print the call that would be made without making it")
	backupVspcCmd.AddCommand(backupVspcEditCmd)

	// What hangs off a VSPC tenant
	backupCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "policies",
		Short: "List the retention policies an agent can be put on",
		Run:   backupservices.ListPolicies,
	}))

	backupCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "agents",
		Short:   "List every backup agent, and what each one protects",
		Aliases: []string{"agent-list"},
		Run:     backupservices.ListAgents,
	}))

	backupCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "billing",
		Short: "Show what each part of the backup service costs, and what it has consumed",
		Long: "Show what each part of the backup service costs, and what it has consumed.\n\n" +
			"The backup API carries no price. The plan, the price, the period, the renewal mode and " +
			"the next billing date come from the account's service router, matched to each backup " +
			"resource by its identifier; the consumption comes from the account's current usage.",
		Run: backupservices.ShowBilling,
	}))

	backupCmd.AddCommand(&cobra.Command{
		Use: "deploy-script",
		// download-agent is the name the product brief uses for this; the
		// command is one thing under two names rather than two commands.
		Aliases: []string{"download-agent"},
		Short:   "Show the command that installs the backup agent on a machine",
		Long: "Show the command that installs the backup agent on a machine.\n\n" +
			"An agent created through the API exists as an object and protects nothing until this " +
			"script has run on the machine. The links carry their own authorisation.",
		Run: backupservices.ShowDeployScript,
	})

	backupLicenseCmd := &cobra.Command{
		Use:   "licenses",
		Short: "Show the Veeam licences held by a VSPC tenant",
	}
	backupCmd.AddCommand(backupLicenseCmd)

	backupLicenseCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List the Veeam licences",
		Run:     backupservices.ListLicenses,
	}))

	backupLicenseCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "servers <license_id>",
		Short: "List the backup servers driven by a licence",
		Args:  cobra.ExactArgs(1),
		Run:   backupservices.ListLicenseServers,
	}))

	rootCmd.AddCommand(backupCmd)
}
