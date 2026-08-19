// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"strings"
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/spf13/cobra"
)

func TestMarkAsInputFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "", "Name flag")

	markAsInputFlag(cmd, "name")

	f := cmd.Flags().Lookup("name")
	if f == nil {
		t.Fatal("flag 'name' not found")
	}
	if _, ok := f.Annotations[inputFlagAnnotation]; !ok {
		t.Error("expected flag to have input annotation")
	}
}

func TestMarkAsInputFlag_NonExistentFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	// Should not panic
	markAsInputFlag(cmd, "nonexistent")
}

func TestFlagUsagesByAnnotation(t *testing.T) {
	cmd := &cobra.Command{Use: "test", Run: func(*cobra.Command, []string) {}}
	cmd.Flags().String("name", "", "Name of the resource")
	cmd.Flags().String("from-file", "", "File containing parameters")
	markAsInputFlag(cmd, "from-file")

	// Non-input flags should include "name" but not "from-file"
	nonInput := flagUsagesByAnnotation(cmd, false)
	if !strings.Contains(nonInput, "--name") {
		t.Error("expected non-input flags to contain --name")
	}
	if strings.Contains(nonInput, "--from-file") {
		t.Error("expected non-input flags to NOT contain --from-file")
	}

	// Input flags should include "from-file" but not "name"
	input := flagUsagesByAnnotation(cmd, true)
	if !strings.Contains(input, "--from-file") {
		t.Error("expected input flags to contain --from-file")
	}
	if strings.Contains(input, "--name") {
		t.Error("expected input flags to NOT contain --name")
	}
}

func TestFlagUsagesByAnnotation_HiddenFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test", Run: func(*cobra.Command, []string) {}}
	cmd.Flags().String("from-file", "", "File containing parameters")
	markAsInputFlag(cmd, "from-file")

	f := cmd.Flags().Lookup("from-file")
	f.Hidden = true

	input := flagUsagesByAnnotation(cmd, true)
	if strings.Contains(input, "--from-file") {
		t.Error("expected hidden input flag to be excluded")
	}
}

func TestCreateCmdUsageTemplate_SeparatesSections(t *testing.T) {
	cmd := &cobra.Command{Use: "create", Short: "Create a resource", Run: func(*cobra.Command, []string) {}}
	cmd.Flags().String("name", "", "Name of the resource")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("from-file", "", "File containing parameters")
	cmd.Flags().Bool("editor", false, "Use a text editor to define parameters")

	markAsInputFlag(cmd, "from-file")
	markAsInputFlag(cmd, "editor")
	applyInputFlagsTemplate(cmd)

	usage := cmd.UsageString()

	// Should contain both sections
	if !strings.Contains(usage, "\nFlags:\n") {
		t.Error("expected usage to contain 'Flags:' section")
	}
	if !strings.Contains(usage, "\nInput Flags:\n") {
		t.Error("expected usage to contain 'Input Flags:' section")
	}

	// The "Flags:" section should appear before "Input Flags:"
	flagsIdx := strings.Index(usage, "\nFlags:\n")
	inputFlagsIdx := strings.Index(usage, "\nInput Flags:\n")
	if flagsIdx >= inputFlagsIdx {
		t.Error("expected 'Flags:' to appear before 'Input Flags:'")
	}

	// Extract the Flags section content (between "Flags:" and "Input Flags:")
	flagsSection := usage[flagsIdx:inputFlagsIdx]
	if !strings.Contains(flagsSection, "--name") {
		t.Error("expected 'Flags:' section to contain --name")
	}
	if !strings.Contains(flagsSection, "--description") {
		t.Error("expected 'Flags:' section to contain --description")
	}
	if strings.Contains(flagsSection, "--from-file") {
		t.Error("expected 'Flags:' section to NOT contain --from-file")
	}
	if strings.Contains(flagsSection, "--editor") {
		t.Error("expected 'Flags:' section to NOT contain --editor")
	}

	// Extract the Input Flags section content (after "Input Flags:")
	inputSection := usage[inputFlagsIdx:]
	if !strings.Contains(inputSection, "--from-file") {
		t.Error("expected 'Input Flags:' section to contain --from-file")
	}
	if !strings.Contains(inputSection, "--editor") {
		t.Error("expected 'Input Flags:' section to contain --editor")
	}
	if strings.Contains(inputSection, "--name") {
		t.Error("expected 'Input Flags:' section to NOT contain --name")
	}
}

func TestCreateCmdUsageTemplate_NoInputFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "list", Short: "List resources", Run: func(*cobra.Command, []string) {}}
	cmd.Flags().String("filter", "", "Filter results")
	applyInputFlagsTemplate(cmd)

	usage := cmd.UsageString()

	// Should have Flags section but no Input Flags section
	if !strings.Contains(usage, "\nFlags:\n") {
		t.Error("expected usage to contain 'Flags:' section")
	}
	if strings.Contains(usage, "Input Flags:") {
		t.Error("expected usage to NOT contain 'Input Flags:' when no input flags exist")
	}
	if !strings.Contains(usage, "--filter") {
		t.Error("expected --filter to appear in Flags section")
	}
}

func TestAddParameterFileFlags_FromFileFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "create", Run: func(*cobra.Command, []string) {}}

	addParameterFileFlags(cmd, true, nil, "", "", "", nil)

	f := cmd.Flags().Lookup("from-file")
	if f == nil {
		t.Fatal("expected --from-file flag to exist")
	}
	if _, ok := f.Annotations[inputFlagAnnotation]; !ok {
		t.Error("expected --from-file to have input annotation")
	}
}

func TestAddInteractiveEditorFlag_AnnotatesAndAppliesTemplate(t *testing.T) {
	cmd := &cobra.Command{Use: "create", Run: func(*cobra.Command, []string) {}}

	addInteractiveEditorFlag(cmd)

	f := cmd.Flags().Lookup("editor")
	if f == nil {
		t.Fatal("expected --editor flag to exist")
	}
	if _, ok := f.Annotations[inputFlagAnnotation]; !ok {
		t.Error("expected --editor to have input annotation")
	}
}

func TestAddParameterFileFlags_InitFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "create", Run: func(*cobra.Command, []string) {}}

	addParameterFileFlags(cmd, false, nil, "/test", "post", "{}", nil)

	for _, name := range []string{"init-file", "replace"} {
		f := cmd.Flags().Lookup(name)
		if f == nil {
			t.Fatalf("expected --%s flag to exist", name)
		}
		if _, ok := f.Annotations[inputFlagAnnotation]; !ok {
			t.Errorf("expected --%s to have input annotation", name)
		}
	}
}

func TestAllInputFlagsHidden_NoInputFlagsSection(t *testing.T) {
	cmd := &cobra.Command{Use: "create", Run: func(*cobra.Command, []string) {}}
	cmd.Flags().String("name", "", "Name")
	cmd.Flags().String("from-file", "", "File containing parameters")
	cmd.Flags().Bool("editor", false, "Use editor")

	markAsInputFlag(cmd, "from-file")
	markAsInputFlag(cmd, "editor")
	applyInputFlagsTemplate(cmd)

	// Hide all input flags (simulates WASM mode)
	cmd.Flags().Lookup("from-file").Hidden = true
	cmd.Flags().Lookup("editor").Hidden = true

	usage := cmd.UsageString()

	if strings.Contains(usage, "Input Flags:") {
		t.Error("expected no 'Input Flags:' section when all input flags are hidden")
	}
}

// --init-file is how somebody writing a script discovers the shape of a request
// body, and it could not run in a script: it always opened a picker, which
// needs a terminal, so it failed with "could not open a new TTY" and wrote
// nothing. Outside a terminal there is nobody to ask, so one example is taken —
// and which one is taken has to be decided rather than left to a map.
func TestExampleForUnattendedRunPrefersTheOneThisCLIShips(t *testing.T) {
	assert := td.Assert(t)

	// "default" is the sample shipped with the command. For a reinstall it is
	// the only one carrying a storage block, which is the whole reason a script
	// asks for this file.
	withDefault := map[string]string{
		"Example 1 - Simple Linux Installation": `{"operatingSystem": "debian12_64"}`,
		"Example 4 - BringYourOwnLinux":         `{"operatingSystem": "byolinux_64"}`,
		"default":                               `{"storage": [{"partitioning": {}}]}`,
	}
	assert.Cmp(exampleForUnattendedRun(withDefault), "default")
}

// examples is a map, so ranging over it picks a different entry each run. A
// parameter file that changes under a script is worse than one it did not get
// to choose.
func TestExampleForUnattendedRunIsStableAcrossRuns(t *testing.T) {
	assert := td.Assert(t)

	schemaOnly := map[string]string{
		"Example 3 - Windows": `{"operatingSystem": "win2022"}`,
		"Example 1 - Linux":   `{"operatingSystem": "debian12_64"}`,
		"Example 2 - BSD":     `{"operatingSystem": "freebsd13"}`,
	}

	first := exampleForUnattendedRun(schemaOnly)
	for range 20 {
		assert.Cmp(exampleForUnattendedRun(schemaOnly), first)
	}
	assert.Cmp(first, "Example 1 - Linux", "and the choice is the sorted first, not an arbitrary one")

	assert.Cmp(exampleForUnattendedRun(nil), "", "no example is not a crash")
}
