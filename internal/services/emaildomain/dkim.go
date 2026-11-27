// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package emaildomain

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
)

//go:embed templates/dkim.tmpl
var dkimTemplate string

// Values of email.domain.DKIMStatusEnum.
const (
	dkimStatusEnabled  = "enabled"
	dkimStatusDisabled = "disabled"

	// dkimStatusError is reported both by a change that failed and by a
	// configuration that was already broken before the command ran.
	dkimStatusError = "error"

	// dkimStatusModifying is the only status that proves a change started.
	dkimStatusModifying = "modifying"

	// dkimStatusToConfigure is read as an enabled DKIM waiting for its
	// selectors to be published in DNS, which is something the API cannot do
	// on its own: an end of the wait rather than a step towards "enabled".
	// That reading comes from the name of the enum value and from the
	// selectors the same answer carries; it has not been checked against the
	// live API.
	dkimStatusToConfigure = "toConfigure"
)

func dkimEndpoint(domain string) string {
	return fmt.Sprintf("/email/domain/%s/dkim", url.PathEscape(domain))
}

func GetDKIM(_ *cobra.Command, args []string) {
	domain := args[0]

	var dkim map[string]any
	if err := httpLib.Client.Get(dkimEndpoint(domain), &dkim); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch the DKIM configuration of %s: %s", domain, err)
		return
	}

	display.OutputObject(dkim, domain, dkimTemplate, &flags.OutputFormatConfig)
}

func EnableDKIM(_ *cobra.Command, args []string) {
	// "toConfigure" is read as a resting state of an enabled DKIM, so the wait
	// ends there too rather than run to its timeout. See dkimStatusToConfigure
	// for what that reading rests on.
	switchDKIM(args[0], "enable", dkimStatusEnabled, []string{dkimStatusEnabled, dkimStatusToConfigure})
}

func DisableDKIM(_ *cobra.Command, args []string) {
	switchDKIM(args[0], "disable", dkimStatusDisabled, []string{dkimStatusDisabled})
}

// switchDKIM turns DKIM signing on or off for a domain.
//
// Both endpoints are PUTs that answer with a task, where the other PUTs of
// this API answer with nothing at all. The identifier in that answer is a task
// identifier and never says anything about DKIM itself. Unlike the redirection
// queue, this task carries a status (email.domain.task), which gives --wait
// something to read instead of a disappearance.
//
// The API exposes one queue per family and none of them is named after DKIM;
// "all" is the one that serves these tasks, as the live API confirms. The wait
// still ends on the DKIM configuration itself: a task reaching "done" says the
// operation ran, not what state DKIM came to rest in, and "toConfigure" in
// particular is only ever visible on the object.
func switchDKIM(domain, action, target string, settled []string) {
	endpoint := fmt.Sprintf("%s/%s", dkimEndpoint(domain), action)

	var task map[string]any
	if err := httpLib.Client.Put(endpoint, nil, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to %s DKIM on %s: %s", action, domain, err)
		return
	}

	taskID := taskIdentifier(task)

	if !flags.WaitForTask {
		if taskID == "" {
			display.OutputInfo(&flags.OutputFormatConfig, task,
				"⏳ Request to %s DKIM on %s accepted (the API returned no task to follow)", action, domain)
			return
		}

		display.OutputInfo(&flags.OutputFormatConfig, task,
			"⏳ Request to %s DKIM on %s accepted (task ID: %s). Use --wait to wait for it to be applied.", action, domain, taskID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TaskTimeout)
	defer cancel()

	if taskID != "" {
		if err := waitForTask(ctx, domain, taskFamilyAll, taskID); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "error waiting for DKIM to be %s on %s: %s%s",
				target, domain, err, dkimStatusSuffix(domain))
			return
		}
	}

	dkim, status, err := waitForDKIM(ctx, domain, target, settled)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error waiting for DKIM to be %s on %s: %s", target, domain, err)
		return
	}

	if status != target {
		// Reached a resting state that is not the one asked for: the request
		// went through, but something outside the API has to happen before
		// DKIM signs anything. Saying "done" here would be a lie.
		if records := dkimSelectorsToSet(dkim); records != "" {
			display.OutputInfo(&flags.OutputFormatConfig, dkim,
				"⚠️ DKIM on %s is in status %q, not %q yet: the API reports these selectors as not set in DNS: %s",
				domain, status, target, records)
			return
		}

		display.OutputInfo(&flags.OutputFormatConfig, dkim,
			"⚠️ DKIM on %s is in status %q, not %q yet", domain, status, target)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, dkim, "✅ DKIM is now %s on %s", status, domain)
}

// dkimStatusSuffix reads where DKIM stands right now, to be appended to an
// error about a task that failed or was abandoned.
//
// A task that did not run to its end says the operation stopped, not what became
// of DKIM, and the two readings only mean something together: the live API was
// seen cancelling a "dkim/enable" and leaving DKIM "disabled", i.e. nothing
// changed. Without this the caller is told an opaque identifier failed and has to
// go and look for the rest.
//
// This only decorates a message that is already an error, so it never produces
// one of its own: an answer that does not come, or that carries no status, simply
// adds nothing. It is equally not a second opinion on the outcome — a status that
// happens to match what was asked for does not turn a failed task into a success,
// since what makes the API give up on a task has not been established.
func dkimStatusSuffix(domain string) string {
	var dkim map[string]any
	if err := httpLib.Client.Get(dkimEndpoint(domain), &dkim); err != nil {
		return ""
	}

	status, _ := dkim["status"].(string)
	if status == "" {
		return ""
	}

	return fmt.Sprintf(" (DKIM on %s is in status %q)", domain, status)
}

// dkimSelectorsToSet lists the selectors the API does not report as "set", as
// "<name>: <cname>" pairs. The values are repeated as the API wrote them: what
// a given zone has to hold is the registrar's business, not something to guess
// here.
func dkimSelectorsToSet(dkim map[string]any) string {
	selectors, _ := dkim["selectors"].([]any)

	var records []string
	for _, entry := range selectors {
		selector, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if status, _ := selector["status"].(string); status == "set" {
			continue
		}

		name, _ := selector["selectorName"].(string)
		cname, _ := selector["cname"].(string)
		if name == "" && cname == "" {
			continue
		}

		records = append(records, fmt.Sprintf("%s: %s", name, cname))
	}

	return strings.Join(records, ", ")
}

// waitForDKIM polls the DKIM configuration of a domain until its status is
// target or one of the other settled statuses, and returns that configuration
// along with the status it stopped on.
//
// Only target is believed on sight. Every other reading may still describe the
// domain as it was before the command ran: the task reaching "done" does not
// guarantee this view of the configuration has caught up with it, and when the
// API answers with no task at all there was nothing to wait on in the first
// place.
//
//   - another resting status, such as "toConfigure" after an enable, is where
//     the change may have settled, or where the domain already was;
//   - "error" may be this change failing, or a setup that was already broken
//     before it — and answering "your command failed" to someone who is
//     precisely trying to repair or turn off a broken setup would be wrong.
//
// Both are therefore acted on only once the change has been seen under way
// ("modifying"), or once the same status has held for maxStillPolls polls,
// which means nothing is moving. waitForCloudResourceDeleted already reads
// deletions this way.
func waitForDKIM(ctx context.Context, domain, target string, settled []string) (map[string]any, string, error) {
	const maxStillPolls = 3

	// Read while the whole budget is still ahead: once the wait gives up there
	// is nothing left to measure.
	gaveUp := gaveUpWaiting(ctx)
	endpoint := dkimEndpoint(domain)

	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	var (
		status     string
		lastStatus string
		stillPolls int
		seenModify bool
	)

	for {
		var dkim map[string]any
		if err := httpLib.Client.Get(endpoint, &dkim); err != nil {
			return nil, status, fmt.Errorf("failed to fetch the DKIM configuration of %s: %w", domain, err)
		}

		status, _ = dkim["status"].(string)

		// The status that was asked for needs no corroboration: it is the end
		// of the change, and reading it before the change started only means
		// the domain was already there.
		if status == target {
			return dkim, status, nil
		}

		if status != lastStatus {
			lastStatus, stillPolls = status, 0
		}
		stillPolls++

		settledEnough := seenModify || stillPolls >= maxStillPolls

		switch {
		case status == dkimStatusModifying:
			seenModify = true

		case slices.Contains(settled, status):
			if settledEnough {
				return dkim, status, nil
			}

		case status == dkimStatusError:
			if seenModify {
				return dkim, status, fmt.Errorf("DKIM of %s ended in status %q", domain, status)
			}
			if settledEnough {
				return dkim, status, fmt.Errorf("DKIM of %s was already in status %q and never started changing", domain, status)
			}
		}

		select {
		case <-ctx.Done():
			// The wait ran out, not the operation: the status named here was
			// read a whole interval ago and is given as the last thing seen.
			return dkim, status, fmt.Errorf("%s: DKIM of %s was in status %q on the last check", gaveUp, domain, status)
		case <-ticker.C:
			log.Printf("Still waiting for the DKIM of %s to settle (status=%s)…", domain, status)
		}
	}
}
