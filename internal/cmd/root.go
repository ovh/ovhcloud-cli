package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"stash.ovh.net/api/ovh-cli/internal/config"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/http"
)

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "ovh-cli",
		Short: "CLI to manage your OVHcloud services",
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	http.InitClient()

	// Load configuration files by order of increasing priority. All configuration
	// files are optional. Only load file from user home if home could be resolve
	flags.CliConfig, flags.CliConfigPath = config.LoadINI()

	rootCmd.PersistentFlags().BoolVar(&flags.Debug, "debug", false, "Activate debug mode (will log all HTTP requests details)")
	rootCmd.PersistentFlags().BoolVar(&flags.IgnoreErrors, "ignore-errors", false, "Ignore errors in API calls when it is not fatal to the execution")
	rootCmd.PersistentFlags().BoolVar(&flags.OutputFormatConfig.JsonOutput, "json", false, "Output in JSON")
	rootCmd.PersistentFlags().BoolVar(&flags.OutputFormatConfig.YamlOutput, "yaml", false, "Output in YAML")
	rootCmd.PersistentFlags().BoolVar(&flags.OutputFormatConfig.InteractiveOutput, "interactive", false, "Interactive output")
	rootCmd.PersistentFlags().StringVar(&flags.OutputFormatConfig.CustomFormat, "format", "", "Output value according to given format (expression using gval format)")
	rootCmd.MarkFlagsMutuallyExclusive("json", "yaml", "interactive", "format")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if http.Client == nil {
			display.ExitError("API client is not initialized, please run `ovh-cli login` to authenticate")
		}
	}
}

func removeRootFlagsFromCommand(subCommand *cobra.Command) {
	subCommand.SetHelpFunc(func(command *cobra.Command, strings []string) {
		rootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
			if flag.Name != "debug" {
				flag.Hidden = true
			}
		})
		command.Parent().HelpFunc()(command, strings)
	})
}

func withFilterFlag(c *cobra.Command) *cobra.Command {
	c.PersistentFlags().StringArrayVar(
		&flags.GenericFilters,
		"filter",
		nil,
		"Filter results by any property using https://github.com/PaesslerAG/gval syntax'",
	)

	return c
}
