// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"errors"
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// The route answers 404 outside an installation, and that sentence is the only
// thing separating "nothing is running" from "the call failed".
func TestIsNotBeingInstalledRecognisesTheIdleAnswer(t *testing.T) {
	td.Cmp(t, isNotBeingInstalled(
		errors.New(`Error 404: "Server is not being installed or reinstalled at the moment"`)), true)

	td.Cmp(t, isNotBeingInstalled(
		errors.New(`Error 403: "This call has not been granted"`)), false,
		"a real failure must stay one")

	td.Cmp(t, isNotBeingInstalled(nil), false)
}

// A reinstall runs for tens of minutes, so the number somebody reads has to be
// minutes. The API counts in seconds.
func TestFormatElapsedReadsAsTime(t *testing.T) {
	td.Cmp(t, formatElapsed(0), "0s")
	td.Cmp(t, formatElapsed(59), "59s")
	td.Cmp(t, formatElapsed(60), "1m00s")
	td.Cmp(t, formatElapsed(754), "12m34s")
	td.Cmp(t, formatElapsed(3600), "60m00s")
}

// Measured on a real reinstall, twice: elapsedTime answered -1935 during the
// hardware reboot and -87 again after the final one. A negative age is not a
// duration, so it is not printed as one — and never as a large positive
// number, which is what an unsigned reading of it would have produced.
func TestFormatElapsedRefusesToInventANegativeDuration(t *testing.T) {
	td.Cmp(t, formatElapsed(-1935), unknownElapsed)
	td.Cmp(t, formatElapsed(-87), unknownElapsed)
	td.Cmp(t, formatElapsed(-1), unknownElapsed)
}

// The running step is the one the operator is waiting on.
func TestCurrentStepIsTheOneBeingWorkedOn(t *testing.T) {
	progress := InstallationProgress{}
	progress.Progress = append(progress.Progress,
		step("Checking BIOS version", "done", ""),
		step("Running Hardware Reboot", "doing", ""),
		step("Setting up hardware raid", "todo", ""))

	comment, position, ok := progress.current()

	td.Cmp(t, ok, true)
	td.Cmp(t, comment, "Running Hardware Reboot")
	td.Cmp(t, position, 2)
}

// Every step done and none running is what an installation looks like on the
// poll after the last one finished. Reporting a running step there would name
// one that is not.
func TestCurrentStepIsAbsentWhenNothingIsRunning(t *testing.T) {
	progress := InstallationProgress{}
	progress.Progress = append(progress.Progress,
		step("Checking BIOS version", "done", ""),
		step("Rebooting", "done", ""))

	_, _, ok := progress.current()

	td.Cmp(t, ok, false)
}

// A failed step carries why in a field of its own, and that is the whole value
// of reading the status at all when something goes wrong.
func TestFailedStepCarriesItsReason(t *testing.T) {
	progress := InstallationProgress{}
	progress.Progress = append(progress.Progress,
		step("Checking BIOS version", "done", ""),
		step("Preparing disks for new Partitioning", "error", "disk 2 is not present"))

	comment, reason, ok := progress.failed()

	td.Cmp(t, ok, true)
	td.Cmp(t, comment, "Preparing disks for new Partitioning")
	td.Cmp(t, reason, "disk 2 is not present")
}

func step(comment, status, failure string) struct {
	Comment string `json:"comment"`
	Error   string `json:"error"`
	Status  string `json:"status"`
} {
	return struct {
		Comment string `json:"comment"`
		Error   string `json:"error"`
		Status  string `json:"status"`
	}{Comment: comment, Error: failure, Status: status}
}
