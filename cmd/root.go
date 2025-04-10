/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"

	"github.com/ovh/go-ovh/ovh"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"

	"stash.ovh.net/api/ovh-cli/internal/config"
)

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "ovh-cli",
		Short: "CLI to manage your OVHcloud services",
	}
	// OVH API client
	client *ovh.Client

	// INI configuration file and its path
	cliConfig     *ini.File
	cliConfigPath string

	// Common flags used by all subcommands to control output format (json, yaml)
	jsonOutput, yamlOutput bool

	// Common filters that can be used in all listing commands
	genericFilters []string
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

	// Init API client
	client, err = ovh.NewDefaultClient()
	if err != nil {
		log.Print(`OVHcloud API client not initialized, please run "ovh-cli login" to authenticate`)
	}

	// Load configuration files by order of increasing priority. All configuration
	// files are optional. Only load file from user home if home could be resolve
	cliConfig, cliConfigPath = config.LoadINI()
	if err != nil {
		log.Printf("cannot load configuration: %s", err)
	}

	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON")
	rootCmd.PersistentFlags().BoolVar(&yamlOutput, "yaml", false, "Output in YAML")
	rootCmd.MarkFlagsMutuallyExclusive("json", "yaml")
}
