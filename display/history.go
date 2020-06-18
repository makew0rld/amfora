package display

// Tab number mapped to list of URLs ordered from first to most recent.
var tabHist = make(map[int][]string)

// Tab number mapped to where in its history you are.
// The value is a valid index of the string slice above.
var tabHistPos = make(map[int]int)

// addToHist adds the given URL to history.
// It assumes the URL is currently being loaded and displayed on the page.
func addToHist(u string) {
	if tabHistPos[curTab] < len(tabHist[curTab])-1 {
		// We're somewhere in the middle of the history instead, with URLs ahead and behind.
		// The URLs ahead need to be removed so this new URL is the most recent item in the history
		tabHist[curTab] = tabHist[curTab][:tabHistPos[curTab]+1]
	}
	tabHist[curTab] = append(tabHist[curTab], u)
	tabHistPos[curTab]++
}

func histForward() {
	if tabHistPos[curTab] >= len(tabHist[curTab])-1 {
		// Already on the most recent URL in the history
		return
	}
	tabHistPos[curTab]++
	go func() {
		handleURL(tabHist[curTab][tabHistPos[curTab]])
		applyScroll()
	}()
}

func histBack() {
	if tabHistPos[curTab] <= 0 {
		// First tab in history
		return
	}
	tabHistPos[curTab]--
	go func() {
		handleURL(tabHist[curTab][tabHistPos[curTab]])
		applyScroll()
	}()
}
