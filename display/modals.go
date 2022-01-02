package display

import (
	"fmt"
	"strings"
	"time"

	"code.rocketnine.space/tslocum/cview"
	humanize "github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/spf13/viper"
)

// This file contains code for the popups / modals used in the display.
// The bookmark modal is in bookmarks.go

var infoModal = cview.NewModal()
var errorModal = cview.NewModal()
var inputModal = cview.NewModal()
var yesNoModal = cview.NewModal()

var inputCh = make(chan string)
var yesNoCh = make(chan bool)

var inputModalText string // The current text of the input field in the modal

// Internal channel used to know when a modal has been dismissed
var modalDone = make(chan struct{})

func modalInit() {
	infoModal.AddButtons([]string{"Ok"})
	errorModal.AddButtons([]string{"Ok"})
	yesNoModal.AddButtons([]string{"Yes", "No"})

	panels.AddPanel(PanelInfoModal, infoModal, false, false)
	panels.AddPanel(PanelErrorModal, errorModal, false, false)
	panels.AddPanel(PanelInputModal, inputModal, false, false)
	panels.AddPanel(PanelYesNoModal, yesNoModal, false, false)

	// Color setup
	if viper.GetBool("a-general.color") {
		m := infoModal
		m.SetBackgroundColor(config.GetColor("info_modal_bg"))
		m.SetButtonBackgroundColor(config.GetColor("btn_bg"))
		m.SetButtonTextColor(config.GetColor("btn_text"))
		m.SetTextColor(config.GetColor("info_modal_text"))
		form := m.GetForm()
		form.SetButtonBackgroundColorFocused(config.GetColor("btn_text"))
		form.SetButtonTextColorFocused(config.GetTextColor("btn_bg", "btn_text"))
		frame := m.GetFrame()
		frame.SetBorderColor(config.GetColor("info_modal_text"))
		frame.SetTitleColor(config.GetColor("info_modal_text"))

		m = errorModal
		m.SetBackgroundColor(config.GetColor("error_modal_bg"))
		m.SetButtonBackgroundColor(config.GetColor("btn_bg"))
		m.SetButtonTextColor(config.GetColor("btn_text"))
		m.SetTextColor(config.GetColor("error_modal_text"))
		form = m.GetForm()
		form.SetButtonBackgroundColorFocused(config.GetColor("btn_text"))
		form.SetButtonTextColorFocused(config.GetTextColor("btn_bg", "btn_text"))
		frame = errorModal.GetFrame()
		frame.SetBorderColor(config.GetColor("error_modal_text"))
		frame.SetTitleColor(config.GetColor("error_modal_text"))

		m = inputModal
		m.SetBackgroundColor(config.GetColor("input_modal_bg"))
		m.SetButtonBackgroundColor(config.GetColor("btn_bg"))
		m.SetButtonTextColor(config.GetColor("btn_text"))
		m.SetTextColor(config.GetColor("input_modal_text"))
		frame = inputModal.GetFrame()
		frame.SetBorderColor(config.GetColor("input_modal_text"))
		frame.SetTitleColor(config.GetColor("input_modal_text"))
		form = inputModal.GetForm()
		form.SetFieldBackgroundColor(config.GetColor("input_modal_field_bg"))
		form.SetFieldTextColor(config.GetColor("input_modal_field_text"))
		form.SetButtonBackgroundColorFocused(config.GetColor("btn_text"))
		form.SetButtonTextColorFocused(config.GetTextColor("btn_bg", "btn_text"))

		m = yesNoModal
		m.SetButtonBackgroundColor(config.GetColor("btn_bg"))
		m.SetButtonTextColor(config.GetColor("btn_text"))
		form = m.GetForm()
		form.SetButtonBackgroundColorFocused(config.GetColor("btn_text"))
		form.SetButtonTextColorFocused(config.GetTextColor("btn_bg", "btn_text"))
	} else {
		m := infoModal
		m.SetBackgroundColor(tcell.ColorBlack)
		m.SetButtonBackgroundColor(tcell.ColorWhite)
		m.SetButtonTextColor(tcell.ColorBlack)
		m.SetTextColor(tcell.ColorWhite)
		form := m.GetForm()
		form.SetButtonBackgroundColorFocused(tcell.ColorBlack)
		form.SetButtonTextColorFocused(tcell.ColorWhite)
		frame := infoModal.GetFrame()
		frame.SetBorderColor(tcell.ColorWhite)
		frame.SetTitleColor(tcell.ColorWhite)

		m = errorModal
		m.SetBackgroundColor(tcell.ColorBlack)
		m.SetButtonBackgroundColor(tcell.ColorWhite)
		m.SetButtonTextColor(tcell.ColorBlack)
		m.SetTextColor(tcell.ColorWhite)
		form = m.GetForm()
		form.SetButtonBackgroundColorFocused(tcell.ColorBlack)
		form.SetButtonTextColorFocused(tcell.ColorWhite)
		frame = errorModal.GetFrame()
		frame.SetBorderColor(tcell.ColorWhite)
		frame.SetTitleColor(tcell.ColorWhite)

		m = inputModal
		m.SetBackgroundColor(tcell.ColorBlack)
		m.SetButtonBackgroundColor(tcell.ColorWhite)
		m.SetButtonTextColor(tcell.ColorBlack)
		m.SetTextColor(tcell.ColorWhite)
		frame = inputModal.GetFrame()
		frame.SetBorderColor(tcell.ColorWhite)
		frame.SetTitleColor(tcell.ColorWhite)
		form = inputModal.GetForm()
		form.SetFieldBackgroundColor(tcell.ColorWhite)
		form.SetFieldTextColor(tcell.ColorBlack)
		form.SetButtonBackgroundColorFocused(tcell.ColorBlack)
		form.SetButtonTextColorFocused(tcell.ColorWhite)

		// YesNo background color is changed in funcs
		m = yesNoModal
		m.SetButtonBackgroundColor(tcell.ColorWhite)
		m.SetButtonTextColor(tcell.ColorBlack)
		form = m.GetForm()
		form.SetButtonBackgroundColorFocused(tcell.ColorBlack)
		form.SetButtonTextColorFocused(tcell.ColorWhite)
	}

	// Modal functions that can't be added up above, because they return the wrong type

	infoModal.SetBorder(true)
	frame := infoModal.GetFrame()
	frame.SetTitleAlign(cview.AlignCenter)
	frame.SetTitle(" Info ")
	infoModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		panels.HidePanel(PanelInfoModal)
		App.SetFocus(tabs[curTab].view)
		App.Draw()
		modalDone <- struct{}{}
	})

	errorModal.SetBorder(true)
	errorModal.GetFrame().SetTitleAlign(cview.AlignCenter)
	errorModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		panels.HidePanel(PanelErrorModal)
		App.SetFocus(tabs[curTab].view)
		App.Draw()
		modalDone <- struct{}{}
	})

	inputModal.SetBorder(true)
	frame = inputModal.GetFrame()
	frame.SetTitleAlign(cview.AlignCenter)
	frame.SetTitle(" Input ")
	inputModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Send" {
			inputCh <- inputModalText
			return
		}
		// Empty string indicates no input
		inputCh <- ""
	})

	yesNoModal.SetBorder(true)
	yesNoModal.GetFrame().SetTitleAlign(cview.AlignCenter)
	yesNoModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Yes" {
			yesNoCh <- true
			return
		}
		yesNoCh <- false
	})

	bkmkInit()
	dlInit()
}

// Error displays an error on the screen in a modal, and blocks until dismissed by the user.
func Error(title, text string) {
	if text == "" {
		text = "No additional information."
	} else {
		text = strings.ToUpper(string([]rune(text)[0])) + text[1:]
		if !strings.HasSuffix(text, ".") && !strings.HasSuffix(text, "!") && !strings.HasSuffix(text, "?") {
			text += "."
		}
	}
	// Add spaces to title for aesthetic reasons
	title = " " + strings.TrimSpace(title) + " "

	errorModal.GetFrame().SetTitle(title)
	errorModal.SetText(text)
	panels.ShowPanel(PanelErrorModal)
	panels.SendToFront(PanelErrorModal)
	App.SetFocus(errorModal)
	App.Draw()

	<-modalDone
}

// Info displays some info on the screen in a modal, and blocks until dismissed by the user.
func Info(s string) {
	infoModal.SetText(s)
	panels.ShowPanel(PanelInfoModal)
	panels.SendToFront(PanelInfoModal)
	App.SetFocus(infoModal)
	App.Draw()

	<-modalDone
}

// Input pulls up a modal that asks for input, waits for that input, and returns it.
// It returns an bool indicating if the user chose to send input or not.
func Input(prompt string, sensitive bool) (string, bool) {
	// Remove elements and re-add them - to clear input text and keep input in focus
	inputModal.ClearButtons()
	inputModal.GetForm().Clear(false)

	inputModal.AddButtons([]string{"Send", "Cancel"})
	inputModalText = ""

	if sensitive {
		// TODO use bullet characters if user wants it once bug is fixed - see NOTES.md
		inputModal.GetForm().AddPasswordField("", "", 0, '*',
			func(text string) {
				// Store for use later
				inputModalText = text
			})
	} else {
		inputModal.GetForm().AddInputField("", "", 0, nil,
			func(text string) {
				inputModalText = text
			})
	}

	inputModal.SetText(prompt + " ")
	panels.ShowPanel(PanelInputModal)
	panels.SendToFront(PanelInputModal)
	App.SetFocus(inputModal)
	App.Draw()

	resp := <-inputCh

	panels.HidePanel(PanelInputModal)
	App.SetFocus(tabs[curTab].view)
	App.Draw()

	if resp == "" {
		return "", false
	}
	return resp, true
}

// YesNo displays a modal asking a yes-or-no question, waits for an answer, then returns it as a bool.
func YesNo(prompt string) bool {
	if viper.GetBool("a-general.color") {
		m := yesNoModal
		m.SetBackgroundColor(config.GetColor("yesno_modal_bg"))
		m.SetTextColor(config.GetColor("yesno_modal_text"))
		frame := yesNoModal.GetFrame()
		frame.SetBorderColor(config.GetColor("yesno_modal_text"))
		frame.SetTitleColor(config.GetColor("yesno_modal_text"))
	} else {
		m := yesNoModal
		m.SetBackgroundColor(tcell.ColorBlack)
		m.SetTextColor(tcell.ColorWhite)
		frame := yesNoModal.GetFrame()
		frame.SetBorderColor(tcell.ColorWhite)
		frame.SetTitleColor(tcell.ColorWhite)
	}
	yesNoModal.GetFrame().SetTitle("")
	yesNoModal.SetText(prompt)
	panels.ShowPanel(PanelYesNoModal)
	panels.SendToFront(PanelYesNoModal)
	App.SetFocus(yesNoModal)
	App.Draw()

	resp := <-yesNoCh
	panels.HidePanel(PanelYesNoModal)
	App.SetFocus(tabs[curTab].view)
	App.Draw()
	return resp
}

// Tofu displays the TOFU warning modal.
// It blocks then returns a bool indicating whether the user wants to continue.
func Tofu(host string, expiry time.Time) bool {
	// Reuses yesNoModal, with error color

	m := yesNoModal
	frame := yesNoModal.GetFrame()
	if viper.GetBool("a-general.color") {
		m.SetBackgroundColor(config.GetColor("tofu_modal_bg"))
		m.SetTextColor(config.GetColor("tofu_modal_text"))
		frame.SetBorderColor(config.GetColor("tofu_modal_text"))
		frame.SetTitleColor(config.GetColor("tofu_modal_text"))
	} else {
		m.SetBackgroundColor(tcell.ColorBlack)
		m.SetTextColor(tcell.ColorWhite)
		m.SetBorderColor(tcell.ColorWhite)
		m.SetTitleColor(tcell.ColorWhite)
	}
	frame.SetTitle(" TOFU ")
	m.SetText(
		//nolint:lll
		fmt.Sprintf("%s's certificate has changed, possibly indicating a security issue. The certificate would have expired %s. Are you sure you want to continue? ",
			host,
			humanize.Time(expiry),
		),
	)
	panels.ShowPanel(PanelYesNoModal)
	panels.SendToFront(PanelYesNoModal)
	App.SetFocus(yesNoModal)
	App.Draw()

	resp := <-yesNoCh
	panels.HidePanel(PanelYesNoModal)
	App.SetFocus(tabs[curTab].view)
	App.Draw()
	return resp
}
