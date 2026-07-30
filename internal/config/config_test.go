// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"gopkg.in/ini.v1"
)

func newTestConfig(content string) *ini.File {
	cfg, err := ini.Load([]byte(content))
	if err != nil {
		panic(err)
	}
	return cfg
}

func TestGetActiveProfileName_Legacy(t *testing.T) {
	cfg := newTestConfig(`
[default]
endpoint=ovh-eu

[ovh-eu]
application_key=test
`)

	name := GetActiveProfileName(cfg, "")
	td.Cmp(t, name, "")
}

func TestGetActiveProfileName_FromConfig(t *testing.T) {
	cfg := newTestConfig(`
[default]
profile=work

[profile:work]
endpoint=ovh-eu
application_key=test
`)

	name := GetActiveProfileName(cfg, "")
	td.Cmp(t, name, "work")
}

func TestGetActiveProfileName_FlagOverride(t *testing.T) {
	cfg := newTestConfig(`
[default]
profile=work
`)

	name := GetActiveProfileName(cfg, "personal")
	td.Cmp(t, name, "personal")
}

func TestGetActiveProfileName_EnvVar(t *testing.T) {
	cfg := newTestConfig(`
[default]
profile=work
`)

	os.Setenv("OVH_PROFILE", "fromenv")
	defer os.Unsetenv("OVH_PROFILE")

	// Env should override config
	name := GetActiveProfileName(cfg, "")
	td.Cmp(t, name, "fromenv")

	// Flag should override env
	name = GetActiveProfileName(cfg, "fromflag")
	td.Cmp(t, name, "fromflag")
}

func TestIsProfileMode(t *testing.T) {
	legacyCfg := newTestConfig(`
[default]
endpoint=ovh-eu
`)
	profileCfg := newTestConfig(`
[default]
profile=work
`)

	td.CmpFalse(t, IsProfileMode(legacyCfg, ""))
	td.CmpTrue(t, IsProfileMode(profileCfg, ""))
	td.CmpTrue(t, IsProfileMode(legacyCfg, "override"))
}

func TestGetProfileCredentials(t *testing.T) {
	cfg := newTestConfig(`
[profile:work]
endpoint=ovh-eu
application_key=ak123
application_secret=as456
consumer_key=ck789
`)

	endpoint, appKey, appSecret, consumerKey, err := GetProfileCredentials(cfg, "work")
	td.Require(t).CmpNoError(err)
	td.Cmp(t, endpoint, "ovh-eu")
	td.Cmp(t, appKey, "ak123")
	td.Cmp(t, appSecret, "as456")
	td.Cmp(t, consumerKey, "ck789")
}

func TestGetProfileCredentials_NotFound(t *testing.T) {
	cfg := newTestConfig(`
[default]
endpoint=ovh-eu
`)

	_, _, _, _, err := GetProfileCredentials(cfg, "nonexistent")
	td.CmpError(t, err)
}

func TestListProfiles(t *testing.T) {
	cfg := newTestConfig(`
[default]
profile=work

[profile:work]
endpoint=ovh-eu

[profile:personal]
endpoint=ovh-ca

[ovh-eu]
application_key=legacy
`)

	profiles := ListProfiles(cfg)
	td.Cmp(t, profiles, td.Bag("work", "personal"))
}

func TestSetActiveProfile(t *testing.T) {
	cfg := newTestConfig(`
[default]
endpoint=ovh-eu

[profile:work]
endpoint=ovh-eu
`)

	tmpFile := filepath.Join(t.TempDir(), "test.conf")

	td.Require(t).CmpNoError(SetActiveProfile(cfg, tmpFile, "work"))
	td.Cmp(t, GetActiveProfileName(cfg, ""), "work")
}

func TestDeleteProfile(t *testing.T) {
	cfg := newTestConfig(`
[default]
profile=work

[profile:work]
endpoint=ovh-eu
application_key=test

[profile:personal]
endpoint=ovh-ca
`)

	tmpFile := filepath.Join(t.TempDir(), "test.conf")

	// Delete the active profile
	td.Require(t).CmpNoError(DeleteProfile(cfg, tmpFile, "work"))

	// Active profile should be cleared
	td.Cmp(t, GetActiveProfileName(cfg, ""), "")

	// Only personal should remain
	td.Cmp(t, ListProfiles(cfg), td.Bag("personal"))
}

func TestDeleteProfile_NotFound(t *testing.T) {
	cfg := newTestConfig(`
[default]
endpoint=ovh-eu
`)

	tmpFile := filepath.Join(t.TempDir(), "test.conf")

	td.CmpError(t, DeleteProfile(cfg, tmpFile, "nonexistent"))
}

func TestSetProfileConfigValue(t *testing.T) {
	cfg := newTestConfig(`
[profile:work]
endpoint=ovh-eu
`)

	tmpFile := filepath.Join(t.TempDir(), "test.conf")

	td.Require(t).CmpNoError(SetProfileConfigValue(cfg, tmpFile, "work", "default_cloud_project", "proj123"))
	td.Cmp(t, GetProfileConfigValue(cfg, "work", "default_cloud_project"), "proj123")
}

func TestSetProfileConfigValue_NewProfile(t *testing.T) {
	cfg := ini.Empty()
	tmpFile := filepath.Join(t.TempDir(), "test.conf")

	td.Require(t).CmpNoError(SetProfileConfigValue(cfg, tmpFile, "newprofile", "endpoint", "ovh-eu"))
	td.Cmp(t, ListProfiles(cfg), td.Bag("newprofile"))
}

func TestGetConfigValue_ProfileMode(t *testing.T) {
	cfg := newTestConfig(`
[default]
profile=work

[profile:work]
default_cloud_project=proj-from-profile

[ovh-cli]
default_cloud_project=proj-from-legacy
`)

	val, err := GetConfigValue(cfg, "", "default_cloud_project")
	td.Require(t).CmpNoError(err)
	td.Cmp(t, val, "proj-from-profile")
}

func TestGetConfigValue_ProfileMode_NoFallback(t *testing.T) {
	// When a named profile is active but doesn't have a key,
	// it should NOT fall back to the legacy section
	cfg := newTestConfig(`
[default]
profile=work

[profile:work]
endpoint=ovh-eu

[ovh-cli]
default_cloud_project=proj-from-legacy
`)

	val, err := GetConfigValue(cfg, "", "default_cloud_project")
	td.Require(t).CmpNoError(err)
	td.Cmp(t, val, "")
}

func TestGetConfigValue_LegacyMode(t *testing.T) {
	cfg := newTestConfig(`
[default]
endpoint=ovh-eu

[ovh-cli]
default_cloud_project=proj-from-legacy
`)

	val, err := GetConfigValue(cfg, "", "default_cloud_project")
	td.Require(t).CmpNoError(err)
	td.Cmp(t, val, "proj-from-legacy")
}

func TestIsDefaultProfile(t *testing.T) {
	td.CmpTrue(t, IsDefaultProfile("default"))
	td.CmpFalse(t, IsDefaultProfile("work"))
	td.CmpFalse(t, IsDefaultProfile(""))
}

func TestGetConfigValue_DefaultProfileFallsToLegacy(t *testing.T) {
	// When profile is set to "default", GetConfigValue should read from legacy sections
	cfg := newTestConfig(`
[default]
profile=default

[ovh-cli]
default_cloud_project=proj-from-legacy
`)

	val, err := GetConfigValue(cfg, "", "default_cloud_project")
	td.Require(t).CmpNoError(err)
	td.Cmp(t, val, "proj-from-legacy")
}

func TestListProfiles_DoesNotIncludeDefault(t *testing.T) {
	// ListProfiles only returns named profiles from [profile:*] sections,
	// not the virtual "default" profile
	cfg := newTestConfig(`
[default]
endpoint=ovh-eu

[profile:work]
endpoint=ovh-eu
`)

	profiles := ListProfiles(cfg)
	td.Cmp(t, profiles, td.Bag("work"))
	td.CmpNot(t, profiles, td.Contains(DefaultProfileName))
}

func TestGetCustomHeaders(t *testing.T) {
	cfg := newTestConfig(`
[https://api.eu.ovhcloud.com/1.0]
application_key      = ak123
header.X-Routing-Key  = internal-build-eu
header.X-Debug-Bypass = true
`)

	headers := GetCustomHeaders(cfg, "https://api.eu.ovhcloud.com/1.0")
	td.Cmp(t, headers, map[string]string{
		"X-Routing-Key":  "internal-build-eu",
		"X-Debug-Bypass": "true",
	})
}

func TestGetCustomHeaders_NoHeaders(t *testing.T) {
	cfg := newTestConfig(`
[ovh-eu]
application_key = ak123
`)

	td.Cmp(t, GetCustomHeaders(cfg, "ovh-eu"), map[string]string{})
}

func TestGetCustomHeaders_NilOrEmpty(t *testing.T) {
	td.Cmp(t, GetCustomHeaders(nil, "ovh-eu"), map[string]string{})
	td.Cmp(t, GetCustomHeaders(newTestConfig(""), ""), map[string]string{})
	td.Cmp(t, GetCustomHeaders(newTestConfig(""), "does-not-exist"), map[string]string{})
}

func TestGetProfileCustomHeaders(t *testing.T) {
	cfg := newTestConfig(`
[profile:work]
endpoint             = https://api.eu.ovhcloud.com/1.0
application_key      = ak123
header.X-Routing-Key = internal-build-eu
`)

	td.Cmp(t, GetProfileCustomHeaders(cfg, "work"), map[string]string{"X-Routing-Key": "internal-build-eu"})
}

func TestGetProfileCustomHeaders_NotFound(t *testing.T) {
	cfg := newTestConfig(`
[default]
endpoint = ovh-eu
`)

	td.Cmp(t, GetProfileCustomHeaders(cfg, "nonexistent"), map[string]string{})
	td.Cmp(t, GetProfileCustomHeaders(nil, "nonexistent"), map[string]string{})
}

func TestDeleteConfigValue(t *testing.T) {
	cfg := newTestConfig(`
[ovh-eu]
application_key      = ak123
header.X-Routing-Key = internal-build-eu
`)
	tmpFile := filepath.Join(t.TempDir(), "test.conf")

	td.Require(t).CmpNoError(DeleteConfigValue(cfg, tmpFile, "ovh-eu", "header.X-Routing-Key"))
	td.Cmp(t, GetCustomHeaders(cfg, "ovh-eu"), map[string]string{})

	// Unrelated key untouched
	val, err := GetConfigValue(cfg, "ovh-eu", "application_key")
	td.Require(t).CmpNoError(err)
	td.Cmp(t, val, "ak123")
}

func TestDeleteProfileConfigValue(t *testing.T) {
	cfg := newTestConfig(`
[profile:work]
application_key      = ak123
header.X-Routing-Key = internal-build-eu
`)
	tmpFile := filepath.Join(t.TempDir(), "test.conf")

	td.Require(t).CmpNoError(DeleteProfileConfigValue(cfg, tmpFile, "work", "header.X-Routing-Key"))
	td.Cmp(t, GetProfileCustomHeaders(cfg, "work"), map[string]string{})
	td.Cmp(t, GetProfileConfigValue(cfg, "work", "application_key"), "ak123")
}

func TestGetConfigValue_ActiveProfileOverride(t *testing.T) {
	// When ActiveProfileOverride is set (simulating --profile flag),
	// GetConfigValue should read from the overridden profile
	cfg := newTestConfig(`
[default]
endpoint=ovh-eu

[profile:work]
default_cloud_project=proj-from-work

[profile:personal]
default_cloud_project=proj-from-personal

[ovh-cli]
default_cloud_project=proj-from-legacy
`)

	// No override: legacy mode, reads from [ovh-cli]
	ActiveProfileOverride = ""
	val, err := GetConfigValue(cfg, "", "default_cloud_project")
	td.Require(t).CmpNoError(err)
	td.Cmp(t, val, "proj-from-legacy")

	// Override to "work": reads from [profile:work]
	ActiveProfileOverride = "work"
	val, err = GetConfigValue(cfg, "", "default_cloud_project")
	td.Require(t).CmpNoError(err)
	td.Cmp(t, val, "proj-from-work")

	// Override to "personal": reads from [profile:personal]
	ActiveProfileOverride = "personal"
	val, err = GetConfigValue(cfg, "", "default_cloud_project")
	td.Require(t).CmpNoError(err)
	td.Cmp(t, val, "proj-from-personal")

	// Override to "default": reads from legacy
	ActiveProfileOverride = "default"
	val, err = GetConfigValue(cfg, "", "default_cloud_project")
	td.Require(t).CmpNoError(err)
	td.Cmp(t, val, "proj-from-legacy")

	// Clean up
	ActiveProfileOverride = ""
}
