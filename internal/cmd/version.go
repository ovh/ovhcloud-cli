package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
)

var (
	version    = "undefined"
	lastCommit = "undefined"
)

func init() {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Get ovh-cli version",
		Run: func(_ *cobra.Command, _ []string) {
			data := []map[string]any{
				{
					"version":     version,
					"last_commit": lastCommit,
				},
			}

			display.RenderTable(data, []string{"version", "last_commit"}, &flags.OutputFormatConfig)
		},
	}

	// Disable parent pre-run that verifies if the API client is correctly initialized
	versionCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {}

	rootCmd.AddCommand(versionCmd)
}
