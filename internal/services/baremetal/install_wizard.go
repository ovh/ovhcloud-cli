// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"slices"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/utils"
)

// A reinstall is decided by answers that depend on each other, and the CLI made
// the operator discover that dependency the hard way.
//
// `list-compatible-os` and `list-partition-schemes` each answer one question,
// but the second cannot be asked before the first: the schemes are a property
// of the template. So the sequence was: run one command, read the answer, copy
// a name into the next command, read that answer, copy two names into a third.
// Four steps, three of them clerical, to find out what a machine will accept.
// And it has to happen BEFORE the reinstall: finding out at install time means
// the disks are already wiped.
//
// --wizard asks the questions in the only order the API allows, with each list
// fetched for the answer just given.

// ReinstallViaWizard is set by --wizard.
var ReinstallViaWizard bool

// choicePicker is the interactive prompt, indirected so the tests can drive the
// wizard without a terminal. Every other selector in this CLI calls
// display.RunGenericChoicePicker directly and is therefore untestable; the cost
// of testing this one is a single package-level variable.
var choicePicker = display.RunGenericChoicePicker

// interactive reports whether there is somebody at the keyboard, indirected for
// the same reason: `go test` has no terminal, so a wizard that asked
// utils.IsInteractiveTerminal directly could only ever be tested refusing.
var interactive = utils.IsInteractiveTerminal

// reinstallChoices is what the wizard settles, in the shape the reinstall body
// wants it.
type reinstallChoices struct {
	OperatingSystem string
	SchemeName      string
}

// fetchCompatibleTemplates lists the templates this server can take.
//
// The answer has two lists, and they are not interchangeable: `personal` holds
// the account's own templates, which are the ones somebody deliberately built.
// They are labelled as such rather than merged, because a personal template and
// an OVH one can carry the same name and picking the wrong one installs the
// wrong system.
func fetchCompatibleTemplates(server string) (map[string]string, error) {
	path := fmt.Sprintf("/v1/dedicated/server/%s/install/compatibleTemplates",
		url.PathEscape(server))

	var answer struct {
		Ovh      []string `json:"ovh"`
		Personal []string `json:"personal"`
	}
	if err := httpLib.Client.Get(path, &answer); err != nil {
		return nil, fmt.Errorf("failed to fetch the templates %s can take: %w", server, err)
	}

	choices := make(map[string]string, len(answer.Ovh)+len(answer.Personal))
	for _, name := range answer.Ovh {
		choices[name] = name
	}
	for _, name := range answer.Personal {
		choices[name+" (personal)"] = name
	}

	if len(choices) == 0 {
		return nil, fmt.Errorf("no template can be installed on %s", server)
	}

	return choices, nil
}

// fetchPartitionSchemes lists the schemes a template allows on this server.
func fetchPartitionSchemes(server, template string) ([]string, error) {
	path := fmt.Sprintf("/v1/dedicated/server/%s/install/compatibleTemplatePartitionSchemes?templateName=%s",
		url.PathEscape(server), url.QueryEscape(template))

	var schemes []string
	if err := httpLib.Client.Get(path, &schemes); err != nil {
		return nil, fmt.Errorf("failed to fetch the partition schemes of %s for %s: %w",
			server, template, err)
	}

	return schemes, nil
}

// reportRaidCapability says what the machine will accept in a hardwareRaid
// block, and never stops the wizard.
//
// It is a statement, not a question, because there is nothing to choose: with
// no controller the block is illegal, and with one the wizard still has no way
// to size a configuration — install/hardwareRaidSize is in the schema, badged
// "Stable production version", and answers 404 on every server tried (see
// install.go). Inventing a layout from a controller model would be a guess
// written to disk.
//
// A discovery call that fails must not block a reinstall the operator has
// already decided on, so this one only ever prints.
func reportRaidCapability(server string) {
	path := fmt.Sprintf("/v1/dedicated/server/%s/install/hardwareRaidProfile",
		url.PathEscape(server))

	var profile struct {
		Controllers []struct {
			Model string `json:"model"`
		} `json:"controllers"`
	}

	if err := httpLib.Client.Get(path, &profile); err != nil {
		if isUnsupportedHardwareRaid(err) {
			log.Printf("ℹ️  %s has no hardware RAID controller: software RAID goes in the partitioning layout, via --from-file.", server)
			return
		}
		log.Printf("ℹ️  Could not read the RAID profile of %s (%s); the reinstall does not depend on it.", server, err)
		return
	}

	if len(profile.Controllers) == 0 {
		return
	}

	models := make([]string, 0, len(profile.Controllers))
	for _, controller := range profile.Controllers {
		models = append(models, controller.Model)
	}
	log.Printf("ℹ️  %s has a hardware RAID controller (%v). This wizard does not build a hardwareRaid block: describe it in a file and pass --from-file.",
		server, models)
}

// runReinstallWizard asks what has to be known before the disks are wiped.
func runReinstallWizard(server string) (reinstallChoices, error) {
	// A picker with nobody at the keyboard would draw an alternate screen over
	// a pipeline's logs and wait forever. Refusing here, with the flags that
	// replace it, is the same rule --yes follows for the confirmation.
	if !interactive() {
		return reinstallChoices{}, errors.New(
			"--wizard has questions to ask and there is no terminal to ask them on; " +
				"pass --os, and --from-file for the storage layout")
	}

	templates, err := fetchCompatibleTemplates(server)
	if err != nil {
		return reinstallChoices{}, err
	}

	_, template, err := choicePicker(
		fmt.Sprintf("Which OS should %s be reinstalled with?", server), templates, 20)
	if err != nil {
		return reinstallChoices{}, err
	}
	if template == "" {
		return reinstallChoices{}, errors.New("no operating system chosen, so nothing was done")
	}

	schemes, err := fetchPartitionSchemes(server, template)
	if err != nil {
		return reinstallChoices{}, err
	}

	choices := reinstallChoices{OperatingSystem: template}

	switch len(schemes) {
	case 0:
		// The template imposes its own layout. Sending an empty schemeName
		// would be a different request from sending none.
		log.Printf("ℹ️  %s offers no partitioning scheme to choose from; its default layout is used.", template)
	case 1:
		// A question with one possible answer is not a question. Measured on
		// this account: every one of the eight servers checked answers exactly
		// one scheme, "default", for the same template. Asking would add a
		// keystroke and no information — but the answer is still printed,
		// because it goes into the request.
		choices.SchemeName = schemes[0]
		log.Printf("ℹ️  Partitioning scheme: %s — the only one %s allows here.", schemes[0], template)
	default:
		options := make(map[string]string, len(schemes))
		for _, scheme := range slices.Sorted(slices.Values(schemes)) {
			options[scheme] = scheme
		}
		_, scheme, err := choicePicker(
			fmt.Sprintf("Which partitioning scheme, of the %d %s allows?", len(schemes), template),
			options, 10)
		if err != nil {
			return reinstallChoices{}, err
		}
		if scheme == "" {
			return reinstallChoices{}, errors.New("no partitioning scheme chosen, so nothing was done")
		}
		choices.SchemeName = scheme
	}

	reportRaidCapability(server)

	return choices, nil
}

// applyWizardChoices runs the wizard and reports what the reinstall should
// carry.
//
// It exists as its own function so the flag has one meaning in one place: the
// answers overwrite --os, because somebody who both passed a flag and asked to
// be walked through the choices meant the choice they just made.
func applyWizardChoices(server string) error {
	answers, err := runReinstallWizard(server)
	if err != nil {
		return err
	}

	OperatingSystem = answers.OperatingSystem
	if answers.SchemeName != "" {
		WizardStorage = []reinstallStorage{{
			Partitioning: reinstallPartitioning{SchemeName: answers.SchemeName},
		}}
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil,
		"Reinstalling %s with %s%s.", server, answers.OperatingSystem,
		schemeSuffix(answers.SchemeName))

	return nil
}

func schemeSuffix(scheme string) string {
	if scheme == "" {
		return ""
	}
	return ", partitioning scheme " + scheme
}
