// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"bytes"
	"errors"
	"log"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/go-ovh/ovh"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// asked records every question the wizard puts to the operator, and answers
// each one from a script. A question that was never asked, or asked with the
// wrong options, is the failure these tests are looking for.
type asked struct {
	questions []string
	options   []map[string]string
	answers   []string
	err       error
}

func (a *asked) pick(question string, choices map[string]string, _ int) (string, string, error) {
	a.questions = append(a.questions, question)
	a.options = append(a.options, choices)
	if a.err != nil {
		return "", "", a.err
	}
	if len(a.answers) == 0 {
		return "", "", errors.New("the wizard asked one question too many: " + question)
	}
	answer := a.answers[0]
	a.answers = a.answers[1:]
	return answer, choices[answer], nil
}

// withWizardAPI points the shared client at httpmock and puts a scripted
// operator at the keyboard.
func withWizardAPI(t *testing.T, answers ...string) (*asked, *bytes.Buffer) {
	t.Helper()
	httpmock.Activate(t)

	client, err := ovh.NewClient("ovh-eu", "app_key", "app_secret", "consumer_key")
	td.Require(t).CmpNoError(err)

	operator := &asked{answers: answers}
	var logs bytes.Buffer

	origClient, origPicker, origFlags := httpLib.Client, choicePicker, log.Flags()
	origInteractive := interactive
	httpLib.Client = client
	choicePicker = operator.pick
	// `go test` has no terminal. Without this the wizard would refuse before
	// asking anything, and every test below would pass for the wrong reason.
	interactive = func() bool { return true }
	log.SetOutput(&logs)
	log.SetFlags(0)
	t.Cleanup(func() {
		httpLib.Client = origClient
		choicePicker = origPicker
		interactive = origInteractive
		log.SetOutput(nil)
		log.SetFlags(origFlags)
	})

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
		httpmock.NewStringResponder(200, "0"))

	return operator, &logs
}

func templatesAre(body string) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/srv/install/compatibleTemplates",
		httpmock.NewStringResponder(200, body))
}

func schemesAre(template, body string) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/srv/install/compatibleTemplatePartitionSchemes?templateName="+template,
		httpmock.NewStringResponder(200, body))
}

func noRaidController() {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/srv/install/hardwareRaidProfile",
		httpmock.NewStringResponder(403, `{"message": "Hardware RAID is not supported by this server"}`))
}

// The whole point of the wizard: the second list is fetched for the answer just
// given. Asking for the schemes of no particular template returns something,
// which is what makes this worth a test — the request must carry the choice.
func TestWizardAsksTheSchemesOfTheChosenTemplate(t *testing.T) {
	operator, _ := withWizardAPI(t, "debian12_64")
	templatesAre(`{"ovh": ["debian12_64", "ubuntu2404-server_64"]}`)
	schemesAre("debian12_64", `["default"]`)
	noRaidController()

	choices, err := runReinstallWizard("srv")

	td.Require(t).CmpNoError(err)
	td.Cmp(t, choices.OperatingSystem, "debian12_64")
	td.Cmp(t, choices.SchemeName, "default")

	info := httpmock.GetCallCountInfo()
	td.Cmp(t, info["GET https://eu.api.ovh.com/v1/dedicated/server/srv/install/compatibleTemplatePartitionSchemes?templateName=debian12_64"], 1,
		"the schemes are asked for the template that was just chosen")
	td.Cmp(t, operator.options[0], td.ContainsKey("ubuntu2404-server_64"),
		"the list offered is the one the server can actually take")
}

// Measured on the eight servers of the test account: every template answers
// exactly one scheme, "default". A picker with one entry is a keystroke that
// carries no information — but the value still goes into the request, so it is
// stated rather than hidden.
func TestWizardDoesNotAskAQuestionWithOneAnswer(t *testing.T) {
	operator, logs := withWizardAPI(t, "debian12_64")
	templatesAre(`{"ovh": ["debian12_64"]}`)
	schemesAre("debian12_64", `["default"]`)
	noRaidController()

	choices, err := runReinstallWizard("srv")

	td.Require(t).CmpNoError(err)
	td.Cmp(t, len(operator.questions), 1, "only the OS is worth asking about here")
	td.Cmp(t, choices.SchemeName, "default")
	td.Cmp(t, logs.String(), td.Contains("Partitioning scheme: default"),
		"what was decided for the operator is still said out loud")
}

// Two schemes is a real choice, and it must be put.
func TestWizardAsksWhenThereIsSomethingToChoose(t *testing.T) {
	operator, _ := withWizardAPI(t, "debian12_64", "custom")
	templatesAre(`{"ovh": ["debian12_64"]}`)
	schemesAre("debian12_64", `["default", "custom"]`)
	noRaidController()

	choices, err := runReinstallWizard("srv")

	td.Require(t).CmpNoError(err)
	td.Cmp(t, len(operator.questions), 2)
	td.Cmp(t, operator.questions[1], td.Contains("partitioning scheme"))
	td.Cmp(t, choices.SchemeName, "custom")
}

// A personal template and an OVH one can carry the same name, and they install
// different systems. Merging the two lists would make the picker lie.
func TestWizardTellsPersonalTemplatesApart(t *testing.T) {
	operator, _ := withWizardAPI(t, "mine_64 (personal)")
	templatesAre(`{"ovh": ["debian12_64"], "personal": ["mine_64"]}`)
	schemesAre("mine_64", `[]`)
	noRaidController()

	choices, err := runReinstallWizard("srv")

	td.Require(t).CmpNoError(err)
	td.Cmp(t, operator.options[0], td.ContainsKey("mine_64 (personal)"))
	td.Cmp(t, choices.OperatingSystem, "mine_64",
		"the label is for reading; what is sent is the template name")
}

// No scheme at all is not the empty scheme: sending schemeName:"" would be a
// different request from sending none.
func TestWizardSendsNoSchemeWhenTheTemplateOffersNone(t *testing.T) {
	_, logs := withWizardAPI(t, "byolinux_64")
	templatesAre(`{"ovh": ["byolinux_64"]}`)
	schemesAre("byolinux_64", `[]`)
	noRaidController()

	choices, err := runReinstallWizard("srv")

	td.Require(t).CmpNoError(err)
	td.Cmp(t, choices.SchemeName, "")
	td.Cmp(t, logs.String(), td.Contains("no partitioning scheme"))
}

// The 403 is the answer 35 servers out of 35 give. It tells the operator where
// software RAID goes, and it must not stop a reinstall that has nothing to do
// with RAID.
func TestWizardReportsTheMissingRaidControllerWithoutFailing(t *testing.T) {
	_, logs := withWizardAPI(t, "debian12_64")
	templatesAre(`{"ovh": ["debian12_64"]}`)
	schemesAre("debian12_64", `["default"]`)
	noRaidController()

	_, err := runReinstallWizard("srv")

	td.Require(t).CmpNoError(err)
	td.Cmp(t, logs.String(), td.Contains("no hardware RAID controller"))
	td.Cmp(t, logs.String(), td.Contains("--from-file"))
}

// A controller exists, and the wizard still cannot size a configuration:
// install/hardwareRaidSize answers 404 on every server tried. Saying so beats
// writing a guess to the disks.
func TestWizardDoesNotInventAHardwareRaidBlock(t *testing.T) {
	_, logs := withWizardAPI(t, "debian12_64")
	templatesAre(`{"ovh": ["debian12_64"]}`)
	schemesAre("debian12_64", `["default"]`)
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/srv/install/hardwareRaidProfile",
		httpmock.NewStringResponder(200, `{"controllers": [{"model": "PERC H740P"}]}`))

	choices, err := runReinstallWizard("srv")

	td.Require(t).CmpNoError(err)
	td.Cmp(t, logs.String(), td.Contains("PERC H740P"))
	td.Cmp(t, logs.String(), td.Contains("does not build a hardwareRaid block"))
	td.Cmp(t, choices.SchemeName, "default", "the rest of the answers still stand")
}

// A discovery call that breaks must not stop a reinstall the operator has
// already decided on: the RAID profile is context, not a dependency.
func TestWizardSurvivesARaidProfileError(t *testing.T) {
	_, logs := withWizardAPI(t, "debian12_64")
	templatesAre(`{"ovh": ["debian12_64"]}`)
	schemesAre("debian12_64", `["default"]`)
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/dedicated/server/srv/install/hardwareRaidProfile",
		httpmock.NewStringResponder(500, `{"message": "internal error"}`))

	choices, err := runReinstallWizard("srv")

	td.Require(t).CmpNoError(err)
	td.Cmp(t, choices.OperatingSystem, "debian12_64")
	td.Cmp(t, logs.String(), td.Contains("does not depend on it"))
}

// Closing the picker without choosing is an answer: do nothing. Treating it as
// "no OS given" and carrying on would reinstall the server with the API's idea
// of a default.
func TestWizardStopsWhenNothingIsChosen(t *testing.T) {
	withWizardAPI(t, "")
	templatesAre(`{"ovh": ["debian12_64"]}`)

	_, err := runReinstallWizard("srv")

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains("nothing was done"))
}

// A server with nothing installable is a dead end, and saying so beats an empty
// list nobody can act on.
func TestWizardSaysWhenThereIsNothingToInstall(t *testing.T) {
	withWizardAPI(t)
	templatesAre(`{"ovh": []}`)

	_, err := runReinstallWizard("srv")

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains("no template"))
}

// And the guard itself: with nobody at the keyboard the wizard refuses instead
// of drawing an alternate screen over a pipeline's logs and waiting for an
// answer that cannot come. It names the two flags that replace it.
func TestWizardRefusesWithoutATerminal(t *testing.T) {
	operator, _ := withWizardAPI(t, "debian12_64")
	interactive = func() bool { return false }
	templatesAre(`{"ovh": ["debian12_64"]}`)

	_, err := runReinstallWizard("srv")

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains("no terminal"))
	td.Cmp(t, err.Error(), td.Contains("--from-file"))
	td.Cmp(t, len(operator.questions), 0, "it refuses before asking, not halfway through")
	td.Cmp(t, httpmock.GetTotalCallCount(), 0, "and before reaching the API")
}
