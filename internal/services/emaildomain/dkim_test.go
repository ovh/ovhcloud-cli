// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package emaildomain

import (
	"context"
	"testing"
	"time"

	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
)

const dkimURL = "https://eu.api.ovh.com/1.0/email/domain/example.com/dkim"

var enableSettles = []string{dkimStatusEnabled, dkimStatusToConfigure}

// A DKIM that was already broken answers "error" from the very first poll. The
// wait must not read that as the command having failed: the person running
// "dkim disable" on a broken setup is doing the one thing that fixes it.
func TestWaitForDKIM_AnErrorAlreadyThereIsNotTheCommandFailing(t *testing.T) {
	withTaskAPI(t)
	respondInSequence(dkimURL,
		answer{200, `{"status": "error"}`},
		answer{200, `{"status": "error"}`},
		answer{200, `{"status": "modifying"}`},
		answer{200, `{"status": "disabled"}`},
	)

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	dkim, status, err := waitForDKIM(ctx, "example.com", dkimStatusDisabled, []string{dkimStatusDisabled})

	td.Require(t).CmpNoError(err)
	td.Cmp(t, status, dkimStatusDisabled)
	td.Cmp(t, dkim["status"], dkimStatusDisabled)
}

// Nothing ever starts: the status does not budge off "error". Waiting the full
// ten minutes for a change that was refused helps nobody, so the wait stops and
// says the change never started — not that it failed.
func TestWaitForDKIM_AnErrorThatNeverMovesStopsTheWait(t *testing.T) {
	withTaskAPI(t)
	respondInSequence(dkimURL, answer{200, `{"status": "error"}`})

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	_, status, err := waitForDKIM(ctx, "example.com", dkimStatusEnabled, enableSettles)

	td.Require(t).CmpError(err)
	td.Cmp(t, status, dkimStatusError)
	td.Cmp(t, err.Error(), td.Contains("never started changing"))
	td.Cmp(t, err.Error(), td.Not(td.Contains("gave up waiting")),
		"the wait stopped on what it read, it did not run out of time")
}

// Once the change has been seen under way, "error" can only be that change
// failing, and it is reported as such.
func TestWaitForDKIM_AnErrorAfterTheChangeStartedIsFatal(t *testing.T) {
	withTaskAPI(t)
	respondInSequence(dkimURL,
		answer{200, `{"status": "modifying"}`},
		answer{200, `{"status": "error"}`},
	)

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	_, _, err := waitForDKIM(ctx, "example.com", dkimStatusEnabled, enableSettles)

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains(`ended in status "error"`))
}

// "toConfigure" is read as a resting state, so the wait ends there instead of
// spending the whole budget on a change that is not coming.
func TestWaitForDKIM_ToConfigureEndsTheWait(t *testing.T) {
	withTaskAPI(t)
	respondInSequence(dkimURL,
		answer{200, `{"status": "modifying"}`},
		answer{200, `{"status": "toConfigure"}`},
	)

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	_, status, err := waitForDKIM(ctx, "example.com", dkimStatusEnabled, enableSettles)

	td.Require(t).CmpNoError(err)
	td.Cmp(t, status, dkimStatusToConfigure)
}

// A domain can already be in "toConfigure" before the command runs. Taking the
// first poll at its word would report that resting state as the outcome of a
// change that had not started — the very trap the "error" handling avoids.
func TestWaitForDKIM_AToConfigureAlreadyThereIsNotTheOutcome(t *testing.T) {
	withTaskAPI(t)
	respondInSequence(dkimURL,
		answer{200, `{"status": "toConfigure"}`},
		answer{200, `{"status": "modifying"}`},
		answer{200, `{"status": "enabled"}`},
	)

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	_, status, err := waitForDKIM(ctx, "example.com", dkimStatusEnabled, enableSettles)

	td.Require(t).CmpNoError(err)
	td.Cmp(t, status, dkimStatusEnabled)
}

// Running out of time is neither a failure nor knowledge of the present state:
// the status quoted was read an interval ago, and is given as such.
func TestWaitForDKIM_TimeoutNamesTheLastStatusSeen(t *testing.T) {
	withTaskAPI(t)
	respondInSequence(dkimURL, answer{200, `{"status": "modifying"}`})

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()

	_, _, err := waitForDKIM(ctx, "example.com", dkimStatusEnabled, enableSettles)

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains("gave up waiting after"))
	td.Cmp(t, err.Error(), td.Contains(`was in status "modifying" on the last check`))
	td.Cmp(t, err.Error(), td.Not(td.Contains("failed")))
}

// Only the selectors the API does not report as set are worth printing, and
// they are printed as the API wrote them.
func TestDKIMSelectorsToSet(t *testing.T) {
	td.Cmp(t, dkimSelectorsToSet(map[string]any{
		"selectors": []any{
			map[string]any{"selectorName": "ovhse1", "cname": "ovhse1.dkim.mail.ovh.net", "status": "toSet"},
			map[string]any{"selectorName": "ovhse2", "cname": "ovhse2.dkim.mail.ovh.net", "status": "set"},
		},
	}), "ovhse1: ovhse1.dkim.mail.ovh.net")

	td.Cmp(t, dkimSelectorsToSet(map[string]any{}), "",
		"a configuration without selectors has nothing to list")
}

// A DKIM that is off carries no selector, no active selector and no autoconfig
// flag. Chaining into them anyway is how a template ends up printing
// "<no value>", or failing outright on an index into nil.
func TestDKIMTemplate_RendersWithoutTheNullableFields(t *testing.T) {
	prevExitFunc := display.ExitFunc
	display.ExitFunc = func(int) {}
	t.Cleanup(func() { display.ExitFunc = prevExitFunc })

	display.ResultError = nil
	display.OutputObject(map[string]any{"status": "disabled"}, "example.com", dkimTemplate, &flags.OutputFormatConfig)

	td.CmpNoError(t, display.ResultError)
	td.Cmp(t, display.ResultString, td.Contains("disabled"))
	td.Cmp(t, display.ResultString, td.Not(td.Contains("<no value>")))
	td.Cmp(t, display.ResultString, td.Not(td.Contains("<nil>")))
}

func TestDKIMTemplate_ListsTheSelectors(t *testing.T) {
	prevExitFunc := display.ExitFunc
	display.ExitFunc = func(int) {}
	t.Cleanup(func() { display.ExitFunc = prevExitFunc })

	display.ResultError = nil
	display.OutputObject(map[string]any{
		"status":         "enabled",
		"activeSelector": "ovhse1",
		"autoconfig":     false,
		"selectors": []any{
			map[string]any{"selectorName": "ovhse1", "cname": "ovhse1.dkim.mail.ovh.net", "status": "set"},
		},
	}, "example.com", dkimTemplate, &flags.OutputFormatConfig)

	td.CmpNoError(t, display.ResultError)
	td.Cmp(t, display.ResultString, td.Contains("ovhse1.dkim.mail.ovh.net"))
	td.Cmp(t, display.ResultString, td.Contains("false"),
		"an autoconfig that is off is the one worth reading, and must not be hidden")
}
