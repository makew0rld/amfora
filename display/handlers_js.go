package display

import "github.com/makeworld-the-better-one/amfora/structs"
import "syscall/js"

func init() {
	js.Global().Set("amforaAPI", makeJSAPI())
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

func observePage(page *structs.Page) {
	if jsFunc := js.Global().Get("observePage"); js.TypeFunction == jsFunc.Type() {
		jsFunc.Invoke(page.URL, page.Title())
	}
}

func jsError(msg string) js.Value {
	return js.Global().Get("Error").New(msg)
}

