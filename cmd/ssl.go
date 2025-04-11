package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	sslColumnsToDisplay = []string{ "serviceName","type","authority","status" }

	//go:embed templates/ssl.tmpl
	sslTemplate string
)

func listSsl(_ *cobra.Command, _ []string) {
	manageListRequest("/ssl", sslColumnsToDisplay, genericFilters)
}

func getSsl(_ *cobra.Command, args []string) {
	manageObjectRequest("/ssl", args[0], sslTemplate)
}

func init() {
	sslCmd := &cobra.Command{
		Use:   "ssl",
		Short: "Retrieve information and manage your Ssl services",
	}

	// Command to list Ssl services
	sslListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Ssl services",
		Run:   listSsl,
	}
	sslListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	sslCmd.AddCommand(sslListCmd)

	// Command to get a single Ssl
	sslCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Ssl",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getSsl,
	})

	rootCmd.AddCommand(sslCmd)
}
