package structs

// Page is for storing UTF-8 text/gemini pages, as well as text/plain pages.
type Page struct {
	Url     string
	Raw     string   // The raw response, as received over the network
	Content string   // The processed content, NOT raw. Uses cview colour tags. All link/link texts must have region tags. It will also have a left margin.
	Links   []string // URLs, for each region in the content.
	Row     int      // Scroll position
	Column  int      // ditto
	Width   int      // The width of the terminal at the time when the Content was set. This is to know when reformatting should happen.
}

// Size returns an approx. size of a Page in bytes.
func (p *Page) Size() int {
	b := len(p.Raw) + len(p.Content) + len(p.Url)
	for i := range p.Links {
		b += len(p.Links[i])
	}
	return b
}
