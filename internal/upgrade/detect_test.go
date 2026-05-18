// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package upgrade

import "testing"

func TestDetect(t *testing.T) {
	cases := []struct {
		name    string
		exePath string
		gobin   string
		gopath  string
		home    string
		want    Method
	}{
		{
			name:    "homebrew cellar",
			exePath: "/opt/homebrew/Cellar/ovhcloud-cli/1.2.3/bin/ovhcloud",
			home:    "/Users/alice",
			want:    MethodBrew,
		},
		{
			name:    "homebrew caskroom",
			exePath: "/usr/local/Caskroom/ovhcloud-cli/1.2.3/ovhcloud",
			home:    "/Users/alice",
			want:    MethodBrew,
		},
		{
			name:    "go install via GOBIN",
			exePath: "/home/alice/go-bin/ovhcloud",
			gobin:   "/home/alice/go-bin",
			home:    "/home/alice",
			want:    MethodGoInstall,
		},
		{
			name:    "go install via GOPATH",
			exePath: "/home/alice/gopath/bin/ovhcloud",
			gopath:  "/home/alice/gopath",
			home:    "/home/alice",
			want:    MethodGoInstall,
		},
		{
			name:    "go install via default $HOME/go/bin",
			exePath: "/home/alice/go/bin/ovhcloud",
			home:    "/home/alice",
			want:    MethodGoInstall,
		},
		{
			name:    "install.sh in ~/.local/bin",
			exePath: "/home/alice/.local/bin/ovhcloud",
			home:    "/home/alice",
			want:    MethodBinary,
		},
		{
			name:    "manual /usr/local/bin",
			exePath: "/usr/local/bin/ovhcloud",
			home:    "/home/alice",
			want:    MethodBinary,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := detect(tc.exePath, tc.gobin, tc.gopath, tc.home)
			if got != tc.want {
				t.Fatalf("detect(%q) = %v, want %v", tc.exePath, got, tc.want)
			}
		})
	}
}
