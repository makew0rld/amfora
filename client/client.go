// Package client retrieves data over Gemini and implements a TOFU system.
package client

import (
	"github.com/makeworld-the-better-one/go-gemini"
)

// Fetch returns response data and an error.
// The error text is human friendly and should be displayed.
func Fetch(url string) (*gemini.Response, error) {
	resp, err := gemini.Fetch(url)
	if err != nil {
		return nil, err
	}
	ok := handleTofu(resp.Cert)
	if !ok {
		return nil, ErrTofu
	}
	return resp, err
}
