//go:build !(js && wasm)

package display

import (
	"encoding/json"

	fxdisplay "github.com/amstuta/fx/display"
)

func displayInteractive(value any) {
	bytes, err := json.Marshal(value)
	if err != nil {
		ExitError("error preparing interactive output: %s", err)
	}
	fxdisplay.Display(bytes, "")
}
