// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"maps"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/editor"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
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

	// Installation flags
	OperatingSystem string
	Customizations  baremetalCustomizations

	// Virtual Network Interfaces Aggregation flags
	BaremetalOLAInterfaces []string
	BaremetalOLAName       string

	// IPMI flags
	BaremetalIpmiTTL        int
	BaremetalIpmiAccessType string
	BaremetalIpmiIP         string
	BaremetalIpmiSshKey     string

	// The two booleans are pointers, and the two plain bools below them are what
	// cobra writes into.
	//
	// `omitempty` on a bool drops false, so `--monitoring=false` and
	// `--no-intervention=false` sent nothing at all — and EditResource reads the
	// object, merges the (empty) command-line map into it and PUTs the result, so
	// the current value came straight back and the command answered "✅ Resource
	// updated successfully". Turning a flag off was not a no-op, it was a no-op
	// that claimed to have worked. On a pointer, omitempty drops only nil, so an
	// explicit false travels.
	//
	// Monitoring had the same defect and looked fine: the only caller that names
	// it, the doctor's remedy for monitoring being off, happens to pass true.
	EditBaremetalParams struct {
		BootId            int    `json:"bootId,omitempty"`
		BootScript        string `json:"bootScript,omitempty"`
		EfiBootloaderPath string `json:"efiBootloaderPath,omitempty"`
		Monitoring        *bool  `json:"monitoring,omitempty"`
		NoIntervention    *bool  `json:"noIntervention,omitempty"`
		RescueMail        string `json:"rescueMail,omitempty"`
		RescueSshKey      string `json:"rescueSshKey,omitempty"`
		RootDevice        string `json:"rootDevice,omitempty"`
		State             string `json:"state,omitempty"`
	}

	// EditBaremetalMonitoring and EditBaremetalNoIntervention hold what cobra
	// parsed; EditBaremetal decides whether it was asked for.
	EditBaremetalMonitoring     bool
	EditBaremetalNoIntervention bool
)

func ListBaremetal(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/dedicated/server", "", baremetalColumnsToDisplay, flags.GenericFilters)
}

func ListBaremetalTasks(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/v1/dedicated/server/%s/task", args[0])
	common.ManageListRequest(url, "", []string{"taskId", "function", "status", "startDate", "doneDate"}, flags.GenericFilters)
}

func GetBaremetal(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v1/dedicated/server/%s", url.PathEscape(args[0]))

	// Fetch dedicated server
	var object map[string]any
	if err := httpLib.Client.Get(path, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching %s: %s", path, err)
		return
	}

	// Fetch running tasks
	path = fmt.Sprintf("/v1/dedicated/server/%s/task", url.PathEscape(args[0]))
	tasks, err := httpLib.FetchExpandedArray(path, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching tasks for %s: %s", args[0], err)
		return
	}
	object["tasks"] = tasks

	// Fetch network information
	path = fmt.Sprintf("/v1/dedicated/server/%s/specifications/network", url.PathEscape(args[0]))
	var network map[string]any
	if err := httpLib.Client.Get(path, &network); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching network specifications for %s: %s", args[0], err)
		return
	}
	object["network"] = network

	path = fmt.Sprintf("/v1/dedicated/server/%s/serviceInfos", url.PathEscape(args[0]))
	var serviceInfo map[string]any
	if err := httpLib.Client.Get(path, &serviceInfo); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching billing information for %s: %s", args[0], err)
		return
	}
	object["serviceInfo"] = serviceInfo

	display.OutputObject(object, args[0], baremetalTemplate, &flags.OutputFormatConfig)
}

func EditBaremetal(cmd *cobra.Command, args []string) {
	// A boolean flag has no absent value of its own, so the question "was it
	// typed" is asked of cobra rather than of the value.
	//
	// Assigned both ways round, never only on the true branch: the wasm build
	// keeps one process across invocations, and a pointer left over from an
	// earlier command would send a value nobody asked for this time.
	EditBaremetalParams.Monitoring = nil
	EditBaremetalParams.NoIntervention = nil
	if cmd.Flags().Changed("monitoring") {
		EditBaremetalParams.Monitoring = &EditBaremetalMonitoring
	}
	if cmd.Flags().Changed("no-intervention") {
		EditBaremetalParams.NoIntervention = &EditBaremetalNoIntervention
	}

	if err := common.EditResource(
		cmd,
		"/dedicated/server/{serviceName}",
		fmt.Sprintf("/v1/dedicated/server/%s", url.PathEscape(args[0])),
		&EditBaremetalParams,
		assets.BaremetalOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func RebootBaremetal(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/v1/dedicated/server/%s/reboot", url.PathEscape(args[0]))

	if !common.ConfirmAction(common.Disruptive, args[0],
		fmt.Sprintf("Rebooting %s interrupts everything running on it.", args[0])) {
		display.OutputError(&flags.OutputFormatConfig, "reboot of %s cancelled", args[0])
		return
	}

	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: url}) {
		return
	}

	if err := httpLib.Client.Post(url, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error rebooting server %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Reboot launched…")
}

func RebootRescueBaremetal(cmd *cobra.Command, args []string) {
	if !common.ConfirmAction(common.Disruptive, args[0], fmt.Sprintf(
		"Rebooting %s into rescue mode interrupts everything running on it, and it stays in rescue until the boot is set back to disk.",
		args[0])) {
		display.OutputError(&flags.OutputFormatConfig, "reboot of %s into rescue mode cancelled", args[0])
		return
	}

	// Three calls, not one: the boot is read, written to the server, and only
	// then is the server rebooted. The write is what outlives the reboot.
	server := fmt.Sprintf("/v1/dedicated/server/%s", url.PathEscape(args[0]))
	if common.ReportDryRun(
		common.Call{Method: "GET", Endpoint: server + "/boot?bootType=rescue"},
		common.Call{Method: "PUT", Endpoint: server + "  (bootId of the rescue entry)"},
		common.Call{Method: "POST", Endpoint: server + "/reboot"},
	) {
		return
	}

	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/boot?bootType=rescue", url.PathEscape(args[0]))

	var boots []int
	if err := httpLib.Client.Get(endpoint, &boots); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch boot options: %s", err)
		return
	}

	if len(boots) == 0 {
		display.OutputError(&flags.OutputFormatConfig, "no boot found for rescue mode")
		return
	}

	// Update server with boot ID
	endpoint = fmt.Sprintf("/v1/dedicated/server/%s", url.PathEscape(args[0]))
	if err := httpLib.Client.Put(endpoint, map[string]any{
		"bootId": boots[0],
	}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to set boot ID %d for server: %s", boots[0], err)
		return
	}

	// Reboot server
	endpoint += "/reboot"

	var task map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to reboot server: %s", err)
		return
	}

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Reboot in rescue mode is started…")
		return
	}

	if err := waitForDedicatedServerTask(args[0], task["taskId"], nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for server to be rebooted: %s", err)
		return
	}

	log.Println("⚡️ Reboot done, fetching new authentication secrets…")

	// Fetch new secrets
	GetBaremetalAuthenticationSecrets(cmd, args)
}

// taskPollInterval and taskPollAttempts bound how long --wait follows a task.
// Past that the task is not cancelled: only the waiting stops, which is what
// the message has to say. They are variables rather than constants so that the
// test covering that message does not have to wait fifty minutes for it.
var (
	taskPollInterval = 30 * time.Second
	taskPollAttempts = 100
)

// describeTask names a task the way its owner would look it up: its
// identifier, and the operation it performs when the API says which.
//
// The identifier is formatted with %v and not %d: the API client decodes with
// UseNumber, so it arrives as a json.Number and %d rendered it as
// "%!d(json.Number=156839472)" — in the one message written to explain a
// failure.
func describeTask(taskID any, task map[string]any) string {
	if function, ok := task["function"].(string); ok && function != "" {
		return fmt.Sprintf("%v (%s)", taskID, function)
	}

	return fmt.Sprintf("%v", taskID)
}

// taskFailureReason returns what the API said about a failure, or "" when it
// said nothing. Both fields are free text filled in by the platform.
func taskFailureReason(task map[string]any) string {
	for _, field := range []string{"comment", "note"} {
		if reason, ok := task[field].(string); ok && strings.TrimSpace(reason) != "" {
			return strings.TrimSpace(reason)
		}
	}

	return ""
}

// waitForDedicatedServerTask follows a task until it ends.
//
// note, when given, is asked for one extra line on every poll. The task API
// says whether the work is finished and nothing else, so a reinstall that
// takes half an hour printed the same sentence sixty times; the caller that
// has a second source of progress passes it here. It is decoration: a note
// that cannot be produced is an empty string, never an error, because failing
// to describe an installation must not fail the wait for it.
func waitForDedicatedServerTask(serviceName string, taskID any, note func() string) error {
	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/task/%s", url.PathEscape(serviceName), taskID)
	followUp := fmt.Sprintf("follow it with: ovhcloud baremetal list-tasks %s", serviceName)

	var (
		lastStatus      any
		lastDescription = fmt.Sprintf("%v", taskID)
	)

	for retry := 0; retry < taskPollAttempts; retry++ {
		var task map[string]any

		if err := httpLib.Client.Get(endpoint, &task); err != nil {
			return fmt.Errorf("failed to fetch task: %w", err)
		}

		lastStatus = task["status"]
		lastDescription = describeTask(taskID, task)

		switch task["status"] {
		case "done":
			return nil

		case "todo", "init", "doing":
			if note != nil {
				if progress := note(); progress != "" {
					log.Printf("%s — %s", progress, describeTask(taskID, task))
					time.Sleep(taskPollInterval)
					continue
				}
			}

			log.Printf("Still waiting for task %s to complete (status=%v)…",
				describeTask(taskID, task), task["status"])
			time.Sleep(taskPollInterval)

		// These are the terminal failures of the API enum. Calling them an
		// invalid state told the operator the CLI had not understood, when
		// what had happened is that their operation failed.
		case "cancelled", "customerError", "ovhError":
			if reason := taskFailureReason(task); reason != "" {
				return fmt.Errorf("task %s ended with status %v: %s",
					describeTask(taskID, task), task["status"], reason)
			}

			return fmt.Errorf("task %s ended with status %v, and the API gave no reason; %s",
				describeTask(taskID, task), task["status"], followUp)

		// A status this CLI does not know about is not a failure of the task:
		// the enum can grow. Say what was received rather than guess.
		default:
			return fmt.Errorf("task %s reported the unexpected status %v; %s",
				describeTask(taskID, task), task["status"], followUp)
		}
	}

	// Only the waiting stopped; the task was not cancelled. The status is
	// quoted as what was last *seen*, not as what is true now: the loop sleeps
	// one interval before giving up, so it is up to taskPollInterval old.
	return fmt.Errorf("stopped waiting for task %s after %s; it had not finished (last status seen: %v), %s",
		lastDescription, time.Duration(taskPollAttempts)*taskPollInterval, lastStatus, followUp)
}

func BaremetalGetIPMIAccess(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v1/dedicated/server/%s/features/ipmi/access", url.PathEscape(args[0]))

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
		display.OutputError(&flags.OutputFormatConfig, "failed to request IMPI access: %s", err)
		return
	}

	if err := waitForDedicatedServerTask(args[0], task["taskId"], nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed waiting for task: %s", err)
		return
	}

	path += "?type=" + url.QueryEscape(BaremetalIpmiAccessType)

	var accessDetails map[string]any
	if err := httpLib.Client.Get(path, &accessDetails); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch IPMI access information: %s", err)
		return
	}

	output := fmt.Sprintf("⚡️ IPMI access: %s", accessDetails["value"])
	if expiration, ok := accessDetails["expiration"]; ok {
		output += fmt.Sprintf(" (expires at %s)", expiration)
	}

	display.OutputInfo(&flags.OutputFormatConfig, accessDetails, "%s", output)
}

func BaremetalResetIPMISessions(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v1/dedicated/server/%s/features/ipmi/resetSessions", url.PathEscape(args[0]))

	if err := httpLib.Client.Post(path, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to reset IPMI sessions for %s: %s", args[0], err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ IPMI sessions reset for %s", args[0])
}

func ListBaremetalInterventions(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v1/dedicated/server/%s/intervention", args[0])

	interventions, err := httpLib.FetchExpandedArray(path, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch past interventions: %s", err)
		return
	}

	for _, inter := range interventions {
		inter["status"] = "done"
	}

	path = fmt.Sprintf("/v1/dedicated/server/%s/plannedIntervention", args[0])
	plannedInterventions, err := httpLib.FetchExpandedArray(path, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch planned interventions: %s", err)
		return
	}

	for _, inter := range plannedInterventions {
		inter["date"] = inter["wantedStartDate"]
	}

	plannedInterventions = append(plannedInterventions, interventions...)

	plannedInterventions, err = filtersLib.FilterLines(plannedInterventions, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(plannedInterventions, []string{"type", "date", "status"}, &flags.OutputFormatConfig)
}

func ListBaremetalBoots(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v1/dedicated/server/%s/boot", url.PathEscape(args[0]))

	boots, err := httpLib.FetchExpandedArray(path, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching boot options for server %q: %s", args[0], err)
		return
	}

	for _, boot := range boots {
		path = fmt.Sprintf("/v1/dedicated/server/%s/boot/%s/option", url.PathEscape(args[0]), boot["bootId"])

		options, err := httpLib.FetchExpandedArray(path, "")
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "error fetching options of boot %d for server %s: %s", boot["bootId"], args[0], err)
			return
		}

		boot["options"] = options
	}

	boots, err = filtersLib.FilterLines(boots, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(boots, []string{"bootId", "bootType", "description", "kernel"}, &flags.OutputFormatConfig)
}

func SetBaremetalBootId(_ *cobra.Command, args []string) {
	bootID, err := strconv.Atoi(args[1])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "invalid boot ID given, expected a number")
		return
	}

	url := fmt.Sprintf("/v1/dedicated/server/%s", url.PathEscape(args[0]))
	if err := httpLib.Client.Put(url, map[string]any{
		"bootId": bootID,
	}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error setting boot ID: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Boot ID %d correctly configured", bootID)
}

func SetBaremetalBootScript(_ *cobra.Command, args []string) {
	var (
		script []byte
		err    error
	)

	if EditBaremetalParams.BootScript != "" {
		script = []byte(EditBaremetalParams.BootScript)
	} else if flags.ParametersViaEditor {
		script, err = editor.EditValueWithEditor(nil)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to edit installation parameters using editor: %s", err)
			return
		}
	} else {
		fd, err := os.Open(flags.ParametersFile)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to open given file: %s", err)
			return
		}
		defer fd.Close()

		script, err = io.ReadAll(fd)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to read installation file: %s", err)
			return
		}
	}

	url := fmt.Sprintf("/v1/dedicated/server/%s", url.PathEscape(args[0]))
	if err := httpLib.Client.Put(url, map[string]any{
		"bootScript": string(script),
	}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error setting boot script: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Boot script correctly configured")
}

func ListBaremetalVNIs(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/v1/dedicated/server/%s/virtualNetworkInterface", args[0])
	common.ManageListRequest(url, "", []string{"uuid", "name", "mode", "vrack", "enabled"}, flags.GenericFilters)
}

func CreateBaremetalOLAAggregation(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/v1/dedicated/server/%s/ola/aggregation", url.PathEscape(args[0]))
	if err := httpLib.Client.Post(url, map[string]any{
		"name":                     BaremetalOLAName,
		"virtualNetworkInterfaces": BaremetalOLAInterfaces,
	}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create OLA aggregation: %s", err)
	}
}

func ResetBaremetalOLAAggregation(_ *cobra.Command, args []string) {
	url := fmt.Sprintf("/v1/dedicated/server/%s/ola/reset", url.PathEscape(args[0]))

	// Resetting an aggregation takes the interfaces down and back up: on a
	// server reached over that link, the operator is cutting the branch.
	if !common.ConfirmAction(common.Disruptive, args[0], fmt.Sprintf(
		"Resetting %d interface(s) of %s to their default configuration interrupts the network of the server.",
		len(BaremetalOLAInterfaces), args[0])) {
		display.OutputError(&flags.OutputFormatConfig, "interface reset on %s cancelled", args[0])
		return
	}

	// One call per interface, so the preview lists them rather than implying a
	// single request.
	calls := make([]common.Call, 0, len(BaremetalOLAInterfaces))
	for _, itf := range BaremetalOLAInterfaces {
		calls = append(calls, common.Call{Method: "POST", Endpoint: url + "  (virtualNetworkInterface " + itf + ")"})
	}
	if common.ReportDryRun(calls...) {
		return
	}

	for _, itf := range BaremetalOLAInterfaces {
		if err := httpLib.Client.Post(url, map[string]string{
			"virtualNetworkInterface": itf,
		}, nil); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to reset interface %s: %s", itf, err)
			return
		}
		log.Printf("✅ Interface %s reset to default configuration…", itf)
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ All interfaces reset to default configuration")
}

func ReinstallBaremetal(cmd *cobra.Command, args []string) {
	// No server ID given, print usage and exit
	if len(args) == 0 {
		cmd.Help()
		display.OutputError(&flags.OutputFormatConfig, "reinstall command requires a server ID as the first argument")
		return
	}

	// Reinstalling wipes every disk of the server, so this one asks for the
	// server's name rather than a yes.
	warning := fmt.Sprintf("Reinstalling %s wipes every disk of the server. This cannot be undone.", args[0])
	if OperatingSystem != "" {
		warning += fmt.Sprintf("\n   Operating system to install: %s", OperatingSystem)
	}

	if !common.ConfirmAction(common.Destructive, args[0], warning) {
		display.OutputError(&flags.OutputFormatConfig, "reinstallation of %s cancelled", args[0])
		return
	}

	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/reinstall", url.PathEscape(args[0]))
	task, err := common.CreateResource(
		cmd,
		"/dedicated/server/{serviceName}/reinstall",
		endpoint,
		BaremetalInstallationExample,
		struct {
			OS             string                  `json:"operatingSystem,omitempty"`
			Customizations baremetalCustomizations `json:"customizations,omitzero"`
		}{
			OS:             OperatingSystem,
			Customizations: Customizations,
		},
		assets.BaremetalOpenapiSchema,
		[]string{"operatingSystem"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error reinstalling server: %s", err)
		return
	}

	// Nothing was sent in dry-run mode, so there is no task to follow.
	if flags.DryRun {
		return
	}

	log.Println("⚡️ Reinstallation started…")

	// The wait times itself from here. install/status reports an elapsedTime,
	// but it answered -1935 seconds on the reinstall this was measured
	// against, counting up correctly from an origin about half an hour ahead:
	// its differences are true and its absolute value is not.
	startedAt := time.Now()

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Reinstallation is started…")
		return
	}

	if err := waitForDedicatedServerTask(args[0], task["taskId"], func() string {
		return installProgressNote(args[0], startedAt)
	}); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for server to be reinstalled: %s", err)
		return
	}

	log.Println("⚡️ Reinstall done, fetching new authentication secrets…")

	// Fetch new secrets
	GetBaremetalAuthenticationSecrets(cmd, args)
}

func GetBaremetalRelatedIPs(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v1/ip?routedTo.serviceName=%s", url.QueryEscape(args[0]))

	var ips []any
	if err := httpLib.Client.Get(path, &ips); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch IPs related to baremetal %s: %s", args[0], err)
		return
	}

	ipsExpanded, err := httpLib.FetchObjectsParallel[map[string]any]("/ip/%s", ips, flags.IgnoreErrors)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch objects for each IP: %s", err)
		return
	}

	ipsExpanded, err = filtersLib.FilterLines(ipsExpanded, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(ipsExpanded, []string{"ip", "type", "description", "campus"}, &flags.OutputFormatConfig)
}

func GetBaremetalAuthenticationSecrets(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v1/dedicated/server/%s/authenticationSecret", url.PathEscape(args[0]))

	var allSecrets []map[string]any
	if err := httpLib.Client.Post(path, nil, &allSecrets); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch secrets IDs: %s", err)
		return
	}

	for _, secret := range allSecrets {
		if secretID, ok := secret["password"]; ok {
			var secretValue map[string]any
			if err := httpLib.Client.Post("/v1/secret/retrieve", map[string]any{
				"id": secretID,
			}, &secretValue); err != nil {
				display.OutputError(&flags.OutputFormatConfig, "failed to retrieve secret value: %s", err)
				return
			}
			maps.Copy(secret, secretValue)
		}
	}

	allSecrets, err := filtersLib.FilterLines(allSecrets, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(allSecrets, []string{"type", "url", "user", "secret", "expiration"}, &flags.OutputFormatConfig)
}

func GetBaremetalCompatibleOses(_ *cobra.Command, args []string) {
	path := fmt.Sprintf("/v1/dedicated/server/%s/install/compatibleTemplates", url.PathEscape(args[0]))

	var oses map[string]any
	if err := httpLib.Client.Get(path, &oses); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch compatible OSes: %s", err)
		return
	}

	var formattedValues []map[string]any
	for _, os := range oses["ovh"].([]any) {
		formattedValues = append(formattedValues, map[string]any{
			"source": "ovh",
			"name":   os,
		})
	}

	if personalOSes, ok := oses["personal"]; ok {
		for _, os := range personalOSes.([]any) {
			formattedValues = append(formattedValues, map[string]any{
				"source": "personal",
				"name":   os,
			})
		}
	}

	formattedValues, err := filtersLib.FilterLines(formattedValues, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(formattedValues, []string{"source", "name"}, &flags.OutputFormatConfig)
}
