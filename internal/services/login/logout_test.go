// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"gopkg.in/ini.v1"
)

func TestIsInvalidCredentialError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"forbidden (403)", &ovh.APIError{Code: http.StatusForbidden}, true},
		{"unauthorized (401)", &ovh.APIError{Code: http.StatusUnauthorized}, true},
		{"other API error (500)", &ovh.APIError{Code: http.StatusInternalServerError}, false},
		{"not found (404)", &ovh.APIError{Code: http.StatusNotFound}, false},
		{"wrapped forbidden", fmt.Errorf("revoke failed: %w", &ovh.APIError{Code: http.StatusForbidden}), true},
		{"plain error", errors.New("network unreachable"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			td.Cmp(t, isInvalidCredentialError(tc.err), tc.expected)
		})
	}
}

// exitSentinel stands in for os.Exit so a test can tell "the command stopped
// here" from "the command carried on".
type exitSentinel int

// runLogoutWithRealExit runs Logout with display.ExitFunc panicking instead of
// exiting, and reports whether it stopped early.
//
// The suite-wide stub replaces ExitFunc with a no-op, which makes every
// terminating output look like a plain print — and that stub is exactly what hid
// this defect for as long as it existed. So this test does not use it.
func runLogoutWithRealExit(t *testing.T) (stopped bool) {
	t.Helper()
	saved := display.ExitFunc
	t.Cleanup(func() { display.ExitFunc = saved })
	display.ExitFunc = func(int) { panic(exitSentinel(0)) }

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(exitSentinel); !ok {
				panic(r)
			}
			stopped = true
		}
	}()

	Logout(nil, nil)

	return false
}

// The credentials on this disk are what `logout` exists to remove. The remote
// revocation is a courtesy, and it used to be able to cancel the command: each of
// its three unhappy paths called display.OutputWarning, which does not return —
// OutputWithFormat finishes on ExitFunc(0), which is os.Exit. So a revoked key
// produced "🟠 credentials were already invalid or revoked, skipping remote
// revocation", exit 0, and the key still in the file.
//
// Reproduced against the built binary before this fix, with a bogus key in
// ./ovh.conf: exit 0 and consumer_key untouched. Verified after: removed.
func TestLogoutRemovesTheKeyEvenWhenRevocationFails(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
	}{
		{"already revoked", http.StatusForbidden},
		{"unauthorized", http.StatusUnauthorized},
		{"API failing for another reason", http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "ovh.conf")
			td.Require(t).CmpNoError(os.WriteFile(path, []byte(
				"[default]\nendpoint=ovh-eu\n\n[ovh-eu]\napplication_key=k\napplication_secret=s\nconsumer_key=SENTINEL\n"), 0o600))

			cfg, err := ini.Load(path)
			td.Require(t).CmpNoError(err)

			httpmock.Activate(t)
			client, err := ovh.NewClient("ovh-eu", "k", "s", "c")
			td.Require(t).CmpNoError(err)
			httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
				httpmock.NewStringResponder(200, "0"))
			httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/1.0/auth/logout",
				httpmock.NewStringResponder(tc.status, `{"message":"nope"}`))

			savedClient, savedCfg, savedPath, savedYes := httpLib.Client, flags.CliConfig, flags.CliConfigPath, LogoutAssumeYes
			httpLib.Client, flags.CliConfig, flags.CliConfigPath, LogoutAssumeYes = client, cfg, path, true
			t.Cleanup(func() {
				httpLib.Client, flags.CliConfig, flags.CliConfigPath, LogoutAssumeYes = savedClient, savedCfg, savedPath, savedYes
			})

			stopped := runLogoutWithRealExit(t)

			td.Cmp(t, stopped, false, "the command must not stop before removing the key")

			after, err := os.ReadFile(path)
			td.Require(t).CmpNoError(err)
			td.Cmp(t, strings.Contains(string(after), "SENTINEL"), false,
				"the credential is still on disk: %s", after)
		})
	}
}
