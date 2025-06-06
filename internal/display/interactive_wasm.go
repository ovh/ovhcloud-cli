//go:build js && wasm

package display

func displayInteractive(_ any) {
	ExitError("interactive mode not available in WASM binary")
}
