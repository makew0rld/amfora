package display

import (
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
)

// This file contains code for all the popups / modals used in the display

var infoModal = cview.NewModal().
	SetBackgroundColor(tcell.ColorGray).
	SetButtonBackgroundColor(tcell.ColorNavy).
	SetButtonTextColor(tcell.ColorWhite).
	SetTextColor(tcell.ColorWhite).
	AddButtons([]string{"Ok"})

var errorModal = cview.NewModal().
	SetBackgroundColor(tcell.ColorMaroon).
	SetButtonBackgroundColor(tcell.ColorNavy).
	SetButtonTextColor(tcell.ColorWhite).
	SetTextColor(tcell.ColorWhite).
	AddButtons([]string{"Ok"})

// TODO: Support input
var inputModal = cview.NewModal().
	SetBackgroundColor(tcell.ColorGreen).
	SetButtonBackgroundColor(tcell.ColorNavy).
	SetButtonTextColor(tcell.ColorWhite).
	SetTextColor(tcell.ColorWhite).
	AddButtons([]string{"Send", "Cancel"})

var inputCh = make(chan string)
var inputModalText string // The current text of the input field in the modal

var yesNoModal = cview.NewModal().
	SetBackgroundColor(tcell.ColorPurple).
	SetButtonBackgroundColor(tcell.ColorNavy).
	SetButtonTextColor(tcell.ColorWhite).
	SetTextColor(tcell.ColorWhite).
	AddButtons([]string{"Yes", "No"})

// Channel to recieve yesNo answer on
var yesNoCh = make(chan bool)

func modalInit() {
	// Modal functions that can't be added up above, because they return the wrong type
	infoModal.SetBorder(true)
	infoModal.SetBorderColor(tcell.ColorWhite)
	infoModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		tabPages.SwitchToPage(strconv.Itoa(curTab))
	})

	errorModal.SetBorder(true)
	errorModal.SetBorderColor(tcell.ColorWhite)
	errorModal.SetTitleColor(tcell.ColorWhite)
	errorModal.SetTitleAlign(cview.AlignCenter)
	errorModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		tabPages.SwitchToPage(strconv.Itoa(curTab))
	})

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

	errorModal.SetTitle(title)
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
	// Remove and re-add input field - to clear the old text
	if inputModal.GetForm().GetFormItemCount() > 0 {
		inputModal.GetForm().RemoveFormItem(0)
	}
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
	yesNoModal.SetText(prompt)
	tabPages.ShowPage("yesno")
	tabPages.SendToFront("yesno")
	App.SetFocus(yesNoModal)
	App.Draw()

	resp := <-yesNoCh
	tabPages.SwitchToPage(strconv.Itoa(curTab))
	return resp
}
