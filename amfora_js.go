//go:build js
// +build js

package main

import "github.com/makeworld-the-better-one/amfora/display"
import "syscall/js"

func init() {
	js.Global().Set("amforaAPI", makeJSAPI())
}

func makeJSAPI() js.Value {
	var api = map[string]interface{}{
		"navigateTo": js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
			if len(args) == 0 {
				return jsError("musr provide a URL argument")
			}

			display.URL(args[0].String()) // does the navigation
			return nil
		}),
	}

	return js.ValueOf(api)
}

func jsError(msg string) js.Value {
	return js.Global().Get("Error").New(msg)
}
