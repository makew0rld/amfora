package display

import "syscall/js"

func init() {
	js.Global().Set("amforaAPI", makeJSAPI())
}

func observeURL(u string) {
	if js.Global().Get("observeURL").IsNull() {
		return
	}

	js.Global().Call("observeURL", u)
}

func makeJSAPI() js.Value {
	api := map[string]interface{}{
		"navigateTo": js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
			if len(args) == 0 {
				return jsError("musr provide a URL argument")
			}

			URL(args[0].String()) // does the navigation
			return nil
		}),
	}

	return js.ValueOf(api)
}

func jsError(msg string) js.Value {
	return js.Global().Get("Error").New(msg)
}

