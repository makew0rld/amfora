package display

func histForward() {
	if tabs[curTab].history.pos >= len(tabs[curTab].history.urls)-1 {
		// Already on the most recent URL in the history
		return
	}
	tabs[curTab].history.pos++
	go func(tab int) {
		handleURL(tab, tabs[tab].history.urls[tabs[tab].history.pos]) // Load that position in history
		tabs[tab].applyScroll()
		if tab == curTab {
			// Display the bottomBar state that handleURL set
			tabs[tab].applyBottomBar()
		}
	}(curTab)
}

func histBack() {
	if tabs[curTab].history.pos <= 0 {
		// First tab in history
		return
	}
	tabs[curTab].history.pos--
	go func(tab int) {
		handleURL(tab, tabs[tab].history.urls[tabs[tab].history.pos]) // Load that position in history
		tabs[tab].applyScroll()
		if tab == curTab {
			// Display the bottomBar state that handleURL set
			tabs[tab].applyBottomBar()
		}
	}(curTab)
}
