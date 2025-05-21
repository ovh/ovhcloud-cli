package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/domainzone"
)

func init() {
	domainzoneCmd := &cobra.Command{
		Use:   "domain-zone",
		Short: "Retrieve information and manage your domain zones",
	}

	// Command to list DomainZone services
	domainzoneListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your domain zones",
		Run:   domainzone.ListDomainZone,
	}
	domainzoneCmd.AddCommand(withFilterFlag(domainzoneListCmd))

	// Command to get a single DomainZone
	domainzoneCmd.AddCommand(&cobra.Command{
		Use:   "get <zone_name>",
		Short: "Retrieve information of a specific domain zone",
		Args:  cobra.ExactArgs(1),
		Run:   domainzone.GetDomainZone,
	})

	rootCmd.AddCommand(domainzoneCmd)
}
