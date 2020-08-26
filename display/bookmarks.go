package display

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
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
	if viper.GetBool("a-general.color") {
		bkmkModal.SetBackgroundColor(config.GetColor("bkmk_modal_bg")).
			SetButtonBackgroundColor(config.GetColor("btn_bg")).
			SetButtonTextColor(config.GetColor("btn_text")).
			SetTextColor(config.GetColor("bkmk_modal_text"))
		bkmkModal.GetForm().
			SetLabelColor(config.GetColor("bkmk_modal_label")).
			SetFieldBackgroundColor(config.GetColor("bkmk_modal_field_bg")).
			SetFieldTextColor(config.GetColor("bkmk_modal_field_text"))
		bkmkModal.GetFrame().
			SetBorderColor(config.GetColor("bkmk_modal_text")).
			SetTitleColor(config.GetColor("bkmk_modal_text"))
	} else {
		bkmkModal.SetBackgroundColor(tcell.ColorBlack).
			SetButtonBackgroundColor(tcell.ColorWhite).
			SetButtonTextColor(tcell.ColorBlack).
			SetTextColor(tcell.ColorWhite)
		bkmkModal.GetForm().
			SetLabelColor(tcell.ColorWhite).
			SetFieldBackgroundColor(tcell.ColorWhite).
			SetFieldTextColor(tcell.ColorBlack)
		bkmkModal.GetFrame().
			SetBorderColor(tcell.ColorWhite).
			SetTitleColor(tcell.ColorWhite)
	}

	bkmkModal.SetBorder(true)
	bkmkModal.GetFrame().
		SetTitleAlign(cview.AlignCenter).
		SetTitle(" Add Bookmark ")
	bkmkModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		switch buttonLabel {
		case "Add":
			bkmkCh <- 1
		case "Change":
			bkmkCh <- 1
		case "Remove":
			bkmkCh <- -1
		case "Cancel":
			bkmkCh <- 0
		}
	})
}

// Bkmk displays the "Add a bookmark" modal.
// It accepts the default value for the bookmark name that will be displayed, but can be changed by the user.
// It also accepts a bool indicating whether this page already has a bookmark.
// It returns the bookmark name and the bookmark action:
// 1, 0, -1 for add/update, cancel, and remove
func openBkmkModal(name string, exists bool) (string, int) {
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
	bkmkModalText = ""
	bkmkModal.GetForm().AddInputField("Name: ", name, 0, nil,
		func(text string) {
			// Store for use later
			bkmkModalText = text
		})

	tabPages.ShowPage("bkmk")
	tabPages.SendToFront("bkmk")
	App.SetFocus(bkmkModal)
	App.Draw()

	action := <-bkmkCh
	tabPages.SwitchToPage(strconv.Itoa(curTab))
	App.SetFocus(tabs[curTab].view)
	App.Draw()

	return bkmkModalText, action
}

// Bookmarks displays the bookmarks page on the current tab.
func Bookmarks(t *tab) {
	// Gather bookmarks
	rawContent := "# Bookmarks\r\n\r\n"
	m, keys := bookmarks.All()
	for i := range keys {
		rawContent += fmt.Sprintf("=> %s %s\r\n", keys[i], m[keys[i]])
	}
	// Render and display
	content, links := renderer.RenderGemini(rawContent, textWidth(), leftMargin())
	page := structs.Page{
		Raw:       rawContent,
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
	if !strings.HasPrefix(tabs[curTab].page.URL, "gemini://") {
		// Can't make bookmarks for other kinds of URLs
		return
	}

	name, exists := bookmarks.Get(tabs[curTab].page.URL)
	// Open a bookmark modal with the current name of the bookmark, if it exists
	newName, action := openBkmkModal(name, exists)
	switch action {
	case 1:
		// Add/change the bookmark
		bookmarks.Set(tabs[curTab].page.URL, newName)
	case -1:
		bookmarks.Remove(tabs[curTab].page.URL)
	}
	// Other case is action = 0, meaning "Cancel", so nothing needs to happen
}
