// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/ovh/go-ovh/ovh"
	"github.com/spf13/cobra"
)

// Terminating a service is the same two-step everywhere in this API: a request
// that stops nothing and mails a token, then a confirmation that ends the
// contract. The survey fields are shared types — service.TerminationReasonEnum
// and service.TerminationFutureUseEnum — so the code that checks and previews
// them belongs next to the guard rather than in each service package. A copy
// of a routine that withholds a credential does not expire: it stays wrong in
// silence while the original is fixed.

// CompleteEnum offers the values a flag accepts on <tab>.
//
// The reader is passed in rather than the values themselves because every
// caller reads them from the specification embedded in the binary, behind a
// sync.OnceValues: registering flags happens on every invocation of the CLI,
// and only a completion or an actual termination needs the list.
func CompleteEnum(read func() ([]string, error)) ([]string, cobra.ShellCompDirective) {
	values, err := read()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return values, cobra.ShellCompDirectiveNoFileComp
}

// CheckEnumFlag rejects a value the API would reject, and names the ones it
// would accept. A 400 from the other side of the network says "invalid value"
// and stops there.
func CheckEnumFlag(name, value string, read func() ([]string, error)) error {
	if value == "" {
		return nil
	}

	accepted, err := read()
	if err != nil {
		return fmt.Errorf("failed to read the values accepted for --%s: %w", name, err)
	}

	if slices.Contains(accepted, value) {
		return nil
	}

	return fmt.Errorf("--%s does not accept %q; accepted values are: %s",
		name, value, strings.Join(accepted, ", "))
}

// Fingerprint identifies a credential without reproducing it. Withholding it
// entirely was the first answer, and it costs the one thing an operator
// legitimately checks in a preview: that the shell handed over the token they
// pasted, rather than one it truncated or expanded. Four characters and a
// length settle that question and reconstruct nothing.
func Fingerprint(token string) string {
	if len(token) < 8 {
		return fmt.Sprintf("(%d characters, too short to show)", len(token))
	}
	return fmt.Sprintf("%s… (%d characters)", token[:4], len(token))
}

// DescribeTerminationBody names every field a confirmation would send and
// withholds one value: the token is a single-use credential, and a --dry-run is
// the one command an operator runs with the output on screen or in a pipeline
// log.
func DescribeTerminationBody(body map[string]any) string {
	fields := make([]string, 0, len(body))
	for _, name := range []string{"token", "reason", "futureUse", "commentary"} {
		value, set := body[name]
		if !set {
			continue
		}
		if name == "token" {
			fields = append(fields, fmt.Sprintf("token: %s", Fingerprint(fmt.Sprint(value))))
			continue
		}
		fields = append(fields, fmt.Sprintf("%s: %v", name, value))
	}
	return strings.Join(fields, ", ")
}

// IsNotFound tells the API saying "this does not exist" apart from the API
// failing to answer.
//
// The difference decides what a command does next, and getting it wrong is not
// theoretical: treating every failure as a 404 has already sent a create where
// an update was due, and announced "no migration token exists" after a 403.
func IsNotFound(err error) bool {
	var apiErr *ovh.APIError
	return errors.As(err, &apiErr) && apiErr.Code == http.StatusNotFound
}
