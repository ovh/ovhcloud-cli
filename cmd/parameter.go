package cmd

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/openapi"
)

var (
	paramFile        string
	replaceParamFile bool
)

func addInitParameterFileFlag(cmd *cobra.Command, openapiSchema []byte, path, method, defaultContent string) {
	cmd.Flags().StringVar(&paramFile, "init-file", "", "Create a file with example parameters")
	cmd.Flags().BoolVar(&replaceParamFile, "replace", false, "Replace parameters file if it already exists")
	cmd.PreRun = func(_ *cobra.Command, _ []string) {
		if paramFile == "" {
			return
		}

		if !replaceParamFile {
			if _, err := os.Stat(paramFile); !errors.Is(err, os.ErrNotExist) {
				display.ExitError("file %q already exists", paramFile)
			}
		}

		examples, err := openapi.GetOperationRequestExamples(openapiSchema, path, method)
		if err != nil {
			display.ExitError("failed to fetch parameter file examples: %s", err)
		}

		var choice string
		if len(examples) > 0 {
			_, choice, err = display.RunGenericChoicePicker("Please select a parameter example", examples)
			if err != nil {
				display.ExitError(err.Error())
			}
		}

		if choice == "" {
			if defaultContent == "" {
				display.ExitWarning("No example selected, exiting...")
			} else {
				log.Print("No example chosen, using default value")
				choice = defaultContent
			}
		}

		tmplFile, err := os.Create(paramFile)
		if err != nil {
			display.ExitError("failed to create parameter file: %s", err)
		}
		defer tmplFile.Close()

		if _, err := tmplFile.WriteString(choice); err != nil {
			display.ExitError("error writing parameter file: %s", err)
		}

		fmt.Printf("\n⚡️ Parameter file written at %s\n", paramFile)
		os.Exit(0)
	}
}
