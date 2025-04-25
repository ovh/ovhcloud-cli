package cmd

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
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

func ipSetReverse(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/ip/%s/reverse", url.PathEscape(args[0]))
	if err := client.Post(url, map[string]string{
		"ipReverse": args[1],
		"reverse":   args[2],
	}, nil); err != nil {
		display.ExitError(err.Error())
	}

	fmt.Println("\n⚡️ Reverse correctly set")
}

func ipGetReverse(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/ip/%s/reverse", url.PathEscape(args[0]))
	manageListRequest(url, []string{"ipReverse", "reverse"}, genericFilters)
}

func ipDeleteReverse(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/ip/%s/reverse/%s", url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := client.Delete(url, nil); err != nil {
		display.ExitError(err.Error())
	}

	fmt.Println("\n⚡️ Reverse correctly deleted")
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

	ipReverseCmd := &cobra.Command{
		Use:   "reverse",
		Short: "Manage reverses on the given IP",
	}
	ipCmd.AddCommand(ipReverseCmd)

	ipReverseCmd.AddCommand(&cobra.Command{
		Use:        "set",
		Short:      "Set reverse on the given IP",
		Args:       cobra.ExactArgs(3),
		ArgAliases: []string{"service_name", "ip", "reverse"},
		Run:        ipSetReverse,
	})

	ipReverseCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "List reverse on the given IP range",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        ipGetReverse,
	})

	ipReverseCmd.AddCommand(&cobra.Command{
		Use:        "delete",
		Short:      "Delete reverse on the given IP",
		Args:       cobra.ExactArgs(2),
		ArgAliases: []string{"service_name", "ip"},
		Run:        ipDeleteReverse,
	})

	rootCmd.AddCommand(ipCmd)
}
