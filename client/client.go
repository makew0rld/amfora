// Package client retrieves data over Gemini and implements a TOFU system.
package client

import (
	"io/ioutil"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/makeworld-the-better-one/go-gemini"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	certCache   = make(map[string][][]byte)
	certCacheMu = &sync.RWMutex{}

	fetchClient *gemini.Client
)

func Init() {
	fetchClient = &gemini.Client{
		ConnectTimeout: 10 * time.Second, // Default is 15
		ReadTimeout:    time.Duration(viper.GetInt("a-general.page_max_time")) * time.Second,
	}
}

func clientCert(host string) ([]byte, []byte) {
	certCacheMu.RLock()
	pair, ok := certCache[host]
	certCacheMu.RUnlock()
	if ok {
		return pair[0], pair[1]
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
		certCacheMu.Lock()
		certCache[host] = [][]byte{nil, nil}
		certCacheMu.Unlock()
		return nil, nil
	}

	cert, err := ioutil.ReadFile(certPath)
	if err != nil {
		certCacheMu.Lock()
		certCache[host] = [][]byte{nil, nil}
		certCacheMu.Unlock()
		return nil, nil
	}
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		certCacheMu.Lock()
		certCache[host] = [][]byte{nil, nil}
		certCacheMu.Unlock()
		return nil, nil
	}

	certCacheMu.Lock()
	certCache[host] = [][]byte{cert, key}
	certCacheMu.Unlock()
	return cert, key
}

// HasClientCert returns whether or not a client certificate exists for a host.
func HasClientCert(host string) bool {
	cert, _ := clientCert(host)
	return cert != nil
}

func fetch(u string, c *gemini.Client) (*gemini.Response, error) {
	parsed, _ := url.Parse(u)
	cert, key := clientCert(parsed.Host)

	var res *gemini.Response
	var err error
	if cert != nil {
		res, err = c.FetchWithCert(u, cert, key)
	} else {
		res, err = c.Fetch(u)
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

// Fetch returns response data and an error.
// The error text is human friendly and should be displayed.
func Fetch(u string) (*gemini.Response, error) {
	return fetch(u, fetchClient)
}

func fetchWithProxy(proxyHostname, proxyPort, u string, c *gemini.Client) (*gemini.Response, error) {
	parsed, _ := url.Parse(u)
	cert, key := clientCert(parsed.Host)

	var res *gemini.Response
	var err error
	if cert != nil {
		res, err = c.FetchWithHostAndCert(net.JoinHostPort(proxyHostname, proxyPort), u, cert, key)
	} else {
		res, err = c.FetchWithHost(net.JoinHostPort(proxyHostname, proxyPort), u)
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

// FetchWithProxy is the same as Fetch, but uses a proxy.
func FetchWithProxy(proxyHostname, proxyPort, u string) (*gemini.Response, error) {
	return fetchWithProxy(proxyHostname, proxyPort, u, fetchClient)
}
