// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"testing"

	"github.com/spf13/cobra"
)

// newEditTestCmd builds a command exposing the same kinds of flags an edit
// command has: a project selector, --wait, global flags, and actual content
// flags (--name, --size).
func newEditTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "edit"}
	cmd.Flags().String("cloud-project", "", "")
	cmd.Flags().Bool("wait", false, "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Int("size", 0, "")
	return cmd
}

func TestEditHasContentFlags(t *testing.T) {
	tests := []struct {
		name string
		set  map[string]string
		want bool
	}{
		{
			name: "nothing set",
			set:  nil,
			want: false,
		},
		{
			name: "only selector and wait",
			set:  map[string]string{"cloud-project": "proj", "wait": "true"},
			want: false,
		},
		{
			name: "content flag set",
			set:  map[string]string{"cloud-project": "proj", "name": "renamed"},
			want: true,
		},
		{
			name: "only a numeric content flag",
			set:  map[string]string{"size": "20"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newEditTestCmd()
			for k, v := range tt.set {
				if err := cmd.Flags().Set(k, v); err != nil {
					t.Fatalf("failed to set flag %s: %v", k, err)
				}
			}
			if got := editHasContentFlags(cmd); got != tt.want {
				t.Errorf("editHasContentFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}
