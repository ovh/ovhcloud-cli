// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"os"
	"runtime"

	"dario.cat/mergo"
)

func MergeMaps(left, right map[string]any, opts ...func(*mergo.Config)) error {
	opts = append([]func(*mergo.Config){mergo.WithOverride, mergo.WithAppendSlice}, opts...)
	if err := mergo.Merge(&left, right, opts...); err != nil {
		return fmt.Errorf("error merging maps: %w", err)
	}

	return nil
}

// ReplaceSlices is a MergeMaps option making a slice of right replace the one of left.
func ReplaceSlices(config *mergo.Config) {
	config.AppendSlice = false
}

func IsInputFromPipe() bool {
	if runtime.GOARCH == "wasm" && runtime.GOOS == "js" {
		return false
	}

	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}
