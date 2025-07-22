package main

import (
	"log"

	"stash.ovh.net/api/ovh-cli/internal/cmd"

	"github.com/spf13/cobra/doc"
)

func main() {
	rootCmd := cmd.GetRootCommand()
	err := doc.GenMarkdownTree(rootCmd, "./doc/")
	if err != nil {
		log.Fatal(err)
	}
}
