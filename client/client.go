// Package client retrieves data over Gemini and implements a TOFU system.
package client

import (
	"net"
	"net/url"

	"github.com/makeworld-the-better-one/go-gemini"
)

// Fetch returns response data and an error.
// The error text is human friendly and should be displayed.
func Fetch(u string) (*gemini.Response, error) {

	res, err := gemini.Fetch(u)
	if err != nil {
		return nil, err
	}

	parsed, _ := url.Parse(u)

	ok := handleTofu(parsed.Hostname(), parsed.Port(), res.Cert)
	if !ok {
		return res, ErrTofu
	}

	return res, err
}

// FetchWithProxy is the same as Fetch, but uses a proxy.
func FetchWithProxy(proxyHostname, proxyPort, u string) (*gemini.Response, error) {
	res, err := gemini.FetchWithHost(net.JoinHostPort(proxyHostname, proxyPort), u)
	if err != nil {
		return nil, err
	}

	// Only associate the returned cert with the proxy
	ok := handleTofu(proxyHostname, proxyPort, res.Cert)
	if !ok {
		return res, ErrTofu
	}

	return res, nil
}
