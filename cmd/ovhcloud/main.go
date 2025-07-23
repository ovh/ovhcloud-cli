//go:build !(js && wasm)

package main

import (
	"os"

	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func main() {
	if _, err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
