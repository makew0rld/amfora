// Package client retrieves data over Gemini and implements a TOFU system.
package client

import (
	"net/url"

	"github.com/makeworld-the-better-one/go-gemini"
)

// Fetch returns response data and an error.
// The error text is human friendly and should be displayed.
func Fetch(u string) (*gemini.Response, error) {
	resp, err := gemini.Fetch(u)
	if err != nil {
		return nil, err
	}

	parsed, _ := url.Parse(u)
	ok := handleTofu(resp.Cert, parsed.Port())
	if !ok {
		return nil, ErrTofu
	}
	return resp, err
}
