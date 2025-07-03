package utils

import (
	"fmt"
	"os"
	"runtime"

	"dario.cat/mergo"
)

func MergeMaps(left, right map[string]any) error {
	if err := mergo.Merge(&left, right, mergo.WithOverride, mergo.WithAppendSlice); err != nil {
		return fmt.Errorf("error merging maps: %w", err)
	}

	return nil
}

func IsInputFromPipe() bool {
	if runtime.GOARCH == "wasm" && runtime.GOOS == "js" {
		return false
	}

	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}
