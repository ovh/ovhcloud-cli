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

	domainZoneRecordPostCmd := &cobra.Command{
		Use:   "create <zone_name>",
		Short: "Create a single DNS record in your zone",
		Args:  cobra.ExactArgs(1),
		Run:   domainzone.CreateRecord,
	}
	domainZoneRecordPostCmd.Flags().StringVar(&domainzone.CreateRecordSpec.FieldType, "field-type", "", "Record type (A, AAAA, CAA, CNAME, DKIM, DMARC, DNAME, HTTPS, LOC, MX, NAPTR, NS, PTR, RP, SPF, SRV, SSHFP, SVCB, TLSA, TXT)")
	domainZoneRecordPostCmd.Flags().StringVar(&domainzone.CreateRecordSpec.SubDomain, "sub-domain", "", "Record subDomain")
	domainZoneRecordPostCmd.Flags().StringVar(&domainzone.CreateRecordSpec.Target, "target", "", "Target of the record")
	domainZoneRecordPostCmd.Flags().IntVar(&domainzone.CreateRecordSpec.TTL, "ttl", 0, "TTL of the record")

	addInitParameterFileFlag(domainZoneRecordPostCmd, assets.DomainOpenapiSchema, "/domain/zone/{zoneName}/record", "post", domainzone.RecordCreateExample, nil)
	addInteractiveEditorFlag(domainZoneRecordPostCmd)
	addFromFileFlag(domainZoneRecordPostCmd)
	domainZoneRecordPostCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	domainZoneRecordCmd.AddCommand(domainZoneRecordPostCmd)

	domainZoneRecordPutCmd := &cobra.Command{
		Use:   "update <zone_name> <record_id>",
		Short: "Update a single DNS record from your zone",
		Args:  cobra.ExactArgs(2),
		Run:   domainzone.UpdateRecord,
	}
	domainZoneRecordPutCmd.Flags().StringVar(&domainzone.UpdateRecordSpec.SubDomain, "sub-domain", "", "Subdomain to update")
	domainZoneRecordPutCmd.Flags().StringVar(&domainzone.UpdateRecordSpec.Target, "target", "", "New target to apply")
	domainZoneRecordPutCmd.Flags().IntVar(&domainzone.UpdateRecordSpec.TTL, "ttl", 0, "New TTL to apply")

	addInitParameterFileFlag(domainZoneRecordPutCmd, assets.DomainOpenapiSchema, "/domain/zone/{zoneName}/record/{id}", "put", domainzone.RecordUpdateExample, nil)
	addInteractiveEditorFlag(domainZoneRecordPutCmd)
	addFromFileFlag(domainZoneRecordPutCmd)
	domainZoneRecordPutCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	domainZoneRecordCmd.AddCommand(domainZoneRecordPutCmd)

	domainZoneRecordDeleteCmd := &cobra.Command{
		Use:   "delete <zone_name> <record_id>",
		Short: "Delete a single DNS record from your zone",
		Args:  cobra.ExactArgs(2),
		Run:   domainzone.DeleteRecord,
	}
	domainZoneRecordCmd.AddCommand(domainZoneRecordDeleteCmd)

	rootCmd.AddCommand(domainzoneCmd)
}
