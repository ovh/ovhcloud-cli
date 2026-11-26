// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package emaildomain

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/ovh/go-ovh/ovh"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// Families of tasks exposed under /email/domain/{domain}/task. The API keeps
// one queue per family and an identifier is only meaningful inside its own, so
// the family has to travel with the identifier.
const (
	taskFamilyRedirection = "redirection"

	// taskFamilyAll is the queue holding every task of a domain. It is the
	// only one for operations that have no family of their own, such as
	// switching DKIM on or off.
	taskFamilyAll = "all"
)

var (
	// PollInterval is the delay between two polls. It is a variable so that the
	// tests do not have to wait for real seconds.
	PollInterval = 5 * time.Second

	// TaskTimeout bounds how long --wait follows a task. Reaching it cancels
	// nothing: only the waiting stops.
	TaskTimeout = 10 * time.Minute
)

// taskIdentifier returns the identifier of a task as the API wrote it, or ""
// when the response carries none.
//
// It is formatted with %v rather than with a numeric verb: the email domain API
// types these identifiers as strings while other task families use integers,
// and the client decodes numbers with UseNumber, so %d would print
// "%!d(json.Number=156839472)" in the middle of the one message meant to tell
// the user what to follow.
func taskIdentifier(task map[string]any) string {
	if task == nil {
		return ""
	}

	id, ok := task["id"]
	if !ok || id == nil {
		return ""
	}

	return fmt.Sprintf("%v", id)
}

// isNotFound reports whether an API call failed because the object is not (or
// no longer) known.
func isNotFound(err error) bool {
	apiErr, ok := errors.AsType[*ovh.APIError](err)
	return ok && apiErr.Code == http.StatusNotFound
}

// gaveUpWaiting opens the message of a wait that ran out of time.
//
// The duration comes from the context's own deadline rather than from
// TaskTimeout, so a wait that was given five seconds does not report having
// waited ten minutes.
func gaveUpWaiting(ctx context.Context) string {
	deadline, ok := ctx.Deadline()
	if !ok {
		return "gave up waiting"
	}

	remaining := time.Until(deadline)
	if remaining > time.Second {
		// Whole seconds read better than "9m59.998s"; a budget shorter than a
		// second is only ever set by a test, and rounding it would print "0s".
		remaining = remaining.Round(time.Second)
	}

	return fmt.Sprintf("gave up waiting after %s", remaining)
}

// waitForTask follows an email domain task until it leaves its queue.
//
// The schema of a redirection task has no status field, so there is nothing to
// read that would say the task is over. What is taken as the end of it is the
// API no longer serving it: a 404 is read as "left the queue". That is what the
// live API was seen doing — a redirection task ends up answering 404.
//
// It is still not the last word, and callers confirm the outcome on the object
// itself: a task that is gone does not prove the operation succeeded, and 404
// is also what a wrong path answers. A status is honoured whenever the API does
// send one (the /task/all family has one), so a family that reports its state
// is not waited on for a disappearance that may never come.
func waitForTask(ctx context.Context, domain, family, taskID string) error {
	endpoint := fmt.Sprintf("/email/domain/%s/task/%s/%s",
		url.PathEscape(domain), url.PathEscape(family), url.PathEscape(taskID))

	// Read while the full budget is still ahead: by the time the wait gives up
	// there is nothing left to measure.
	gaveUp := gaveUpWaiting(ctx)

	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	for {
		var task map[string]any

		err := httpLib.Client.Get(endpoint, &task)
		if isNotFound(err) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to fetch task %s: %w", taskID, err)
		}

		status, _ := task["status"].(string)
		switch status {
		case "done":
			return nil

		// The terminal failures of the API enum. They mean the operation
		// failed, not that the CLI failed to understand the answer.
		case "cancelled", "customerError", "ovhError":
			return fmt.Errorf("task %s ended with status %q", taskID, status)
		}

		select {
		case <-ctx.Done():
			// What ran out is the wait, not the task. The status named here was
			// read a whole interval ago, so it is given as the last thing seen
			// and not as the state of the task now.
			seen := ""
			if status != "" {
				seen = fmt.Sprintf(" (last status seen: %s)", status)
			}
			return fmt.Errorf("%s: task %s had not completed%s, and it was not cancelled", gaveUp, taskID, seen)

		case <-ticker.C:
			if status != "" {
				log.Printf("Still waiting for task %s to complete (status=%s)…", taskID, status)
			} else {
				log.Printf("Still waiting for task %s to complete…", taskID)
			}
		}
	}
}

// waitUntilGone polls an object until the API no longer knows it. It is how a
// deletion is confirmed on the object itself, rather than on the disappearance
// of the task that was supposed to perform it.
func waitUntilGone(ctx context.Context, endpoint, description string) error {
	gaveUp := gaveUpWaiting(ctx)

	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	for {
		err := httpLib.Client.Get(endpoint, nil)
		if isNotFound(err) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to fetch %s: %w", description, err)
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("%s: %s was still there on the last check", gaveUp, description)
		case <-ticker.C:
			log.Printf("Still waiting for %s to be deleted…", description)
		}
	}
}
