package bookmarks

// Structs and code for the XBEL XML bookmark format.
// https://github.com/makeworld-the-better-one/amfora/issues/68

import (
	"encoding/xml"
)

var xbelHeader = []byte(xml.Header + `<!DOCTYPE xbel
  PUBLIC "+//IDN python.org//DTD XML Bookmark Exchange Language 1.1//EN//XML"
         "http://www.python.org/topics/xml/dtds/xbel-1.1.dtd">
`)

const xbelVersion = "1.1"

type xbelBookmark struct {
	XMLName xml.Name `xml:"bookmark"`
	URL     string   `xml:"href,attr"`
	Name    string   `xml:"title"`
}

// xbelFolder is unused as folders aren't supported by the UI yet.
// Follow #56 for details.
// https://github.com/makeworld-the-better-one/amfora/issues/56
type xbelFolder struct {
	XMLName   xml.Name        `xml:"folder"`
	Version   string          `xml:"version,attr"`
	Folded    string          `xml:"folded,attr"` // Idk if this will be used or not
	Name      string          `xml:"title"`
	Bookmarks []*xbelBookmark `xml:"bookmark"`
	Folders   []*xbelFolder   `xml:"folder"`
}

type xbel struct {
	XMLName   xml.Name        `xml:"xbel"`
	Version   string          `xml:"version,attr"`
	Bookmarks []*xbelBookmark `xml:"bookmark"`
	// Later: Folders []*xbelFolder
}

// Instance of xbel - loaded from bookmarks file
var data xbel
