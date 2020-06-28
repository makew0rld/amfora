package structs

// Page is for storing UTF-8 text/gemini pages, as well as text/plain pages.
type Page struct {
	Url        string
	Content    string   // The processed content, NOT raw. Uses cview colour tags. All link/link texts must have region tags.
	Links      []string // URLs, for each region in the content.
	Row        int      // Scroll position
	Column     int      // ditto
	LeftMargin int      // <1 when the content is unmodified. Otherwise it indicates how many spaces have been prepended to each line
}

// Size returns an approx. size of a Page in bytes.
func (p *Page) Size() int {
	b := len(p.Content) + len(p.Url)
	for i := range p.Links {
		b += len(p.Links[i])
	}
	return b
}
