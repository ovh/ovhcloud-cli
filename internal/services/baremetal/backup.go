// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

// A dedicated server comes with a Backup FTP space, and the CLI could neither
// see it nor say who was allowed to reach it. Nine routes, none of them wired:
// the space itself, the access list that guards it, the two passwords, and the
// cloud backup beside it.

//go:embed templates/backup_ftp.tmpl
var backupFtpTemplate string

//go:embed templates/backup_password.tmpl
var backupPasswordTemplate string

// backupPollInterval and backupPollAttempts bound how long --wait follows a
// creation or a deletion. Variables rather than constants so a test can
// exercise the timeout in milliseconds instead of in ten minutes.
var (
	backupPollInterval = 5 * time.Second
	backupPollAttempts = 120
)

var (
	// BackupAclFtp, BackupAclNfs and BackupAclCifs are the protocols an access
	// rule opens.
	BackupAclFtp  bool
	BackupAclNfs  bool
	BackupAclCifs bool

	// BackupWait follows the task until the space exists, or is gone.
	BackupWait bool

	// RevealBackupPassword prints the new passwords instead of their
	// fingerprints.
	RevealBackupPassword bool

	// BackupCloudProjectId and BackupCloudProjectDescription place the cloud
	// backup containers in a public cloud project.
	BackupCloudProjectId          string
	BackupCloudProjectDescription string
)

// ShowBackupFtp shows the Backup FTP space of a server.
func ShowBackupFtp(_ *cobra.Command, args []string) {
	server := args[0]

	space, err := backupFtpSpace(server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputObject(space, server, backupFtpTemplate, &flags.OutputFormatConfig)
}

// CreateBackupFtp creates the Backup FTP space included with the server.
//
// The request carries no body: there is nothing to choose. The paid capacities
// are a separate order, which `backup orderable` reads.
func CreateBackupFtp(_ *cobra.Command, args []string) {
	server := args[0]

	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/features/backupFTP", url.PathEscape(server))
	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	var task map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to create the Backup FTP space of %s: %s", server, err)
		return
	}

	if !BackupWait {
		display.OutputInfo(&flags.OutputFormatConfig, task,
			"⚡️ The Backup FTP space of %s is being created. Follow it with: ovhcloud baremetal list-tasks %s",
			server, server)
		return
	}

	if err := waitForBackupFtp(server, true); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	space, err := backupFtpSpace(server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputObject(space, server, backupFtpTemplate, &flags.OutputFormatConfig)
}

// DeleteBackupFtp terminates the Backup FTP space of a server.
func DeleteBackupFtp(_ *cobra.Command, args []string) {
	server := args[0]

	// Destructive, in the API's own words: "Terminate your Backup FTP service,
	// ALL DATA WILL BE PERMANENTLY DELETED". This is the one command of the lot
	// that loses something no other command can bring back, so it takes the
	// strongest guard the CLI has — the server's name, typed.
	if !common.ConfirmAction(common.Destructive, server, fmt.Sprintf(
		"This terminates the Backup FTP space of %s. Everything stored on it is permanently deleted.",
		server)) {
		display.OutputError(&flags.OutputFormatConfig, "deletion of the Backup FTP space of %s cancelled", server)
		return
	}

	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/features/backupFTP", url.PathEscape(server))
	if common.ReportDryRun(common.Call{Method: "DELETE", Endpoint: endpoint}) {
		return
	}

	var task map[string]any
	if err := httpLib.Client.Delete(endpoint, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to delete the Backup FTP space of %s: %s", server, err)
		return
	}

	if !BackupWait {
		display.OutputInfo(&flags.OutputFormatConfig, task,
			"⚡️ The Backup FTP space of %s is being deleted. Follow it with: ovhcloud baremetal list-tasks %s",
			server, server)
		return
	}

	if err := waitForBackupFtp(server, false); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"serviceName": server},
		"✅ The Backup FTP space of %s is gone.", server)
}

// ChangeBackupFtpPassword changes the Backup FTP password.
//
// The response is a task, not the password: the API mails it. Printing the task
// as though it carried the credential is how somebody ends up pasting a task
// identifier into an FTP client.
func ChangeBackupFtpPassword(_ *cobra.Command, args []string) {
	server := args[0]

	if !common.ConfirmAction(common.Disruptive, server, fmt.Sprintf(
		"Changing the Backup FTP password of %s stops every job still using the old one.", server)) {
		display.OutputError(&flags.OutputFormatConfig, "password change on %s cancelled", server)
		return
	}

	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/features/backupFTP/password", url.PathEscape(server))
	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	var task map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to change the Backup FTP password of %s: %s", server, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, task,
		"⚡️ The Backup FTP password of %s is being changed. The new one is emailed to the administrative contact.",
		server)
}

// ListBackupFtpAcl lists who may reach the Backup FTP space, and how.
func ListBackupFtpAcl(_ *cobra.Command, args []string) {
	server := args[0]

	blocks, err := backupFtpAclBlocks(server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if len(blocks) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"serviceName": server, "acl": []any{}},
			"No IP block is allowed to reach the Backup FTP space of %s.", server)
		return
	}

	base := fmt.Sprintf("/v1/dedicated/server/%s/features/backupFTP/access", url.PathEscape(server))
	rows := make([]map[string]any, 0, len(blocks))
	for _, block := range blocks {
		var acl map[string]any
		if err := httpLib.Client.Get(fmt.Sprintf("%s/%s", base, url.PathEscape(block)), &acl); err != nil {
			display.OutputError(&flags.OutputFormatConfig,
				"failed to read the access rule of %s: %s", block, err)
			return
		}
		acl["protocols"] = aclProtocols(acl)
		rows = append(rows, acl)
	}

	sort.Slice(rows, func(i, j int) bool {
		return fmt.Sprint(rows[i]["ipBlock"]) < fmt.Sprint(rows[j]["ipBlock"])
	})

	filtered, err := filtersLib.FilterLines(rows, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(filtered,
		[]string{"ipBlock", "protocols", "isApplied", "lastUpdate"},
		&flags.OutputFormatConfig)
}

// GetBackupFtpAcl shows one access rule.
func GetBackupFtpAcl(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(
		fmt.Sprintf("/v1/dedicated/server/%s/features/backupFTP/access", url.PathEscape(args[0])),
		args[1], "")
}

// AddBackupFtpAcl opens the Backup FTP space to an IP block.
func AddBackupFtpAcl(_ *cobra.Command, args []string) {
	server, block := args[0], args[1]

	if err := checkAclProtocols(); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	allowed, err := authorizableBlocks(server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	chosen, ok := pickBlock(allowed, block)
	if !ok {
		display.OutputError(&flags.OutputFormatConfig, "%s", unauthorizableBlock(server, block, allowed))
		return
	}

	if !common.ConfirmAction(common.Disruptive, server, fmt.Sprintf(
		"This opens the Backup FTP space of %s to %s over %s.", server, chosen, aclProtocolList())) {
		display.OutputError(&flags.OutputFormatConfig, "access rule on %s cancelled", server)
		return
	}

	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/features/backupFTP/access", url.PathEscape(server))
	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	body := map[string]any{
		"ipBlock": chosen,
		"ftp":     BackupAclFtp,
		"nfs":     BackupAclNfs,
		"cifs":    BackupAclCifs,
	}

	var task map[string]any
	if err := httpLib.Client.Post(endpoint, body, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to open the Backup FTP space to %s: %s", chosen, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, task,
		"⚡️ %s is being allowed on the Backup FTP space of %s over %s.",
		chosen, server, aclProtocolList())
}

// SetBackupFtpAcl changes the protocols an existing access rule opens.
func SetBackupFtpAcl(_ *cobra.Command, args []string) {
	server, block := args[0], args[1]

	if err := checkAclProtocols(); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if !common.ConfirmAction(common.Disruptive, server, fmt.Sprintf(
		"The Backup FTP space of %s becomes reachable from %s over %s, and over nothing else.",
		server, block, aclProtocolList())) {
		display.OutputError(&flags.OutputFormatConfig, "change on %s cancelled", block)
		return
	}

	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/features/backupFTP/access/%s",
		url.PathEscape(server), url.PathEscape(block))
	if common.ReportDryRun(common.Call{Method: "PUT", Endpoint: endpoint}) {
		return
	}

	body := map[string]any{"ftp": BackupAclFtp, "nfs": BackupAclNfs, "cifs": BackupAclCifs}
	if err := httpLib.Client.Put(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to change the access rule of %s: %s", block, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"serviceName": server, "ipBlock": block, "protocols": aclProtocolList()},
		"✅ %s now reaches the Backup FTP space of %s over %s.", block, server, aclProtocolList())
}

// DeleteBackupFtpAcl closes the Backup FTP space to an IP block.
func DeleteBackupFtpAcl(_ *cobra.Command, args []string) {
	server, block := args[0], args[1]

	if !common.ConfirmAction(common.Disruptive, server, fmt.Sprintf(
		"This stops %s reaching the Backup FTP space of %s. Backup jobs running from it will fail.",
		block, server)) {
		display.OutputError(&flags.OutputFormatConfig, "revocation of %s cancelled", block)
		return
	}

	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/features/backupFTP/access/%s",
		url.PathEscape(server), url.PathEscape(block))
	if common.ReportDryRun(common.Call{Method: "DELETE", Endpoint: endpoint}) {
		return
	}

	var task map[string]any
	if err := httpLib.Client.Delete(endpoint, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to revoke the access of %s: %s", block, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, task,
		"⚡️ %s is being revoked from the Backup FTP space of %s.", block, server)
}

// ListAuthorizableBlocks lists the IP blocks that may be put in an access rule.
//
// The list is per server and regional: measured on a real account, the European
// servers offer around thirty blocks and the Canadian ones five — their own.
func ListAuthorizableBlocks(_ *cobra.Command, args []string) {
	server := args[0]

	blocks, err := authorizableBlocks(server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if len(blocks) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"serviceName": server, "blocks": []any{}},
			"No IP block of this account can be allowed on the Backup FTP space of %s.", server)
		return
	}

	rows := make([]map[string]any, 0, len(blocks))
	for _, block := range blocks {
		rows = append(rows, map[string]any{"ipBlock": block})
	}

	filtered, err := filtersLib.FilterLines(rows, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(filtered, []string{"ipBlock"}, &flags.OutputFormatConfig)
}

// ShowBackupCloud shows the cloud backup containers of a server.
func ShowBackupCloud(_ *cobra.Command, args []string) {
	server := args[0]

	var backup map[string]any
	path := fmt.Sprintf("/v1/dedicated/server/%s/features/backupCloud", url.PathEscape(server))
	if err := httpLib.Client.Get(path, &backup); err != nil {
		if common.IsNotFound(err) {
			display.OutputError(&flags.OutputFormatConfig,
				"%s has no cloud backup: %s\n   Create one with: ovhcloud baremetal backup cloud create %s",
				server, err, server)
			return
		}

		display.OutputError(&flags.OutputFormatConfig,
			"failed to read the cloud backup of %s: %s", server, err)
		return
	}

	display.OutputObject(maskBackupCloud(backup), server, "", &flags.OutputFormatConfig)
}

// ShowBackupCloudOffer shows what the cloud backup of this server would hold.
//
// A server the offer does not cover answers HTTP 403 — measured on 34 of the 35
// servers of a real account, while the thirty-fifth answered its sizes. It is a
// business answer, not a permission problem, and it is reported as one.
func ShowBackupCloudOffer(_ *cobra.Command, args []string) {
	server := args[0]

	var offer map[string]any
	path := fmt.Sprintf("/v1/dedicated/server/%s/backupCloudOfferDetails", url.PathEscape(server))
	if err := httpLib.Client.Get(path, &offer); err != nil {
		// Matched on the message and not on the status alone. The 403 the
		// business answer carries says "Not available for this server" —
		// recorded on the 34 servers of the account that answered it — and an
		// API key simply lacking the right on this route answers 403 too. Read
		// by status, that key was told no offer covers the machine, and
		// -o json handed a script offered:false, which is a claim about the
		// catalogue rather than about the caller's rights.
		var apiErr *ovh.APIError
		if errors.As(err, &apiErr) && apiErr.Code == http.StatusForbidden {
			if strings.Contains(strings.ToLower(apiErr.Message), "not available for this server") {
				display.OutputInfo(&flags.OutputFormatConfig,
					map[string]any{"serviceName": server, "offered": false},
					"No cloud backup offer covers %s.", server)
				return
			}

			display.OutputError(&flags.OutputFormatConfig,
				"not allowed to read the cloud backup offer of %s: %s\n   This is a rights answer, not a catalogue one: check the API key's grant on dedicatedServer:apiovh:backupCloudOfferDetails/get.",
				server, err)
			return
		}

		display.OutputError(&flags.OutputFormatConfig,
			"failed to read the cloud backup offer of %s: %s", server, err)
		return
	}

	display.OutputObject(offer, server, "", &flags.OutputFormatConfig)
}

// CreateBackupCloud creates the cloud backup containers of a server.
func CreateBackupCloud(_ *cobra.Command, args []string) {
	server := args[0]

	if !common.ConfirmAction(common.Disruptive, server, fmt.Sprintf(
		"This creates the cloud backup containers of %s in a public cloud project.", server)) {
		display.OutputError(&flags.OutputFormatConfig, "cloud backup of %s cancelled", server)
		return
	}

	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/features/backupCloud", url.PathEscape(server))
	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	body := map[string]any{}
	if BackupCloudProjectId != "" {
		body["cloudProjectId"] = BackupCloudProjectId
	}
	if BackupCloudProjectDescription != "" {
		body["projectDescription"] = BackupCloudProjectDescription
	}

	var created map[string]any
	if err := httpLib.Client.Post(endpoint, body, &created); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to create the cloud backup of %s: %s", server, err)
		return
	}

	display.OutputObject(maskBackupCloud(created), server, "", &flags.OutputFormatConfig)
}

// DeleteBackupCloud deactivates the cloud backup of a server.
//
// Disruptive rather than Destructive, on the API's own statement: "This does
// not delete container data." The containers stay in the cloud project; what
// stops is the server's link to them.
func DeleteBackupCloud(_ *cobra.Command, args []string) {
	server := args[0]

	if !common.ConfirmAction(common.Disruptive, server, fmt.Sprintf(
		"This deactivates the cloud backup of %s. The containers and their data stay in the cloud project.",
		server)) {
		display.OutputError(&flags.OutputFormatConfig, "deactivation on %s cancelled", server)
		return
	}

	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/features/backupCloud", url.PathEscape(server))
	if common.ReportDryRun(common.Call{Method: "DELETE", Endpoint: endpoint}) {
		return
	}

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to deactivate the cloud backup of %s: %s", server, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"serviceName": server},
		"✅ The cloud backup of %s is deactivated. Its containers were not deleted.", server)
}

// ChangeBackupCloudPassword resets the four cloud backup passwords.
//
// Unlike the Backup FTP password, which the API mails, this response carries
// the credentials themselves — four of them. They are withheld behind --reveal,
// and the substitution happens on the object so -o json is covered by the same
// decision as the table: a password reset is exactly the command somebody runs
// inside a pipeline whose output is kept.
func ChangeBackupCloudPassword(_ *cobra.Command, args []string) {
	server := args[0]

	if !common.ConfirmAction(common.Disruptive, server, fmt.Sprintf(
		"Resetting the cloud backup passwords of %s stops every job still using the old ones.", server)) {
		display.OutputError(&flags.OutputFormatConfig, "password reset on %s cancelled", server)
		return
	}

	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/features/backupCloud/password", url.PathEscape(server))
	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	var passwords map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &passwords); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to reset the cloud backup passwords of %s: %s", server, err)
		return
	}

	display.OutputObject(backupPasswordView(passwords), server, backupPasswordTemplate, &flags.OutputFormatConfig)
}

// ShowOrderableBackupStorage shows the paid capacities this server accepts.
func ShowOrderableBackupStorage(_ *cobra.Command, args []string) {
	server := args[0]

	var orderable map[string]any
	path := fmt.Sprintf("/v1/dedicated/server/%s/orderable/backupStorage", url.PathEscape(server))
	if err := httpLib.Client.Get(path, &orderable); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to read the orderable backup storage of %s: %s", server, err)
		return
	}

	capacities := numberSlice(orderable["capacities"])
	if len(capacities) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"serviceName": server, "orderable": false},
			"No backup storage can be ordered for %s.", server)
		return
	}

	sizes := make([]string, 0, len(capacities))
	for _, capacity := range capacities {
		sizes = append(sizes, readableCapacity(capacity))
	}

	display.OutputInfo(&flags.OutputFormatConfig, orderable,
		"%s accepts backup storage of %s.", server, strings.Join(sizes, ", "))
}

// backupFtpSpace reads the Backup FTP space of a server.
//
// A server without one answers 404, which is a state and not a failure: only
// seven of thirty-five servers on a real account had a space. It is reported as
// the state it is, with the command that creates one.
func backupFtpSpace(server string) (map[string]any, error) {
	var space map[string]any

	path := fmt.Sprintf("/v1/dedicated/server/%s/features/backupFTP", url.PathEscape(server))
	if err := httpLib.Client.Get(path, &space); err != nil {
		if common.IsNotFound(err) {
			return nil, fmt.Errorf("%s has no Backup FTP space.\n   Create the one included with the server: ovhcloud baremetal backup ftp create %s",
				server, server)
		}

		return nil, fmt.Errorf("failed to read the Backup FTP space of %s: %w", server, err)
	}

	return space, nil
}

func backupFtpAclBlocks(server string) ([]string, error) {
	var blocks []string

	path := fmt.Sprintf("/v1/dedicated/server/%s/features/backupFTP/access", url.PathEscape(server))
	if err := httpLib.Client.Get(path, &blocks); err != nil {
		if common.IsNotFound(err) {
			return nil, fmt.Errorf("%s has no Backup FTP space, so nothing can reach it.\n   Create the one included with the server: ovhcloud baremetal backup ftp create %s",
				server, server)
		}

		return nil, fmt.Errorf("failed to read the access list of %s: %w", server, err)
	}

	return blocks, nil
}

func authorizableBlocks(server string) ([]string, error) {
	var blocks []string

	path := fmt.Sprintf("/v1/dedicated/server/%s/features/backupFTP/authorizableBlocks", url.PathEscape(server))
	if err := httpLib.Client.Get(path, &blocks); err != nil {
		if common.IsNotFound(err) {
			return nil, fmt.Errorf("%s has no Backup FTP space.\n   Create the one included with the server: ovhcloud baremetal backup ftp create %s",
				server, server)
		}

		return nil, fmt.Errorf("failed to read the blocks allowed on %s: %w", server, err)
	}

	sort.Strings(blocks)

	return blocks, nil
}

// waitForBackupFtp follows a creation or a deletion by reading the state of the
// space, not by reading the task.
//
// Same reason as everywhere else in this CLI: one request can produce more than
// one task, so a wait that trusts the task it was handed can report success
// while the other half is still running.
//
// But "the object answers 200" is not the state to wait for here, and that cost
// a run to find out. On a real creation the space appeared after 7 minutes with
// an empty ftpBackupName, and the API then refused every write against it —
// "You have a backupFTP create task pending, you cannot release your backupFtp
// until it's done." A wait that stopped there handed back a space nothing could
// be done with. The name is what says the space has been placed on a storage
// server, so that is what is waited for.
// The two directions are not the same question, and reusing one predicate for
// both would have made `delete --wait` report success while the space was still
// being created: a half-provisioned space is "not ready", but it is very much
// still there.
func waitForBackupFtp(server string, want bool) error {
	for attempt := 0; attempt < backupPollAttempts; attempt++ {
		space, err := backupFtpSpace(server)
		if backupFtpSettled(space, err == nil, want) {
			return nil
		}

		time.Sleep(backupPollInterval)
	}

	state := "created"
	if !want {
		state = "deleted"
	}

	return fmt.Errorf("stopped waiting after %s; the Backup FTP space of %s is not %s yet, follow it with: ovhcloud baremetal list-tasks %s",
		time.Duration(backupPollAttempts)*backupPollInterval, server, state, server)
}

// backupFtpReady answers whether the space is there and usable.
//
// A space with no ftpBackupName exists as an object and cannot be reached, nor
// changed: the API refuses writes while its creation task runs. Both halves are
// the same question — is there a Backup FTP space one can do something with.
// backupFtpSettled answers whether a reading shows the wanted state reached.
//
// It is a plain function of the reading so both directions can be exercised
// without provisioning a space — one of them took sixteen minutes to observe
// once, and its lesson is the whole reason this predicate exists.
func backupFtpSettled(space map[string]any, found, want bool) bool {
	if !want {
		return !found
	}

	return found && backupFtpIsUsable(space)
}

// backupFtpIsUsable says whether a space can be reached and changed.
//
// A space with no ftpBackupName exists as an object and is neither: the API
// refuses writes while its creation task runs.
func backupFtpIsUsable(space map[string]any) bool {
	name, _ := space["ftpBackupName"].(string)

	return name != ""
}

// checkAclProtocols refuses a rule that opens nothing.
//
// The API takes three booleans and accepts all three false, which creates a
// rule that appears in the access list, says an IP block is allowed, and lets
// it reach nothing. That is worse than no rule at all: it reads as protection
// that is in place.
func checkAclProtocols() error {
	if BackupAclFtp || BackupAclNfs || BackupAclCifs {
		return nil
	}

	return errors.New("an access rule has to open at least one protocol; give --ftp, --nfs or --cifs")
}

func aclProtocolList() string {
	var protocols []string
	if BackupAclFtp {
		protocols = append(protocols, "FTP")
	}
	if BackupAclNfs {
		protocols = append(protocols, "NFS")
	}
	if BackupAclCifs {
		protocols = append(protocols, "CIFS")
	}

	return strings.Join(protocols, ", ")
}

// aclProtocols renders which protocols a rule opens, so the table says what the
// rule does rather than showing three booleans to cross-read.
func aclProtocols(acl map[string]any) string {
	var protocols []string
	for _, protocol := range []struct {
		field string
		label string
	}{{"ftp", "FTP"}, {"nfs", "NFS"}, {"cifs", "CIFS"}} {
		if allowed, _ := acl[protocol.field].(bool); allowed {
			protocols = append(protocols, protocol.label)
		}
	}

	if len(protocols) == 0 {
		return "none"
	}

	return strings.Join(protocols, ", ")
}

// pickBlock finds a block among the ones this server accepts.
func pickBlock(allowed []string, wanted string) (string, bool) {
	for _, candidate := range allowed {
		if strings.EqualFold(candidate, wanted) {
			return candidate, true
		}
	}

	return "", false
}

// unauthorizableBlock refuses a block this server will not accept, and says how
// many it would.
//
// The list is not printed: a European server measured for this offered thirty.
// The count and the command that prints them answer better than thirty lines in
// an error message.
func unauthorizableBlock(server, block string, allowed []string) error {
	return fmt.Errorf("%s cannot be allowed on the Backup FTP space of %s.\n   %d block(s) can — list them with: ovhcloud baremetal backup ftp authorizable-blocks %s",
		block, server, len(allowed), server)
}

// maskBackupCloud withholds the credentials the cloud backup object carries.
//
// The reset command already withholds four passwords, on the reasoning that a
// password reset is exactly what somebody runs with the output on screen or in
// a build log. The same four are in the object every `show` and `create`
// prints, one level down: archive and storage each carry an sftp and a swift
// block, and each of those a password — `format: password` in the schema, all
// four of them. Masking on one command and printing on the two others is not a
// policy, it is an oversight; and `show` is the one that can be run again and
// again, so it is the likelier leak of the three.
//
// The walk keys on the name rather than on a list of the four paths, because
// the object is a tree the API grows: a third container, or a third protocol
// under an existing one, would otherwise be printed in the clear by a masker
// that still looked like it was doing its job.
func maskBackupCloud(object map[string]any) map[string]any {
	if RevealBackupPassword {
		return object
	}

	masked, changed := maskPasswords(object)
	view, _ := masked.(map[string]any)
	if view == nil {
		return object
	}
	if changed {
		view["hidden"] = true
	}
	return view
}

// maskPasswords copies the tree, replacing every value under a "password" key
// with its fingerprint. It copies rather than edits in place: the caller's
// object is also what -o json would have rendered, and a masker that mutated
// its input would be a masker whose correctness depended on call order.
func maskPasswords(value any) (any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		copied := make(map[string]any, len(typed))
		changed := false
		for key, item := range typed {
			if key == "password" {
				if text, ok := item.(string); ok && text != "" {
					copied[key] = common.Fingerprint(text)
					changed = true
					continue
				}
			}
			sub, subChanged := maskPasswords(item)
			copied[key] = sub
			changed = changed || subChanged
		}
		return copied, changed
	case []any:
		copied := make([]any, len(typed))
		changed := false
		for i, item := range typed {
			sub, subChanged := maskPasswords(item)
			copied[i] = sub
			changed = changed || subChanged
		}
		return copied, changed
	default:
		return value, false
	}
}

// backupPasswordView replaces each password with its fingerprint unless
// --reveal.
func backupPasswordView(passwords map[string]any) map[string]any {
	view := map[string]any{}
	for key, value := range passwords {
		view[key] = value
	}

	if RevealBackupPassword {
		return view
	}

	for _, field := range []string{"sftpArchive", "sftpStorage", "swiftArchive", "swiftStorage"} {
		if value, set := view[field]; set {
			view[field] = common.Fingerprint(fmt.Sprint(value))
		}
	}
	view["hidden"] = true

	return view
}

// readableCapacity says a size in gigabytes the way a catalogue does.
func readableCapacity(gigabytes int64) string {
	if gigabytes >= 1000 && gigabytes%1000 == 0 {
		return fmt.Sprintf("%d TB", gigabytes/1000)
	}

	return fmt.Sprintf("%d GB", gigabytes)
}

func numberSlice(value any) []int64 {
	items, ok := value.([]any)
	if !ok {
		return nil
	}

	numbers := make([]int64, 0, len(items))
	for _, item := range items {
		switch number := item.(type) {
		case json.Number:
			if parsed, err := number.Int64(); err == nil {
				numbers = append(numbers, parsed)
			}
		case float64:
			numbers = append(numbers, int64(number))
		case int64:
			numbers = append(numbers, number)
		}
	}

	return numbers
}
