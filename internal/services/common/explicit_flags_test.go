// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/spf13/cobra"
)

type nestedSpec struct {
	RoutingAsDefault bool `json:"routingAsDefault,omitempty"`
}

type editSpec struct {
	Monitoring     bool       `json:"monitoring,omitempty"`
	NoIntervention bool       `json:"noIntervention,omitempty"`
	Retries        int        `json:"retries,omitempty"`
	Comment        string     `json:"comment,omitempty"`
	Network        nestedSpec `json:"network,omitempty"`
}

func newCommand(spec *editSpec) *cobra.Command {
	cmd := &cobra.Command{Use: "edit"}
	cmd.Flags().BoolVar(&spec.Monitoring, "monitoring", false, "")
	cmd.Flags().BoolVar(&spec.NoIntervention, "no-intervention", false, "")
	cmd.Flags().IntVar(&spec.Retries, "retries", 0, "")
	cmd.Flags().StringVar(&spec.Comment, "comment", "", "")
	// Deliberately unrelated flag name: matching by name would miss this one.
	cmd.Flags().BoolVar(&spec.Network.RoutingAsDefault, "default-route", false, "")

	return cmd
}

// A boolean explicitly set to false must reach the request body, otherwise the
// merge with the fetched resource silently restores the remote value.
func TestAddExplicitlySetFlags_FalseBooleanIsKept(t *testing.T) {
	var spec editSpec
	cmd := newCommand(&spec)
	td.Require(t).CmpNoError(cmd.Flags().Parse([]string{"--monitoring=false"}))

	body := map[string]any{}
	addExplicitlySetFlags(cmd, &spec, body)

	td.Cmp(t, body, map[string]any{"monitoring": false})
}

// Fields the user did not mention must stay out of the body: an edit is a
// partial update.
func TestAddExplicitlySetFlags_UntouchedFieldsAreLeftOut(t *testing.T) {
	var spec editSpec
	cmd := newCommand(&spec)
	td.Require(t).CmpNoError(cmd.Flags().Parse([]string{"--monitoring=false"}))

	body := map[string]any{}
	addExplicitlySetFlags(cmd, &spec, body)

	td.Cmp(t, body["noIntervention"], nil)
	td.Cmp(t, body["comment"], nil)
	td.Cmp(t, body["retries"], nil)
}

// Zero values of other kinds are dropped by omitempty just the same.
func TestAddExplicitlySetFlags_ZeroValuesOfEveryKind(t *testing.T) {
	var spec editSpec
	cmd := newCommand(&spec)
	td.Require(t).CmpNoError(cmd.Flags().Parse([]string{"--retries=0", "--comment="}))

	body := map[string]any{}
	addExplicitlySetFlags(cmd, &spec, body)

	td.Cmp(t, body["retries"], 0)
	td.Cmp(t, body["comment"], "")
}

// Flags are matched by address, so a flag whose name has nothing to do with
// the field it writes to is still restored — and lands at the right depth.
func TestAddExplicitlySetFlags_NestedFieldUnderUnrelatedFlagName(t *testing.T) {
	var spec editSpec
	cmd := newCommand(&spec)
	td.Require(t).CmpNoError(cmd.Flags().Parse([]string{"--default-route=false"}))

	body := map[string]any{}
	addExplicitlySetFlags(cmd, &spec, body)

	td.Cmp(t, body, map[string]any{"network": map[string]any{"routingAsDefault": false}})
}

// An existing value at the same path must not be clobbered by a sibling.
func TestAddExplicitlySetFlags_KeepsSiblingsOfNestedPath(t *testing.T) {
	var spec editSpec
	cmd := newCommand(&spec)
	td.Require(t).CmpNoError(cmd.Flags().Parse([]string{"--default-route=false"}))

	body := map[string]any{"network": map[string]any{"vlanId": float64(42)}}
	addExplicitlySetFlags(cmd, &spec, body)

	td.Cmp(t, body, map[string]any{
		"network": map[string]any{"vlanId": float64(42), "routingAsDefault": false},
	})
}

func TestAddExplicitlySetFlags_ToleratesNilInputs(t *testing.T) {
	var spec editSpec
	cmd := newCommand(&spec)

	addExplicitlySetFlags(nil, &spec, map[string]any{})
	addExplicitlySetFlags(cmd, nil, map[string]any{})
	addExplicitlySetFlags(cmd, &spec, nil)
	addExplicitlySetFlags(cmd, editSpec{}, map[string]any{}) // not a pointer
}
