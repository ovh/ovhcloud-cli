package cmd

import (
	"runtime/debug"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/version"
)

func init() {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Get OVHcloud CLI version",
		Run: func(_ *cobra.Command, _ []string) {
			data := map[string]any{
				"version": version.Version,
			}

			// Retrieve last commit information
			if info, ok := debug.ReadBuildInfo(); ok {
				for _, setting := range info.Settings {
					switch setting.Key {
					case "vcs.revision":
						data["last_commit"] = setting.Value
					case "vcs.time":
						data["last_commit_time"] = setting.Value
					}
				}
			}

			display.RenderTable([]map[string]any{data}, []string{"version", "last_commit", "last_commit_time"}, &flags.OutputFormatConfig)
		},
	}

	// Disable parent pre-run that verifies if the API client is correctly initialized
	versionCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {}

	rootCmd.AddCommand(versionCmd)
}
