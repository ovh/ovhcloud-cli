package main

import (
	"log"

	"github.com/ovh/ovhcloud-cli/internal/cmd"

	"github.com/spf13/cobra/doc"
)

func main() {
	rootCmd := cmd.GetRootCommand()
	err := doc.GenMarkdownTree(rootCmd, "./doc/")
	if err != nil {
		log.Fatal(err)
	}
}
