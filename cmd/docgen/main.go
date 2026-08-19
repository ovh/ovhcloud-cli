// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"

	"github.com/ovh/ovhcloud-cli/internal/cmd"

	"github.com/spf13/cobra/doc"
)

// outputFlagSummary replaces the --output usage while the pages are written.
//
// cobra repeats the whole inherited-flags block on every page it generates,
// and the --output usage carries a dozen worked examples, so those examples
// were stored 934 times: two thirds of doc/ was one paragraph, copied. Worse
// than the size, it made the flag unchangeable — adding a single format
// rewrote every page in the repository, which put the change past the 300
// files a review can look at.
//
// Only the generated pages are shortened. `ovhcloud --help` still prints the
// examples in full, which is where somebody looking for them actually is.
// The usage must not contain backquotes: cobra reads the first backquoted
// word as the flag's value placeholder, which turned the -o line into
// "--output ovhcloud --help".
const outputFlagSummary = "Output format: json, yaml, interactive, or a custom format expression. " +
	"Run 'ovhcloud --help' for the full list with examples."

func main() {
	rootCmd := cmd.GetRootCommand()

	if flag := rootCmd.PersistentFlags().Lookup("output"); flag != nil {
		flag.Usage = outputFlagSummary
	}

	err := doc.GenMarkdownTree(rootCmd, "./doc/")
	if err != nil {
		log.Fatal(err)
	}
}
