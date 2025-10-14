// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/helpers/tdsuite"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
	httplib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type MockSuite struct{}

func (ms *MockSuite) Setup(t *td.T) error {
	httpmock.Activate(t)

	oldStdOut := os.Stdout
	t.Cleanup(func() { os.Stdout = oldStdOut })
	os.Stdout = nil

	client, err := ovh.NewClient("ovh-eu", "app_key", "app_secret", "consumer_key")
	if err != nil {
		return err
	}

	httplib.Client = client

	return nil
}

func (ms *MockSuite) PreTest(t *td.T, testName string) error {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
		httpmock.NewStringResponder(200, "0"))

	return nil
}

func (ms *MockSuite) PostTest(_ *td.T, _ string) error {
	httpmock.Reset()
	cmd.PostExecute()
	resetSubCommandFlagValues(cmd.GetRootCommand())
	return nil
}

func TestMockSuite(t *testing.T) {
	tdsuite.Run(t, &MockSuite{})
}

// resetSubCommandFlagValues resets all flags of all subcommands of the given root command to their default values.
func resetSubCommandFlagValues(root *cobra.Command) {
	for _, c := range root.Commands() {
		c.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Changed {
				if f.Value.Type() == "stringArray" {
					// Special handling for stringArray for which we cannot
					// use DefValue since it is equal to "[]".
					if r, ok := f.Value.(pflag.SliceValue); ok {
						r.Replace(nil)
					}
				} else {
					f.Value.Set(f.DefValue)
				}
				f.Changed = false
			}
		})
		resetSubCommandFlagValues(c)
	}
}

func cleanWhitespacesHelper(s string) string {
	return regexp.MustCompile(` +\n`).ReplaceAllString(s, "\n")
}
