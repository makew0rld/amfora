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
	h.Write(cert.Raw)
	return fmt.Sprintf("%X", h.Sum(nil))
}

func saveTofuEntry(domain, port string, cert *x509.Certificate) {
	tofuStore.Set(idKey(domain, port), certID(cert))
	tofuStore.Set(expiryKey(domain, port), cert.NotAfter.UTC())
	tofuStore.WriteConfig()
}

// handleTofu is the abstracted interface for taking care of TOFU.
// A cert is provided and storage, checking, etc, are taken care of.
// It returns a bool indicating if the cert is valid according to
// the TOFU database.
// If false is returned, the connection should not go ahead.
func handleTofu(domain, port string, cert *x509.Certificate) bool {
	id, expiry, err := loadTofuEntry(domain, port)
	if err != nil {
		// Cert isn't in database or data is malformed
		// So it can't be checked and anything is valid
		saveTofuEntry(domain, port, cert)
		return true
	}
	if certID(cert) == id {
		// Same cert as the one stored

		// Store expiry again in case it changed
		tofuStore.Set(expiryKey(domain, port), cert.NotAfter.UTC())
		tofuStore.WriteConfig()

		return true
	}
	if origCertID(cert) == id {
		// Valid but uses old ID type
		saveTofuEntry(domain, port, cert)
		return true
	}
	if time.Now().After(expiry) {
		// Old cert expired, so anything is valid
		saveTofuEntry(domain, port, cert)
		return true
	}
	return false
}

// ResetTofuEntry forces the cert passed to be valid, overwriting any previous TOFU entry.
// The port string can be empty, to indicate port 1965.
func ResetTofuEntry(domain, port string, cert *x509.Certificate) {
	saveTofuEntry(domain, port, cert)
}

// GetExpiry returns the stored expiry date for the given host.
// The time will be empty (zero) if there is not expiry date stored for that host.
func GetExpiry(domain, port string) time.Time {
	return tofuStore.GetTime(expiryKey(domain, port))
}
