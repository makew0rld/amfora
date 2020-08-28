package display

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/gdamore/tcell"
	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/spf13/viper"
	"gitlab.com/tslocum/cview"
)

// This file contains code for the popups / modals used in the display.
// The bookmark modal is in bookmarks.go

var infoModal = cview.NewModal().
	AddButtons([]string{"Ok"})

var errorModal = cview.NewModal().
	AddButtons([]string{"Ok"})

var inputModal = cview.NewModal()
var inputCh = make(chan string)
var inputModalText string // The current text of the input field in the modal

var yesNoModal = cview.NewModal().
	AddButtons([]string{"Yes", "No"})

// Channel to receive yesNo answer on
var yesNoCh = make(chan bool)

func modalInit() {
	tabPages.AddPage("info", infoModal, false, false).
		AddPage("error", errorModal, false, false).
		AddPage("input", inputModal, false, false).
		AddPage("yesno", yesNoModal, false, false).
		AddPage("bkmk", bkmkModal, false, false).
		AddPage("dlChoice", dlChoiceModal, false, false).
		AddPage("dl", dlModal, false, false)

	// Color setup
	if viper.GetBool("a-general.color") {
		infoModal.SetBackgroundColor(config.GetColor("info_modal_bg")).
			SetButtonBackgroundColor(config.GetColor("btn_bg")).
			SetButtonTextColor(config.GetColor("btn_text")).
			SetTextColor(config.GetColor("info_modal_text"))
		infoModal.GetFrame().
			SetBorderColor(config.GetColor("info_modal_text")).
			SetTitleColor(config.GetColor("info_modal_text"))

		errorModal.SetBackgroundColor(config.GetColor("error_modal_bg")).
			SetButtonBackgroundColor(config.GetColor("btn_bg")).
			SetButtonTextColor(config.GetColor("btn_text")).
			SetTextColor(config.GetColor("error_modal_text"))
		errorModal.GetFrame().
			SetBorderColor(config.GetColor("error_modal_text")).
			SetTitleColor(config.GetColor("error_modal_text"))

		inputModal.SetBackgroundColor(config.GetColor("input_modal_bg")).
			SetButtonBackgroundColor(config.GetColor("btn_bg")).
			SetButtonTextColor(config.GetColor("btn_text")).
			SetTextColor(config.GetColor("input_modal_text"))
		inputModal.GetFrame().
			SetBorderColor(config.GetColor("input_modal_text")).
			SetTitleColor(config.GetColor("input_modal_text"))
		inputModal.GetForm().
			SetFieldBackgroundColor(config.GetColor("input_modal_field_bg")).
			SetFieldTextColor(config.GetColor("input_modal_field_text"))

		yesNoModal.SetButtonBackgroundColor(config.GetColor("btn_bg")).
			SetButtonTextColor(config.GetColor("btn_text"))
	} else {
		infoModal.SetBackgroundColor(tcell.ColorBlack).
			SetButtonBackgroundColor(tcell.ColorWhite).
			SetButtonTextColor(tcell.ColorBlack).
			SetTextColor(tcell.ColorWhite)
		infoModal.GetFrame().
			SetBorderColor(tcell.ColorWhite).
			SetTitleColor(tcell.ColorWhite)

		errorModal.SetBackgroundColor(tcell.ColorBlack).
			SetButtonBackgroundColor(tcell.ColorWhite).
			SetButtonTextColor(tcell.ColorBlack).
			SetTextColor(tcell.ColorWhite)
		errorModal.GetFrame().
			SetBorderColor(tcell.ColorWhite).
			SetTitleColor(tcell.ColorWhite)

		inputModal.SetBackgroundColor(tcell.ColorBlack).
			SetButtonBackgroundColor(tcell.ColorWhite).
			SetButtonTextColor(tcell.ColorBlack).
			SetTextColor(tcell.ColorWhite)
		inputModal.GetFrame().
			SetBorderColor(tcell.ColorWhite).
			SetTitleColor(tcell.ColorWhite)
		inputModal.GetForm().
			SetFieldBackgroundColor(tcell.ColorWhite).
			SetFieldTextColor(tcell.ColorBlack)

		// YesNo background color is changed in funcs
		yesNoModal.SetButtonBackgroundColor(tcell.ColorWhite).
			SetButtonTextColor(tcell.ColorBlack)
	}

	// Modal functions that can't be added up above, because they return the wrong type

	infoModal.SetBorder(true)
	infoModal.GetFrame().
		SetTitleAlign(cview.AlignCenter).
		SetTitle(" Info ")
	infoModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		tabPages.SwitchToPage(strconv.Itoa(curTab))
		App.SetFocus(tabs[curTab].view)
		App.Draw()
	})

	errorModal.SetBorder(true)
	errorModal.GetFrame().SetTitleAlign(cview.AlignCenter)
	errorModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		tabPages.SwitchToPage(strconv.Itoa(curTab))
		App.SetFocus(tabs[curTab].view)
		App.Draw()
	})

	inputModal.SetBorder(true)
	inputModal.GetFrame().
		SetTitleAlign(cview.AlignCenter).
		SetTitle(" Input ")
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
	feedInit()
}

// Error displays an error on the screen in a modal.
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
	tabPages.ShowPage("error")
	tabPages.SendToFront("error")
	App.SetFocus(errorModal)
	App.Draw()
}

// Info displays some info on the screen in a modal.
func Info(s string) {
	infoModal.SetText(s)
	tabPages.ShowPage("info")
	tabPages.SendToFront("info")
	App.SetFocus(infoModal)
	App.Draw()
}

// Input pulls up a modal that asks for input, and returns the user's input.
// It returns an bool indicating if the user chose to send input or not.
func Input(prompt string) (string, bool) {
	// Remove elements and re-add them - to clear input text and keep input in focus
	inputModal.ClearButtons()
	inputModal.GetForm().Clear(false)

	inputModal.AddButtons([]string{"Send", "Cancel"})
	inputModalText = ""
	inputModal.GetForm().AddInputField("", "", 0, nil,
		func(text string) {
			// Store for use later
			inputModalText = text
		})

	inputModal.SetText(prompt + " ")
	tabPages.ShowPage("input")
	tabPages.SendToFront("input")
	App.SetFocus(inputModal)
	App.Draw()

	resp := <-inputCh

	tabPages.SwitchToPage(strconv.Itoa(curTab))
	App.SetFocus(tabs[curTab].view)
	App.Draw()

	if resp == "" {
		return "", false
	}
	return resp, true
}

// YesNo displays a modal asking a yes-or-no question.
func YesNo(prompt string) bool {
	if viper.GetBool("a-general.color") {
		yesNoModal.
			SetBackgroundColor(config.GetColor("yesno_modal_bg")).
			SetTextColor(config.GetColor("yesno_modal_text"))
		yesNoModal.GetFrame().
			SetBorderColor(config.GetColor("yesno_modal_text")).
			SetTitleColor(config.GetColor("yesno_modal_text"))
	} else {
		yesNoModal.
			SetBackgroundColor(tcell.ColorBlack).
			SetTextColor(tcell.ColorWhite)
		yesNoModal.GetFrame().
			SetBorderColor(tcell.ColorWhite).
			SetTitleColor(tcell.ColorWhite)
	}
	yesNoModal.GetFrame().SetTitle("")
	yesNoModal.SetText(prompt)
	tabPages.ShowPage("yesno")
	tabPages.SendToFront("yesno")
	App.SetFocus(yesNoModal)
	App.Draw()

	resp := <-yesNoCh
	tabPages.SwitchToPage(strconv.Itoa(curTab))
	App.SetFocus(tabs[curTab].view)
	App.Draw()
	return resp
}

// Tofu displays the TOFU warning modal.
// It returns a bool indicating whether the user wants to continue.
func Tofu(host string, expiry time.Time) bool {
	// Reuses yesNoModal, with error color

	if viper.GetBool("a-general.color") {
		yesNoModal.
			SetBackgroundColor(config.GetColor("tofu_modal_bg")).
			SetTextColor(config.GetColor("tofu_modal_text"))
		yesNoModal.GetFrame().
			SetBorderColor(config.GetColor("tofu_modal_text")).
			SetTitleColor(config.GetColor("tofu_modal_text"))
	} else {
		yesNoModal.
			SetBackgroundColor(tcell.ColorBlack).
			SetTextColor(tcell.ColorWhite)
		yesNoModal.
			SetBorderColor(tcell.ColorWhite).
			SetTitleColor(tcell.ColorWhite)
	}
	yesNoModal.GetFrame().SetTitle(" TOFU ")
	yesNoModal.SetText(
		//nolint:lll
		fmt.Sprintf("%s's certificate has changed, possibly indicating an security issue. The certificate would have expired %s. Are you sure you want to continue? ",
			host,
			humanize.Time(expiry),
		),
	)
	tabPages.ShowPage("yesno")
	tabPages.SendToFront("yesno")
	App.SetFocus(yesNoModal)
	App.Draw()

	resp := <-yesNoCh
	tabPages.SwitchToPage(strconv.Itoa(curTab))
	App.SetFocus(tabs[curTab].view)
	App.Draw()
	return resp
}
