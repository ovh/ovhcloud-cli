//go:build !(js && wasm)

package main

import (
	"os"

	"stash.ovh.net/api/ovh-cli/internal/cmd"
)

func main() {
	if _, err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
