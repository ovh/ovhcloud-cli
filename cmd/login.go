package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/config"
	"stash.ovh.net/api/ovh-cli/internal/display"
)

func login(_ *cobra.Command, _ []string) {
	selectedRegion := display.RunChoicePicker("Which OVHcloud API do you want to login to ?", []string{"EU", "CA", "US"})

	if selectedRegion == "" {
		return
	}

	credentials := display.RunLoginInput()

	for k, v := range credentials {
		if v == "" {
			display.ExitError("no value provided for %q", k)
		}
	}

	regionConfigKey := fmt.Sprintf("ovh-%s", strings.ToLower(selectedRegion))

	if err := config.SetConfigValue(cliConfig, cliConfigPath, "", "endpoint", regionConfigKey); err != nil {
		display.ExitError("failed to write endpoint in configuration: %s", err)
	}

	for k, v := range credentials {
		if err := config.SetConfigValue(cliConfig, cliConfigPath, regionConfigKey, k, v); err != nil {
			display.ExitError("failed to write configuration %q: %s", k, err)
		}
	}
}

func init() {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login to your OVHcloud account to create API credentials",
		Run:   login,
	}

	// Disable parent pre-run that verifies if the API client is correctly initialized
	loginCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {}

	rootCmd.AddCommand(loginCmd)
}
