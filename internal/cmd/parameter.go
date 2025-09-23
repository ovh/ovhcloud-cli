// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	_ "embed"
	"errors"
	"os"
	"runtime"

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
