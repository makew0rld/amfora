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
// There is ongoing TOFU discussion on the mailing list about better
// ways to do this, and I will update this file once those are decided on.
// Update: See #7 for some small improvements made.

var ErrTofu = errors.New("server cert does not match TOFU database")

var tofuStore = config.TofuStore

// idKey returns the config/viper key needed to retrieve
// a cert's ID / fingerprint.
func idKey(domain string, port string) string {
	if port == "1965" || port == "" {
		return strings.ReplaceAll(domain, ".", "/")
	}
	return strings.ReplaceAll(domain, ".", "/") + ":" + port
}

func expiryKey(domain string, port string) string {
	if port == "1965" || port == "" {
		return strings.ReplaceAll(strings.TrimSuffix(domain, "."), ".", "/") + "/expiry"
	}
	return strings.ReplaceAll(strings.TrimSuffix(domain, "."), ".", "/") + "/expiry" + ":" + port
}

func loadTofuEntry(domain string, port string) (string, time.Time, error) {
	id := tofuStore.GetString(idKey(domain, port)) // Fingerprint
	if len(id) != 64 {
		// Not set, or invalid
		return "", time.Time{}, errors.New("not found")
	}

	expiry := tofuStore.GetTime(expiryKey(domain, port))
	if expiry.IsZero() {
		// Not set
		return id, time.Time{}, errors.New("not found")
	}
	return id, expiry, nil
}

// certID returns a generic string representing a cert or domain.
func certID(cert *x509.Certificate) string {
	h := sha256.New()
	h.Write(cert.RawSubjectPublicKeyInfo) // Better than cert.Raw, see #7
	return fmt.Sprintf("%X", h.Sum(nil))
}

// origCertID uses cert.Raw, which was used in v1.0.0 of the app.
func origCertID(cert *x509.Certificate) string {
	h := sha256.New()
	h.Write(cert.Raw) // Better than cert.Raw, see #7
	return fmt.Sprintf("%X", h.Sum(nil))
}

func saveTofuEntry(cert *x509.Certificate, port string) {
	tofuStore.Set(idKey(cert.Subject.CommonName, port), certID(cert))
	tofuStore.Set(expiryKey(cert.Subject.CommonName, port), cert.NotAfter.UTC())
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
func handleTofu(cert *x509.Certificate, port string) bool {
	id, expiry, err := loadTofuEntry(cert.Subject.CommonName, port)
	if err != nil {
		// Cert isn't in database or data is malformed
		// So it can't be checked and anything is valid
		saveTofuEntry(cert, port)
		return true
	}
	if time.Now().After(expiry) {
		// Old cert expired, so anything is valid
		saveTofuEntry(cert, port)
		return true
	}
	if certID(cert) == id {
		// Same cert as the one stored
		return true
	}
	if origCertID(cert) == id {
		// Valid but uses old ID type
		saveTofuEntry(cert, port)
		return true
	}
	return false
}
