package display

func histForward(t *tab) {
	if t.history.pos >= len(t.history.urls)-1 {
		// Already on the most recent URL in the history
		return
	}
	t.history.pos++
	go func(tt *tab) {
		handleURL(tt, tt.history.urls[tt.history.pos]) // Load that position in history
		tt.applyScroll()
		tt.applySelected()
		if tt == tabs[curTab] {
			// Display the bottomBar state that handleURL set
			tt.applyBottomBar()
		}
	}(t)
}

func histBack(t *tab) {
	if t.history.pos <= 0 {
		// First tab in history
		return
	}
	t.history.pos--
	go func(tt *tab) {
		handleURL(tt, tt.history.urls[tt.history.pos]) // Load that position in history
		tt.applyScroll()
		tt.applySelected()
		if tt == tabs[curTab] {
			// Display the bottomBar state that handleURL set
			tt.applyBottomBar()
		}
	}(t)
}
