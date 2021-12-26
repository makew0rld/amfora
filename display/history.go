package display

// applyHist is a history.go internal function, to load a URL in the history.
func applyHist(t *tab) {
	handleURL(t, t.history.urls[t.history.pos], 0) // Load that position in history

	// Set page's scroll and link info from history cache, in case it didn't have it in the page already
	// Like for non-cached pages like about: pages
	// This fixes #122
	pg := t.history.pageCache[t.history.pos]
	p := t.page
	p.Row = pg.row
	p.Column = pg.column
	p.Selected = pg.selected
	p.SelectedID = pg.selectedID
	p.Mode = pg.mode

	t.applyAll()
}

func histForward(t *tab) {
	if t.history.pos >= len(t.history.urls)-1 {
		// Already on the most recent URL in the history
		return
	}

	// Update page cache in history for #122
	t.historyCachePage()

	t.history.pos++
	go applyHist(t)
}

func histBack(t *tab) {
	if t.history.pos <= 0 {
		// First tab in history
		return
	}

	// Update page cache in history for #122
	t.historyCachePage()

	t.history.pos--
	go applyHist(t)
}
