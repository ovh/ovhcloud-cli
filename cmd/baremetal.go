package cmd

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	filtersLib "stash.ovh.net/api/ovh-cli/internal/filters"
	"stash.ovh.net/api/ovh-cli/internal/openapi"
	"stash.ovh.net/api/ovh-cli/internal/utils"
)

type baremetalCustomizations struct {
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

var (
	baremetalColumnsToDisplay = []string{"name", "region", "os", "powerState", "state"}

	//go:embed templates/baremetal.tmpl
	baremetalTemplate string

	//go:embed parameter-samples/baremetal.json
	baremetalInstallationExample string

	//go:embed api-schemas/baremetal.json
	baremetalOpenapiSchema []byte

	// Installation flags
	installationFile string
	installViaEditor bool
	operatingSystem  string
	customizations   baremetalCustomizations

	// Virtual Network Interfaces Aggregation flags
	baremetalOLAInterfaces []string
	baremetalOLAName       string

	// IPMI flags
	baremetalIpmiTTL        int
	baremetalIpmiAccessType string
	baremetalIpmiIP         string
	baremetalIpmiSshKey     string
)

func listBaremetal(_ *cobra.Command, _ []string) {
	manageListRequest("/dedicated/server", "", baremetalColumnsToDisplay, genericFilters)
}

func listBaremetalTasks(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/dedicated/server/%s/task", args[0])
	manageListRequest(url, "", []string{"taskId", "function", "status", "startDate", "doneDate"}, genericFilters)
}

func getBaremetal(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/dedicated/server/%s", url.PathEscape(args[0]))

	// Fetch dedicated server
	var object map[string]any
	if err := client.Get(path, &object); err != nil {
		display.ExitError("error fetching %s: %s\n", path, err)
	}

	// Fetch running tasks
	path = fmt.Sprintf("/dedicated/server/%s/task", url.PathEscape(args[0]))
	tasks, err := fetchExpandedArray(path, "")
	if err != nil {
		display.ExitError("error fetching tasks for %s: %s", args[0], err)
	}
	object["tasks"] = tasks

	// Fetch network information
	path = fmt.Sprintf("/dedicated/server/%s/specifications/network", url.PathEscape(args[0]))
	var network map[string]any
	if err := client.Get(path, &network); err != nil {
		display.ExitError("error fetching network specifications for %s: %s\n", args[0], err)
	}
	object["network"] = network

	path = fmt.Sprintf("/dedicated/server/%s/serviceInfos", url.PathEscape(args[0]))
	var serviceInfo map[string]any
	if err := client.Get(path, &serviceInfo); err != nil {
		display.ExitError("error fetching billing information for %s: %s\n", args[0], err)
	}
	object["serviceInfo"] = serviceInfo

	display.OutputObject(object, args[0], baremetalTemplate, &outputFormatConfig)
}

func editBaremetal(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/dedicated/server/%s", url.PathEscape(args[0]))
	editor.EditResource(client, "/dedicated/server/{serviceName}", url, baremetalOpenapiSchema)
}

func rebootBaremetal(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/dedicated/server/%s/reboot", url.PathEscape(args[0]))

	if err := client.Post(url, nil, nil); err != nil {
		display.ExitError("error rebooting server %s: %s\n", args[0], err)
	}

	fmt.Println("\n⚡️ Reboot is started ...")
}

func rebootRescueBaremetal(cmd *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/dedicated/server/%s/boot?bootType=rescue", url.PathEscape(args[0]))

	var boots []int
	if err := client.Get(endpoint, &boots); err != nil {
		display.ExitError("failed to fetch boot options: %s", err)
	}

	if len(boots) == 0 {
		display.ExitError("no boot found for rescue mode")
	}

	// Update server with boot ID
	endpoint = fmt.Sprintf("/dedicated/server/%s", url.PathEscape(args[0]))
	if err := client.Put(endpoint, map[string]any{
		"bootId": boots[0],
	}, nil); err != nil {
		display.ExitError("failed to set boot ID %d for server: %s", boots[0], err)
	}

	// Reboot server
	endpoint += "/reboot"

	var task map[string]any
	if err := client.Post(endpoint, nil, &task); err != nil {
		display.ExitError("failed to reboot server: %s", err)
	}

	if !waitForTask {
		fmt.Println("\n⚡️ Reboot in rescue mode is started ...")
		return
	}

	if err := waitForDedicatedServerTask(args[0], task["taskId"]); err != nil {
		display.ExitError("failed to wait for server to be rebooted: %s", err)
	}

	fmt.Println("\n⚡️ Reboot done, fetching new authentication secrets...")

	// Fetch new secrets
	getBaremetalAuthenticationSecrets(cmd, args)
}

func waitForDedicatedServerTask(serviceName string, taskID any) error {
	endpoint := fmt.Sprintf("/dedicated/server/%s/task/%s", url.PathEscape(serviceName), taskID)

	for retry := 0; retry < 100; retry++ {
		var task map[string]any

		if err := client.Get(endpoint, &task); err != nil {
			return fmt.Errorf("failed to fetch task: %w", err)
		}

		switch task["status"] {
		case "done":
			return nil
		case "todo", "init", "doing":
			log.Printf("Still waiting for task to complete (status=%s) ...", task["status"])
			time.Sleep(30 * time.Second)
		default:
			return fmt.Errorf("invalid state for task %d: %s", taskID, task["status"])
		}
	}

	return fmt.Errorf("timeout waiting for task %d to be completed", taskID)
}

func baremetalGetIPMIAccess(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/dedicated/server/%s/features/ipmi/access", url.PathEscape(args[0]))

	parameters := map[string]any{
		"type": baremetalIpmiAccessType,
		"ttl":  baremetalIpmiTTL,
	}
	if baremetalIpmiIP != "" {
		parameters["ipToAllow"] = baremetalIpmiIP
	}
	if baremetalIpmiSshKey != "" {
		parameters["sshKey"] = baremetalIpmiSshKey
	}

	var task map[string]any
	if err := client.Post(path, parameters, &task); err != nil {
		display.ExitError("failed to request IMPI access: %s", err)
	}

	if err := waitForDedicatedServerTask(args[0], task["taskId"]); err != nil {
		display.ExitError("failed waiting for task: %s", err)
	}

	path += "?type=" + url.QueryEscape(baremetalIpmiAccessType)

	var accessDetails map[string]any
	if err := client.Get(path, &accessDetails); err != nil {
		display.ExitError("failed to fetch IPMI access information: %s", err)
	}

	output := fmt.Sprintf("\n⚡️ IPMI access: %s", accessDetails["value"])
	if expiration, ok := accessDetails["expiration"]; ok {
		output += fmt.Sprintf(" (expires at %s)", expiration)
	}

	fmt.Println(output)
}

func listBaremetalInterventions(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/dedicated/server/%s/intervention", args[0])

	interventions, err := fetchExpandedArray(path, "")
	if err != nil {
		display.ExitError("failed to fetch past interventions: %s", err)
	}

	for _, inter := range interventions {
		inter["status"] = "done"
	}

	path = fmt.Sprintf("/dedicated/server/%s/plannedIntervention", args[0])
	plannedInterventions, err := fetchExpandedArray(path, "")
	if err != nil {
		display.ExitError("failed to fetch planned interventions: %s", err)
	}

	for _, inter := range plannedInterventions {
		inter["date"] = inter["wantedStartDate"]
	}

	plannedInterventions = append(plannedInterventions, interventions...)

	plannedInterventions, err = filtersLib.FilterLines(plannedInterventions, genericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
	}

	display.RenderTable(plannedInterventions, []string{"type", "date", "status"}, &outputFormatConfig)
}

func listBaremetalBoots(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/dedicated/server/%s/boot", url.PathEscape(args[0]))

	boots, err := fetchExpandedArray(path, "")
	if err != nil {
		display.ExitError("error fetching boot options for server %q: %s", args[0], err)
	}

	for _, boot := range boots {
		path = fmt.Sprintf("/dedicated/server/%s/boot/%s/option", url.PathEscape(args[0]), boot["bootId"])

		options, err := fetchExpandedArray(path, "")
		if err != nil {
			display.ExitError("error fetching options of boot %d for server %s: %s", boot["bootId"], args[0], err)
		}

		boot["options"] = options
	}

	boots, err = filtersLib.FilterLines(boots, genericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
	}

	display.RenderTable(boots, []string{"bootId", "bootType", "description", "kernel"}, &outputFormatConfig)
}

func setBaremetalBootId(_ *cobra.Command, args []string) {
	bootID, err := strconv.Atoi(args[1])
	if err != nil {
		display.ExitError("invalid boot ID given, expected a number")
	}

	url := fmt.Sprintf("/dedicated/server/%s", url.PathEscape(args[0]))
	if err := client.Put(url, map[string]any{
		"bootId": bootID,
	}, nil); err != nil {
		display.ExitError("error setting boot ID: %s", err)
	}

	fmt.Printf("\n✅ Boot ID %d correctly configured\n", bootID)
}

func listBaremetalVNIs(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/dedicated/server/%s/virtualNetworkInterface", args[0])
	manageListRequest(url, "", []string{"uuid", "name", "mode", "vrack", "enabled"}, genericFilters)
}

func createBaremetalOLAAggregation(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/dedicated/server/%s/ola/aggregation", url.PathEscape(args[0]))
	if err := client.Post(url, map[string]any{
		"name":                     baremetalOLAName,
		"virtualNetworkInterfaces": baremetalOLAInterfaces,
	}, nil); err != nil {
		display.ExitError("failed to create OLA aggregation: %s", err)
	}
}

func resetBaremetalOLAAggregation(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/dedicated/server/%s/ola/reset", url.PathEscape(args[0]))

	for _, itf := range baremetalOLAInterfaces {
		if err := client.Post(url, map[string]string{
			"virtualNetworkInterface": itf,
		}, nil); err != nil {
			display.ExitError("failed to reset interface %s: %s", itf, err)
		}
		fmt.Printf("✅ Interface %s reset to default configuration ...\n", itf)
	}
}

func reinstallBaremetal(cmd *cobra.Command, args []string) {
	// No server ID given, print usage and exit
	if len(args) == 0 {
		cmd.Help()
		os.Exit(1)
	}

	// Create object from parameters given on command line
	jsonCliParameters, err := json.Marshal(struct {
		OS             string                  `json:"operatingSystem,omitempty"`
		Customizations baremetalCustomizations `json:"customizations"`
	}{
		OS:             operatingSystem,
		Customizations: customizations,
	})
	if err != nil {
		display.ExitError("failed to prepare arguments from command line: %s", err)
	}
	var cliParameters map[string]any
	if err := json.Unmarshal(jsonCliParameters, &cliParameters); err != nil {
		display.ExitError("failed to parse arguments from command line: %s", err)
	}

	var parameters map[string]any

	if isInputFromPipe() { // Install data given through a pipe
		var stdin []byte
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			stdin = append(stdin, scanner.Bytes()...)
		}
		if err := scanner.Err(); err != nil {
			display.ExitError(err.Error())
		}

		if err := json.Unmarshal(stdin, &parameters); err != nil {
			display.ExitError("failed to parse given installation data: %s", err)
		}
	} else if installViaEditor {
		log.Print("Flag --editor used, all other flags will override the example values")

		examples, err := openapi.GetOperationRequestExamples(baremetalOpenapiSchema, "/dedicated/server/{serviceName}/reinstall", "post", cliParameters)
		if err != nil {
			display.ExitError("failed to fetch API call examples: %s", err)
		}

		_, choice, err := display.RunGenericChoicePicker("Please select an installation example", examples)
		if err != nil {
			display.ExitError(err.Error())
		}

		if choice == "" {
			display.ExitWarning("No installation example selected, exiting...")
		}

		newValue, err := editor.EditValueWithEditor([]byte(choice))
		if err != nil {
			display.ExitError("failed to edit installation parameters using editor: %s", err)
		}

		if err := json.Unmarshal(newValue, &parameters); err != nil {
			display.ExitError("failed to parse given installation parameters: %s", err)
		}
	} else if installationFile != "" { // Install data given in a file
		log.Print("Flag --installation-file used, all other flags will override the file values")

		fd, err := os.Open(installationFile)
		if err != nil {
			display.ExitError("failed to open given file: %s", err)
		}
		defer fd.Close()

		content, err := io.ReadAll(fd)
		if err != nil {
			display.ExitError("failed to read installation file: %s", err)
		}

		if err := json.Unmarshal(content, &parameters); err != nil {
			display.ExitError("failed to parse given installation file: %s", err)
		}
	}

	// Only merge CLI parameters with other ones if not in --editor mode.
	// In this case, the CLI parameters have already been merged with the
	// request examples coming from API schemas.
	if !installViaEditor {
		parameters = utils.MergeMaps(cliParameters, parameters)
	}

	// Check if at least an OS was provided as it is mandatory
	if os, ok := parameters["operatingSystem"]; !ok || os == "" {
		display.ExitError("operating system parameter is mandatory to trigger a reinstallation")
	}

	out, err := json.MarshalIndent(parameters, "", " ")
	if err != nil {
		display.ExitError("installation parameters cannot be marshalled: %s", err)
	}

	log.Println("Installation parameters: \n" + string(out))

	var task map[string]any
	url := fmt.Sprintf("/dedicated/server/%s/reinstall", url.PathEscape(args[0]))
	if err := client.Post(url, parameters, &task); err != nil {
		display.ExitError("error reinstalling server %s: %s\n", args[0], err)
	}

	fmt.Println("\n⚡️ Reinstallation started ...")

	if !waitForTask {
		return
	}

	if err := waitForDedicatedServerTask(args[0], task["taskId"]); err != nil {
		display.ExitError("failed to wait for server to be reinstalled: %s", err)
	}

	fmt.Println("\n⚡️ Reinstall done, fetching new authentication secrets...")

	// Fetch new secrets
	getBaremetalAuthenticationSecrets(cmd, args)
}

func getBaremetalRelatedIPs(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/ip?routedTo.serviceName=%s", url.QueryEscape(args[0]))

	var ips []any
	if err := client.Get(path, &ips); err != nil {
		display.ExitError("failed to fetch IPs related to baremetal %s: %s", args[0], err)
	}

	ipsExpanded, err := fetchObjectsParallel[map[string]any]("/ip/%s", ips, false)
	if err != nil {
		display.ExitError("failed to fetch objects for each IP: %s", err)
	}

	ipsExpanded, err = filtersLib.FilterLines(ipsExpanded, genericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
	}

	display.RenderTable(ipsExpanded, []string{"ip", "type", "description", "campus"}, &outputFormatConfig)
}

func getBaremetalAuthenticationSecrets(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/dedicated/server/%s/authenticationSecret", url.PathEscape(args[0]))

	var allSecrets []map[string]any
	if err := client.Post(path, nil, &allSecrets); err != nil {
		display.ExitError("failed to fetch secrets IDs: %s", err)
	}

	for _, secret := range allSecrets {
		if secretID, ok := secret["password"]; ok {
			var secretValue map[string]any
			if err := client.Post("/secret/retrieve", map[string]any{
				"id": secretID,
			}, &secretValue); err != nil {
				display.ExitError("failed to retrieve secret value: %s", err)
			}
			maps.Copy(secret, secretValue)
		}
	}

	allSecrets, err := filtersLib.FilterLines(allSecrets, genericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
	}

	display.RenderTable(allSecrets, []string{"type", "url", "user", "secret", "expiration"}, &outputFormatConfig)
}

func getBaremetalCompatibleOses(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/dedicated/server/%s/install/compatibleTemplates", url.PathEscape(args[0]))

	var oses map[string]any
	if err := client.Get(path, &oses); err != nil {
		display.ExitError("failed to fetch compatible OSes: %s", err)
	}

	var formattedValues []map[string]any
	for _, os := range oses["ovh"].([]any) {
		formattedValues = append(formattedValues, map[string]any{
			"source": "ovh",
			"name":   os,
		})
	}
	for _, os := range oses["personal"].([]any) {
		formattedValues = append(formattedValues, map[string]any{
			"source": "personal",
			"name":   os,
		})
	}

	formattedValues, err := filtersLib.FilterLines(formattedValues, genericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
	}

	display.RenderTable(formattedValues, []string{"source", "name"}, &outputFormatConfig)
}

func init() {
	baremetalCmd := &cobra.Command{
		Use:   "baremetal",
		Short: "Retrieve information and manage your baremetal services",
	}

	// Command to list Baremetal services
	baremetalListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Baremetal services",
		Run:   listBaremetal,
	}
	baremetalCmd.AddCommand(withFilterFlag(baremetalListCmd))

	// Command to get a single Baremetal
	baremetalCmd.AddCommand(&cobra.Command{
		Use:        "get <service_name>",
		Short:      "Retrieve information of a specific baremetal",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getBaremetal,
	})

	// Command to edit a single Baremetal
	baremetalCmd.AddCommand(&cobra.Command{
		Use:        "edit <service_name>",
		Short:      "Update the given baremetal",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        editBaremetal,
	})

	// Command to list baremetal tasks
	baremetalListTasksCmd := &cobra.Command{
		Use:        "list-tasks <service_name>",
		Short:      "Retrieve tasks of the given baremetal",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        listBaremetalTasks,
	}
	baremetalCmd.AddCommand(withFilterFlag(baremetalListTasksCmd))

	// Command to reboot a baremetal
	baremetalRebootCmd := &cobra.Command{
		Use:        "reboot <service_name>",
		Short:      "Reboot the given baremetal",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        rebootBaremetal,
	}
	removeRootFlagsFromCommand(baremetalRebootCmd)
	baremetalCmd.AddCommand(baremetalRebootCmd)

	// Command to reboot a baremetal in rescue mode
	baremetalRebootRescueCmd := &cobra.Command{
		Use:   "reboot-rescue <service_name>",
		Short: "Reboot the given baremetal in rescue mode",
		Args:  cobra.ExactArgs(1),
		Run:   rebootRescueBaremetal,
	}
	removeRootFlagsFromCommand(baremetalRebootRescueCmd)
	baremetalRebootRescueCmd.Flags().BoolVar(&waitForTask, "wait", false, "Wait for reboot to be done before exiting")
	baremetalCmd.AddCommand(baremetalRebootRescueCmd)

	// Command to reinstall a baremetal
	reinstallBaremetalCmd := &cobra.Command{
		Use:   "reinstall <service_name>",
		Short: "Reinstall the given baremetal",
		Long: `Use this command to reinstall the given dedicated server.
There are three ways to define the installation parameters:

1. Using only CLI flags:

  ovh-cli baremetal reinstall ns1234.ip-11.22.33.net --os byolinux_64 --language fr-fr --image-url https://...

2. Using a configuration file

  First you can generate an example of installation file using the following command:

	ovh-cli baremetal reinstall --init-file ./install.json

  You will be able to choose from several installation examples. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct installation parameters, run:

	ovh-cli baremetal reinstall ns1234.ip-11.22.33.net --from-file ./install.json

  Note that you can also pipe the content of the file to reinstall, like the following:

	cat ./install.json | ovh-cli baremetal reinstall ns1234.ip-11.22.33.net

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovh-cli baremetal reinstall ns1234.ip-11.22.33.net --from-file ./install.json --hostname new-hostname

3. Using your default text editor

  ovh-cli baremetal reinstall ns1234.ip-11.22.33.net --editor

  You will be able to choose from several installation examples. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the reinstallation will be run.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovh-cli baremetal reinstall ns1234.ip-11.22.33.net --editor --os debian12_64

You can visit https://eu.api.ovh.com/console/?section=%2Fdedicated%2Fserver&branch=v1#post-/dedicated/server/-serviceName-/reinstall
to see all the available parameters and real life examples.

Please note that all parameters are not compatible with all OSes.
`,
		Args:       cobra.MaximumNArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        reinstallBaremetal,
	}

	addInitParameterFileFlag(reinstallBaremetalCmd, baremetalOpenapiSchema, "/dedicated/server/{serviceName}/reinstall", "post", baremetalInstallationExample)
	reinstallBaremetalCmd.Flags().StringVar(&installationFile, "from-file", "", "File containing installation parameters")
	reinstallBaremetalCmd.Flags().BoolVar(&installViaEditor, "editor", false, "Use a text editor to define installation parameters")
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
	reinstallBaremetalCmd.Flags().BoolVar(&waitForTask, "wait", false, "Wait for reinstall to be done before exiting")
	removeRootFlagsFromCommand(reinstallBaremetalCmd)
	reinstallBaremetalCmd.MarkFlagsMutuallyExclusive("from-file", "editor")
	baremetalCmd.AddCommand(reinstallBaremetalCmd)

	// List boots and their options
	baremetalBootCmd := &cobra.Command{
		Use:   "boot",
		Short: "Manage boot options for the given baremetal",
	}
	baremetalCmd.AddCommand(baremetalBootCmd)
	baremetalListBootsCmd := &cobra.Command{
		Use:        "list <service_name>",
		Short:      "List boot options for the given baremetal",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        listBaremetalBoots,
	}
	baremetalBootCmd.AddCommand(withFilterFlag(baremetalListBootsCmd))
	baremetalBootCmd.AddCommand(&cobra.Command{
		Use:        "set <service_name> <boot_id>",
		Short:      "Configure a boot ID on the given baremetal",
		Args:       cobra.ExactArgs(2),
		ArgAliases: []string{"service_name", "boot_id"},
		Run:        setBaremetalBootId,
	})

	baremetalListInterventionsCmd := &cobra.Command{
		Use:        "list-interventions <service_name>",
		Short:      "List past and planned interventions for the given baremetal",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        listBaremetalInterventions,
	}
	baremetalCmd.AddCommand(withFilterFlag(baremetalListInterventionsCmd))

	baremetalCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list-ips <service_name>",
		Short: "List all IPs that are routed to the given baremetal",
		Args:  cobra.ExactArgs(1),
		Run:   getBaremetalRelatedIPs,
	}))

	baremetalCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list-secrets <service_name>",
		Short: "Retrieve secrets to connect to the server",
		Args:  cobra.ExactArgs(1),
		Run:   getBaremetalAuthenticationSecrets,
	}))

	baremetalCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list-compatible-os <service_name>",
		Short: "Retrieve OSes that can be installed on this baremetal",
		Args:  cobra.ExactArgs(1),
		Run:   getBaremetalCompatibleOses,
	}))

	// Commands to manage virtual network interfaces
	baremetalVNICmd := &cobra.Command{
		Use:   "vni",
		Short: "Manage Virtual Network Interfaces of the given baremetal",
	}
	baremetalCmd.AddCommand(baremetalVNICmd)
	baremetalVNICmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:        "list <service_name>",
		Short:      "List Virtual Network Interfaces of the given baremetal",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        listBaremetalVNIs,
	}))
	baremetalVNICreateOLAAggregationCmd := &cobra.Command{
		Use:        "ola-create-aggregation <service_name> --name <name> --interface <uuid> --interface <uuid>",
		Short:      "Group interfaces into an aggregation",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        createBaremetalOLAAggregation,
	}
	baremetalVNICreateOLAAggregationCmd.Flags().StringArrayVar(&baremetalOLAInterfaces, "interface", nil, "Interfaces to group")
	baremetalVNICreateOLAAggregationCmd.MarkFlagRequired("interface")
	baremetalVNICreateOLAAggregationCmd.Flags().StringVar(&baremetalOLAName, "name", "", "Name of the aggregation")
	baremetalVNICreateOLAAggregationCmd.MarkFlagRequired("name")
	baremetalVNICmd.AddCommand(baremetalVNICreateOLAAggregationCmd)

	baremetalVNIResetOLAAggregationCmd := &cobra.Command{
		Use:        "ola-reset <service_name> --interface <uuid> --interface <uuid>",
		Short:      "Reset interfaces to default configuration",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        resetBaremetalOLAAggregation,
	}
	baremetalVNIResetOLAAggregationCmd.Flags().StringArrayVar(&baremetalOLAInterfaces, "interface", nil, "Interfaces to group")
	baremetalVNIResetOLAAggregationCmd.MarkFlagRequired("interface")
	baremetalVNICmd.AddCommand(baremetalVNIResetOLAAggregationCmd)

	baremetalIPMICmd := &cobra.Command{
		Use:   "ipmi",
		Short: "Manage IPMI on your baremetal",
	}
	baremetalCmd.AddCommand(baremetalIPMICmd)

	baremetalIPMIGetAccessCmd := &cobra.Command{
		Use:   "get-access <service_name> --type serialOverLanURL --ttl 5",
		Short: "Request an acces on KVM IPMI interface",
		Args:  cobra.ExactArgs(1),
		Run:   baremetalGetIPMIAccess,
	}
	baremetalIPMIGetAccessCmd.Flags().IntVar(&baremetalIpmiTTL, "ttl", 1, "Time to live in minutes for cache (1, 3, 5, 10, 15)")
	baremetalIPMIGetAccessCmd.MarkFlagRequired("ttl")
	baremetalIPMIGetAccessCmd.Flags().StringVar(&baremetalIpmiAccessType, "type", "", "Distinct way to acces a KVM IPMI session (kvmipHtml5URL, kvmipJnlp, serialOverLanSshKey, serialOverLanURL)")
	baremetalIPMIGetAccessCmd.MarkFlagRequired("type")
	baremetalIPMIGetAccessCmd.Flags().StringVar(&baremetalIpmiIP, "allowed-ip", "", "IPv4 address that can use the access")
	baremetalIPMIGetAccessCmd.Flags().StringVar(&baremetalIpmiSshKey, "ssh-key", "", "Public SSH key for Serial Over Lan SSH access")
	baremetalIPMICmd.AddCommand(baremetalIPMIGetAccessCmd)

	rootCmd.AddCommand(baremetalCmd)
}
