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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ovhcloud",
	Short: "CLI to manage your OVHcloud services",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(args ...string) (string, error) {
	if len(args) != 0 {
		rootCmd.SetArgs(args)
	}

	err := rootCmd.Execute()
	if err != nil {
		return display.ResultString, err
	}

	return display.ResultString, display.ResultError
}

func init() {
	http.InitClient()

	// Load configuration files by order of increasing priority. All configuration
	// files are optional. Only load file from user home if home could be resolve
	flags.CliConfig, flags.CliConfigPath = config.LoadINI()

	rootCmd.PersistentFlags().BoolVarP(&flags.Debug, "debug", "d", false, "Activate debug mode (will log all HTTP requests details)")
	rootCmd.PersistentFlags().BoolVarP(&flags.IgnoreErrors, "ignore-errors", "e", false, "Ignore errors in API calls when it is not fatal to the execution")
	rootCmd.PersistentFlags().BoolVarP(&flags.OutputFormatConfig.JsonOutput, "json", "j", false, "Output in JSON")
	rootCmd.PersistentFlags().BoolVarP(&flags.OutputFormatConfig.YamlOutput, "yaml", "y", false, "Output in YAML")
	rootCmd.PersistentFlags().BoolVarP(&flags.OutputFormatConfig.InteractiveOutput, "interactive", "i", false, "Interactive output")
	rootCmd.PersistentFlags().StringVarP(&flags.OutputFormatConfig.CustomFormat, "format", "f", "", "Output value according to given format (expression using gval format)")
	rootCmd.MarkFlagsMutuallyExclusive("json", "yaml", "interactive", "format")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if http.Client == nil {
			display.ExitError("API client is not initialized, please run `ovhcloud login` to authenticate")
			os.Exit(1) // Force os.Exit even in WASM mode
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
