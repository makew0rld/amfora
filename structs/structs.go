package structs

type Mediatype string

const (
	TextGemini Mediatype = "text/gemini"
	TextPlain  Mediatype = "text/plain"
	TextAnsi   Mediatype = "text/x-ansi"
)

type PageMode int

const (
	ModeOff        PageMode = iota // Regular mode
	ModeLinkSelect                 // When the enter key is pressed, allow for tab-based link navigation
	ModeSearch                     // When a keyword is being searched in a page - TODO: NOT USED YET
)

// Page is for storing UTF-8 text/gemini pages, as well as text/plain pages.
type Page struct {
	Url        string
	Mediatype  Mediatype
	Raw        string   // The raw response, as received over the network
	Content    string   // The processed content, NOT raw. Uses cview color tags. All link/link texts must have region tags. It will also have a left margin.
	Links      []string // URLs, for each region in the content.
	Row        int      // Scroll position
	Column     int      // ditto
	Width      int      // The width of the terminal at the time when the Content was set. This is to know when reformatting should happen.
	Selected   string   // The current text or link selected
	SelectedID string   // The cview region ID for the selected text/link
	Mode       PageMode
}

// Size returns an approx. size of a Page in bytes.
func (p *Page) Size() int {
	n := len(p.Raw) + len(p.Content) + len(p.Url) + len(p.Selected) + len(p.SelectedID)
	for i := range p.Links {
		n += len(p.Links[i])
	}
	return n
}
