package display

import "github.com/makeworld-the-better-one/amfora/structs"
import "syscall/js"

func observePage(page *structs.Page) {
	if jsFunc := js.Global().Get("observePage"); js.TypeFunction == jsFunc.Type() {
		jsFunc.Invoke(page.URL, page.Title())
	}
}
