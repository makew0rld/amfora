package client

import (
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/makeworld-the-better-one/amfora/config"
)

// TOFU implementation.
// Stores cert hash and expiry for now, like Bombadillo.
// There is ongoing TOFU discussiong on the mailing list about better
// ways to do this, and I will update this file once those are decided on.

var ErrTofu = errors.New("server cert does not match TOFU database")

var tofuStore = config.TofuStore

// idKey returns the config/viper key needed to retrieve
// a cert's ID / fingerprint.
func idKey(domain string) string {
	return strings.ReplaceAll(domain, ".", "/")
}

func expiryKey(domain string) string {
	return strings.ReplaceAll(strings.TrimSuffix(domain, "."), ".", "/") + "/expiry"
}

func loadTofuEntry(domain string) (string, time.Time, error) {
	id := tofuStore.GetString(idKey(domain)) // Fingerprint
	if len(id) != 64 {
		// Not set, or invalid
		return "", time.Time{}, errors.New("not found")
	}

	expiry := tofuStore.GetTime(expiryKey(domain))
	if expiry.IsZero() {
		// Not set
		return id, time.Time{}, errors.New("not found")
	}
	return id, expiry, nil
}

// certID returns a generic string representing a cert or domain.
func certID(cert *x509.Certificate) string {
	h := sha256.New()
	h.Write(cert.Raw)
	return fmt.Sprintf("%X", h.Sum(nil))

	// The old way that uses the cert public key:
	// b, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	// h := sha256.New()
	// if err != nil {
	// 	// Unsupported key type - try to store a hash of the struct instead
	// 	h.Write([]byte(fmt.Sprint(cert.PublicKey)))
	// 	return fmt.Sprintf("%X", h.Sum(nil))
	// }
	// h.Write(b)
	// return fmt.Sprintf("%X", h.Sum(nil))
}

func saveTofuEntry(cert *x509.Certificate) {
	tofuStore.Set(idKey(cert.Subject.CommonName), certID(cert))
	tofuStore.Set(expiryKey(cert.Subject.CommonName), cert.NotAfter.UTC())
	err := tofuStore.WriteConfig()
	if err != nil {
		panic(err)
	}
}

// handleTofu is the abstracted interface for taking care of TOFU.
// A cert is provided and storage, checking, etc, are taken care of.
// It returns a bool indicating if the cert is valid according to
// the TOFU database.
// If false is returned, the connection should not go ahead.
func handleTofu(cert *x509.Certificate) bool {
	id, expiry, err := loadTofuEntry(cert.Subject.CommonName)
	if err != nil {
		// Cert isn't in database or data is malformed
		// So it can't be checked and anything is valid
		saveTofuEntry(cert)
		return true
	}
	if certID(cert) == id {
		// Save cert as the one stored
		return true
	}
	if time.Now().After(expiry) {
		// Old cert expired, so anything is valid
		saveTofuEntry(cert)
		return true
	}
	return false
}
