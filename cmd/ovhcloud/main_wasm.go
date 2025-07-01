//go:build js && wasm

package main

import (
	"syscall/js"

	shellwords "github.com/mattn/go-shellwords"
	"stash.ovh.net/api/ovh-cli/internal/cmd"
)

func execCLI() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		cmdLine, err := shellwords.Parse(args[0].String())
		if err != nil {
			errorConstructor := js.Global().Get("Error")
			errorObject := errorConstructor.New(err.Error())
			return js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, args []js.Value) any {
				args[1].Invoke(errorObject)
				return nil
			}))
		}

		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			reject := args[1]

			go func() {
				defer cmd.PostExecute()

				out, err := cmd.Execute(cmdLine...)
				if err != nil {
					errorConstructor := js.Global().Get("Error")
					errorObject := errorConstructor.New(err.Error())
					reject.Invoke(errorObject)
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
	registerCallbacks()
	select {}
}
