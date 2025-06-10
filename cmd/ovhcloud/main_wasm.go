//go:build js && wasm

package main

import (
	"strings"
	"syscall/js"

	"stash.ovh.net/api/ovh-cli/internal/cmd"
)

func execCLI() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		cmdLine := strings.Split(args[0].String(), " ")

		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			reject := args[1]

			go func() {
				out, err := cmd.Execute(cmdLine...)
				if err != nil {
					reject.Invoke(err)
					return
				}

				arrayConstructor := js.Global().Get("Uint8Array")
				dataJS := arrayConstructor.New(len(out))
				js.CopyBytesToJS(dataJS, []byte(out))

				responseConstructor := js.Global().Get("Response")
				response := responseConstructor.New(dataJS)

				resolve.Invoke(response)
			}()
			return nil
		})
		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}

func registerCallbacks() {
	js.Global().Set("exec", execCLI())
}

func main() {
	c := make(chan struct{})
	registerCallbacks()
	<-c
}
