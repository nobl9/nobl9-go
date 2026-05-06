package sdk

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
)

// newCustomCATransport returns an [http.RoundTripper] that trusts the certificates
// loaded from the PEM file at caCertFile in addition to clearing any default trust.
// When caCertFile is empty it returns nil so callers fall back to [http.DefaultTransport].
//
// Setting [tls.Config.RootCAs] forces Go's pure-Go [crypto/x509] verifier to be used
// instead of the platform verifier (Security.framework on macOS, wincrypt on Windows).
// This is the supported escape hatch for environments where the platform verifier
// rejects a chain that Go's verifier (and the loaded roots) accept - e.g. corporate
// laptops with MDM-installed trust profiles that interfere with SecTrustEvaluateWithError.
func newCustomCATransport(caCertFile string) (http.RoundTripper, error) {
	if caCertFile == "" {
		return nil, nil
	}
	pemBytes, err := os.ReadFile(caCertFile) // nolint: gosec
	if err != nil {
		return nil, fmt.Errorf("failed to read CA bundle %q: %w", caCertFile, err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemBytes) {
		return nil, fmt.Errorf("failed to parse any PEM certificates from %q", caCertFile)
	}
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("http.DefaultTransport is not *http.Transport")
	}
	cloned := transport.Clone()
	cloned.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    pool,
	}
	return cloned, nil
}
