/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/ovh/go-ovh/ovh"
	"github.com/spf13/cobra"
)

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "ovh-cli",
		Short: "CLI to manage your OVHcloud services",
	}
	// OVH API client
	client *ovh.Client
	// Common flags used by all subcommands to control output format (json, yaml)
	jsonOutput, yamlOutput bool
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
	var err error
	client, err = ovh.NewDefaultClient()
	if err != nil {
		fmt.Printf("error initializing client: %s\n", err)
		os.Exit(1)
	}

	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON")
	rootCmd.PersistentFlags().BoolVar(&yamlOutput, "yaml", false, "Output in YAML")
	rootCmd.MarkFlagsMutuallyExclusive("json", "yaml")
}
