// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/services/domainzone"
	"github.com/spf13/cobra"
)

func init() {
	domainzoneCmd := &cobra.Command{
		Use:   "domain-zone",
		Short: "Retrieve information and manage your domain zones",
	}

	// Command to list DomainZone services
	domainzoneListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your domain zones",
		Run:     domainzone.ListDomainZone,
	}
	domainzoneCmd.AddCommand(withFilterFlag(domainzoneListCmd))

	// Command to get a single DomainZone
	domainzoneCmd.AddCommand(&cobra.Command{
		Use:   "get <zone_name>",
		Short: "Retrieve information of a specific domain zone",
		Args:  cobra.ExactArgs(1),
		Run:   domainzone.GetDomainZone,
	})

	domainzoneCmd.AddCommand(&cobra.Command{
		Use:   "refresh <zone_name>",
		Short: "Refresh the given zone",
		Args:  cobra.ExactArgs(1),
		Run:   domainzone.RefreshZone,
	})

	domainZoneRecordCmd := &cobra.Command{
		Use:   "record",
		Short: "Retrieve information and manage your DNS records within a zone",
	}
	domainzoneCmd.AddCommand(domainZoneRecordCmd)

	domainZoneRecordGetCmd := &cobra.Command{
		Use:   "get <zone_name> <record_id>",
		Short: "Get a single DNS record from your zone",
		Args:  cobra.ExactArgs(2),
		Run:   domainzone.GetRecord,
	}
	domainZoneRecordCmd.AddCommand(domainZoneRecordGetCmd)

	domainZoneRecordPutCmd := &cobra.Command{
		Use:   "update <zone_name> <record_id>",
		Short: "Update a single DNS record from your zone",
		Args:  cobra.ExactArgs(2),
		Run:   domainzone.UpdateRecord,
	}
	domainZoneRecordPutCmd.Flags().StringVar(&domainzone.UpdateRecordSpec.SubDomain, "subdomain", "", "Subdomain to update")
	domainZoneRecordPutCmd.Flags().StringVar(&domainzone.UpdateRecordSpec.Target, "target", "", "New target to apply")
	domainZoneRecordPutCmd.Flags().IntVar(&domainzone.UpdateRecordSpec.TTL, "ttl", 0, "New TTL to apply")

	addInitParameterFileFlag(domainZoneRecordPutCmd, assets.DomainOpenapiSchema, "/domain/zone/{zoneName}/record/{id}", "put", domainzone.RecordUpdateExample, nil)
	addInteractiveEditorFlag(domainZoneRecordPutCmd)
	addFromFileFlag(domainZoneRecordPutCmd)
	domainZoneRecordPutCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	domainZoneRecordCmd.AddCommand(domainZoneRecordPutCmd)

	rootCmd.AddCommand(domainzoneCmd)
}
