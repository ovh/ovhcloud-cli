package cmd

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
)

var (
	//go:embed templates/cloud_quota.tmpl
	cloudQuotaTemplate string
)

func getCloudQuota(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	url := fmt.Sprintf("/cloud/project/%s/region/%s/quota", projectID, url.PathEscape(args[0]))

	var object map[string]any
	if err := client.Get(url, &object); err != nil {
		display.ExitError("error fetching quotas for region %s: %s", args[0], err)
	}

	display.OutputObject(object, args[0], cloudQuotaTemplate, &outputFormatConfig)
}

func initCloudQuotaCommand(cloudCmd *cobra.Command) {
	quotaCmd := &cobra.Command{
		Use:   "quota",
		Short: "Check quotas in the given cloud project",
	}
	quotaCmd.PersistentFlags().StringVar(&cloudProject, "cloud-project", "", "Cloud project ID")

	quotaCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Get quotas for a specific region",
		Run:        getCloudQuota,
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"region"},
	})

	cloudCmd.AddCommand(quotaCmd)
}
