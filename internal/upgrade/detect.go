// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package upgrade

import (
	"os"
	"path/filepath"
	"strings"
)

// DetectInstallMethod inspects the running binary's path and environment to
// guess how the CLI was installed.
func DetectInstallMethod() Method {
	exePath, err := os.Executable()
	if err != nil {
		return MethodBinary
	}
	if resolved, err := filepath.EvalSymlinks(exePath); err == nil {
		exePath = resolved
	}
	return detect(exePath, os.Getenv("GOBIN"), os.Getenv("GOPATH"), os.Getenv("HOME"))
}

func detect(exePath, gobin, gopath, home string) Method {
	if strings.Contains(exePath, "/Cellar/") || strings.Contains(exePath, "/Caskroom/") {
		return MethodBrew
	}

	dir := filepath.Dir(exePath)
	if gobin != "" && dir == gobin {
		return MethodGoInstall
	}
	if gopath != "" && dir == filepath.Join(gopath, "bin") {
		return MethodGoInstall
	}
	if home != "" && dir == filepath.Join(home, "go", "bin") {
		return MethodGoInstall
	}

	return MethodBinary
}
