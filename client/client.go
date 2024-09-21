// Package client retrieves data over Gemini and implements a TOFU system.
package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/makeworld-the-better-one/amfora/logger"
	"github.com/makeworld-the-better-one/amfora/openssl"
	"github.com/makeworld-the-better-one/go-gemini"
	gemsocks5 "github.com/makeworld-the-better-one/go-gemini-socks5"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Simple key for certCache map and others, instead of a full URL
// Only uses the part of the URL relevant to matching certs to a URL
type certMapKey struct {
	host string
	path string
}

var (
	// [auth] section of config put into maps
	confCerts = make(map[certMapKey]string)
	confKeys  = make(map[certMapKey]string)

	// Cache the cert and key assigned to different URLs
	certCache   = make(map[certMapKey][][]byte)
	certCacheMu = &sync.RWMutex{}

	fetchClient *gemini.Client
)

func Init() error {
	fetchClient = &gemini.Client{
		ConnectTimeout: 10 * time.Second, // Default is 15
		ReadTimeout:    time.Duration(viper.GetInt("a-general.page_max_time")) * time.Second,
	}

	if socksHost := os.Getenv("AMFORA_SOCKS5"); socksHost != "" {
		fetchClient.Proxy = gemsocks5.ProxyFunc(socksHost, nil)
	}

	// Populate config maps

	certsViper := viper.Sub("auth.certs")
	for _, certURL := range certsViper.AllKeys() {
		// Normalize URL so that it can be matched no matter how it was written
		// in the config
		pu, _ := normalizeURL(FixUserURL(certURL))
		if pu == nil {
			//nolint:goerr113
			return errors.New("[auth.certs]: couldn't normalize URL: " + certURL)
		}
		confCerts[certMapKey{pu.Host, pu.Path}] = certsViper.GetString(certURL)
	}

	keysViper := viper.Sub("auth.keys")
	for _, keyURL := range keysViper.AllKeys() {
		pu, _ := normalizeURL(FixUserURL(keyURL))
		if pu == nil {
			//nolint:goerr113
			return errors.New("[auth.keys]: couldn't normalize URL: " + keyURL)
		}
		confKeys[certMapKey{pu.Host, pu.Path}] = keysViper.GetString(keyURL)
	}

	return nil
}

// getCertPath returns the path of the cert from the config.
// It returns "" if no config value exists.
func getCertPath(host string, path string) string {
	for k, v := range confCerts {
		if k.host == host && (k.path == path || strings.HasPrefix(path, k.path)) {
			// Either exact match to what's in config, or a subpath
			return v
		}
	}
	// No matches
	return ""
}

// getKeyPath returns the path of the key from the config.
// It returns "" if no config value exists.
func getKeyPath(host string, path string) string {
	for k, v := range confKeys {
		if k.host == host && (k.path == path || strings.HasPrefix(path, k.path)) {
			// Either exact match to what's in config, or a subpath
			return v
		}
	}
	// No matches
	return ""
}

func clientCert(host string, path string) ([]byte, []byte) {
	mkey := certMapKey{host, path}

	certCacheMu.RLock()
	pair, ok := certCache[mkey]
	certCacheMu.RUnlock()
	if ok {
		return pair[0], pair[1]
	}

	ogCertPath := getCertPath(host, path)
	// Expand paths starting with ~/
	certPath, err := homedir.Expand(ogCertPath)
	if err != nil {
		certPath = ogCertPath
	}
	ogKeyPath := getKeyPath(host, path)
	keyPath, err := homedir.Expand(ogKeyPath)
	if err != nil {
		keyPath = ogKeyPath
	}

	if certPath == "" && keyPath == "" {
		certCacheMu.Lock()
		certCache[mkey] = [][]byte{nil, nil}
		certCacheMu.Unlock()
		return nil, nil
	}

	cert, err := ioutil.ReadFile(certPath)
	if err != nil {
		certCacheMu.Lock()
		certCache[mkey] = [][]byte{nil, nil}
		certCacheMu.Unlock()
		return nil, nil
	}
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		certCacheMu.Lock()
		certCache[mkey] = [][]byte{nil, nil}
		certCacheMu.Unlock()
		return nil, nil
	}

	certCacheMu.Lock()
	certCache[mkey] = [][]byte{cert, key}
	certCacheMu.Unlock()
	return cert, key
}

// HasClientCert returns whether or not a client certificate exists for a host and path.
func HasClientCert(host string, path string) bool {
	cert, _ := clientCert(host, path)
	return cert != nil
}

func fetch(u string, c *gemini.Client) (*gemini.Response, error) {
	parsed, _ := url.Parse(u)
	cert, key := clientCert(parsed.Host, parsed.Path)

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
	cert, key := clientCert(parsed.Host, parsed.Path)

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

func CreateNewCertRow(url string) error {
	cut, ok := strings.CutPrefix(url, "gemini://")
	if !ok {
		return errors.New(fmt.Sprint("invalid url", url))
	}
	url = cut

	dir, err := openssl.GetCertsDir()
	if err != nil {
		logger.Logger.Fatal(err)
	}

	certPath := filepath.Join(dir, url, "cert.pem")
	keyPath := filepath.Join(dir, url, "key.pem")

	viper.Set("auth.certs", map[string]string{url: certPath})
	viper.Set("auth.keys", map[string]string{url: keyPath})

	if err = viper.WriteConfig(); err != nil {
		logger.Logger.Println(err)
		return err
	}
	if err = viper.WriteConfig(); err != nil {
		logger.Logger.Println(err)
		return err
	}

	return nil
}
