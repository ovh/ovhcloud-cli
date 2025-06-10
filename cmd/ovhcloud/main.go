//go:build js && wasm

package main

import (
	"strings"
	"syscall/js"

	"stash.ovh.net/api/ovh-cli/internal/cmd"
)

func exec(this js.Value, args []js.Value) interface{} {
	ret := strings.Split(args[0].String(), " ")
	cmd.Execute(ret...)
	return js.ValueOf("cmd.Execute")
}

func registerCallbacks() {
	js.Global().Set("exec", js.FuncOf(exec))
}

func main() {
	c := make(chan struct{})
	registerCallbacks()
	// cmd.Execute()
	<-c
}
