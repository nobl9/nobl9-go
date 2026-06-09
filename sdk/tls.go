package sdk

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
)

// newCustomCATransport returns an [http.RoundTripper] that uses the PEM
// certificates loaded from caCertFile as explicit TLS roots.
//
// The transport starts from [x509.SystemCertPool] and appends the configured
// bundle. On macOS and Windows, platform-verifier failures can fall through to
// Go verification with the appended roots, so callers using this as a
// platform-verifier escape hatch should provide a complete bundle containing
// the public roots required by Nobl9 and Okta plus any private corporate CAs.
func newCustomCATransport(caCertFile string) (http.RoundTripper, error) {
	if caCertFile == "" {
		return nil, nil
	}
	pemBytes, err := os.ReadFile(caCertFile) // nolint: gosec
	if err != nil {
		return nil, fmt.Errorf("failed to read CA bundle from %q: %w", caCertFile, err)
	}
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("failed to load system cert pool: %w", err)
	}
	if !pool.AppendCertsFromPEM(pemBytes) {
		return nil, fmt.Errorf("failed to parse any PEM certificates from %q", caCertFile)
	}
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, errors.New("http.DefaultTransport is not *http.Transport")
	}
	cloned := transport.Clone()
	cloned.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    pool,
	}
	return cloned, nil
}
