// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"path/filepath"
	"testing"

	"github.com/maxatome/go-testdeep/td"
	baseconfig "github.com/ovh/ovhcloud-cli/internal/config"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"gopkg.in/ini.v1"
)

// withTestConfig points flags.CliConfig/CliConfigPath at a throwaway ini
// file for the duration of the test, and stubs display.ExitFunc so that
// display.OutputError (called on validation failure) doesn't os.Exit the
// test binary. Everything is restored afterwards.
func withTestConfig(t *testing.T) {
	t.Helper()

	prevConfig, prevPath, prevExitFunc := flags.CliConfig, flags.CliConfigPath, display.ExitFunc
	flags.CliConfig = ini.Empty()
	flags.CliConfigPath = filepath.Join(t.TempDir(), "test.conf")
	display.ExitFunc = func(int) {}

	t.Cleanup(func() {
		flags.CliConfig, flags.CliConfigPath, display.ExitFunc = prevConfig, prevPath, prevExitFunc
	})
}

func TestSetEndpoint_CustomURL(t *testing.T) {
	withTestConfig(t)

	// Regression test: a fully qualified custom endpoint URL must be
	// accepted, not rejected as having an "invalid scheme".
	SetEndpoint(nil, []string{"https://api.eu.ovhcloud.com/1.0"})

	val, err := flags.CliConfig.Section("default").GetKey("endpoint")
	td.Require(t).CmpNoError(err)
	td.Cmp(t, val.String(), "https://api.eu.ovhcloud.com/1.0")
}

func TestSetEndpoint_Region(t *testing.T) {
	withTestConfig(t)

	SetEndpoint(nil, []string{"EU"})

	val, err := flags.CliConfig.Section("default").GetKey("endpoint")
	td.Require(t).CmpNoError(err)
	td.Cmp(t, val.String(), "ovh-eu")
}

func TestSetEndpoint_InvalidScheme(t *testing.T) {
	withTestConfig(t)

	// A value that isn't one of the known region names and isn't a
	// valid http(s) URL (e.g. the menu label "Custom endpoint" that
	// used to leak into this call from the login flow) must not be stored.
	SetEndpoint(nil, []string{"Custom endpoint"})

	_, err := flags.CliConfig.Section("default").GetKey("endpoint")
	td.CmpError(t, err)
}

func TestSetHeader_LegacyMode(t *testing.T) {
	withTestConfig(t)

	SetEndpoint(nil, []string{"https://api.eu.ovhcloud.com/1.0"})
	SetHeader(nil, []string{"X-Routing-Key", "internal-build-eu"})

	val, err := flags.CliConfig.Section("https://api.eu.ovhcloud.com/1.0").GetKey("header.X-Routing-Key")
	td.Require(t).CmpNoError(err)
	td.Cmp(t, val.String(), "internal-build-eu")
}

func TestSetHeader_LegacyMode_NoEndpoint(t *testing.T) {
	withTestConfig(t)

	// No endpoint configured yet: setting a header should fail cleanly,
	// not write anything.
	SetHeader(nil, []string{"X-Routing-Key", "internal-build-eu"})

	val, err := baseconfig.GetConfigValue(flags.CliConfig, "default", "endpoint")
	td.Require(t).CmpNoError(err)
	td.Cmp(t, val, "")
}

func TestUnsetHeader_LegacyMode(t *testing.T) {
	withTestConfig(t)

	SetEndpoint(nil, []string{"https://api.eu.ovhcloud.com/1.0"})
	SetHeader(nil, []string{"X-Routing-Key", "internal-build-eu"})
	UnsetHeader(nil, []string{"X-Routing-Key"})

	td.CmpFalse(t, flags.CliConfig.Section("https://api.eu.ovhcloud.com/1.0").HasKey("header.X-Routing-Key"))
}

func TestSetHeader_ProfileMode(t *testing.T) {
	withTestConfig(t)
	flags.CliConfig.Section("default").Key("profile").SetValue("work")
	flags.CliConfig.Section("profile:work").Key("endpoint").SetValue("ovh-eu")

	SetHeader(nil, []string{"X-Routing-Key", "internal-build-eu"})

	val, err := flags.CliConfig.Section("profile:work").GetKey("header.X-Routing-Key")
	td.Require(t).CmpNoError(err)
	td.Cmp(t, val.String(), "internal-build-eu")
}

func TestUnsetHeader_ProfileMode(t *testing.T) {
	withTestConfig(t)
	flags.CliConfig.Section("default").Key("profile").SetValue("work")
	flags.CliConfig.Section("profile:work").Key("endpoint").SetValue("ovh-eu")

	SetHeader(nil, []string{"X-Routing-Key", "internal-build-eu"})
	UnsetHeader(nil, []string{"X-Routing-Key"})

	td.CmpFalse(t, flags.CliConfig.Section("profile:work").HasKey("header.X-Routing-Key"))
}
