package subscriptions

import (
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/makeworld-the-better-one/amfora/logger"
)

// This file contains funcs for creating PageEntries, which
// are consumed by display/subscriptions.go

// getURL returns a URL to be used in a PageEntry, from a
// list of URLs for that item. It prefers gemini URLs, then
// HTTP(S), then by order.
func getURL(urls []string) string {
	if len(urls) == 0 {
		return ""
	}

	var firstHTTP string
	for _, u := range urls {
		if strings.HasPrefix(u, "gemini://") {
			return u
		}
		if (strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")) && firstHTTP == "" {
			// First HTTP(S) URL in the list
			firstHTTP = u
		}
	}
	if firstHTTP != "" {
		return firstHTTP
	}
	return urls[0]
}

// GetPageEntries returns the current list of PageEntries
// for use in rendering a page.
// The contents of the returned entries will never change,
// so this function needs to be called again to get updates.
// It always returns sorted entries - by post time, from newest to oldest.
func GetPageEntries() *PageEntries {
	logger.Log.Println("subscriptions.GetPageEntries called")

	var pe PageEntries

	data.RLock()

	for _, feed := range data.Feeds {
		for _, item := range feed.Items {
			if item.Links == nil || len(item.Links) == 0 {
				// Ignore items without links
				continue
			}

			// Set pub

			var pub time.Time

			// Try to use updated time first, then published

			if !item.UpdatedParsed.IsZero() {
				pub = *item.UpdatedParsed
			} else if !item.PublishedParsed.IsZero() {
				pub = *item.PublishedParsed
			} else {
				// No time on the post
				pub = time.Now()
			}

			// Set prefix

			// Prefer using the feed title over anything else.
			// Many feeds in Gemini only have this due to gemfeed's default settings.
			prefix := feed.Title

			if prefix == "" {
				// feed.Title was empty

				if item.Author != nil {
					// Prefer using the item author over the feed author
					prefix = item.Author.Name
				} else {
					if feed.Author != nil {
						prefix = feed.Author.Name
					} else {
						prefix = "[author unknown]"
					}
				}
			} else {
				// There's already a title, so add the author (if exists) to
				// the end of the title in parentheses.
				// Don't add the author if it's the same as the title.

				if item.Author != nil && item.Author.Name != prefix {
					// Prefer using the item author over the feed author
					prefix += " (" + item.Author.Name + ")"
				} else if feed.Author != nil && feed.Author.Name != prefix {
					prefix += " (" + feed.Author.Name + ")"
				}
			}

			pe.Entries = append(pe.Entries, &PageEntry{
				Prefix:    prefix,
				Title:     item.Title,
				URL:       getURL(item.Links),
				Published: pub,
			})
		}
	}

	for u, page := range data.Pages {
		parsed, _ := url.Parse(u)

		// Path is title
		title := parsed.Path
		if strings.HasPrefix(title, "/~") {
			// A user dir
			title = title[2:] // Remove beginning slash and tilde
			// Remove trailing slash if the root of a user dir is being tracked
			if strings.Count(title, "/") <= 1 && title[len(title)-1] == '/' {
				title = title[:len(title)-1]
			}
		} else if strings.HasPrefix(title, "/users/") {
			// "/users/" is removed for aesthetics when tracking hosted users
			title = strings.TrimPrefix(title, "/users/")
			title = strings.TrimPrefix(title, "~") // Remove leading tilde
			// Remove trailing slash if the root of a user dir is being tracked
			if strings.Count(title, "/") <= 1 && title[len(title)-1] == '/' {
				title = title[:len(title)-1]
			}
		}

		pe.Entries = append(pe.Entries, &PageEntry{
			Prefix:    parsed.Host,
			Title:     title,
			URL:       u,
			Published: page.Changed,
		})
	}

	data.RUnlock()

	sort.Sort(&pe)
	return &pe
}
