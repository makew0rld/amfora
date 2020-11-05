// Package client retrieves data over Gemini and implements a TOFU system.
package client

import (
	"io/ioutil"
	"net"
	"net/url"

	"github.com/makeworld-the-better-one/go-gemini"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var certCache = make(map[string][][]byte)

func clientCert(host string) ([]byte, []byte) {
	if cert := certCache[host]; cert != nil {
		return cert[0], cert[1]
	}

	// Expand paths starting with ~/
	certPath, err := homedir.Expand(viper.GetString("auth.certs." + host))
	if err != nil {
		certPath = viper.GetString("auth.certs." + host)
	}
	keyPath, err := homedir.Expand(viper.GetString("auth.keys." + host))
	if err != nil {
		keyPath = viper.GetString("auth.keys." + host)
	}
	if certPath == "" && keyPath == "" {
		certCache[host] = [][]byte{nil, nil}
		return nil, nil
	}

	cert, err := ioutil.ReadFile(certPath)
	if err != nil {
		certCache[host] = [][]byte{nil, nil}
		return nil, nil
	}
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		certCache[host] = [][]byte{nil, nil}
		return nil, nil
	}

	certCache[host] = [][]byte{cert, key}
	return cert, key
}

// HasClientCert returns whether or not a client certificate exists for a host.
func HasClientCert(host string) bool {
	cert, _ := clientCert(host)
	return cert != nil
}

// Fetch returns response data and an error.
// The error text is human friendly and should be displayed.
func Fetch(u string) (*gemini.Response, error) {
	parsed, _ := url.Parse(u)
	cert, key := clientCert(parsed.Hostname())

	var res *gemini.Response
	var err error
	if cert != nil {
		res, err = gemini.FetchWithCert(u, cert, key)
	} else {
		res, err = gemini.Fetch(u)
	}
	if err != nil {
		return nil, err
	}

	ok := handleTofu(parsed.Hostname(), parsed.Port(), res.Cert)
	if !ok {
		return res, ErrTofu
	}

	return res, err
}

// FetchWithProxy is the same as Fetch, but uses a proxy.
func FetchWithProxy(proxyHostname, proxyPort, u string) (*gemini.Response, error) {
	parsed, _ := url.Parse(u)
	cert, key := clientCert(parsed.Host)

	var res *gemini.Response
	var err error
	if cert != nil {
		res, err = gemini.FetchWithHostAndCert(net.JoinHostPort(proxyHostname, proxyPort), u, cert, key)
	} else {
		res, err = gemini.FetchWithHost(net.JoinHostPort(proxyHostname, proxyPort), u)
	}
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
