package baremetal

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
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/openapi"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
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
	baremetalColumnsToDisplay = []string{"name", "region", "iam.displayName displayName", "os", "state"}

	//go:embed templates/baremetal.tmpl
	baremetalTemplate string

	//go:embed parameter-samples/baremetal.json
	BaremetalInstallationExample string

	//go:embed api-schemas/baremetal.json
	BaremetalOpenapiSchema []byte

	// Installation flags
	InstallationFile string
	InstallViaEditor bool
	OperatingSystem  string
	Customizations   baremetalCustomizations

	// Virtual Network Interfaces Aggregation flags
	BaremetalOLAInterfaces []string
	BaremetalOLAName       string

	// IPMI flags
	BaremetalIpmiTTL        int
	BaremetalIpmiAccessType string
	BaremetalIpmiIP         string
	BaremetalIpmiSshKey     string
)

func ListBaremetal(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/dedicated/server", "", baremetalColumnsToDisplay, flags.GenericFilters)
}

func ListBaremetalTasks(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/dedicated/server/%s/task", args[0])
	common.ManageListRequest(url, "", []string{"taskId", "function", "status", "startDate", "doneDate"}, flags.GenericFilters)
}

func GetBaremetal(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/dedicated/server/%s", url.PathEscape(args[0]))

	// Fetch dedicated server
	var object map[string]any
	if err := httpLib.Client.Get(path, &object); err != nil {
		display.ExitError("error fetching %s: %s\n", path, err)
		return
	}

	// Fetch running tasks
	path = fmt.Sprintf("/dedicated/server/%s/task", url.PathEscape(args[0]))
	tasks, err := httpLib.FetchExpandedArray(path, "")
	if err != nil {
		display.ExitError("error fetching tasks for %s: %s", args[0], err)
		return
	}
	object["tasks"] = tasks

	// Fetch network information
	path = fmt.Sprintf("/dedicated/server/%s/specifications/network", url.PathEscape(args[0]))
	var network map[string]any
	if err := httpLib.Client.Get(path, &network); err != nil {
		display.ExitError("error fetching network specifications for %s: %s\n", args[0], err)
		return
	}
	object["network"] = network

	path = fmt.Sprintf("/dedicated/server/%s/serviceInfos", url.PathEscape(args[0]))
	var serviceInfo map[string]any
	if err := httpLib.Client.Get(path, &serviceInfo); err != nil {
		display.ExitError("error fetching billing information for %s: %s\n", args[0], err)
		return
	}
	object["serviceInfo"] = serviceInfo

	display.OutputObject(object, args[0], baremetalTemplate, &flags.OutputFormatConfig)
}

func EditBaremetal(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/dedicated/server/%s", url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/dedicated/server/{serviceName}", url, BaremetalOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}

func RebootBaremetal(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/dedicated/server/%s/reboot", url.PathEscape(args[0]))

	if err := httpLib.Client.Post(url, nil, nil); err != nil {
		display.ExitError("error rebooting server %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Reboot launched ...")
}

func RebootRescueBaremetal(cmd *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/dedicated/server/%s/boot?bootType=rescue", url.PathEscape(args[0]))

	var boots []int
	if err := httpLib.Client.Get(endpoint, &boots); err != nil {
		display.ExitError("failed to fetch boot options: %s", err)
		return
	}

	if len(boots) == 0 {
		display.ExitError("no boot found for rescue mode")
		return
	}

	// Update server with boot ID
	endpoint = fmt.Sprintf("/dedicated/server/%s", url.PathEscape(args[0]))
	if err := httpLib.Client.Put(endpoint, map[string]any{
		"bootId": boots[0],
	}, nil); err != nil {
		display.ExitError("failed to set boot ID %d for server: %s", boots[0], err)
		return
	}

	// Reboot server
	endpoint += "/reboot"

	var task map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &task); err != nil {
		display.ExitError("failed to reboot server: %s", err)
		return
	}

	if !flags.WaitForTask {
		fmt.Println("\n⚡️ Reboot in rescue mode is started ...")
		return
	}

	if err := waitForDedicatedServerTask(args[0], task["taskId"]); err != nil {
		display.ExitError("failed to wait for server to be rebooted: %s", err)
		return
	}

	fmt.Println("\n⚡️ Reboot done, fetching new authentication secrets...")

	// Fetch new secrets
	GetBaremetalAuthenticationSecrets(cmd, args)
}

func waitForDedicatedServerTask(serviceName string, taskID any) error {
	endpoint := fmt.Sprintf("/dedicated/server/%s/task/%s", url.PathEscape(serviceName), taskID)

	for retry := 0; retry < 100; retry++ {
		var task map[string]any

		if err := httpLib.Client.Get(endpoint, &task); err != nil {
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

func BaremetalGetIPMIAccess(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/dedicated/server/%s/features/ipmi/access", url.PathEscape(args[0]))

	parameters := map[string]any{
		"type": BaremetalIpmiAccessType,
		"ttl":  BaremetalIpmiTTL,
	}
	if BaremetalIpmiIP != "" {
		parameters["ipToAllow"] = BaremetalIpmiIP
	}
	if BaremetalIpmiSshKey != "" {
		parameters["sshKey"] = BaremetalIpmiSshKey
	}

	var task map[string]any
	if err := httpLib.Client.Post(path, parameters, &task); err != nil {
		display.ExitError("failed to request IMPI access: %s", err)
		return
	}

	if err := waitForDedicatedServerTask(args[0], task["taskId"]); err != nil {
		display.ExitError("failed waiting for task: %s", err)
		return
	}

	path += "?type=" + url.QueryEscape(BaremetalIpmiAccessType)

	var accessDetails map[string]any
	if err := httpLib.Client.Get(path, &accessDetails); err != nil {
		display.ExitError("failed to fetch IPMI access information: %s", err)
		return
	}

	output := fmt.Sprintf("\n⚡️ IPMI access: %s", accessDetails["value"])
	if expiration, ok := accessDetails["expiration"]; ok {
		output += fmt.Sprintf(" (expires at %s)", expiration)
	}

	fmt.Println(output)
}

func ListBaremetalInterventions(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/dedicated/server/%s/intervention", args[0])

	interventions, err := httpLib.FetchExpandedArray(path, "")
	if err != nil {
		display.ExitError("failed to fetch past interventions: %s", err)
		return
	}

	for _, inter := range interventions {
		inter["status"] = "done"
	}

	path = fmt.Sprintf("/dedicated/server/%s/plannedIntervention", args[0])
	plannedInterventions, err := httpLib.FetchExpandedArray(path, "")
	if err != nil {
		display.ExitError("failed to fetch planned interventions: %s", err)
		return
	}

	for _, inter := range plannedInterventions {
		inter["date"] = inter["wantedStartDate"]
	}

	plannedInterventions = append(plannedInterventions, interventions...)

	plannedInterventions, err = filtersLib.FilterLines(plannedInterventions, flags.GenericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
		return
	}

	display.RenderTable(plannedInterventions, []string{"type", "date", "status"}, &flags.OutputFormatConfig)
}

func ListBaremetalBoots(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/dedicated/server/%s/boot", url.PathEscape(args[0]))

	boots, err := httpLib.FetchExpandedArray(path, "")
	if err != nil {
		display.ExitError("error fetching boot options for server %q: %s", args[0], err)
		return
	}

	for _, boot := range boots {
		path = fmt.Sprintf("/dedicated/server/%s/boot/%s/option", url.PathEscape(args[0]), boot["bootId"])

		options, err := httpLib.FetchExpandedArray(path, "")
		if err != nil {
			display.ExitError("error fetching options of boot %d for server %s: %s", boot["bootId"], args[0], err)
			return
		}

		boot["options"] = options
	}

	boots, err = filtersLib.FilterLines(boots, flags.GenericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
		return
	}

	display.RenderTable(boots, []string{"bootId", "bootType", "description", "kernel"}, &flags.OutputFormatConfig)
}

func SetBaremetalBootId(_ *cobra.Command, args []string) {
	bootID, err := strconv.Atoi(args[1])
	if err != nil {
		display.ExitError("invalid boot ID given, expected a number")
		return
	}

	url := fmt.Sprintf("/dedicated/server/%s", url.PathEscape(args[0]))
	if err := httpLib.Client.Put(url, map[string]any{
		"bootId": bootID,
	}, nil); err != nil {
		display.ExitError("error setting boot ID: %s", err)
		return
	}

	fmt.Printf("\n✅ Boot ID %d correctly configured\n", bootID)
}

func SetBaremetalBootScript(_ *cobra.Command, args []string) {
	var (
		script []byte
		err    error
	)

	if InstallViaEditor {
		script, err = editor.EditValueWithEditor(nil)
		if err != nil {
			display.ExitError("failed to edit installation parameters using editor: %s", err)
			return
		}
	} else {
		fd, err := os.Open(InstallationFile)
		if err != nil {
			display.ExitError("failed to open given file: %s", err)
			return
		}
		defer fd.Close()

		script, err = io.ReadAll(fd)
		if err != nil {
			display.ExitError("failed to read installation file: %s", err)
			return
		}
	}

	url := fmt.Sprintf("/dedicated/server/%s", url.PathEscape(args[0]))
	if err := httpLib.Client.Put(url, map[string]any{
		"bootScript": string(script),
	}, nil); err != nil {
		display.ExitError("error setting boot script: %s", err)
		return
	}

	fmt.Println("\n✅ Boot script correctly configured")
}

func ListBaremetalVNIs(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/dedicated/server/%s/virtualNetworkInterface", args[0])
	common.ManageListRequest(url, "", []string{"uuid", "name", "mode", "vrack", "enabled"}, flags.GenericFilters)
}

func CreateBaremetalOLAAggregation(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/dedicated/server/%s/ola/aggregation", url.PathEscape(args[0]))
	if err := httpLib.Client.Post(url, map[string]any{
		"name":                     BaremetalOLAName,
		"virtualNetworkInterfaces": BaremetalOLAInterfaces,
	}, nil); err != nil {
		display.ExitError("failed to create OLA aggregation: %s", err)
	}
}

func ResetBaremetalOLAAggregation(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/dedicated/server/%s/ola/reset", url.PathEscape(args[0]))

	for _, itf := range BaremetalOLAInterfaces {
		if err := httpLib.Client.Post(url, map[string]string{
			"virtualNetworkInterface": itf,
		}, nil); err != nil {
			display.ExitError("failed to reset interface %s: %s", itf, err)
			return
		}
		fmt.Printf("✅ Interface %s reset to default configuration ...\n", itf)
	}
}

func ReinstallBaremetal(cmd *cobra.Command, args []string) {
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
		OS:             OperatingSystem,
		Customizations: Customizations,
	})
	if err != nil {
		display.ExitError("failed to prepare arguments from command line: %s", err)
		return
	}
	var cliParameters map[string]any
	if err := json.Unmarshal(jsonCliParameters, &cliParameters); err != nil {
		display.ExitError("failed to parse arguments from command line: %s", err)
		return
	}

	var parameters map[string]any

	if utils.IsInputFromPipe() { // Install data given through a pipe
		var stdin []byte
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			stdin = append(stdin, scanner.Bytes()...)
		}
		if err := scanner.Err(); err != nil {
			display.ExitError(err.Error())
			return
		}

		if err := json.Unmarshal(stdin, &parameters); err != nil {
			display.ExitError("failed to parse given installation data: %s", err)
			return
		}
	} else if InstallViaEditor {
		log.Print("Flag --editor used, all other flags will override the example values")

		examples, err := openapi.GetOperationRequestExamples(BaremetalOpenapiSchema, "/dedicated/server/{serviceName}/reinstall", "post", cliParameters)
		if err != nil {
			display.ExitError("failed to fetch API call examples: %s", err)
			return
		}

		_, choice, err := display.RunGenericChoicePicker("Please select an installation example", examples)
		if err != nil {
			display.ExitError(err.Error())
			return
		}

		if choice == "" {
			display.ExitWarning("No installation example selected, exiting...")
			return
		}

		newValue, err := editor.EditValueWithEditor([]byte(choice))
		if err != nil {
			display.ExitError("failed to edit installation parameters using editor: %s", err)
			return
		}

		if err := json.Unmarshal(newValue, &parameters); err != nil {
			display.ExitError("failed to parse given installation parameters: %s", err)
			return
		}
	} else if InstallationFile != "" { // Install data given in a file
		log.Print("Flag --installation-file used, all other flags will override the file values")

		fd, err := os.Open(InstallationFile)
		if err != nil {
			display.ExitError("failed to open given file: %s", err)
			return
		}
		defer fd.Close()

		content, err := io.ReadAll(fd)
		if err != nil {
			display.ExitError("failed to read installation file: %s", err)
			return
		}

		if err := json.Unmarshal(content, &parameters); err != nil {
			display.ExitError("failed to parse given installation file: %s", err)
			return
		}
	}

	// Only merge CLI parameters with other ones if not in --editor mode.
	// In this case, the CLI parameters have already been merged with the
	// request examples coming from API schemas.
	if !InstallViaEditor {
		parameters = utils.MergeMaps(cliParameters, parameters)
	}

	// Check if at least an OS was provided as it is mandatory
	if os, ok := parameters["operatingSystem"]; !ok || os == "" {
		display.ExitError("operating system parameter is mandatory to trigger a reinstallation")
		return
	}

	out, err := json.MarshalIndent(parameters, "", " ")
	if err != nil {
		display.ExitError("installation parameters cannot be marshalled: %s", err)
		return
	}

	log.Println("Installation parameters: \n" + string(out))

	var task map[string]any
	url := fmt.Sprintf("/dedicated/server/%s/reinstall", url.PathEscape(args[0]))
	if err := httpLib.Client.Post(url, parameters, &task); err != nil {
		display.ExitError("error reinstalling server %s: %s\n", args[0], err)
		return
	}

	fmt.Println("\n⚡️ Reinstallation started ...")

	if !flags.WaitForTask {
		return
	}

	if err := waitForDedicatedServerTask(args[0], task["taskId"]); err != nil {
		display.ExitError("failed to wait for server to be reinstalled: %s", err)
		return
	}

	fmt.Println("\n⚡️ Reinstall done, fetching new authentication secrets...")

	// Fetch new secrets
	GetBaremetalAuthenticationSecrets(cmd, args)
}

func GetBaremetalRelatedIPs(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/ip?routedTo.serviceName=%s", url.QueryEscape(args[0]))

	var ips []any
	if err := httpLib.Client.Get(path, &ips); err != nil {
		display.ExitError("failed to fetch IPs related to baremetal %s: %s", args[0], err)
		return
	}

	ipsExpanded, err := httpLib.FetchObjectsParallel[map[string]any]("/ip/%s", ips, flags.IgnoreErrors)
	if err != nil {
		display.ExitError("failed to fetch objects for each IP: %s", err)
		return
	}

	ipsExpanded, err = filtersLib.FilterLines(ipsExpanded, flags.GenericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
		return
	}

	display.RenderTable(ipsExpanded, []string{"ip", "type", "description", "campus"}, &flags.OutputFormatConfig)
}

func GetBaremetalAuthenticationSecrets(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/dedicated/server/%s/authenticationSecret", url.PathEscape(args[0]))

	var allSecrets []map[string]any
	if err := httpLib.Client.Post(path, nil, &allSecrets); err != nil {
		display.ExitError("failed to fetch secrets IDs: %s", err)
		return
	}

	for _, secret := range allSecrets {
		if secretID, ok := secret["password"]; ok {
			var secretValue map[string]any
			if err := httpLib.Client.Post("/secret/retrieve", map[string]any{
				"id": secretID,
			}, &secretValue); err != nil {
				display.ExitError("failed to retrieve secret value: %s", err)
				return
			}
			maps.Copy(secret, secretValue)
		}
	}

	allSecrets, err := filtersLib.FilterLines(allSecrets, flags.GenericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
		return
	}

	display.RenderTable(allSecrets, []string{"type", "url", "user", "secret", "expiration"}, &flags.OutputFormatConfig)
}

func GetBaremetalCompatibleOses(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/dedicated/server/%s/install/compatibleTemplates", url.PathEscape(args[0]))

	var oses map[string]any
	if err := httpLib.Client.Get(path, &oses); err != nil {
		display.ExitError("failed to fetch compatible OSes: %s", err)
		return
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

	formattedValues, err := filtersLib.FilterLines(formattedValues, flags.GenericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
		return
	}

	display.RenderTable(formattedValues, []string{"source", "name"}, &flags.OutputFormatConfig)
}
