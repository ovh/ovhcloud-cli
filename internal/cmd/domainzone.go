package cmd

import (
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

	rootCmd.AddCommand(domainzoneCmd)
}
