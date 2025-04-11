package cmd

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var (
	baremetalColumnsToDisplay = []string{"name", "region", "os", "powerState", "state"}

	//go:embed templates/baremetal.tmpl
	baremetalTemplate string

	// Installation flags
	customizations struct {
		ConfigDriveUserData             string            `json:"configDriveUserData,omitempty"`
		EfiBootloaderPath               string            `json:"efiBootloaderPath,omitempty"`
		Hostname                        string            `json:"hostname,omitempty"`
		HttpHeaders                     map[string]string `json:"httpHeaders,omitempty"`
		ImageCheckSum                   string            `json:"imageCheckSum,omitempty"`
		ImageCheckSumType               string            `json:"imageCheckSumType,omitempty"`
		ImageType                       string            `json:"imageType,omitempty"`
		ImageURL                        string            `json:"imageURL,omitempty"`
		Language                        string            `json:"language,omitempty"`
		PostInstallationScript          string            `json:"postInstallationScript,omitempty"`
		PostInstallationScriptExtension string            `json:"postInstallationScriptExtension,omitempty"`
		SshKey                          string            `json:"sshKey,omitempty"`
	}
)

func listBaremetal(_ *cobra.Command, _ []string) {
	manageListRequest("/dedicated/server", baremetalColumnsToDisplay, genericFilters)
}

func listBaremetalTasks(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/dedicated/server/%s/task", args[0])
	manageListRequest(url, []string{"taskId", "function", "status", "startDate", "doneDate"}, genericFilters)
}

func getBaremetal(_ *cobra.Command, args []string) {
	manageObjectRequest("/dedicated/server", args[0], baremetalTemplate)
}

func rebootBaremetal(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/dedicated/server/%s/reboot", args[0])

	if err := client.Post(url, nil, nil); err != nil {
		log.Fatalf("error rebooting server %s: %s\n", args[0], err)
	}

	fmt.Println("\n⚡️ Reboot is started ...")
}

func reinstallBaremetal(_ *cobra.Command, args []string) {
	os := args[1]
	url := fmt.Sprintf("/dedicated/server/%s/reinstall", args[0])
	parameters := map[string]any{
		"operatingSystem": os,
		"customizations":  customizations,
	}

	if err := client.Post(url, parameters, nil); err != nil {
		log.Fatalf("error reinstalling server %s: %s\n", args[0], err)
	}

	fmt.Println("\n⚡️ Reinstallation started ...")
}

func init() {
	baremetalCmd := &cobra.Command{
		Use:   "baremetal",
		Short: "Retrieve information and manage your Baremetal services",
	}

	// Command to list Baremetal services
	baremetalListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Baremetal services",
		Run:   listBaremetal,
	}
	baremetalListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	baremetalCmd.AddCommand(baremetalListCmd)

	// Command to get a single Baremetal
	baremetalCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Baremetal",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getBaremetal,
	})

	// Command to list baremetal tasks
	baremetalListTasksCmd := &cobra.Command{
		Use:        "list-tasks",
		Short:      "Retrieve tasks of a specific Baremetal",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        listBaremetalTasks,
	}
	baremetalListTasksCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	baremetalCmd.AddCommand(baremetalListTasksCmd)

	// Command to reboot a baremetal
	baremetalCmd.AddCommand(&cobra.Command{
		Use:        "reboot",
		Short:      "Reboot a specific Baremetal",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        rebootBaremetal,
	})

	// Command to reinstall a baremetal
	reinstallBaremetalCmd := &cobra.Command{
		Use:        "reinstall",
		Short:      "Reinstall a specific Baremetal",
		Args:       cobra.ExactArgs(2),
		ArgAliases: []string{"service_name", "operating_system"},
		Run:        reinstallBaremetal,
	}
	reinstallBaremetalCmd.Flags().StringVar(&customizations.ConfigDriveUserData, "config-drive-user-data", "", "Config Drive UserData")
	reinstallBaremetalCmd.Flags().StringVar(&customizations.EfiBootloaderPath, "efi-bootloader-path", "", "Path of the EFI bootloader from the OS installed on the server")
	reinstallBaremetalCmd.Flags().StringVar(&customizations.Hostname, "hostname", "", "Custom hostname")
	reinstallBaremetalCmd.Flags().StringToStringVar(&customizations.HttpHeaders, "http-headers", nil, "Image HTTP headers")
	reinstallBaremetalCmd.Flags().StringVar(&customizations.ImageCheckSum, "image-checksum", "", "Image checksum")
	reinstallBaremetalCmd.Flags().StringVar(&customizations.ImageCheckSumType, "image-checksum-type", "", "Image checksum type")
	reinstallBaremetalCmd.Flags().StringVar(&customizations.ImageType, "image-type", "", "Image type (qcow, raw)")
	reinstallBaremetalCmd.Flags().StringVar(&customizations.ImageURL, "image-url", "", "Image URL")
	reinstallBaremetalCmd.Flags().StringVar(&customizations.Language, "language", "", "Display language")
	reinstallBaremetalCmd.Flags().StringVar(&customizations.PostInstallationScript, "post-installation-script", "", "Post-installation script")
	reinstallBaremetalCmd.Flags().StringVar(&customizations.PostInstallationScriptExtension, "post-installation-script-extension", "", "Post-installation script extension (cmd, ps1)")
	reinstallBaremetalCmd.Flags().StringVar(&customizations.SshKey, "ssh-key", "", "SSH public key")
	baremetalCmd.AddCommand(reinstallBaremetalCmd)

	rootCmd.AddCommand(baremetalCmd)
}
