// Package client retrieves data over Gemini and implements a TOFU system.
package client

import (
	"net/url"

	"github.com/makeworld-the-better-one/amfora/config"
	"github.com/makeworld-the-better-one/go-gemini"
	"github.com/spf13/viper"
)

// Fetch returns response data and an error.
// The error text is human friendly and should be displayed.
func Fetch(u string) (*gemini.Response, error) {
	var res *gemini.Response
	var err error

	if config.Proxy == nil {
		res, err = gemini.Fetch(u)
	} else {
		res, err = gemini.FetchWithHost(viper.GetString("a-general.proxy"), u)
	}
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
