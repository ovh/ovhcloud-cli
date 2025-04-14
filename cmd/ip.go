package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	ipColumnsToDisplay = []string{"ip", "rir", "routedTo.serviceName", "country", "description"}

	//go:embed templates/ip.tmpl
	ipTemplate string
)

func listIp(_ *cobra.Command, _ []string) {
	manageListRequest("/ip", ipColumnsToDisplay, genericFilters)
}

func getIp(_ *cobra.Command, args []string) {
	manageObjectRequest("/ip", args[0], ipTemplate)
}

func init() {
	ipCmd := &cobra.Command{
		Use:   "ip",
		Short: "Retrieve information and manage your Ip services",
	}

	// Command to list Ip services
	ipListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Ip services",
		Run:   listIp,
	}
	ipListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	ipCmd.AddCommand(ipListCmd)

	// Command to get a single Ip
	ipCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Ip",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getIp,
	})

	rootCmd.AddCommand(ipCmd)
}
