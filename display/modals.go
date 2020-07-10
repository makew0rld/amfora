package display

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell"
	"github.com/spf13/viper"
	"gitlab.com/tslocum/cview"
)

// This file contains code for the popups / modals used in the display.
// The bookmark modal is in bookmarks.go

var infoModal = cview.NewModal().
	SetTextColor(tcell.ColorWhite).
	AddButtons([]string{"Ok"})

var errorModal = cview.NewModal().
	SetTextColor(tcell.ColorWhite).
	AddButtons([]string{"Ok"})

var inputModal = cview.NewModal().
	SetTextColor(tcell.ColorWhite)
	//AddButtons([]string{"Send", "Cancel"}) - Added in func

var inputCh = make(chan string)
var inputModalText string // The current text of the input field in the modal

var yesNoModal = cview.NewModal().
	SetTextColor(tcell.ColorWhite).
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
		infoModal.SetBackgroundColor(tcell.ColorGray).
			SetButtonBackgroundColor(tcell.ColorNavy).
			SetButtonTextColor(tcell.ColorWhite)
		errorModal.SetBackgroundColor(tcell.ColorMaroon).
			SetButtonBackgroundColor(tcell.ColorNavy).
			SetButtonTextColor(tcell.ColorWhite)
		inputModal.SetBackgroundColor(tcell.ColorGreen).
			SetButtonBackgroundColor(tcell.ColorNavy).
			SetButtonTextColor(tcell.ColorWhite)
		yesNoModal.SetButtonBackgroundColor(tcell.ColorNavy).
			SetButtonTextColor(tcell.ColorWhite)
	} else {
		infoModal.SetBackgroundColor(tcell.ColorBlack).
			SetButtonBackgroundColor(tcell.ColorWhite).
			SetButtonTextColor(tcell.ColorBlack)
		errorModal.SetBackgroundColor(tcell.ColorBlack).
			SetButtonBackgroundColor(tcell.ColorWhite).
			SetButtonTextColor(tcell.ColorBlack)
		inputModal.SetBackgroundColor(tcell.ColorBlack).
			SetButtonBackgroundColor(tcell.ColorWhite).
			SetButtonTextColor(tcell.ColorBlack)
		inputModal.GetForm().
			SetLabelColor(tcell.ColorWhite).
			SetFieldBackgroundColor(tcell.ColorWhite).
			SetFieldTextColor(tcell.ColorBlack)

		// YesNo background color is changed in funcs
		yesNoModal.SetButtonBackgroundColor(tcell.ColorWhite).
			SetButtonTextColor(tcell.ColorBlack)
	}

	// Modal functions that can't be added up above, because they return the wrong type

	infoModal.SetBorder(true)
	infoModal.SetBorderColor(tcell.ColorWhite)
	infoModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		tabPages.SwitchToPage(strconv.Itoa(curTab))
	})
	infoModal.GetFrame().SetTitleColor(tcell.ColorWhite)
	infoModal.GetFrame().SetTitleAlign(cview.AlignCenter)
	infoModal.GetFrame().SetTitle(" Info ")

	errorModal.SetBorder(true)
	errorModal.SetBorderColor(tcell.ColorWhite)
	errorModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		tabPages.SwitchToPage(strconv.Itoa(curTab))
	})
	errorModal.GetFrame().SetTitleColor(tcell.ColorWhite)
	errorModal.GetFrame().SetTitleAlign(cview.AlignCenter)

	inputModal.SetBorder(true)
	inputModal.SetBorderColor(tcell.ColorWhite)
	inputModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Send" {
			inputCh <- inputModalText
			return
		}
		// Empty string indicates no input
		inputCh <- ""

		//tabPages.SwitchToPage(strconv.Itoa(curTab)) - handled in Input()
	})
	inputModal.GetFrame().SetTitleColor(tcell.ColorWhite)
	inputModal.GetFrame().SetTitleAlign(cview.AlignCenter)
	inputModal.GetFrame().SetTitle(" Input ")

	yesNoModal.SetBorder(true)
	yesNoModal.SetBorderColor(tcell.ColorWhite)
	yesNoModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Yes" {
			yesNoCh <- true
			return
		}
		yesNoCh <- false

		//tabPages.SwitchToPage(strconv.Itoa(curTab)) - Handled in YesNo()
	})
	yesNoModal.GetFrame().SetTitleColor(tcell.ColorWhite)
	yesNoModal.GetFrame().SetTitleAlign(cview.AlignCenter)

	bkmkInit()
	dlInit()
}

// Error displays an error on the screen in a modal.
func Error(title, text string) {
	// Capitalize and add period if necessary - because most errors don't do that
	text = strings.ToUpper(string([]rune(text)[0])) + text[1:]
	if !strings.HasSuffix(text, ".") && !strings.HasSuffix(text, "!") && !strings.HasSuffix(text, "?") {
		text += "."
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

	inputModal.SetText(prompt)
	tabPages.ShowPage("input")
	tabPages.SendToFront("input")
	App.SetFocus(inputModal)
	App.Draw()

	resp := <-inputCh
	tabPages.SwitchToPage(strconv.Itoa(curTab))
	if resp == "" {
		return "", false
	}
	return resp, true
}

// YesNo displays a modal asking a yes-or-no question.
func YesNo(prompt string) bool {
	if viper.GetBool("a-general.color") {
		yesNoModal.SetBackgroundColor(tcell.ColorPurple)
	} else {
		yesNoModal.SetBackgroundColor(tcell.ColorBlack)
	}
	yesNoModal.GetFrame().SetTitle("")
	yesNoModal.SetText(prompt)
	tabPages.ShowPage("yesno")
	tabPages.SendToFront("yesno")
	App.SetFocus(yesNoModal)
	App.Draw()

	resp := <-yesNoCh
	tabPages.SwitchToPage(strconv.Itoa(curTab))
	return resp
}

// Tofu displays the TOFU warning modal.
// It returns a bool indicating whether the user wants to continue.
func Tofu(host string, expiry time.Time) bool {
	// Reuses yesNoModal, with error colour

	if viper.GetBool("a-general.color") {
		yesNoModal.SetBackgroundColor(tcell.ColorMaroon)
	} else {
		yesNoModal.SetBackgroundColor(tcell.ColorBlack)
	}
	yesNoModal.GetFrame().SetTitle(" TOFU ")
	yesNoModal.SetText(
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
	return resp
}
