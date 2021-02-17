package display

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/makeworld-the-better-one/amfora/bookmarks"
	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/amfora/renderer"
	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/spf13/viper"
	"gitlab.com/tslocum/cview"
)

// For adding and removing bookmarks, basically a clone of the input modal.
var bkmkModal = cview.NewModal()

// bkmkCh is for the user action
var bkmkCh = make(chan int) // 1, 0, -1 for add/update, cancel, and remove
var bkmkModalText string    // The current text of the input field in the modal

func bkmkInit() {
	panels.AddPanel("bkmk", bkmkModal, false, false)

	m := bkmkModal
	if viper.GetBool("a-general.color") {
		m.SetBackgroundColor(config.GetColor("bkmk_modal_bg"))
		m.SetButtonBackgroundColor(config.GetColor("btn_bg"))
		m.SetButtonTextColor(config.GetColor("btn_text"))
		m.SetTextColor(config.GetColor("bkmk_modal_text"))
		form := m.GetForm()
		form.SetLabelColor(config.GetColor("bkmk_modal_label"))
		form.SetFieldBackgroundColor(config.GetColor("bkmk_modal_field_bg"))
		form.SetFieldTextColor(config.GetColor("bkmk_modal_field_text"))
		form.SetButtonBackgroundColorFocused(config.GetColor("btn_text"))
		form.SetButtonTextColorFocused(config.GetColor("btn_bg"))
		frame := m.GetFrame()
		frame.SetBorderColor(config.GetColor("bkmk_modal_text"))
		frame.SetTitleColor(config.GetColor("bkmk_modal_text"))
	} else {
		m.SetBackgroundColor(tcell.ColorBlack)
		m.SetButtonBackgroundColor(tcell.ColorWhite)
		m.SetButtonTextColor(tcell.ColorBlack)
		m.SetTextColor(tcell.ColorWhite)
		form := m.GetForm()
		form.SetLabelColor(tcell.ColorWhite)
		form.SetFieldBackgroundColor(tcell.ColorWhite)
		form.SetFieldTextColor(tcell.ColorBlack)
		form.SetButtonBackgroundColorFocused(tcell.ColorBlack)
		form.SetButtonTextColorFocused(tcell.ColorWhite)
		frame := m.GetFrame()
		frame.SetBorderColor(tcell.ColorWhite)
		frame.SetTitleColor(tcell.ColorWhite)
	}

	m.SetBorder(true)
	frame := m.GetFrame()
	frame.SetTitleAlign(cview.AlignCenter)
	frame.SetTitle(" Add Bookmark ")
	m.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		switch buttonLabel {
		case "Add":
			bkmkCh <- 1
		case "Change":
			bkmkCh <- 1
		case "Remove":
			bkmkCh <- -1
		case "Cancel":
			bkmkCh <- 0
		case "":
			bkmkCh <- 0
		}
	})
}

// Bkmk displays the "Add a bookmark" modal.
// It accepts the default value for the bookmark name that will be displayed, but can be changed by the user.
// It also accepts a bool indicating whether this page already has a bookmark.
// It returns the bookmark name and the bookmark action:
// 1, 0, -1 for add/update, cancel, and remove
func openBkmkModal(name string, exists bool, favicon string) (string, int) {
	// Basically a copy of Input()

	// Reset buttons before input field, to make sure the input is in focus
	bkmkModal.ClearButtons()
	if exists {
		bkmkModal.SetText("Change or remove the bookmark for the current page?")
		bkmkModal.AddButtons([]string{"Change", "Remove", "Cancel"})
	} else {
		bkmkModal.SetText("Create a bookmark for the current page?")
		bkmkModal.AddButtons([]string{"Add", "Cancel"})
	}

	// Remove and re-add input field - to clear the old text
	bkmkModal.GetForm().Clear(false)
	if favicon != "" && !exists {
		name = favicon + " " + name
	}
	bkmkModalText = name
	bkmkModal.GetForm().AddInputField("Name: ", name, 0, nil,
		func(text string) {
			// Store for use later
			bkmkModalText = text
		})

	panels.ShowPanel("bkmk")
	panels.SendToFront("bkmk")
	App.SetFocus(bkmkModal)
	App.Draw()

	action := <-bkmkCh
	panels.HidePanel("bkmk")
	App.SetFocus(tabs[curTab].view)
	App.Draw()

	return bkmkModalText, action
}

// Bookmarks displays the bookmarks page on the current tab.
func Bookmarks(t *tab) {
	bkmkPageRaw := "# Bookmarks\r\n\r\n"

	// Gather bookmarks
	m, keys := bookmarks.All()
	for i := range keys {
		bkmkPageRaw += fmt.Sprintf("=> %s %s\r\n", keys[i], m[keys[i]])
	}
	// Render and display
	content, links := renderer.RenderGemini(bkmkPageRaw, textWidth(), false)
	page := structs.Page{
		Raw:       bkmkPageRaw,
		Content:   content,
		Links:     links,
		URL:       "about:bookmarks",
		Width:     termW,
		Mediatype: structs.TextGemini,
	}
	setPage(t, &page)
	t.applyBottomBar()
}

// addBookmark goes through the process of adding a bookmark for the current page.
// It is the high-level way of doing it. It should be called in a goroutine.
// It can also be called to edit an existing bookmark.
func addBookmark() {
	t := tabs[curTab]
	p := t.page

	if !t.hasContent() {
		// It's an about: page, or a malformed one
		return
	}
	name, exists := bookmarks.Get(p.URL)
	// Open a bookmark modal with the current name of the bookmark, if it exists
	newName, action := openBkmkModal(name, exists, p.Favicon)
	switch action {
	case 1:
		// Add/change the bookmark
		bookmarks.Set(p.URL, newName)
	case -1:
		bookmarks.Remove(p.URL)
	}
	// Other case is action = 0, meaning "Cancel", so nothing needs to happen
}
