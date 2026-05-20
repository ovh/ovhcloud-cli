// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package upgrade

// Method describes how the CLI was installed.
type Method int

const (
	// MethodBinary is a standalone binary (install.sh, manual download, or unknown).
	MethodBinary Method = iota
	// MethodBrew is a Homebrew cask install.
	MethodBrew
	// MethodGoInstall is a `go install` install.
	MethodGoInstall
)

func (m Method) String() string {
	switch m {
	case MethodBrew:
		return "brew"
	case MethodGoInstall:
		return "go install"
	default:
		return "binary"
	}
}
