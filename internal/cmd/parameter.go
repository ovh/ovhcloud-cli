// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
	"github.com/spf13/cobra"
)

var (
	paramFile        string
	replaceParamFile bool
)

func addInteractiveEditorFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define parameters")
}

func addFromFileFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&flags.ParametersFile, "from-file", "", "File containing parameters")
}

func addInitParameterFileFlag(cmd *cobra.Command, openapiSchema []byte, path, method, defaultContent string, replaceValueFn func(*cobra.Command, []string) (map[string]any, error)) {
	if runtime.GOARCH == "wasm" && runtime.GOOS == "js" {
		return
	}

	cmd.Flags().StringVar(&paramFile, "init-file", "", "Create a file with example parameters")
	cmd.Flags().BoolVar(&replaceParamFile, "replace", false, "Replace parameters file if it already exists")
	cmd.PreRun = func(_ *cobra.Command, args []string) {
		if paramFile == "" {
			return
		}

		if !replaceParamFile {
			if _, err := os.Stat(paramFile); !errors.Is(err, os.ErrNotExist) {
				display.OutputError(&flags.OutputFormatConfig, "file %q already exists", paramFile)
				return
			}
		}

		// Run given func to get replacement values, if not nil
		var (
			replaceValues map[string]any
			err           error
		)
		if replaceValueFn != nil {
			replaceValues, err = replaceValueFn(cmd, args)
			if err != nil {
				display.OutputError(&flags.OutputFormatConfig, "failed to get replacement values: %s", err)
				return
			}
		}

		// Get examples from OpenAPI schema and replace values with provided replacements
		examples, err := openapi.GetOperationRequestExamples(openapiSchema, path, method, defaultContent, replaceValues)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to fetch parameter file examples: %s", err)
			return
		}

		// Run choice picker to select an example
		var choice string
		if len(examples) > 0 {
			_, choice, err = display.RunGenericChoicePicker("Please select a parameter example", examples, 0)
			if err != nil {
				display.OutputError(&flags.OutputFormatConfig, "%s", err)
				return
			}
		}

		if choice == "" {
			display.OutputWarning(&flags.OutputFormatConfig, "No example selected, exiting…")
			return
		}

		// Write the selected example to the parameter file
		tmplFile, err := os.Create(paramFile)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to create parameter file: %s", err)
			return
		}
		defer tmplFile.Close()

		if _, err := tmplFile.WriteString(choice); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "error writing parameter file: %s", err)
			return
		}

		display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Parameter file written at %s", paramFile)
		os.Exit(0)
	}
}

func getGenericCreateCmd(resource, baseCommand, flagsSample, path, bodyExample string,
	openAPISchema []byte, positionalArgs []string, fn func(*cobra.Command, []string),
) *cobra.Command {
	var formattedArgs strings.Builder
	for _, arg := range positionalArgs {
		fmt.Fprintf(&formattedArgs, " <%s>", arg)
	}

	createCmd := &cobra.Command{
		Use:   "create" + formattedArgs.String(),
		Short: fmt.Sprintf("Create a new %s", resource),
		Long: fmt.Sprintf(`Use this command to create a new %[1]s.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud %[2]s %[3]s

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud %[2]s --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud %[2]s --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud %[2]s

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud %[2]s --from-file ./params.json %[3]s

3. Using your default text editor:

	ovhcloud %[2]s --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud %[2]s --editor %[3]s
`, resource, baseCommand, flagsSample),
		Run:  fn,
		Args: cobra.ExactArgs(len(positionalArgs)),
	}

	// Common flags for other means to define parameters
	addInitParameterFileFlag(createCmd, openAPISchema, path, "post", bodyExample, nil)
	addInteractiveEditorFlag(createCmd)
	addFromFileFlag(createCmd)
	createCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return createCmd
}
