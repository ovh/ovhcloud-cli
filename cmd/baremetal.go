package cmd

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
)

var (
	baremetalColumnsToDisplay = []string{"name", "region", "os", "powerState", "state"}

	//go:embed templates/baremetal.tmpl
	baremetalTemplate string

	//go:embed parameter-samples/baremetal.json
	baremetalInstallationExample string

	// Installation flags
	installationFile string
	operatingSystem  string
	customizations   struct {
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
	path := fmt.Sprintf("/dedicated/server/%s", url.PathEscape(args[0]))

	// Fetch dedicated server
	var object map[string]any
	if err := client.Get(path, &object); err != nil {
		log.Fatalf("error fetching %s: %s\n", path, err)
	}

	// Fetch running tasks
	path = fmt.Sprintf("/dedicated/server/%s/task", url.PathEscape(args[0]))
	tasks, err := fetchExpandedArray(path)
	if err != nil {
		log.Fatalf("error fetching tasks for %s: %s", args[0], err)
	}
	object["tasks"] = tasks

	display.OutputObject(object, args[0], baremetalTemplate, jsonOutput, yamlOutput, interactiveOutput)
}

func rebootBaremetal(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/dedicated/server/%s/reboot", args[0])

	if err := client.Post(url, nil, nil); err != nil {
		log.Fatalf("error rebooting server %s: %s\n", args[0], err)
	}

	fmt.Println("\n⚡️ Reboot is started ...")
}

func reinstallBaremetal(cmd *cobra.Command, args []string) {
	// No server ID given, print usage and exit
	if len(args) == 0 {
		cmd.Help()
		os.Exit(1)
	}

	var parameters map[string]any

	if isInputFromPipe() { // Install data given through a pipe
		var stdin []byte
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			stdin = append(stdin, scanner.Bytes()...)
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		if err := json.Unmarshal(stdin, &parameters); err != nil {
			log.Fatalf("failed to parse given installation data: %s", err)
		}
	} else if installationFile != "" { // Install data given in a file
		log.Print("Flag --installation-file used, all other flags will be ignored")

		fd, err := os.Open(installationFile)
		if err != nil {
			log.Fatalf("failed to open given file: %s", err)
		}
		defer fd.Close()

		content, err := io.ReadAll(fd)
		if err != nil {
			log.Fatalf("failed to read installation file: %s", err)
		}

		if err := json.Unmarshal(content, &parameters); err != nil {
			log.Fatalf("failed to parse given installation file: %s", err)
		}
	} else { // Install data given via CLI flags
		if operatingSystem == "" {
			log.Fatalf("operating system parameter is mandatory to trigger a reinstallation")
		}

		parameters = map[string]any{
			"operatingSystem": operatingSystem,
			"customizations":  customizations,
		}
	}

	url := fmt.Sprintf("/dedicated/server/%s/reinstall", args[0])
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
		Use:   "reinstall",
		Short: "Reinstall a specific Baremetal",
		Long: `Use this command to reinstall the given dedicated server.
There are two ways to define the installation parameters:

1. Using CLI flags:

  ovh-cli baremetal reinstall ns1234.ip-11.22.33.net --os byolinux_64 --language fr-fr --image-url https://...

2. Using a configuration file

  First you can generate an example of installation file using the following command:

  ovh-cli baremetal reinstall --init-file ./install.json

  After editing the file to set the correct installation parameters, run:

  ovh-cli baremetal reinstall ns1234.ip-11.22.33.net --from-file ./install.json

  Note that you can also pipe the content of the file to reinstall, like the following:

  cat ./install.json | ovh-cli baremetal reinstall ns1234.ip-11.22.33.net

You can visit https://eu.api.ovh.com/console/?section=%2Fdedicated%2Fserver&branch=v1#post-/dedicated/server/-serviceName-/reinstall
to see all the available parameters and real life examples.
`,
		Args:       cobra.MaximumNArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        reinstallBaremetal,
	}

	addInitParameterFileFlag(reinstallBaremetalCmd, baremetalInstallationExample)
	reinstallBaremetalCmd.Flags().StringVar(&installationFile, "from-file", "", "File containing installation parameters")
	reinstallBaremetalCmd.Flags().StringVar(&operatingSystem, "os", "", "Operating system to install")
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
