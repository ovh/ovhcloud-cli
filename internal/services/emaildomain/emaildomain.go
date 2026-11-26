// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package emaildomain

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	emaildomainColumnsToDisplay = []string{"domain", "status", "offer"}

	//go:embed templates/emaildomain.tmpl
	emaildomainTemplate string

	//go:embed templates/redirection.tmpl
	redirectionTemplate string

	//go:embed parameter-samples/redirection-create.json
	RedirectionCreateExample string

	RedirectionSpec struct {
		From      string `json:"from"`
		To        string `json:"to"`
		LocalCopy bool   `json:"localCopy"`
	}
)

func ListEmailDomain(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/email/domain", "", emaildomainColumnsToDisplay, flags.GenericFilters)
}

func GetEmailDomain(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/email/domain", args[0], emaildomainTemplate)
}

func ListRedirections(_ *cobra.Command, args []string) {
	serviceName := args[0]
	path := fmt.Sprintf("/email/domain/%s/redirection", url.PathEscape(serviceName))
	columnsToDisplay := []string{"id", "from", "to", "localCopy"}
	common.ManageListRequest(path, "", columnsToDisplay, flags.GenericFilters)
}

func GetRedirection(_ *cobra.Command, args []string) {
	serviceName := args[0]
	redirectionID := args[1]
	path := fmt.Sprintf("/email/domain/%s/redirection", url.PathEscape(serviceName))
	common.ManageObjectRequest(path, redirectionID, redirectionTemplate)
}

// errRedirectionNotFound is returned by lookupRedirection while the API does
// not know a redirection between the two addresses yet.
var errRedirectionNotFound = errors.New("redirection not found")

// lookupRedirection returns the redirection between two addresses.
//
// Creating a redirection only answers with a task, so this lookup is the only
// way to the identifier the user needs next. Both addresses are used as
// filters: narrowing on the source alone would match every redirection fanning
// out from it.
func lookupRedirection(serviceName, from, to string) (map[string]any, error) {
	query := url.Values{}
	query.Set("from", from)
	query.Set("to", to)

	endpoint := fmt.Sprintf("/email/domain/%s/redirection?%s", url.PathEscape(serviceName), query.Encode())

	var ids []string
	if err := httpLib.Client.Get(endpoint, &ids); err != nil {
		return nil, fmt.Errorf("failed to list redirections from %s to %s: %w", from, to, err)
	}

	switch len(ids) {
	case 0:
		return nil, errRedirectionNotFound
	case 1:
	default:
		// Guessing one of them would print an identifier that may belong to a
		// redirection the user did not just create.
		return nil, fmt.Errorf("%d redirections go from %s to %s, cannot tell which one was just created", len(ids), from, to)
	}

	endpoint = fmt.Sprintf("/email/domain/%s/redirection/%s", url.PathEscape(serviceName), url.PathEscape(ids[0]))

	var redirection map[string]any
	if err := httpLib.Client.Get(endpoint, &redirection); err != nil {
		return nil, fmt.Errorf("failed to fetch redirection %s: %w", ids[0], err)
	}

	return redirection, nil
}

// waitForRedirection polls until the redirection between the two addresses
// exists. The task being over is the platform's own signal; this confirms the
// result on the object the user asked for, so a task that was never found does
// not get reported as a success.
func waitForRedirection(ctx context.Context, serviceName, from, to string) (map[string]any, error) {
	gaveUp := gaveUpWaiting(ctx)

	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	for {
		redirection, err := lookupRedirection(serviceName, from, to)
		if !errors.Is(err, errRedirectionNotFound) {
			return redirection, err
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%s: no redirection from %s to %s had appeared on the last check", gaveUp, from, to)
		case <-ticker.C:
			log.Printf("Still waiting for the redirection from %s to %s to appear…", from, to)
		}
	}
}

func CreateRedirection(cmd *cobra.Command, args []string) {
	serviceName := args[0]

	// The parameters actually sent are needed to find the created redirection
	// back, because they are not always the ones RedirectionSpec holds: under
	// --editor the request is built from what the user edited and the flag
	// struct is never consulted.
	task, parameters, err := common.CreateResourceWithParameters(
		cmd,
		"/email/domain/{serviceName}/redirection",
		fmt.Sprintf("/email/domain/%s/redirection", url.PathEscape(serviceName)),
		RedirectionCreateExample,
		RedirectionSpec,
		assets.EmaildomainOpenapiSchema,
		[]string{"from", "to"},
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error creating redirection: %s", err)
		return
	}

	// This endpoint creates asynchronously: it answers with the task that will
	// create the redirection, so the "id" of the answer is a task identifier.
	// It used to be printed as the redirection's own, which sent people to an
	// identifier that "email-domain redirection get" does not know.
	taskID := taskIdentifier(task)
	if taskID == "" {
		display.OutputInfo(&flags.OutputFormatConfig, task,
			"⏳ Email redirection creation requested (the API returned no task to follow)")
		return
	}

	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, task,
			"⏳ Email redirection creation requested (task ID: %s). Use --wait to wait for it to be applied.", taskID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TaskTimeout)
	defer cancel()

	if err := waitForTask(ctx, serviceName, taskFamilyRedirection, taskID); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error waiting for the redirection to be created: %s", err)
		return
	}

	from, _ := parameters["from"].(string)
	to, _ := parameters["to"].(string)
	if from == "" || to == "" {
		display.OutputInfo(&flags.OutputFormatConfig, task,
			"✅ Email redirection created successfully (task %s done, addresses unknown so its ID could not be retrieved)", taskID)
		return
	}

	redirection, err := waitForRedirection(ctx, serviceName, from, to)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "task %s is over but the created redirection could not be retrieved: %s", taskID, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, redirection,
		"✅ Email redirection created successfully (ID: %v)", redirection["id"])
}

func DeleteRedirection(_ *cobra.Command, args []string) {
	serviceName := args[0]
	redirectionID := args[1]
	path := fmt.Sprintf("/email/domain/%s/redirection/%s", url.PathEscape(serviceName), url.PathEscape(redirectionID))

	// Deleting is asynchronous as well: the API answers with the task that will
	// remove the redirection.
	var task map[string]any
	if err := httpLib.Client.Delete(path, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete redirection: %s", err)
		return
	}

	taskID := taskIdentifier(task)

	if !flags.WaitForTask {
		if taskID == "" {
			display.OutputInfo(&flags.OutputFormatConfig, task, "⏳ Email redirection %s deletion requested", redirectionID)
			return
		}

		display.OutputInfo(&flags.OutputFormatConfig, task,
			"⏳ Email redirection %s deletion requested (task ID: %s). Use --wait to wait for it to be applied.", redirectionID, taskID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TaskTimeout)
	defer cancel()

	if taskID != "" {
		if err := waitForTask(ctx, serviceName, taskFamilyRedirection, taskID); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "error waiting for the redirection to be deleted: %s", err)
			return
		}
	}

	if err := waitUntilGone(ctx, path, fmt.Sprintf("redirection %s", redirectionID)); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error waiting for the redirection to be deleted: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Email redirection %s deleted successfully", redirectionID)
}
