package display

// applyHist is a history.go internal function, to load a URL in the history.
func applyHist(t *tab) {
	t.handleURL(t.history.urls[t.history.pos], 0) // Load that position in history
	t.applyAll()
}

func (t *tab) histForward() {
	if t.history.pos >= len(t.history.urls)-1 {
		// Already on the most recent URL in the history
		return
	}
	t.history.pos++
	go applyHist(t)
}

func (t *tab) histBack() {
	if t.history.pos <= 0 {
		// First tab in history
		return
	}
	t.history.pos--
	go applyHist(t)
}
