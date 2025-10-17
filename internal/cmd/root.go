// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync/atomic"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slices"

	"github.com/ovh/ovhcloud-cli/internal/config"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httplib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/version"
)

// rootCmd represents the base command when called without any subcommands
var (
	rootCmd = &cobra.Command{
		Use:   "ovhcloud",
		Short: "CLI to manage your OVHcloud services",
	}

	// wasmHiddenFlags are flags that should be hidden in WASM mode
	wasmHiddenFlags = []string{
		"editor",
		"from-file",
		"init-file",
		"replace",
		"interactive",
		"format",
		"debug",
	}

	wasmHiddenCommands = []string{
		"login",
		"config",
	}
)

func GetRootCommand() *cobra.Command {
	return rootCmd
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

func PostExecute() {
	// Reset output variables
	display.ResultString = ""
	display.ResultError = nil

	// Reset all flags to default values
	flags.GenericFilters = nil
	flags.OutputFormatConfig = display.OutputFormat{}
	flags.ParametersViaEditor = false
	flags.ParametersFile = ""

	// Recursively reset all flags of all subcommands to their default values
	resetSubCommandFlagValues(rootCmd)
}

// resetSubCommandFlagValues resets all flags of all subcommands of the given root command to their default values.
func resetSubCommandFlagValues(root *cobra.Command) {
	for _, c := range root.Commands() {
		c.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Changed {
				if f.Value.Type() == "stringArray" {
					// Special handling for stringArray for which we cannot
					// use DefValue since it is equal to "[]".
					if r, ok := f.Value.(pflag.SliceValue); ok {
						r.Replace(nil)
					}
				} else {
					f.Value.Set(f.DefValue)
				}
				f.Changed = false
			}
		})
		resetSubCommandFlagValues(c)
	}
}

func initRootCmd() {
	rootCmd.DisableAutoGenTag = true

	rootCmd.PersistentFlags().BoolVarP(&flags.Debug, "debug", "d", false, "Activate debug mode (will log all HTTP requests details)")
	rootCmd.PersistentFlags().BoolVarP(&flags.IgnoreErrors, "ignore-errors", "e", false, "Ignore errors in API calls when it is not fatal to the execution")
	rootCmd.PersistentFlags().BoolVarP(&flags.OutputFormatConfig.JsonOutput, "json", "j", false, "Output in JSON")
	rootCmd.PersistentFlags().BoolVarP(&flags.OutputFormatConfig.YamlOutput, "yaml", "y", false, "Output in YAML")
	rootCmd.PersistentFlags().BoolVarP(&flags.OutputFormatConfig.InteractiveOutput, "interactive", "i", false, "Interactive output")
	rootCmd.PersistentFlags().StringVarP(&flags.OutputFormatConfig.CustomFormat, "format", "f", "", `Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
Examples:
  --format 'id' (to extract a single field)
  --format 'nested.field.subfield' (to extract a nested field)
  --format '[id, 'name']' (to extract multiple fields as an array)
  --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
  --format 'name+","+type' (to extract and concatenate fields in a string)
  --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)`)
	rootCmd.MarkFlagsMutuallyExclusive("json", "yaml", "interactive", "format")

	var newVersionMessage atomic.Pointer[string]
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Check if a new version is available in a separate goroutine
		// Don't do it when running in WASM binary
		if !(runtime.GOARCH == "wasm" && runtime.GOOS == "js") {
			go func() {
				// Skip version check if version is undefined (development mode)
				if version.Version == "undefined" {
					return
				}

				const latestURL = "https://github.com/ovh/ovhcloud-cli/releases/latest"
				req, err := http.NewRequest("GET", latestURL, nil)
				if err != nil {
					return
				}
				req.Header.Set("Accept", "application/json")
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return
				}
				defer resp.Body.Close()
				var data struct {
					TagName string `json:"tag_name"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
					return
				}
				if data.TagName != "" && data.TagName != version.Version {
					message := fmt.Sprintf("A new version of ovhcloud-cli is available: %s (current: %s)", data.TagName, version.Version)
					newVersionMessage.Store(&message)
				}
			}()
		}

		// Check if the API client is initialized
		if httplib.Client == nil {
			display.OutputError(&flags.OutputFormatConfig, "API client is not initialized, please run `ovhcloud login` to authenticate")
			os.Exit(1) // Force os.Exit even in WASM mode
		}
	}

	// Set PostRun to display the new version message if available
	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		if msg := newVersionMessage.Load(); msg != nil {
			log.Println(*msg)
		}
	}
}

func init() {
	httplib.InitClient()

	// Load configuration files by order of increasing priority. All configuration
	// files are optional. Only load file from user home if home could be resolve
	flags.CliConfig, flags.CliConfigPath = config.LoadINI()

	initRootCmd()
}

func WasmCleanCommands() {
	// Remove commands that are not relevant in WASM mode
	for _, child := range rootCmd.Commands() {
		if slices.Contains(wasmHiddenCommands, child.Name()) {
			rootCmd.RemoveCommand(child)
		}
	}

	// Hide "completion" command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Recursively clean flags that are not relevant in WASM mode
	wasmCleanHelpCommands(rootCmd)
}

func wasmCleanHelpCommands(cmd *cobra.Command) {
	cmd.SetHelpFunc(func(command *cobra.Command, _ []string) {
		// Hide flags that are not relevant in WASM mode
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if slices.Contains(wasmHiddenFlags, flag.Name) {
				flag.Hidden = true
			}
		})

		// Return help string as JSON
		help, err := json.Marshal(map[string]string{
			"help": command.UsageString(),
		})
		if err != nil {
			display.ResultError = fmt.Errorf("failed to marshal help: %w", err)
			return
		}

		display.ResultString = string(help)
	})

	for _, child := range cmd.Commands() {
		wasmCleanHelpCommands(child)
	}
}

func withFilterFlag(c *cobra.Command) *cobra.Command {
	c.PersistentFlags().StringArrayVar(
		&flags.GenericFilters,
		"filter",
		nil,
		`Filter results by any property using https://github.com/PaesslerAG/gval syntax
Examples:
  --filter 'state="running"'
  --filter 'name=~"^my.*"'
  --filter 'nested.property.subproperty>10'
  --filter 'startDate>="2023-12-01"'
  --filter 'name=~"something" && nbField>10'`)

	return c
}
