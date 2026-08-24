// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package vrack

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// waitForAttachment blocks until the vRack really holds — or no longer holds —
// the given interface.
//
// It polls the membership itself and treats the task as a progress hint only.
// That order is deliberate, and it comes from a measurement made on a sibling
// command: a dedicated server reboot task reported `done` at sixty seconds
// while the machine only obeyed at t+143s. A wait that trusts the task proves
// the call was accepted, not that the change happened, which is precisely the
// difference an operator is waiting for.
//
// The task is still read, for one thing the state cannot say: when it ends
// cancelled, the wait stops with that reason instead of running to the timeout.
// vrack.TaskStatusEnum is cancelled, doing, done, init and todo — it has
// neither customerError nor ovhError, unlike the dedicated server enum, so
// there is nothing else to catch here.
func waitForAttachment(vrack, uuid string, wantPresent bool, task map[string]any) error {
	deadline := time.Now().Add(VrackTimeout)

	for attempt := 0; ; attempt++ {
		present, err := isAttached(vrack, uuid)
		if err != nil {
			return err
		}
		if present == wantPresent {
			return nil
		}

		if reason := taskFailure(vrack, task); reason != "" {
			return fmt.Errorf("the vRack operation did not complete: %s", reason)
		}

		if time.Now().After(deadline) {
			return fmt.Errorf(
				"%s did not report interface %s as %s within %s.\nThe operation may still be running; check it with `ovhcloud vrack get %s`",
				vrack, uuid, presence(wantPresent), VrackTimeout, vrack)
		}

		if attempt > 0 {
			log.Printf("Still waiting for %s to be %s %s…", uuid, presence(wantPresent), vrack)
		}
		time.Sleep(taskPollInterval)
	}
}

func presence(wantPresent bool) string {
	if wantPresent {
		return "in"
	}
	return "out of"
}

// isAttached reads the membership the API actually holds.
func isAttached(vrack, uuid string) (bool, error) {
	var uuids []string
	endpoint := fmt.Sprintf("/v1/vrack/%s/dedicatedServerInterface", url.PathEscape(vrack))

	if err := httpLib.Client.Get(endpoint, &uuids); err != nil {
		return false, fmt.Errorf("failed to read the interfaces of %s: %w", vrack, err)
	}

	for _, attached := range uuids {
		if attached == uuid {
			return true, nil
		}
	}
	return false, nil
}

// taskFailure reports why the task stopped, when it stopped badly.
//
// It answers an empty string for everything else, including a task it could not
// read: a wait must not fail because the progress hint was unavailable, when the
// state it really depends on is being read successfully next to it.
func taskFailure(vrack string, task map[string]any) string {
	id, ok := taskID(task)
	if !ok {
		return ""
	}

	var current map[string]any
	endpoint := fmt.Sprintf("/v1/vrack/%s/task/%s", url.PathEscape(vrack), id)
	if err := httpLib.Client.Get(endpoint, &current); err != nil {
		return ""
	}

	if fmt.Sprintf("%v", current["status"]) == "cancelled" {
		return fmt.Sprintf("task %s was cancelled", id)
	}
	return ""
}

// taskID pulls the identifier out of whatever the API returned.
//
// The value arrives as a JSON number, so it lands in an interface{} as a
// float64: formatting it with %v would produce "1.234567e+06" and request a
// task that does not exist. It is formatted as an integer instead, and a body
// without a usable id simply yields no task to follow.
func taskID(task map[string]any) (string, bool) {
	raw, found := task["id"]
	if !found {
		return "", false
	}

	switch value := raw.(type) {
	case json.Number:
		// This is the one that happens. go-ovh decodes responses with
		// UseNumber, so every real task id arrives as a json.Number — which a
		// type switch does not silently widen to float64. Without this case the
		// cancelled-task check below never fires, and a cancelled operation is
		// reported ten minutes later as one that may still be running.
		return value.String(), true
	case float64:
		return fmt.Sprintf("%.0f", value), true
	case int:
		return fmt.Sprintf("%d", value), true
	case string:
		if value == "" {
			return "", false
		}
		return value, true
	default:
		return "", false
	}
}
